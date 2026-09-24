package health

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func FetchReport(ctx context.Context, pool *pgxpool.Pool) (*Report, error) {
	report := &Report{}

	query := `
	WITH conn AS (
		SELECT 
			count(*) AS total,
			count(*) FILTER (WHERE state = 'active') AS active
		FROM pg_stat_activity
	),
	db_cache AS (
		SELECT 
			blks_hit,
			blks_read
		FROM pg_stat_database
		WHERE datname = current_database()
	)
	SELECT 
		current_setting('server_version'),
		current_database(),
		pg_size_pretty(pg_database_size(current_database())),
		date_trunc('second', now() - pg_postmaster_start_time())::text,
		pg_is_in_recovery(),
		current_setting('max_connections')::int,
		conn.total,
		conn.active,
		COALESCE(db_cache.blks_hit, 0),
		COALESCE(db_cache.blks_read, 0)
	FROM conn, db_cache;
	`

	var (
		blksHit  int64
		blksRead int64
	)

	err := pool.QueryRow(ctx, query).Scan(
		&report.Server.PGVersion,
		&report.Server.DatabaseName,
		&report.Server.DatabaseSize,
		&report.Server.Uptime,
		&report.Server.InRecovery,
		&report.Connections.MaxConnections,
		&report.Connections.TotalConnections,
		&report.Connections.Active,
		&blksHit,
		&blksRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch database report: %w", err)
	}

	totalBlocks := blksHit + blksRead
	if totalBlocks > 0 {
		report.Cache.HitRatio = (float64(blksHit) / float64(totalBlocks)) * 100
		report.Cache.MissRatio = 100.0 - report.Cache.HitRatio
	} else {
		report.Cache.HitRatio = 100.0
		report.Cache.MissRatio = 0.0
	}

	return report, nil
}
