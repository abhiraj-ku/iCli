package db

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

var paramCheckRegex = regexp.MustCompile(`\$\d+|\?`)

func sanitizeExplain(query string) string {
	return paramCheckRegex.ReplaceAllString(query, "(SELECT NULL)")
}

func ExplainPlan(ctx context.Context, pool *pgxpool.Pool, query string) (string, error) {
	safeQuery := sanitizeExplain(query)
	explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", safeQuery)

	var jsonStr string
	if err := pool.QueryRow(ctx, explainQuery).Scan(&jsonStr); err != nil {
		return "", fmt.Errorf("failed to generate EXPLAIN plan: %w\nQuery: %s", err, explainQuery)
	}
	return jsonStr, nil
}

func GetExplainPlan(ctx context.Context, pool *pgxpool.Pool, query string) (string, error) {
	return ExplainPlan(ctx, pool, query)
}
