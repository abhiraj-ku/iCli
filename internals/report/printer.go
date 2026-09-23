package report

import (
	"fmt"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Define centralized Lipgloss styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ADD8")). // Go/Postgres Blue
			MarginTop(1).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8C537"))

	alertCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginBottom(1)

	recommendationStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF5F87"))

	sqlSnippetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8CC8C")).
			Italic(true)
)

// PrintQueryIssues formats the bottlenecks into bordered alert cards.
func PrintQueryIssues(queryID int64, queryText string, avgTime float64, issues []analyzer.Issue) {
	fmt.Println(titleStyle.Render(fmt.Sprintf("--- Analyzing Query ID: %d ---", queryID)))

	// Replace newlines and tabs with spaces to keep the query on one line.
	displayQuery := strings.Join(strings.Fields(queryText), " ")
	if len(displayQuery) > 80 {
		displayQuery = displayQuery[:77] + "..."
	}

	fmt.Printf("Query: %s\n", displayQuery)
	fmt.Printf("Avg Execution Time: %s\n\n", warningStyle.Render(fmt.Sprintf("%.2fms", avgTime)))

	if len(issues) == 0 {
		fmt.Println(successStyle.Render("✓ No obvious bottlenecks found in execution plan."))
		return
	}

	for _, issue := range issues {
		// Build the content for the alert card
		var cardContent string
		cardContent += fmt.Sprintf("[%s] %s\n", issue.Severity, issue.Type)
		cardContent += fmt.Sprintf("Relation: %s\n", issue.Relation)
		cardContent += fmt.Sprintf("Details:  %s\n\n", issue.Description)

		switch issue.Type {
		case "Missing Index":
			rec := recommendationStyle.Render("RECOMMENDATION:")
			sql := sqlSnippetStyle.Render(fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_optimizer ON %s (/* columns */);", issue.Relation, issue.Relation))
			cardContent += fmt.Sprintf("%s %s", rec, sql)
		case "Memory Starvation":
			rec := recommendationStyle.Render("RECOMMENDATION:")
			sql := sqlSnippetStyle.Render("Increase work_mem for this session (e.g., SET work_mem = '64MB');")
			cardContent += fmt.Sprintf("%s %s", rec, sql)
		}

		// Render the bordered card
		fmt.Println(alertCardStyle.Render(cardContent))
	}
}

// PrintUnusedIndexes formats the dead-weight indexes into a bordered table.
func PrintUnusedIndexes(indexes []db.UnusedIndex) {
	fmt.Println(titleStyle.Render("=== UNUSED INDEX REPORT ==="))

	if len(indexes) == 0 {
		fmt.Println(successStyle.Render("✓ No unused indexes found. Your write performance is healthy!\n"))
		return
	}

	fmt.Printf("Found %d indexes with 0 scans. These are degrading INSERT/UPDATE performance.\n\n", len(indexes))

	// Initialize a new lipgloss Table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#999999"))).
		Headers("TABLE", "INDEX NAME", "SPACE WASTED").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ADD8")).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

	for _, idx := range indexes {
		t.Row(fmt.Sprintf("%s.%s", idx.Schema, idx.Table), idx.IndexName, idx.Size)
	}

	// Render the table
	fmt.Println(t.Render())

	fmt.Println(titleStyle.Render("\nActionable SQL:"))
	for _, idx := range indexes {
		fmt.Println(sqlSnippetStyle.Render(fmt.Sprintf("DROP INDEX CONCURRENTLY %s.%s;", idx.Schema, idx.IndexName)))
	}
	fmt.Println()
}
