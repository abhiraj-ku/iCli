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
		var b strings.Builder
		b.WriteString(successStyle.Width(iw).Render(
			fmt.Sprintf("✓ All database sequences are operating within safe limits (< %.0f%% capacity). No integer ID overflow risks detected!", minThreshold),
		))
		fmt.Println(queryCardStyle.Width(cw).Render(b.String()))
		return
	}

	criticalCount := 0
	warningCount := 0
	for _, iss := range issues {
		if iss.Severity == analyzer.SeverityCritical {
			criticalCount++
		} else if iss.Severity == analyzer.SeverityWarning {
			warningCount++
		}
	}

	var sum strings.Builder
	if criticalCount > 0 {
		sum.WriteString(highSeverityBadge.Render(fmt.Sprintf(" 🚨 CRITICAL: %d SEQUENCE(S) NEAR OVERFLOW CAPACITY ", criticalCount)))
		sum.WriteString("\n\n")
	} else if warningCount > 0 {
		sum.WriteString(medSeverityBadge.Render(fmt.Sprintf(" ⚠️  WARNING: %d SEQUENCE(S) APPROACHING CAPACITY ", warningCount)))
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

	// Print remediation cards for CRITICAL and WARNING issues
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

func renderProgressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	fillChar := "█"
	emptyChar := "░"

	barColor := successColor
	if pct >= 85.0 {
		barColor = dangerColor
	} else if pct >= 60.0 {
		barColor = warningColor
	}

	fillStr := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat(fillChar, filled))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#444466")).Render(strings.Repeat(emptyChar, empty))
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
	if n >= 1_000_000_000_000 {
		return fmt.Sprintf("%.2fT", float64(n)/1_000_000_000_000.0)
	}
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.2fB", float64(n)/1_000_000_000.0)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000.0)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.2fK", float64(n)/1_000.0)
	}
	return fmt.Sprintf("%d", n)
}
