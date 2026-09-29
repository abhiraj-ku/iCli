package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryStats struct {
	QueryID   int64
	QueryText string
	Calls     int
	TotalTime float64
	AvgTime   float64
}

type UnusedIndex struct {
	Schema    string
	Table     string
	IndexName string
	Size      string
}

type MissingFKIndex struct {
	Schema         string
	ChildTable     string
	ConstraintName string
	FKColumns      string
	ParentTable    string
}

func StatTimeCol(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	var verStr string
	if err := pool.QueryRow(ctx, "SHOW server_version_num;").Scan(&verStr); err != nil {
		return "", fmt.Errorf("failed to fetch server version: %w", err)
	}

	version, _ := strconv.Atoi(verStr)
	if version >= 130000 {
		return "total_exec_time", nil
	}
	return "total_time", nil
}

func TopQueries(ctx context.Context, pool *pgxpool.Pool, limit int) ([]QueryStats, error) {
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
			(%[1]s / calls) AS avg_time
		FROM pg_stat_statements 
		WHERE query NOT ILIKE '%%iCli%%'
		  AND query NOT ILIKE '%%SHOW server_version_num%%'
		  AND calls > 5
		ORDER BY %[1]s DESC 
		LIMIT $1;
	`, timeCol)

	rows, err := pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_statements: %w", err)
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
	return stats, rows.Err()
}

func UnusedIndexes(ctx context.Context, pool *pgxpool.Pool) ([]UnusedIndex, error) {
	query := `
		SELECT 
			s.schemaname,
			s.relname AS table_name,
			s.indexrelname AS index_name,
			pg_size_pretty(pg_relation_size(s.indexrelid)) AS index_size
		FROM pg_stat_user_indexes s
		JOIN pg_index i ON s.indexrelid = i.indexrelid
		WHERE s.idx_scan = 0
		  AND i.indisprimary = false
		  AND i.indisunique = false
		ORDER BY pg_relation_size(s.indexrelid) DESC;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unused indexes: %w", err)
	}
	defer rows.Close()

	var indexes []UnusedIndex
	for rows.Next() {
		var idx UnusedIndex
		if err := rows.Scan(&idx.Schema, &idx.Table, &idx.IndexName, &idx.Size); err != nil {
			return nil, fmt.Errorf("failed to scan unused index row: %w", err)
		}
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func RUnusedIndex(ctx context.Context, pool *pgxpool.Pool) ([]UnusedIndex, error) {
	return UnusedIndexes(ctx, pool)
}

func UnindexedFKs(ctx context.Context, pool *pgxpool.Pool) ([]MissingFKIndex, error) {
	query := `
		SELECT
			n.nspname AS schema_name,
			c_child.relname AS child_table,
			c.conname AS constraint_name,
			ARRAY_TO_STRING(ARRAY(
				SELECT a.attname
				FROM unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
				JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = u.attnum
				ORDER BY u.ord
			), ', ') AS fk_columns,
			c_parent.relname AS parent_table
		FROM pg_constraint c
		JOIN pg_class c_child ON c.conrelid = c_child.oid
		JOIN pg_class c_parent ON c.confrelid = c_parent.oid
		JOIN pg_namespace n ON c_child.relnamespace = n.oid
		WHERE c.contype = 'f'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND NOT EXISTS (
			  SELECT 1
			  FROM pg_index idx
			  WHERE idx.indrelid = c.conrelid
			    AND (idx.indkey::int2[])[1:cardinality(c.conkey)] = c.conkey
		  )
		ORDER BY c_child.relname, c.conname;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query missing foreign key indexes: %w", err)
	}
	defer rows.Close()

	var fks []MissingFKIndex
	for rows.Next() {
		var fk MissingFKIndex
		if err := rows.Scan(&fk.Schema, &fk.ChildTable, &fk.ConstraintName, &fk.FKColumns, &fk.ParentTable); err != nil {
			return nil, fmt.Errorf("failed to scan missing foreign key index row: %w", err)
		}
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func GetMissingFKIndexes(ctx context.Context, pool *pgxpool.Pool) ([]MissingFKIndex, error) {
	return UnindexedFKs(ctx, pool)
}
