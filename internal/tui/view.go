package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI based on current state
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Build the interface
	header := m.HeaderView()
	content := m.ContentView()
	footer := m.FooterView()

	// Combine all parts
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// ContentView renders the main content area based on current view
func (m Model) ContentView() string {
	switch m.currentView {
	case ViewDashboard:
		return m.DashboardView()
	case ViewRepositories:
		return m.RepositoriesView()
	case ViewPullRequests:
		return m.PullRequestsView()
	case ViewConfig:
		return m.ConfigView()
	case ViewHelp:
		return m.HelpView()
	default:
		return m.DashboardView()
	}
}

// DashboardView renders the main dashboard
func (m Model) DashboardView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 15) // Account for header and footer

	var content strings.Builder

	if m.error != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)
		content.WriteString(errorStyle.Render("Error: " + m.error.Error()))
		content.WriteString("\n\n")
	}

	if len(m.results) == 0 {
		if m.loading {
			content.WriteString("Loading repositories and pull requests...\n")
		} else {
			content.WriteString("No repositories configured or found.\n")
			content.WriteString("Use 'git-pr-cli setup' to configure repositories.\n")
		}
	} else {
		// Summary statistics
		totalRepos := len(m.results)
		totalPRs := 0
		readyPRs := 0
		skippedPRs := 0
		errorCount := 0

		for _, result := range m.results {
			if result.Error != nil {
				errorCount++
				continue
			}
			totalPRs += len(result.PullRequests)
			for _, pr := range result.PullRequests {
				if pr.Error != nil {
					errorCount++
				} else if pr.Skipped {
					skippedPRs++
				} else if pr.Ready {
					readyPRs++
				}
			}
		}

		// Dashboard summary
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("📊 Repository Overview"))
		content.WriteString("\n\n")

		summaryStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1).
			Width(m.width - 12)

		var summary strings.Builder
		summary.WriteString(fmt.Sprintf("🏢 Repositories: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render(fmt.Sprintf("%d", totalRepos))))
		summary.WriteString(fmt.Sprintf("📝 Total PRs:    %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true).Render(fmt.Sprintf("%d", totalPRs))))
		summary.WriteString(fmt.Sprintf("✅ Ready PRs:    %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render(fmt.Sprintf("%d", readyPRs))))
		if skippedPRs > 0 {
			summary.WriteString(fmt.Sprintf("⊝  Skipped PRs:  %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render(fmt.Sprintf("%d", skippedPRs))))
		}
		if errorCount > 0 {
			summary.WriteString(fmt.Sprintf("❌ Errors:       %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render(fmt.Sprintf("%d", errorCount))))
		}

		content.WriteString(summaryStyle.Render(summary.String()))
		content.WriteString("\n\n")

		// Quick actions
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("🚀 Quick Actions"))
		content.WriteString("\n\n")

		actionsStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1).
			Width(m.width - 12)

		var actions strings.Builder
		actions.WriteString("🔍 Press [TAB] to browse repositories\n")
		actions.WriteString("📄 Press [TAB] twice to view pull requests\n")
		actions.WriteString("🔄 Press [r] to refresh data\n")
		actions.WriteString("⚙️  Press [TAB] three times for configuration\n")
		actions.WriteString("❓ Press [h] for help\n")

		content.WriteString(actionsStyle.Render(actions.String()))
	}

	return style.Render(content.String())
}

// RepositoriesView renders the repository browser
func (m Model) RepositoriesView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 15)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("🏢 Repository Browser"))
	content.WriteString("\n\n")

	if len(m.repositoryList.Items()) == 0 {
		content.WriteString("No repositories found.\n")
		content.WriteString("Press [r] to refresh or configure repositories with 'git-pr-cli setup'.\n")
	} else {
		content.WriteString("Use ↑/↓ to navigate, [ENTER] to view PRs, [TAB] to change view.\n\n")
		content.WriteString(m.repositoryList.View())
	}

	return style.Render(content.String())
}

// PullRequestsView renders the pull request list
func (m Model) PullRequestsView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("11")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 15)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true).Render("📝 Pull Requests"))
	content.WriteString("\n\n")

	if len(m.prList.Items()) == 0 {
		content.WriteString("No pull requests found for the selected repository.\n")
		content.WriteString("Press [TAB] to go back to repositories or [r] to refresh.\n")
	} else {
		content.WriteString("Use ↑/↓ to navigate, [ENTER] for details, [TAB] to change view.\n\n")
		content.WriteString(m.prList.View())
	}

	return style.Render(content.String())
}

