package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/pr"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewRepositories
	ViewPullRequests
	ViewConfig
	ViewHelp
)

// Model represents the TUI application state
type Model struct {
	// Application state
	ctx         context.Context
	config      *config.Config
	width       int
	height      int
	currentView ViewMode

	// Data
	repositories []common.Repository
	results      []pr.ProcessResult
	lastUpdate   time.Time

	// UI Components
	repositoryList list.Model
	prList         list.Model

	// Status
	loading bool
	error   error

	// Navigation
	selectedRepo int
	selectedPR   int

	// Auto-refresh
	autoRefresh  bool
	refreshTimer *time.Timer
}

// Init initializes the TUI model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		func() tea.Msg { return RefreshDataMsg{} },
	)
}

// Messages for the TUI
type RefreshDataMsg struct{}
type DataLoadedMsg struct {
	Results []pr.ProcessResult
	Error   error
}
type WindowSizeMsg struct {
	Width  int
	Height int
}
type TickMsg time.Time

// New creates a new TUI model
func New(ctx context.Context, cfg *config.Config) Model {
	// Initialize repository list
	repositoryList := list.New([]list.Item{}, NewRepositoryDelegate(), 0, 0)
	repositoryList.Title = "Repositories"
	repositoryList.SetShowStatusBar(false)
	repositoryList.SetFilteringEnabled(false)
	repositoryList.SetShowHelp(false)

	// Initialize PR list
	prList := list.New([]list.Item{}, NewPRDelegate(), 0, 0)
	prList.Title = "Pull Requests"
	prList.SetShowStatusBar(false)
	prList.SetFilteringEnabled(false)
	prList.SetShowHelp(false)

	return Model{
		ctx:            ctx,
		config:         cfg,
		currentView:    ViewDashboard,
		repositoryList: repositoryList,
		prList:         prList,
		autoRefresh:    true,
	}
}
