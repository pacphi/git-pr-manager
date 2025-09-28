package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pacphi/git-pr-manager/pkg/providers/common"
)

// RepositoryItem represents a repository in the list
type RepositoryItem struct {
	Repository common.Repository
	Provider   string
	PRCount    int
}

func (i RepositoryItem) FilterValue() string {
	return i.Repository.FullName
}

// PRItem represents a pull request in the list
type PRItem struct {
	PullRequest common.PullRequest
	Ready       bool
	Skipped     bool
	Reason      string
	Error       error
}

func (i PRItem) FilterValue() string {
	return i.PullRequest.Title
}

// RepositoryDelegate handles rendering of repository items
type RepositoryDelegate struct{}

func NewRepositoryDelegate() RepositoryDelegate {
	return RepositoryDelegate{}
}

func (d RepositoryDelegate) Height() int                               { return 1 }
func (d RepositoryDelegate) Spacing() int                              { return 0 }
func (d RepositoryDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d RepositoryDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	if i, ok := listItem.(RepositoryItem); ok {
		str := fmt.Sprintf("%d. %s", index+1, i.Repository.FullName)

		fn := itemStyle.Render
		if index == m.Index() {
			str = "> " + str
			fn = selectedItemStyle.Render
		}

		fmt.Fprint(w, fn(str))
	}
}

// PRDelegate handles rendering of pull request items
type PRDelegate struct{}

func NewPRDelegate() PRDelegate {
	return PRDelegate{}
}

func (d PRDelegate) Height() int                               { return 2 }
func (d PRDelegate) Spacing() int                              { return 1 }
func (d PRDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d PRDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	if i, ok := listItem.(PRItem); ok {
		var statusIcon string
		var statusColor lipgloss.Color

		if i.Error != nil {
			statusIcon = "✗"
			statusColor = lipgloss.Color("1") // Red
		} else if i.Skipped {
			statusIcon = "⊝"
			statusColor = lipgloss.Color("3") // Yellow
		} else if i.Ready {
			statusIcon = "✓"
			statusColor = lipgloss.Color("2") // Green
		} else {
			statusIcon = "⧖"
			statusColor = lipgloss.Color("6") // Cyan
		}

		title := fmt.Sprintf("#%d %s", i.PullRequest.Number, i.PullRequest.Title)
		reason := i.Reason

		if len(title) > 60 {
			title = title[:57] + "..."
		}
		if len(reason) > 60 {
			reason = reason[:57] + "..."
		}

		str := fmt.Sprintf("%s %s\n   %s",
			lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
			title,
			reason)

		fn := itemStyle.Render
		if index == m.Index() {
			fn = selectedItemStyle.Render
		}

		fmt.Fprint(w, fn(str))
	}
}

// Styles for list items
var (
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))
)
