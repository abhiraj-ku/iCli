package db

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

var paramCheckRegex = regexp.MustCompile(`\$\d+|\?`)

// Thuis removes all parametrized variable with with SELECT NULL
// This prevents Postgres from optimizing away "column = NULL" checks.
func sanitizeExplainQ(qr string) string {
	return paramCheckRegex.ReplaceAllString(qr, "(SELECT NULL)")
}

func GetExplainPlan(ctx context.Context, pool *pgxpool.Pool, query string) (string, error) {
	//  remove missing parameters
	safeQr := sanitizeExplainQ(query)

	explainQr := fmt.Sprintf("explain (format json) %s", safeQr)

	// EXPLAIN (FORMAT JSON) returns a single row with a single string column containing the JSON array.
	var jsonStr string
	err := pool.QueryRow(ctx, explainQr).Scan(&jsonStr)
	if err != nil {
		return "", fmt.Errorf("failed to generate EXPLAIN plan: %w\nAttempted Query: %s", err, explainQr)
	}

	return jsonStr, nil
}
