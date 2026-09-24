package report

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abhiraj-ku/pg_adv/internals/health"

	"github.com/charmbracelet/lipgloss"
)

var (
	healthTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00ADD8")).
				Margin(1, 0, 0, 0)

	healthSectionHeaderStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FF5F87"))

	healthLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A8B0D3")).
				Width(18)

	healthValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E6EEF8")).
				Bold(true)

	healthPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF5F87")).
				Background(lipgloss.Color("#120F1D")).
				Padding(1, 2).
				MarginBottom(1)
)

func RenderHealthReport(r *health.Report) {
	var b strings.Builder
	b.WriteString(healthTitleStyle.Render("DATABASE REPORT"))
	b.WriteString("\n\n")

	b.WriteString(renderHealthSection("Server",
		[2]string{"PostgreSQL", r.Server.PGVersion},
		[2]string{"Database", r.Server.DatabaseName},
		[2]string{"Size", r.Server.DatabaseSize},
		[2]string{"Uptime", formatUptime(r.Server.Uptime)},
		[2]string{"In recovery", func() string {
			if r.Server.InRecovery {
				return "yes"
			}
			return "no"
		}()},
	))
	b.WriteString("\n")

	b.WriteString(renderHealthSection("Connections",
		[2]string{"Connections", fmt.Sprintf("%d / %d", r.Connections.TotalConnections, r.Connections.MaxConnections)},
		[2]string{"Active", fmt.Sprintf("%d", r.Connections.Active)},
	))
	b.WriteString("\n")

	cacheColor := "#04B575"
	if r.Cache.HitRatio < 95.0 {
		cacheColor = "#FFAF00"
	}
	if r.Cache.HitRatio < 90.0 {
		cacheColor = "#FF5F87"
	}
	hitStr := lipgloss.NewStyle().Foreground(lipgloss.Color(cacheColor)).Bold(true).Render(fmt.Sprintf("%.1f%%", r.Cache.HitRatio))
	b.WriteString(renderHealthSection("Cache",
		[2]string{"Cache hit", hitStr},
		[2]string{"Cache miss", fmt.Sprintf("%.1f%%", r.Cache.MissRatio)},
	))

	panelWidth := terminalWidth()
	panel := healthPanelStyle.Width(panelWidth).MaxWidth(panelWidth)

	fmt.Println(panel.Render(b.String()))
}

func terminalWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if width, err := strconv.Atoi(cols); err == nil && width > 0 {
			return width
		}
	}
	return 100
}

func renderHealthSection(title string, rows ...[2]string) string {
	var b strings.Builder
	b.WriteString(healthSectionHeaderStyle.Render(title))
	b.WriteString("\n")
	for _, row := range rows {
		b.WriteString(fmt.Sprintf("  %s %s\n", healthLabelStyle.Render(row[0]), healthValueStyle.Render(row[1])))
	}
	return b.String()
}

func printRow(label, value string) {
	fmt.Printf("  %s %s\n", healthLabelStyle.Render(label), healthValueStyle.Render(value))
}

func formatUptime(uptime string) string {
	clean := strings.ReplaceAll(uptime, " days ", "d ")
	clean = strings.ReplaceAll(uptime, " day ", "d ")
	parts := strings.Split(clean, ":")
	if len(parts) >= 2 {
		return fmt.Sprintf("%sh", parts[0])
	}
	return clean
}
