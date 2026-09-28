package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/health"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate an instant database health and metadata report",
	Long:  "Queries PostgreSQL health statistics including active connection counts, cache hit ratio, server version, and uptime.",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedDSN, err := resolveDSN()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

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
