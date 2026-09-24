package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/health"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

// Global variables

// GoReleaser will automatically overwrite this value during the build
var version = "dev"

// Holds the conn string passed via CLI
var dsn string

func resolveDSN() (string, error) {
	if dsn != "" {
		return dsn, nil
	}

	if envDSN := os.Getenv("PGDSN"); envDSN != "" {
		dsn = envDSN
		return dsn, nil
	}

	return "", fmt.Errorf("database connection string required: pass via -d/--dsn or set PGDSN")
}

var rootCmd = cobra.Command{
	Use:   "iCli",
	Short: "Postgres index & performance profiler",
	Long: `iCli connects to your PostgreSQL database, analyzes execution 
	plans via pg_stat_statements, and flags missing indexes and index bloat.

	Connection string resolution order:
	  1. -d / --dsn
	  2. PGDSN environment variable`,

	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedDSN, err := resolveDSN()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, err := db.Connect(ctx, resolvedDSN)
		if err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
		defer pool.Close()

		// fetch slow queries
		stats, err := db.GetTopQueries(ctx, pool, 3)
		if err != nil {
			return err
		}

		for _, stat := range stats {
			// get json execution plan
			planJSON, err := db.GetExplainPlan(ctx, pool, stat.QueryText)
			if err != nil {
				continue
			}

			// walk the ast for the json
			issues, _ := analyzer.AnalyzePlan(planJSON)

			// print the findings
			report.PrintQueryIssues(stat.QueryID, stat.QueryText, stat.AvgTime, issues)
		}

		// fetch unused indexes
		unused, err := db.RUnusedIndex(ctx, pool)
		if err != nil {
			return err
		}
		report.PrintUnusedIndexes(unused)
		return nil
	},
}

var updateCmd = cobra.Command{
	Use:   "update",
	Short: "Update iCli to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Checking for updates... (Current version: %s)\n", version)

		// init the updater to point to release repo
		updater, err := selfupdate.NewUpdater(selfupdate.Config{})
		if err != nil {
			return fmt.Errorf("failed to create updater: %w", err)
		}

		latest, err := updater.UpdateSelf(cmd.Context(), version, selfupdate.NewRepositorySlug("abhiraj-ku", "iCli"))
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		if latest.Version() == version {
			fmt.Println("✓ You are already on the latest version.")
		} else {
			fmt.Printf("Successfully updated to %s!\n", latest.Version())
			fmt.Println("Release Notes:\n", latest.ReleaseNotes)
		}

		return nil
	},
}

var reportCmd = cobra.Command{
	Use:   "report",
	Short: "Generate an instant database health and metadata report",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedDSN, err := resolveDSN()
		if err != nil {
			return err
		}

		ctx := context.Background()
		pool, err := pgxpool.New(ctx, resolvedDSN)
		if err != nil {
			return fmt.Errorf("unable to connect to database: %w", err)
		}
		defer pool.Close()

		rep, err := health.FetchReport(ctx, pool)
		if err != nil {
			return err
		}

		report.RenderHealthReport(rep)
		return nil
	},
}

func init() {
	// bind flag -d as short for dsn and make it available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&dsn, "dsn", "d", "", "Postgres connection string (priority 1; falls back to PGDSN)")

	// add update command
	rootCmd.AddCommand(&updateCmd)
	rootCmd.AddCommand(&reportCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
