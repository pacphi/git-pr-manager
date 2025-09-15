package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cphillipson/multi-gitter-pr-automation/internal/mcp"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// Version information (set by build flags)
var (
	version   = "dev"
	buildTime = "unknown"
	commitSHA = "unknown"
)

func main() {
	// Command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("git-pr-mcp version %s (build: %s, commit: %s)\n", version, buildTime, commitSHA)
		return
	}

	if *showHelp {
		fmt.Println("Git PR MCP Server - Model Context Protocol server for Git PR automation")
		fmt.Printf("Version: %s\n", version)
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  git-pr-mcp [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --version  Show version information")
		fmt.Println("  --help     Show this help message")
		fmt.Println()
		fmt.Println("The MCP server provides natural language interface to Git PR automation.")
		fmt.Println("Connect it to your AI assistant to manage pull requests across repositories.")
		fmt.Println()
		fmt.Println("See documentation for IDE integration instructions.")
		return
	}

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

	logger.Infof("Starting MCP server version %s (build: %s, commit: %s)", version, buildTime, commitSHA)

	// Create and start MCP server
	server := mcp.NewMCPServer(logger)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		logger.Errorf("MCP server error: %v", err)
		os.Exit(1)
	}

	logger.Info("MCP server shutting down gracefully")
}
