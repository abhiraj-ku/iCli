package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:     "index",
	Aliases: []string{"indexes"},
	Short:   "Scan for unused indexes and missing foreign key indexes",
	Long:    "Inspects PostgreSQL database indexes to discover unused indexes and unindexed foreign keys on child tables.",
	RunE:    runIndexAll,
}

var unusedIndexCmd = &cobra.Command{
	Use:   "unused",
	Short: "Find non-primary, non-unique indexes with zero recorded scans",
	RunE:  runUnusedIndex,
}

var missingFKIndexCmd = &cobra.Command{
	Use:   "missing-fk",
	Short: "Find foreign key constraints without covering indexes on child tables",
	RunE:  runMissingFKIndex,
}

func init() {
	indexCmd.AddCommand(unusedIndexCmd)
	indexCmd.AddCommand(missingFKIndexCmd)
}

func runIndexAll(cmd *cobra.Command, args []string) error {
	if err := runUnusedIndex(cmd, args); err != nil {
		return err
	}
	if err := runMissingFKIndex(cmd, args); err != nil {
		return err
	}
	return nil
}

func runUnusedIndex(cmd *cobra.Command, args []string) error {
	resolvedDSN, err := resolveDSN()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, resolvedDSN)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	// fetch unused indexes
	unused, err := db.RUnusedIndex(ctx, pool)
	if err != nil {
		return err
	}
	report.PrintUnusedIndexes(unused)

	return nil
}

func runMissingFKIndex(cmd *cobra.Command, args []string) error {
	resolvedDSN, err := resolveDSN()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, resolvedDSN)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	// fetch missing foreign key indexes
	missingFKs, err := db.GetMissingFKIndexes(ctx, pool)
	if err != nil {
		return err
	}
	report.PrintMissingFKIndexes(missingFKs)

	return nil
}
