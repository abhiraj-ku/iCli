package report

import (
	"fmt"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func RenderN1(issues []analyzer.N1Issue, threshold int) {
	cw, iw := LayoutBounds()

	fmt.Println(SectionBannerStyle.Width(cw).Render(" ⚡ N+1 QUERY DETECTION "))

	if len(issues) == 0 {
		var b strings.Builder
		b.WriteString(successStyle.Width(iw).Render(
			fmt.Sprintf("✓ No N+1 query loops detected (threshold: >= %d executions). Your application query batching is performing well!", threshold),
		))
		fmt.Println(queryCardStyle.Width(cw).Render(b.String()))
		return
	}

	var sum strings.Builder
	sum.WriteString(lipgloss.NewStyle().Foreground(warningColor).Bold(true).Width(iw).Render(
		fmt.Sprintf("⚠️  Detected %d query loop bottleneck(s) executing repeatedly in application loops (>= %d calls).", len(issues), threshold),
	))
	sum.WriteString("\n\n")

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555577"))).
		Width(iw).
		Headers("QUERY ID", "TARGET TABLE", "CALLS", "TOTAL OVERHEAD", "SEVERITY").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#E6EEF8"))
		})

	for _, iss := range issues {
		t.Row(
			fmt.Sprintf("#%d", iss.QueryID),
			iss.TargetTable,
			fmt.Sprintf("%d", iss.Calls),
			fmt.Sprintf("%.2fms", iss.TotalTime),
			FormatSeverity(iss.Severity),
		)
	}

	sum.WriteString(t.Render())
	fmt.Println(tableCardStyle.Width(cw).Render(sum.String()))

	for i, iss := range issues {
		var card strings.Builder

		badge := queryIdBadge.Render(fmt.Sprintf("N+1 LOOP CANDIDATE #%d (%d of %d)", iss.QueryID, i+1, len(issues)))
		callsStr := warningStyle.Render(fmt.Sprintf("🔁 Calls: %d", iss.Calls))
		timeStr := warningStyle.Render(fmt.Sprintf("⏱  Total Overhead: %.2fms (Avg: %.2fms)", iss.TotalTime, iss.AvgTime))
		card.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, badge, "  ", callsStr, "  ", timeStr))
		card.WriteString("\n\n")

		card.WriteString(lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Render("Pattern Detected: "))
		card.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#E6EEF8")).Render(iss.PatternType))
		card.WriteString("\n")

		if iss.AvgRowsPerCall > 0 {
			card.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Render(
				fmt.Sprintf("Yield Rate: ~%.1f row(s) returned per call across %d executions.", iss.AvgRowsPerCall, iss.Calls),
			))
			card.WriteString("\n")
		}

		sql := strings.Join(strings.Fields(iss.QueryText), " ")
		card.WriteString("\n")
		card.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Render("Loop Query SQL:"))
		card.WriteString("\n")
		card.WriteString(queryTextStyle.Width(iw).Render(sql))
		card.WriteString("\n")

		if iss.ParentCandidate != nil {
			card.WriteString("\n")
			card.WriteString(lipgloss.NewStyle().Bold(true).Foreground(warningColor).Render("🔗 Suspected Parent Query in Loop:"))
			card.WriteString("\n")
			parentSQL := strings.Join(strings.Fields(iss.ParentCandidate.QueryText), " ")
			card.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#E6EEF8")).Italic(true).Width(iw).Render(
				fmt.Sprintf("  Query #%d (Calls: %d, Total Time: %.2fms): %s", iss.ParentCandidate.QueryID, iss.ParentCandidate.Calls, iss.ParentCandidate.TotalTime, parentSQL),
			))
			card.WriteString("\n")
		}

		card.WriteString("\n")
		card.WriteString(recHeaderStyle.Render("💡 Optimization Strategy:"))
		card.WriteString("\n")
		card.WriteString(sqlSnippetStyle.Width(iw).Render(iss.Recommendation))
		card.WriteString("\n")

		fmt.Println(queryCardStyle.Width(cw).Render(card.String()))
	}
}

func PrintN1QueryReports(reports []analyzer.N1Issue, minCalls int) {
	RenderN1(reports, minCalls)
}
