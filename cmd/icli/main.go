package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/spf13/cobra"
)

// Holds the conn string passed via CLI
var dsn string

var rootCmd = cobra.Command{
	Use:   "iCli",
	Short: "Postgres index & performance profiler",
	Long: `iCli connects to your PostgreSQL database, analyzes execution 
plans via pg_stat_statements, and flags missing indexes and index bloat.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, err := db.Connect(ctx, dsn)
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

func init() {
	// bind flag -d as short for dsn
	rootCmd.Flags().StringVarP(&dsn, "dsn", "d", "", "Postgres connection string(required)")

	rootCmd.MarkFlagRequired("dsn")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
