package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SequenceInfo struct {
	SchemaName   string  `json:"schema_name"`
	SequenceName string  `json:"sequence_name"`
	DataType     string  `json:"data_type"`
	CurrentValue int64   `json:"current_value"`
	MaxValue     int64   `json:"max_value"`
	UsagePercent float64 `json:"usage_percent"`
	TableName    string  `json:"table_name"`
	ColumnName   string  `json:"column_name"`
}

func Sequences(ctx context.Context, pool *pgxpool.Pool) ([]SequenceInfo, error) {
	query := `
		SELECT
			s.schemaname,
			s.sequencename,
			s.data_type::text,
			COALESCE(s.last_value, s.start_value),
			s.max_value,
			COALESCE(ROUND((COALESCE(s.last_value, s.start_value)::numeric / NULLIF(s.max_value::numeric, 0)) * 100, 2), 0),
			COALESCE(t.relname, ''),
			COALESCE(a.attname, '')
		FROM pg_sequences s
		JOIN pg_namespace n ON n.nspname = s.schemaname
		JOIN pg_class c ON c.relnamespace = n.oid AND c.relname = s.sequencename
		LEFT JOIN pg_depend d ON d.objid = c.oid AND d.deptype = 'a'
			AND d.classid = 'pg_class'::regclass AND d.refclassid = 'pg_class'::regclass
		LEFT JOIN pg_class t ON t.oid = d.refobjid
		LEFT JOIN pg_attribute a ON a.attrelid = d.refobjid AND a.attnum = d.refobjsubid
		WHERE s.schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY 6 DESC;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sequences: %w", err)
	}
	defer rows.Close()

	var list []SequenceInfo
	for rows.Next() {
		var s SequenceInfo
		if err := rows.Scan(
			&s.SchemaName,
			&s.SequenceName,
			&s.DataType,
			&s.CurrentValue,
			&s.MaxValue,
			&s.UsagePercent,
			&s.TableName,
			&s.ColumnName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sequence row: %w", err)
		}
		list = append(list, s)
	}

	return list, rows.Err()
}

func FetchSequences(ctx context.Context, pool *pgxpool.Pool) ([]SequenceInfo, error) {
	return Sequences(ctx, pool)
}

