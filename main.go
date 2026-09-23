package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := flag.String("dsn", "", "PostgreSQL connection string(e.g., postgres://user:pass@localhost:5432/db)")
	flag.Parse()

	if *dsn == "" {
		fmt.Println("Error: -dsn flag is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// init db pool
	pool, err := db.Connect(ctx, *dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v\n", err)
	}
	defer pool.Close()
	fmt.Println("Connected to database..")
	fmt.Printf("Fetching top %d queries ...\n\n", 3)

	// slow Queries
	stats, _ := db.GetTopQueries(ctx, pool, 3)

	for _, stat := range stats {
		// get JSON Execution Plan
		planJSON, err := db.GetExplainPlan(ctx, pool, stat.QueryText)
		if err != nil {
			continue // Skip queries that cannot be EXPLAINed
		}

		//  walk the AST
		issues, _ := analyzer.AnalyzePlan(planJSON)

		// Print findings
		report.PrintQueryIssues(stat.QueryID, stat.QueryText, stat.AvgTime, issues)
	}

	// unused Indexes
	unused, _ := db.RUnusedIndex(ctx, pool)

	// show index bloat findings
	report.PrintUnusedIndexes(unused)
}
