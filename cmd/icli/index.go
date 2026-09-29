package main

import (
	"context"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:     "index",
	Aliases: []string{"indexes"},
	Short:   "Scan for unused indexes and missing foreign key indexes",
	Long:    "Inspects PostgreSQL database indexes to discover unused indexes and unindexed foreign keys on child tables.",
	RunE:    runIndexAll,
}

var unusedCmd = &cobra.Command{
	Use:   "unused",
	Short: "Find non-primary, non-unique indexes with zero recorded scans",
	RunE:  runUnusedIndex,
}

var missingFKCmd = &cobra.Command{
	Use:   "missing-fk",
	Short: "Find foreign key constraints without covering indexes on child tables",
	RunE:  runMissingFKIndex,
}

func init() {
	indexCmd.AddCommand(unusedCmd)
	indexCmd.AddCommand(missingFKCmd)
}

func runIndexAll(cmd *cobra.Command, args []string) error {
	if err := runUnusedIndex(cmd, args); err != nil {
		return err
	}
	return runMissingFKIndex(cmd, args)
}

func runUnusedIndex(cmd *cobra.Command, args []string) error {
	return withDB(cmd, 10*time.Second, func(ctx context.Context, pool *pgxpool.Pool) error {
		unused, err := db.UnusedIndexes(ctx, pool)
		if err != nil {
			return err
		}
		report.RenderUnused(unused)
		return nil
	})
}

func runMissingFKIndex(cmd *cobra.Command, args []string) error {
	return withDB(cmd, 10*time.Second, func(ctx context.Context, pool *pgxpool.Pool) error {
		fks, err := db.UnindexedFKs(ctx, pool)
		if err != nil {
			return err
		}
		report.RenderMissingFKs(fks)
		return nil
	})
}
