package analyzer

import (
	"fmt"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/db"
)

type SequenceSeverity string

const (
	SeverityCritical SequenceSeverity = "CRITICAL"
	SeverityWarning  SequenceSeverity = "WARNING"
	SeverityHealthy  SequenceSeverity = "HEALTHY"

	CriticalThreshold = 85.0
	WarningThreshold  = 60.0
)

type SequenceIssue struct {
	Sequence       db.SequenceInfo  `json:"sequence"`
	Severity       SequenceSeverity `json:"severity"`
	RemediationSQL string           `json:"remediation_sql"`
	Recommendation string           `json:"recommendation"`
}

func AnalyzeSequences(sequences []db.SequenceInfo, minThreshold float64) []SequenceIssue {
	var issues []SequenceIssue

	for _, seq := range sequences {
		if seq.UsagePercent < minThreshold {
			continue
		}

		sev := CategorizeSeverity(seq.UsagePercent)
		sql, rec := BuildSequenceRemediation(seq)

		issues = append(issues, SequenceIssue{
			Sequence:       seq,
			Severity:       sev,
			RemediationSQL: sql,
			Recommendation: rec,
		})
	}

	return issues
}

func CategorizeSeverity(pct float64) SequenceSeverity {
	switch {
	case pct >= CriticalThreshold:
		return SeverityCritical
	case pct >= WarningThreshold:
		return SeverityWarning
	default:
		return SeverityHealthy
	}
}

func BuildSequenceRemediation(seq db.SequenceInfo) (string, string) {
	if seq.TableName != "" && seq.ColumnName != "" && isSmallIntType(seq.DataType) {
		sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE BIGINT;", quoteIdent(seq.TableName), quoteIdent(seq.ColumnName))
		rec := fmt.Sprintf("Column %s.%s (%s) is at %.1f%% sequence capacity. Upgrade column type to BIGINT to prevent integer overflow outages.", seq.TableName, seq.ColumnName, seq.DataType, seq.UsagePercent)
		return sql, rec
	}

	sql := fmt.Sprintf("ALTER SEQUENCE %s MAXVALUE 9223372036854775807;", quoteIdent(seq.SequenceName))
	if seq.TableName != "" && seq.ColumnName != "" {
		rec := fmt.Sprintf("Sequence %s (%s.%s) has reached %.1f%% of its max value (%d). Increase sequence maxvalue.", seq.SequenceName, seq.TableName, seq.ColumnName, seq.UsagePercent, seq.MaxValue)
		return sql, rec
	}

	rec := fmt.Sprintf("Sequence %s has reached %.1f%% capacity.", seq.SequenceName, seq.UsagePercent)
	return sql, rec
}

func isSmallIntType(dt string) bool {
	dt = strings.ToLower(dt)
	return dt == "integer" || dt == "int4" || dt == "smallint" || dt == "int2"
}

func quoteIdent(name string) string {
	if strings.ContainsAny(name, "- ") || strings.ToUpper(name) == name {
		return fmt.Sprintf("%q", name)
	}
	return name
}

