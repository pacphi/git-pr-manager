package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pacphi/git-pr-manager/pkg/config"
)

// Run starts the TUI application
func Run(ctx context.Context, cfg *config.Config) error {
	// Create the model
	model := New(ctx, cfg)

	// Create the program with mouse support and alt screen
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Set the context for the program
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
