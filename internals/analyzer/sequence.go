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

		severity := SeverityHealthy
		if seq.UsagePercent >= 85.0 {
			severity = SeverityCritical
		} else if seq.UsagePercent >= 60.0 {
			severity = SeverityWarning
		}

		var remediationSQL string
		var recommendation string

		if seq.TableName != "" && seq.ColumnName != "" {
			if strings.EqualFold(seq.DataType, "integer") || strings.EqualFold(seq.DataType, "int4") || strings.EqualFold(seq.DataType, "smallint") {
				remediationSQL = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE BIGINT;", quoteIdentifier(seq.TableName), quoteIdentifier(seq.ColumnName))
				recommendation = fmt.Sprintf("Column %s.%s is typed as %s and is reaching sequence limit (%.1f%%). Upgrade column type to BIGINT to prevent integer overflow outages.", seq.TableName, seq.ColumnName, seq.DataType, seq.UsagePercent)
			} else {
				remediationSQL = fmt.Sprintf("ALTER SEQUENCE %s MAXVALUE 9223372036854775807;", quoteIdentifier(seq.SequenceName))
				recommendation = fmt.Sprintf("Sequence %s has reached %.1f%% of its max value (%d). Increase sequence maxvalue.", seq.SequenceName, seq.UsagePercent, seq.MaxValue)
			}
		} else {
			remediationSQL = fmt.Sprintf("ALTER SEQUENCE %s MAXVALUE 9223372036854775807;", quoteIdentifier(seq.SequenceName))
			recommendation = fmt.Sprintf("Sequence %s has reached %.1f%% capacity.", seq.SequenceName, seq.UsagePercent)
		}

		issues = append(issues, SequenceIssue{
			Sequence:       seq,
			Severity:       severity,
			RemediationSQL: remediationSQL,
			Recommendation: recommendation,
		})
	}

	return issues
}

func quoteIdentifier(name string) string {
	if strings.Contains(name, "-") || strings.Contains(name, " ") {
		return fmt.Sprintf("\"%s\"", name)
	}
	return name
}
