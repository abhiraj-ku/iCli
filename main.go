package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {

	dsn := flag.String("dsn", "", "Postgres connection string(e.g., postgres://user:pass@localhost:5432/db)")
	limit := flag.Int("limit", 5, "Number of queries to analyze")

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
	fmt.Println("Connected to database..")
	fmt.Printf("Fetching top %d queries ...\n\n", *limit)

	// fetch activity
	stats, err := db.GetTopQueries(ctx, pool, *limit)
	if err != nil {
		if err != nil {
			log.Fatalf("Failed to retrieve query stats: %v\n", err)
		}
	}

	for i, stat := range stats {
		fmt.Printf("#%d | Calls: %d | Avg Time: %.2fms\n", i+1, stat.Calls, stat.AvgTime)
		fmt.Printf("Query: %s\n", stat.QueryText)
		fmt.Println("--------------------------------------------------")
	}

}
