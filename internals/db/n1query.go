package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type N1Stat struct {
	QueryID   int64
	QueryText string
	Calls     int64
	TotalTime float64
	AvgTime   float64
	Rows      int64
}

func N1Candidates(ctx context.Context, pool *pgxpool.Pool, minCalls int) ([]N1Stat, error) {
	timeCol, err := StatTimeCol(ctx, pool)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT 
			queryid, 
			query, 
			calls, 
			%[1]s AS total_time, 
			(%[1]s / calls) AS avg_time,
			rows
		FROM pg_stat_statements 
		WHERE query NOT ILIKE '%%iCli%%'
		  AND query NOT ILIKE '%%SHOW server_version_num%%'
		  AND query NOT ILIKE '%%pg_stat_statements%%'
		  AND query NOT ILIKE '%%pg_catalog%%'
		  AND query NOT ILIKE '%%pg_constraint%%'
		  AND calls >= $1
		ORDER BY calls DESC 
		LIMIT 100;
	`, timeCol)

	rows, err := pool.Query(ctx, query, minCalls)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_statements for N+1 candidates: %w", err)
	}
	defer rows.Close()

	var stats []N1Stat
	for rows.Next() {
		var s N1Stat
		if err := rows.Scan(&s.QueryID, &s.QueryText, &s.Calls, &s.TotalTime, &s.AvgTime, &s.Rows); err != nil {
			return nil, fmt.Errorf("failed to scan N+1 candidate query row: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func GetN1CandidateQueries(ctx context.Context, pool *pgxpool.Pool, minCalls int) ([]N1Stat, error) {
	return N1Candidates(ctx, pool, minCalls)
}
