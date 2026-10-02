package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/abhiraj-ku/pg_adv/internals/report"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var seqThreshold float64

var sequenceCmd = &cobra.Command{
	Use:     "sequence",
	Aliases: []string{"seq", "sequences"},
	Short:   "Audit sequence capacity and detect integer ID overflow risks",
	Long:    "Inspects PostgreSQL sequence generators and primary key integer data types to identify sequences reaching capacity limits before integer overflow outages occur.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(cmd, 15*time.Second, func(ctx context.Context, pool *pgxpool.Pool) error {
			seqs, err := db.Sequences(ctx, pool)
			if err != nil {
				return fmt.Errorf("failed to fetch sequences: %w", err)
			}

			issues := analyzer.AnalyzeSequences(seqs, seqThreshold)
			report.RenderSequences(issues, seqThreshold)
			return nil
		})
	},
}

func init() {
	sequenceCmd.Flags().Float64VarP(&seqThreshold, "threshold", "t", 0.0, "Minimum sequence capacity usage percentage threshold to report (0 to 100)")
}
