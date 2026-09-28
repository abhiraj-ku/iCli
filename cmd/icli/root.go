package main

import (
	"fmt"
	"os"

	"github.com/abhiraj-ku/pg_adv/internals/report"
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

var rootCmd = &cobra.Command{
	Use:   "icli",
	Short: "Postgres index & performance profiler",
	Long: `iCli connects to your PostgreSQL database, analyzes execution 
plans via pg_stat_statements, inspects index usage, and flags performance bottlenecks.

Connection string resolution order:
  1. -d / --dsn
  2. PGDSN environment variable`,

	Run: func(cmd *cobra.Command, args []string) {
		// When user types just 'icli' with no subcommands, show the polished commands menu
		report.RenderCLIMenu(version)
	},
}

func init() {
	// Custom help function to render Lipgloss CLI menu for rootCmd
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		report.RenderCLIMenu(version)
	})

	// bind flag -d as short for dsn and make it available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&dsn, "dsn", "d", "", "Postgres connection string (priority 1; falls back to PGDSN)")

	// register subcommands
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(updateCmd)
}
