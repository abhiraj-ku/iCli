package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

func withDB(cmd *cobra.Command, timeout time.Duration, fn func(ctx context.Context, pool *pgxpool.Pool) error) error {
	dsn, err := resolveDSN()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	return fn(ctx, pool)
}

func runWithDB(cmd *cobra.Command, timeout time.Duration, fn func(ctx context.Context, pool *pgxpool.Pool) error) error {
	return withDB(cmd, timeout, fn)
}
