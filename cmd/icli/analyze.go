package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/spf13/cobra"
)

var analyzeLimit int

var analyzeCmd = &cobra.Command{
	Use:     "analyze",
	Aliases: []string{"queries", "query"},
	Short:   "Analyze slow queries and execution plan bottlenecks",
	Long:    "Fetches top slow queries from pg_stat_statements, requests EXPLAIN (FORMAT JSON) execution plans, and reports performance issues.",
	RunE:    runAnalyze,
}

func init() {
	analyzeCmd.Flags().IntVarP(&analyzeLimit, "limit", "n", 3, "Number of top slow queries to analyze")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	resolvedDSN, err := resolveDSN()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, resolvedDSN)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	// fetch slow queries
	stats, err := db.GetTopQueries(ctx, pool, analyzeLimit)
	if err != nil {
		return err
	}

	if len(stats) > 0 {
		fmt.Println(report.SectionBannerStyle.Render(" ⚡ SLOW QUERY & EXECUTION PLAN ANALYSIS "))
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

	return nil
}
