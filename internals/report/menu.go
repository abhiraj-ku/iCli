package report

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderCLIMenu(version string) {
	cardWidth := getCardWidth()
	innerWidth := cardWidth - 4
	if innerWidth < 30 {
		innerWidth = 30
	}

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

	// Header banner
	bannerText := fmt.Sprintf("⚡ iCli — PostgreSQL Query & Index Profiler (%s)", version)
	b.WriteString(headerStyle.Render(bannerText))
	b.WriteString("\n\n")

	introText := "iCli connects to PostgreSQL to profile slow queries, analyze EXPLAIN execution plans, detect unused indexes, and scan for unindexed foreign keys."
	b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Width(innerWidth).Render(introText))
	b.WriteString("\n")

	// Usage section
	b.WriteString(sectionHeaderStyle.Render("USAGE"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#E6EEF8")).Render("  $ icli <command> [flags]"))
	b.WriteString("\n")

	// Commands section
	b.WriteString(sectionHeaderStyle.Render("AVAILABLE COMMANDS"))
	b.WriteString("\n")

	commands := [][2]string{
		{"analyze", "Analyze top slow queries and EXPLAIN (FORMAT JSON) plan bottlenecks"},
		{"index", "Scan database indexes (unused indexes & missing FK constraints)"},
		{"report", "Generate instant database health, connection & cache metrics summary"},
		{"update", "Check and update iCli binary to the latest GitHub release"},
	}

	for _, cmd := range commands {
		descWidth := innerWidth - 22
		if descWidth < 20 {
			descWidth = 20
		}
		b.WriteString(fmt.Sprintf("  %s %s\n",
			cmdTitleStyle.Render(cmd[0]),
			cmdDescStyle.Width(descWidth).Render(cmd[1]),
		))
	}

	// Index subcommands section
	b.WriteString(sectionHeaderStyle.Render("INDEX SUBCOMMANDS"))
	b.WriteString("\n")

	subcommands := [][2]string{
		{"index unused", "Find non-primary, non-unique indexes with 0 recorded scans"},
		{"index missing-fk", "Find foreign key constraints missing covering child indexes"},
	}

	for _, subcmd := range subcommands {
		descWidth := innerWidth - 28
		if descWidth < 20 {
			descWidth = 20
		}
		b.WriteString(fmt.Sprintf("  %s %s\n",
			subCmdTitleStyle.Render(subcmd[0]),
			cmdDescStyle.Width(descWidth).Render(subcmd[1]),
		))
	}

	// Global Flags
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

	// Quick Start Examples
	b.WriteString(sectionHeaderStyle.Render("QUICK START EXAMPLES"))
	b.WriteString("\n")

	examples := []string{
		"# Set connection string in environment",
		"export PGDSN=\"postgres://postgres@localhost:5432/mydb?sslmode=disable\"",
		"",
		"# 1. Profile top 5 slow queries & execution plans",
		"icli analyze -n 5",
		"",
		"# 2. Scan for unused indexes causing write bloat",
		"icli index unused",
		"",
		"# 3. Find unindexed foreign key columns",
		"icli index missing-fk",
		"",
		"# 4. View database health & cache hit ratio",
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
	b.WriteString(lipgloss.NewStyle().Foreground(subtleTextColor).Italic(true).Width(innerWidth).Render(footerText))

	card := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginBottom(1).
		Width(cardWidth)

	fmt.Println(card.Render(b.String()))
}
