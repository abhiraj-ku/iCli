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

func FetchSequences(ctx context.Context, pool *pgxpool.Pool) ([]SequenceInfo, error) {
	query := `
		SELECT
			s.schemaname AS schema_name,
			s.sequencename AS sequence_name,
			s.data_type::text AS data_type,
			COALESCE(s.last_value, s.start_value) AS current_value,
			s.max_value,
			ROUND(
				(COALESCE(s.last_value, s.start_value)::numeric / NULLIF(s.max_value::numeric, 0)) * 100, 
				2
			) AS usage_percent,
			COALESCE(tbl_cls.relname, '') AS table_name,
			COALESCE(att.attname, '') AS column_name
		FROM pg_sequences s
		JOIN pg_class seq_cls 
			ON seq_cls.relname = s.sequencename
		JOIN pg_namespace seq_ns 
			ON seq_ns.oid = seq_cls.relnamespace AND seq_ns.nspname = s.schemaname
		LEFT JOIN pg_depend d 
			ON d.objid = seq_cls.oid AND d.deptype = 'a'
		LEFT JOIN pg_class tbl_cls 
			ON tbl_cls.oid = d.refobjid
		LEFT JOIN pg_attribute att 
			ON att.attrelid = d.refobjid AND att.attnum = d.refobjsubid
		WHERE s.schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY usage_percent DESC;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sequence statistics: %w", err)
	}
	defer rows.Close()

	var sequences []SequenceInfo
	for rows.Next() {
		var s SequenceInfo
		var usagePct *float64
		err := rows.Scan(
			&s.SchemaName,
			&s.SequenceName,
			&s.DataType,
			&s.CurrentValue,
			&s.MaxValue,
			&usagePct,
			&s.TableName,
			&s.ColumnName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sequence row: %w", err)
		}
		if usagePct != nil {
			s.UsagePercent = *usagePct
		}
		sequences = append(sequences, s)
	}

	return sequences, rows.Err()
}
