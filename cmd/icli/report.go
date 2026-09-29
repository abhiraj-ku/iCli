package main

import (
	"context"
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
		return withDB(cmd, 10*time.Second, func(ctx context.Context, pool *pgxpool.Pool) error {
			rep, err := health.Fetch(ctx, pool)
			if err != nil {
				return err
			}

			report.RenderHealth(rep)
			return nil
		})
	},
}
