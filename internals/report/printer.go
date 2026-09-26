package report

import (
	"fmt"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	primaryColor    = lipgloss.Color("#7D56F4")
	cyanColor       = lipgloss.Color("#00ADD8")
	successColor    = lipgloss.Color("#04B575")
	warningColor    = lipgloss.Color("#F8C537")
	dangerColor     = lipgloss.Color("#FF5F87")
	subtleTextColor = lipgloss.Color("#A8B0D3")
	codeColor       = lipgloss.Color("#A8CC8C")

	SectionBannerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Padding(0, 1).
				MarginTop(1).
				MarginBottom(1)

	queryCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginBottom(1)

	tableCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(cyanColor).
			Padding(1, 2).
			MarginBottom(1)

	highSeverityBadge = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(dangerColor).
				Padding(0, 1)

	medSeverityBadge = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(warningColor).
				Padding(0, 1)

	queryIdBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(cyanColor).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	queryTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E6EEF8")).
			Italic(true)

	sqlSnippetStyle = lipgloss.NewStyle().
			Foreground(codeColor).
			Bold(true)

	recHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(dangerColor)
)

func PrintQueryIssues(queryID int64, queryText string, avgTime float64, issues []analyzer.Issue) {
	var b strings.Builder

	badge := queryIdBadge.Render(fmt.Sprintf("QUERY #%d", queryID))
	timeStr := warningStyle.Render(fmt.Sprintf("⏱  Avg Execution: %.2fms", avgTime))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, badge, "  ", timeStr))
	b.WriteString("\n\n")

	displayQuery := strings.Join(strings.Fields(queryText), " ")
	if len(displayQuery) > 100 {
		displayQuery = displayQuery[:97] + "..."
	}
	b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Render("SQL Query:"))
	b.WriteString("\n")
	b.WriteString(queryTextStyle.Render(displayQuery))
	b.WriteString("\n")

	if len(issues) == 0 {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("✓ Execution plan healthy. No obvious bottlenecks detected."))
	} else {
		for _, issue := range issues {
			b.WriteString("\n")
			var sevBadge string
			if strings.EqualFold(issue.Severity, "High") {
				sevBadge = highSeverityBadge.Render(" HIGH ")
			} else {
				sevBadge = medSeverityBadge.Render(" MED ")
			}

			b.WriteString(fmt.Sprintf("%s %s  (Relation: %s)\n", sevBadge, lipgloss.NewStyle().Bold(true).Render(issue.Type), issue.Relation))
			b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Render(fmt.Sprintf("Details: %s", issue.Description)))
			b.WriteString("\n")

			switch issue.Type {
			case "Missing Index":
				sql := fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_optimizer ON %s (/* columns */);", issue.Relation, issue.Relation)
				b.WriteString(recHeaderStyle.Render("💡 Recommendation: "))
				b.WriteString(sqlSnippetStyle.Render(sql))
			case "Memory Starvation":
				sql := "SET work_mem = '64MB';"
				b.WriteString(recHeaderStyle.Render("💡 Recommendation: "))
				b.WriteString(sqlSnippetStyle.Render(sql))
			case "Inefficient Join":
				b.WriteString(recHeaderStyle.Render("💡 Recommendation: "))
				b.WriteString(sqlSnippetStyle.Render("Add an index on the join condition columns to avoid sequential scans."))
			}
			b.WriteString("\n")
		}
	}

	fmt.Println(queryCardStyle.Render(b.String()))
}

func PrintUnusedIndexes(indexes []db.UnusedIndex) {
	fmt.Println(SectionBannerStyle.Render(" 🔍 UNUSED INDEX ANALYSIS "))

	if len(indexes) == 0 {
		fmt.Println(queryCardStyle.Render(successStyle.Render("✓ No unused indexes found. Your write & storage efficiency is optimal!")))
		return
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(warningColor).Render(
		fmt.Sprintf("⚠️  Found %d index(es) with 0 recorded scans. These degrade INSERT/UPDATE throughput and waste disk space.\n\n", len(indexes)),
	))

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555577"))).
		Headers("TABLE", "INDEX NAME", "WASTED SPACE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#E6EEF8"))
		})

	for _, idx := range indexes {
		t.Row(fmt.Sprintf("%s.%s", idx.Schema, idx.Table), idx.IndexName, idx.Size)
	}

	b.WriteString(t.Render())
	b.WriteString("\n\n")
	b.WriteString(recHeaderStyle.Render("💡 Actionable Remediation SQL:"))
	b.WriteString("\n")

	for _, idx := range indexes {
		dropCmd := fmt.Sprintf("DROP INDEX CONCURRENTLY %s.%s;", idx.Schema, idx.IndexName)
		b.WriteString(sqlSnippetStyle.Render("  " + dropCmd))
		b.WriteString("\n")
	}

	fmt.Println(tableCardStyle.Render(b.String()))
}

func PrintMissingFKIndexes(fks []db.MissingFKIndex) {
	fmt.Println(SectionBannerStyle.Render(" 🔗 MISSING FOREIGN KEY INDEX ANALYSIS "))

	if len(fks) == 0 {
		fmt.Println(queryCardStyle.Render(successStyle.Render("✓ All foreign key constraints have covering child indexes. Cascading operations are fast!")))
		return
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(dangerColor).Render(
		fmt.Sprintf("⚠️  Found %d foreign key constraint(s) missing a child table index. Deletes/updates on parent tables will trigger full child table scans.\n\n", len(fks)),
	))

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555577"))).
		Headers("CHILD TABLE", "FK COLUMNS", "CONSTRAINT", "PARENT TABLE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#E6EEF8"))
		})

	for _, fk := range fks {
		t.Row(fmt.Sprintf("%s.%s", fk.Schema, fk.ChildTable), fk.FKColumns, fk.ConstraintName, fk.ParentTable)
	}

	b.WriteString(t.Render())
	b.WriteString("\n\n")
	b.WriteString(recHeaderStyle.Render("💡 Actionable Remediation SQL:"))
	b.WriteString("\n")

	for _, fk := range fks {
		colSanitized := strings.ReplaceAll(strings.ReplaceAll(fk.FKColumns, ", ", "_"), " ", "_")
		indexName := fmt.Sprintf("idx_%s_%s", fk.ChildTable, colSanitized)
		createCmd := fmt.Sprintf("CREATE INDEX CONCURRENTLY %s ON %s.%s (%s);", indexName, fk.Schema, fk.ChildTable, fk.FKColumns)
		b.WriteString(sqlSnippetStyle.Render("  " + createCmd))
		b.WriteString("\n")
	}

	fmt.Println(tableCardStyle.Render(b.String()))
}
