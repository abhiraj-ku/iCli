package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Stats metadata
type QueryStats struct {
	QueryID   int
	QueryText string
	Calls     int
	TotalTime float64
	AvgTime   float64
}

// Unused indexes detection - these are generally those which are not used
type UnusedIndex struct {
	Schema    string
	Table     string
	IndexName string
	Size      string
}

// Retrieves the top-most time consuming queries
// we check for version PG12 vs PG13+ schema change (total_time vs total_exec_time)
// conflict arises with how the pg_stat_statements return the schema
func GetTopQueries(ctx context.Context, pool *pgxpool.Pool, limit int) ([]QueryStats, error) {
	// determine PostgreSQL version to handle the pg_stat_statements schema change
	var verStr string
	err := pool.QueryRow(ctx, "show server_version_num;").Scan(&verStr)
	if err != nil {
		return nil, fmt.Errorf("failed to check pg_ver: %w", err)
	}
	version, _ := strconv.Atoi(verStr)
	timeCol := "total_time"
	if version >= 130000 {
		timeCol = "total_exec_time"
	}

	// 2. Build the query. We filter out our own application_name and internal queries.
	// We also filter out trivial queries (calls > 5) to focus on problematic queries.

	query := fmt.Sprintf(`
		SELECT 
			queryid, 
			query, 
			calls, 
			%[1]s AS total_time, 
			(%[1]s / calls) AS avg_time
		FROM pg_stat_statements 
		WHERE query NOT ILIKE '%%pg-advisor%%' 
		  AND query NOT ILIKE '%%SHOW server_version_num%%'
		  AND calls > 5
		ORDER BY %[1]s DESC 
		LIMIT $1;
	`, timeCol)

	rows, err := pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_statements. Is the extension enabled? %w", err)
	}
	defer rows.Close()

	var stats []QueryStats

	for rows.Next() {
		var s QueryStats
		if err := rows.Scan(&s.QueryID, &s.QueryText, &s.Calls, &s.TotalTime, &s.AvgTime); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading pg_stat_statements: %w", err)
	}
	return stats, nil
}

// retrieve unsued schemas
func RUnusedIndex(ctx context.Context, pool *pgxpool.Pool) ([]UnusedIndex, error) {
	// With pg_stat_user_indexes and join with pg_index.
	// We will not scan Primary keys and unique constraints

	query := `
		select 
			s.schemaname,
			s.relname as Table_Name,
			s.indexrelname as Index_name,
			pg_size_preety(pg_relation_size(s.indexrelid)) as Index_Size
		from pg_stat_user_index s
		join pg_index i on s.indexrelid = i.indexrelid
		where s.idx_scan=0
		and i.indisprimary = false
		and i.indisunique = false
		order by pg_relation_size(s.indexrelid) desc;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unused indexes: %w", err)
	}
	defer rows.Close()

	var uIndex []UnusedIndex
	for rows.Next() {
		var idx UnusedIndex
		if err := rows.Scan(&idx.Schema, &idx.Table, &idx.IndexName, &idx.Size); err != nil {
			return nil, fmt.Errorf("failed to scan unused index row: %w", err)
		}
		uIndex = append(uIndex, idx)
	}

	return uIndex, rows.Err()
}
