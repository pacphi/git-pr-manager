package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/pacphi/git-pr-manager/internal/executor"
	"github.com/pacphi/git-pr-manager/pkg/pr"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update list dimensions
		listWidth := msg.Width - 4
		listHeight := msg.Height - 10 // Reserve space for header and footer

		m.repositoryList.SetWidth(listWidth)
		m.repositoryList.SetHeight(listHeight)
		m.prList.SetWidth(listWidth)
		m.prList.SetHeight(listHeight)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Cycle through views
			switch m.currentView {
			case ViewDashboard:
				m.currentView = ViewRepositories
			case ViewRepositories:
				m.currentView = ViewPullRequests
			case ViewPullRequests:
				m.currentView = ViewConfig
			case ViewConfig:
				m.currentView = ViewHelp
			case ViewHelp:
				m.currentView = ViewDashboard
			}

		case "r":
			// Manual refresh
			cmds = append(cmds, func() tea.Msg { return RefreshDataMsg{} })

		case "h":
			// Toggle help view
			if m.currentView == ViewHelp {
				m.currentView = ViewDashboard
			} else {
				m.currentView = ViewHelp
			}

		case "enter":
			// Handle selection based on current view
			switch m.currentView {
			case ViewRepositories:
				if len(m.repositoryList.Items()) > 0 {
					m.currentView = ViewPullRequests
					// Filter PRs for selected repository
					cmds = append(cmds, m.filterPRsForRepo(m.repositoryList.Index()))
				}
			}
		}

	case RefreshDataMsg:
		m.loading = true
		cmds = append(cmds, m.loadData())

	case DataLoadedMsg:
		m.loading = false
		m.error = msg.Error
		m.results = msg.Results
		m.lastUpdate = time.Now()

		// Update repository list
		repoItems := make([]RepositoryItem, 0)
		for _, result := range msg.Results {
			if result.Error == nil {
				repoItems = append(repoItems, RepositoryItem{
					Repository: result.Repository,
					Provider:   result.Provider,
					PRCount:    len(result.PullRequests),
				})
			}
		}

		// Convert to list items
		items := make([]list.Item, len(repoItems))
		for i, item := range repoItems {
			items[i] = item
		}
		m.repositoryList.SetItems(items)

		// Schedule next refresh if auto-refresh is enabled
		if m.autoRefresh {
			cmds = append(cmds, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
				return RefreshDataMsg{}
			}))
		}

	case TickMsg:
		// Handle periodic refresh
		if m.autoRefresh {
			cmds = append(cmds, func() tea.Msg { return RefreshDataMsg{} })
		}
	}

	// Update current view's list component
	switch m.currentView {
	case ViewRepositories:
		m.repositoryList, cmd = m.repositoryList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewPullRequests:
		m.prList, cmd = m.prList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// loadData loads PR data in the background
func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		logger := utils.GetGlobalLogger()

		// Create executor
		exec, err := executor.New(m.config)
		if err != nil {
			return DataLoadedMsg{Error: err}
		}
		defer func() {
			if closeErr := exec.Close(); closeErr != nil {
				logger.WithError(closeErr).Warn("Failed to close executor")
			}
		}()

		// Process PRs
		opts := pr.ProcessOptions{
			DryRun: false,
		}

		results, err := exec.ProcessPRs(m.ctx, opts)
		return DataLoadedMsg{
			Results: results,
			Error:   err,
		}
	}
}

// filterPRsForRepo filters PRs for a specific repository
func (m Model) filterPRsForRepo(repoIndex int) tea.Cmd {
	return func() tea.Msg {
		if repoIndex >= len(m.results) {
			return nil
		}

		result := m.results[repoIndex]
		items := make([]list.Item, 0, len(result.PullRequests))

		for _, prResult := range result.PullRequests {
			items = append(items, PRItem{
				PullRequest: prResult.PullRequest,
				Ready:       prResult.Ready,
				Skipped:     prResult.Skipped,
				Reason:      prResult.Reason,
				Error:       prResult.Error,
			})
		}

		m.prList.SetItems(items)
		return nil
	}
}
