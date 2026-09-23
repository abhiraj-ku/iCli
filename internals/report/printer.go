package report

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/abhiraj-ku/pg_adv/internals/analyzer"
	"github.com/abhiraj-ku/pg_adv/internals/db"
)

// ANSI color codes for terminal formatting
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

// outputs the bottlneck found in Exlpain ast
func PrintQueryIssues(queryID int64, queryText string, avgTime float64, issues []analyzer.Issue) {
	fmt.Printf("\n%s%s--- Analyzing Query ID: %d ---%s\n", Bold, Cyan, queryID, Reset)

	// reduce long queries for clean CLI display
	displayQuery := queryText
	if len(displayQuery) > 80 {
		displayQuery = displayQuery[:77] + "..."
	}
	fmt.Printf("Query: %s\n", strings.TrimSpace(displayQuery))
	fmt.Printf("Avg Execution Time: %s%.2fms%s\n", Yellow, avgTime, Reset)

	if len(issues) == 0 {
		fmt.Printf("%s✓ No obvious bottlenecks found in execution plan.%s\n", Green, Reset)
		return
	}

	for _, issue := range issues {
		fmt.Printf("\n  %s[%s] %s%s\n", Red, issue.Severity, issue.Type, Reset)
		fmt.Printf("  Relation: %s\n", issue.Relation)
		fmt.Printf("  Details:  %s\n", issue.Description)

		// give actionable SQL advice based on the issue type
		switch issue.Type {
		case "Missing Index":
			fmt.Printf("  %sRECOMMENDATION:%s CREATE INDEX CONCURRENTLY idx_%s_optimizer ON %s (/* columns */);\n",
				Bold, Reset, issue.Relation, issue.Relation)
		case "Memory Starvation":
			fmt.Printf("  %sRECOMMENDATION:%s Increase work_mem for this session (e.g., SET work_mem = '64MB');\n",
				Bold, Reset)
		}
	}
}

// formats the dead-queries indexes into clean table and generate drop index recom
func PrintUnusedIndexes(indexes []db.UnusedIndex) {
	fmt.Printf("\n%s%s=== UNUSED INDEX REPORT ===%s\n", Bold, Yellow, Reset)

	if len(indexes) == 0 {
		fmt.Printf("%s✓ No unused indexes found. Your write performance is healthy!%s\n\n", Green, Reset)
		return
	}

	fmt.Printf("Found %d indexes with 0 scans. These are degrading INSERT/UPDATE performance.\n\n", len(indexes))

	// clean, aligned CLI tables
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%sTable\tIndex Name\tSpace Wasted%s\n", Bold, Reset)
	fmt.Fprintf(w, "-----\t----------\t------------\n")

	for _, idx := range indexes {
		fmt.Fprintf(w, "%s.%s\t%s\t%s\n", idx.Schema, idx.Table, idx.IndexName, idx.Size)
	}
	w.Flush()

	fmt.Printf("\n%sActionable SQL:%s\n", Bold, Reset)
	for _, idx := range indexes {
		fmt.Printf("%sDROP INDEX CONCURRENTLY %s.%s;%s\n", Red, idx.Schema, idx.IndexName, Reset)
	}
	fmt.Println()
}
