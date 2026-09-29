package report

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderMenu(version string) {
	cw, iw := LayoutBounds()

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor).
		Padding(0, 1).
		MarginBottom(1)

	cmdTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cyanColor).
		Width(18)

	cmdDescStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E6EEF8"))

	subCmdTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(warningColor).
		Width(24)

	sectionHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(dangerColor).
		MarginTop(1).
		MarginBottom(1)

	exampleStyle := lipgloss.NewStyle().
		Foreground(codeColor).
		Bold(true)

	var b strings.Builder

	bannerText := fmt.Sprintf("⚡ iCli — PostgreSQL Query & Index Profiler (%s)", version)
	b.WriteString(headerStyle.Render(bannerText))
	b.WriteString("\n\n")

	introText := "iCli profiles slow queries, inspects EXPLAIN plans, detects N+1 query loops, identifies missing or unused indexes, and audits database health metrics."
	b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Width(iw).Render(introText))
	b.WriteString("\n")

	b.WriteString(sectionHeaderStyle.Render("USAGE"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#E6EEF8")).Render("  $ icli <command> [flags]"))
	b.WriteString("\n")

	b.WriteString(sectionHeaderStyle.Render("AVAILABLE COMMANDS"))
	b.WriteString("\n")

	commands := [][2]string{
		{"analyze", "Analyze top slow queries and EXPLAIN (FORMAT JSON) plan bottlenecks"},
		{"index", "Scan database indexes (unused indexes & missing FK constraints)"},
		{"n1query", "Detect N+1 query patterns and high-frequency application loop queries"},
		{"report", "Generate instant database health, connection & cache metrics summary"},
		{"update", "Check and update iCli binary to the latest GitHub release"},
	}

	for _, cmd := range commands {
		descWidth := iw - 22
		if descWidth < 20 {
			descWidth = 20
		}
		b.WriteString(fmt.Sprintf("  %s %s\n",
			cmdTitleStyle.Render(cmd[0]),
			cmdDescStyle.Width(descWidth).Render(cmd[1]),
		))
	}

	b.WriteString(sectionHeaderStyle.Render("INDEX SUBCOMMANDS"))
	b.WriteString("\n")

	subcommands := [][2]string{
		{"index unused", "Find non-primary, non-unique indexes with 0 recorded scans"},
		{"index missing-fk", "Find foreign key constraints missing covering child indexes"},
	}

	for _, subcmd := range subcommands {
		descWidth := iw - 28
		if descWidth < 20 {
			descWidth = 20
		}
		b.WriteString(fmt.Sprintf("  %s %s\n",
			subCmdTitleStyle.Render(subcmd[0]),
			cmdDescStyle.Width(descWidth).Render(subcmd[1]),
		))
	}

	b.WriteString(sectionHeaderStyle.Render("GLOBAL FLAGS"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s %s\n",
		lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Width(24).Render("-d, --dsn string"),
		cmdDescStyle.Render("PostgreSQL connection string (or set PGDSN env var)"),
	))
	b.WriteString(fmt.Sprintf("  %s %s\n",
		lipgloss.NewStyle().Bold(true).Foreground(cyanColor).Width(24).Render("-h, --help"),
		cmdDescStyle.Render("Display command help and usage details"),
	))

	b.WriteString(sectionHeaderStyle.Render("QUICK START EXAMPLES"))
	b.WriteString("\n")

	examples := []string{
		"# Set connection string in environment",
		"export PGDSN=\"postgres://postgres@localhost:5432/mydb?sslmode=disable\"",
		"",
		"# 1. Profile top 5 slow queries & execution plans",
		"icli analyze -n 5",
		"",
		"# 2. Detect N+1 query patterns & loop bottlenecks",
		"icli n1query -m 50",
		"",
		"# 3. Scan for unused indexes causing write bloat",
		"icli index unused",
		"",
		"# 4. Find unindexed foreign key columns",
		"icli index missing-fk",
		"",
		"# 5. View database health & cache hit ratio",
		"icli report",
	}

	for _, ex := range examples {
		if strings.HasPrefix(ex, "#") {
			b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Render("  " + ex))
		} else if ex != "" {
			b.WriteString(exampleStyle.Render("  $ " + ex))
		}
		b.WriteString("\n")
	}

	footerText := "Run 'icli <command> --help' for detailed info on a specific subcommand."
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Italic(true).Width(iw).Render(footerText))

	card := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginBottom(1).
		Width(cw)

	fmt.Println(card.Render(b.String()))
}

func RenderCLIMenu(version string) {
	RenderMenu(version)
}