// ConfigView renders the configuration view
func (m Model) ConfigView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("13")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 15)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Render("⚙️  Configuration"))
	content.WriteString("\n\n")

	if m.config != nil {
		content.WriteString("📁 Configuration loaded\n")
		content.WriteString(fmt.Sprintf("🏢 Total repositories: %d\n", len(m.config.Repositories)))
		content.WriteString(fmt.Sprintf("🔄 Auto-refresh: %t\n", m.autoRefresh))
		hasNotifications := m.config.Notifications.Slack.Enabled || m.config.Notifications.Email.Enabled
		content.WriteString(fmt.Sprintf("📧 Notifications: %t\n", hasNotifications))

		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Use 'git-pr-cli setup' to modify configuration."))
	} else {
		content.WriteString("No configuration loaded.\n")
		content.WriteString("Run 'git-pr-cli setup' to create a configuration.\n")
	}

	return style.Render(content.String())
}

// HelpView renders the help view
func (m Model) HelpView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("14")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 15)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Render("❓ Help & Keyboard Shortcuts"))
	content.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1).
		Width(m.width - 12)

	var help strings.Builder
	help.WriteString(lipgloss.NewStyle().Underline(true).Render("Navigation:"))
	help.WriteString("\n")
	help.WriteString("  [TAB]        - Cycle through views\n")
	help.WriteString("  [↑/↓]        - Navigate lists\n")
	help.WriteString("  [ENTER]      - Select item\n")
	help.WriteString("  [h]          - Toggle help\n")
	help.WriteString("  [r]          - Refresh data\n")
	help.WriteString("  [q/Ctrl+C]   - Quit\n\n")

	help.WriteString(lipgloss.NewStyle().Underline(true).Render("Views:"))
	help.WriteString("\n")
	help.WriteString("  📊 Dashboard     - Overview and statistics\n")
	help.WriteString("  🏢 Repositories  - Browse repositories\n")
	help.WriteString("  📝 Pull Requests - View PRs for selected repo\n")
	help.WriteString("  ⚙️  Configuration - View current settings\n")
	help.WriteString("  ❓ Help          - This help screen\n\n")

	help.WriteString(lipgloss.NewStyle().Underline(true).Render("Status Icons:"))
	help.WriteString("\n")
	help.WriteString("  ✓ Ready to merge\n")
	help.WriteString("  ⧖ Not ready yet\n")
	help.WriteString("  ⊝ Skipped\n")
	help.WriteString("  ✗ Error\n")

	content.WriteString(helpStyle.Render(help.String()))

	return style.Render(content.String())
}

// FooterView renders the status bar footer
func (m Model) FooterView() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("8")).
		Padding(0, 1).
		Width(m.width)

	var footer strings.Builder
	footer.WriteString("git-pr-cli TUI")

	// Add current view indicator
	viewName := map[ViewMode]string{
		ViewDashboard:    "Dashboard",
		ViewRepositories: "Repositories",
		ViewPullRequests: "Pull Requests",
		ViewConfig:       "Config",
		ViewHelp:         "Help",
	}[m.currentView]

	footer.WriteString(" • ")
	footer.WriteString(viewName)

	// Add shortcuts
	footer.WriteString(" • [q]uit [TAB]cycle [r]efresh [h]elp")

	// Pad to full width
	padding := m.width - len(stripAnsi(footer.String()))
	if padding > 0 {
		footer.WriteString(strings.Repeat(" ", padding))
	}

	return style.Render(footer.String())
}
