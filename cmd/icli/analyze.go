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

var limit int

var analyzeCmd = &cobra.Command{
	Use:     "analyze",
	Aliases: []string{"queries", "query"},
	Short:   "Analyze slow queries and execution plan bottlenecks",
	Long:    "Fetches top slow queries from pg_stat_statements, requests EXPLAIN (FORMAT JSON) execution plans, and reports performance issues.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(cmd, 15*time.Second, func(ctx context.Context, pool *pgxpool.Pool) error {
			stats, err := db.TopQueries(ctx, pool, limit)
			if err != nil {
				return err
			}

			if len(stats) > 0 {
				fmt.Println(report.SectionBannerStyle.Render(" ⚡ SLOW QUERY & EXECUTION PLAN ANALYSIS "))
			}

			for _, stat := range stats {
				planJSON, err := db.ExplainPlan(ctx, pool, stat.QueryText)
				if err != nil {
					continue
				}

				issues, _ := analyzer.AnalyzePlan(planJSON)
				report.RenderIssues(stat.QueryID, stat.QueryText, stat.AvgTime, issues)
			}
			return nil
		})
	},
}

func init() {
	analyzeCmd.Flags().IntVarP(&limit, "limit", "n", 3, "Number of top slow queries to analyze")
}
