package report

import (
	"fmt"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func RenderSequences(issues []analyzer.SequenceIssue, minThreshold float64) {
	cw, iw := LayoutBounds()

	fmt.Println(SectionBannerStyle.Width(cw).Render(" ⚡ SEQUENCE & INTEGER ID OVERFLOW AUDIT "))

	if len(issues) == 0 {
		msg := fmt.Sprintf("✓ All database sequences are operating within safe limits (< %.0f%% capacity). No integer ID overflow risks detected!", minThreshold)
		fmt.Println(queryCardStyle.Width(cw).Render(successStyle.Width(iw).Render(msg)))
		return
	}

	var criticals, warnings int
	for _, iss := range issues {
		switch iss.Severity {
		case analyzer.SeverityCritical:
			criticals++
		case analyzer.SeverityWarning:
			warnings++
		}
	}

	var sum strings.Builder
	if criticals > 0 {
		sum.WriteString(highSeverityBadge.Render(fmt.Sprintf(" 🚨 CRITICAL: %d SEQUENCE(S) NEAR OVERFLOW CAPACITY ", criticals)))
		sum.WriteString("\n\n")
	} else if warnings > 0 {
		sum.WriteString(medSeverityBadge.Render(fmt.Sprintf(" ⚠️  WARNING: %d SEQUENCE(S) APPROACHING CAPACITY ", warnings)))
		sum.WriteString("\n\n")
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555577"))).
		Width(iw).
		Headers("SEQUENCE", "TABLE.COLUMN", "TYPE", "CURRENT / MAX", "USAGE", "STATUS").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#E6EEF8"))
		})

	for _, iss := range issues {
		seq := iss.Sequence
		target := "-"
		if seq.TableName != "" && seq.ColumnName != "" {
			target = fmt.Sprintf("%s.%s", seq.TableName, seq.ColumnName)
		}

		currentMaxStr := fmt.Sprintf("%s / %s", formatCompactNumber(seq.CurrentValue), formatCompactNumber(seq.MaxValue))
		progressBar := renderProgressBar(seq.UsagePercent, 10)
		usageStr := fmt.Sprintf("%s %.1f%%", progressBar, seq.UsagePercent)

		t.Row(
			seq.SequenceName,
			target,
			seq.DataType,
			currentMaxStr,
			usageStr,
			formatSequenceSeverity(iss.Severity),
		)
	}

	sum.WriteString(t.Render())
	fmt.Println(tableCardStyle.Width(cw).Render(sum.String()))

	for _, iss := range issues {
		if iss.Severity == analyzer.SeverityHealthy {
			continue
		}

		var card strings.Builder
		seq := iss.Sequence

		title := fmt.Sprintf("SEQUENCE RISK: %s", seq.SequenceName)
		if seq.TableName != "" && seq.ColumnName != "" {
			title = fmt.Sprintf("SEQUENCE OVERFLOW RISK: %s.%s", seq.TableName, seq.ColumnName)
		}

		badge := formatSequenceSeverity(iss.Severity)
		header := lipgloss.JoinHorizontal(lipgloss.Center, badge, "  ", lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(title))
		card.WriteString(header)
		card.WriteString("\n\n")

		card.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Render("Details:"))
		card.WriteString("\n")
		card.WriteString(queryTextStyle.Width(iw).Render(iss.Recommendation))
		card.WriteString("\n\n")

		if iss.RemediationSQL != "" {
			card.WriteString(lipgloss.NewStyle().Bold(true).Foreground(successColor).Render("🔧 Remediation SQL:"))
			card.WriteString("\n")
			card.WriteString(sqlSnippetStyle.Width(iw).Render(iss.RemediationSQL))
		}

		fmt.Println(queryCardStyle.Width(cw).Render(card.String()))
	}
}

func PrintSequences(issues []analyzer.SequenceIssue, minThreshold float64) {
	RenderSequences(issues, minThreshold)
}

func renderProgressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}

	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	barColor := successColor
	if pct >= analyzer.CriticalThreshold {
		barColor = dangerColor
	} else if pct >= analyzer.WarningThreshold {
		barColor = warningColor
	}

	fillStr := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#444466")).Render(strings.Repeat("░", empty))
	return "[" + fillStr + emptyStr + "]"
}

func formatSequenceSeverity(severity analyzer.SequenceSeverity) string {
	switch severity {
	case analyzer.SeverityCritical:
		return highSeverityBadge.Render(" CRITICAL ")
	case analyzer.SeverityWarning:
		return medSeverityBadge.Render(" WARNING ")
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#000000")).Background(successColor).Padding(0, 1).Render(" OK ")
	}
}

func formatCompactNumber(n int64) string {
	fn := float64(n)
	switch {
	case n >= 1_000_000_000_000:
		return fmt.Sprintf("%.2fT", fn/1e12)
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", fn/1e9)
	case n >= 1_000_000:
		return fmt.Sprintf("%.2fM", fn/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.2fK", fn/1e3)
	default:
		return fmt.Sprintf("%d", n)
	}
}

