package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// HeaderView creates a beautiful ASCII art header with graphics
func (m Model) HeaderView() string {
	// ASCII art with Git theme
	logo := `
   ┌─┐┬┌┬┐   ┌─┐┬─┐   ┌─┐┬  ┬
   │ ┬│ │    ├─┘├┬┘───│  │  │
   └─┘┴ ┴    ┴  ┴└─   └─┘┴─┘┴

   ╭─────────────────────╮
   │   Pull Request      │
   │     Manager         │
   ╰─────────────────────╯
`

	// Create gradient colors for the logo
	gradientColors := []lipgloss.Color{
		lipgloss.Color("5"),  // Magenta
		lipgloss.Color("13"), // Bright Magenta
		lipgloss.Color("12"), // Bright Blue
		lipgloss.Color("6"),  // Cyan
		lipgloss.Color("14"), // Bright Cyan
	}

	// Apply gradient coloring to each line
	lines := strings.Split(strings.TrimSpace(logo), "\n")
	coloredLines := make([]string, len(lines))

	for i, line := range lines {
		colorIndex := i % len(gradientColors)
		style := lipgloss.NewStyle().
			Foreground(gradientColors[colorIndex]).
			Bold(true)
		coloredLines[i] = style.Render(line)
	}

	logoColored := strings.Join(coloredLines, "\n")

	// Status indicators
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Faint(true)

	var status strings.Builder
	status.WriteString(statusStyle.Render("┌─ Status "))
	status.WriteString(statusStyle.Render(strings.Repeat("─", m.width-12)))
	status.WriteString(statusStyle.Render("┐\n"))

	// Current view indicator
	viewName := map[ViewMode]string{
		ViewDashboard:    "Dashboard",
		ViewRepositories: "Repositories",
		ViewPullRequests: "Pull Requests",
		ViewConfig:       "Configuration",
		ViewHelp:         "Help",
	}[m.currentView]

	status.WriteString(statusStyle.Render("│ "))
	status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Current View: "))
	status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render(viewName))

	// Add loading indicator if loading
	if m.loading {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinnerChar := spinner[int(time.Now().UnixMilli()/100)%len(spinner)]
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(fmt.Sprintf(" %s Loading...", spinnerChar)))
	}

	status.WriteString(statusStyle.Render(strings.Repeat(" ", max(0, m.width-len(stripAnsi(status.String()))+2))))
	status.WriteString(statusStyle.Render("│\n"))

	// Last update time
	if !m.lastUpdate.IsZero() {
		status.WriteString(statusStyle.Render("│ "))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Last Update: "))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render(m.lastUpdate.Format("15:04:05")))
		status.WriteString(statusStyle.Render(strings.Repeat(" ", max(0, m.width-30))))
		status.WriteString(statusStyle.Render("│\n"))
	}

	// Repository count
	if len(m.results) > 0 {
		totalPRs := 0
		readyPRs := 0
		for _, result := range m.results {
			if result.Error == nil {
				totalPRs += len(result.PullRequests)
				for _, pr := range result.PullRequests {
					if pr.Ready && !pr.Skipped {
						readyPRs++
					}
				}
			}
		}

		status.WriteString(statusStyle.Render("│ "))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Stats: "))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(fmt.Sprintf("%d repos", len(m.results))))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render(", "))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(fmt.Sprintf("%d PRs", totalPRs)))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render(", "))
		status.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(fmt.Sprintf("%d ready", readyPRs)))
		status.WriteString(statusStyle.Render(strings.Repeat(" ", max(0, m.width-50))))
		status.WriteString(statusStyle.Render("│\n"))
	}

	status.WriteString(statusStyle.Render("└"))
	status.WriteString(statusStyle.Render(strings.Repeat("─", m.width-2)))
	status.WriteString(statusStyle.Render("┘"))

	// Combine logo and status
	return lipgloss.JoinVertical(lipgloss.Center, logoColored, status.String())
}

// Helper function to strip ANSI codes for length calculation
func stripAnsi(str string) string {
	// Simple ANSI escape sequence removal (basic implementation)
	result := ""
	inEscape := false
	for _, char := range str {
		if char == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if char == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(char)
	}
	return result
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
