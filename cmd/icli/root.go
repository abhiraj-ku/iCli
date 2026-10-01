package main

import (
	"fmt"
	"os"

	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/spf13/cobra"
)

var version = "dev"
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
		report.RenderCLIMenu(version)
	},
}

func init() {
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		report.RenderCLIMenu(version)
	})

	rootCmd.PersistentFlags().StringVarP(&dsn, "dsn", "d", "", "Postgres connection string (priority 1; falls back to PGDSN)")

	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(n1Cmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(sequenceCmd)
	rootCmd.AddCommand(updateCmd)
}
