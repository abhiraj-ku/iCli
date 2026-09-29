package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var minCalls int

var n1Cmd = &cobra.Command{
	Use:     "n1query",
	Aliases: []string{"n1", "n+1"},
	Short:   "Detect N+1 query patterns and high-frequency application loop queries",
	Long:    "Analyzes pg_stat_statements to identify parameterized queries running repeatedly inside application loops, causing N+1 latency bottlenecks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(cmd, 15*time.Second, func(ctx context.Context, pool *pgxpool.Pool) error {
			stats, err := db.N1Candidates(ctx, pool, minCalls)
			if err != nil {
				return fmt.Errorf("failed to fetch N+1 candidates: %w", err)
			}

			issues := analyzer.DetectN1(stats)
			report.RenderN1(issues, minCalls)
			return nil
		})
	},
}

func init() {
	n1Cmd.Flags().IntVarP(&minCalls, "min-calls", "m", 50, "Minimum execution count threshold to flag N+1 candidate queries")
}
