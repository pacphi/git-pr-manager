package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pacphi/git-pr-manager/internal/cli/commands"
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

	// Create root command
	rootCmd := commands.NewRootCommand(version, buildTime, commitSHA)

	// Execute command
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.WithError(err).Error("Command execution failed")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
