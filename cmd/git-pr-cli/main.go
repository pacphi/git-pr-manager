package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pacphi/git-pr-manager/internal/cli/commands"
	"github.com/pacphi/git-pr-manager/internal/tui"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// Version information (set by build flags)
var (
	version   = "dev"
	buildTime = "unknown"
	commitSHA = "unknown"
)

func main() {
	// Set up context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Initialize logger
	logger := utils.NewLogger()
	utils.SetGlobalLogger(logger)

	// Check if we should launch TUI mode (no arguments provided)
	if len(os.Args) == 1 {
		// Try to load config for TUI mode
		cfg, err := commands.LoadConfig()
		if err != nil {
			// If config loading fails, fall back to CLI help
			fmt.Fprintf(os.Stderr, "Failed to load configuration for TUI mode: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run 'git-pr-cli setup' to create a configuration, or use CLI commands directly.\n")
			fmt.Fprintf(os.Stderr, "Use 'git-pr-cli --help' for available commands.\n")
			os.Exit(1)
		}

		// Launch TUI
		if err := tui.Run(ctx, cfg); err != nil {
			logger.WithError(err).Error("TUI execution failed")
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create root command for CLI mode
	rootCmd := commands.NewRootCommand(version, buildTime, commitSHA)

	// Execute command
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.WithError(err).Error("Command execution failed")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
