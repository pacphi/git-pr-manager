// Package main provides the MCP server entry point for multi-gitter PR automation.
package main

import (
	"fmt"
	"log"

	mcpServer "github.com/cphillipson/multi-gitter-pr-automation/mcp-server/server"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "Multi-Gitter PR Automation"
	serverVersion = "1.0.0"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithResourceCapabilities(false, false),
		server.WithLogging(),
	)

	// Add all tools
	tools := mcpServer.CreateTools()
	for _, tool := range tools {
		if err := addToolHandler(s, tool); err != nil {
			log.Fatalf("Failed to add tool %s: %v", tool.Name, err)
		}
	}

	// Add all resources
	resources := mcpServer.CreateResources()
	for _, resource := range resources {
		if err := addResourceHandler(s, resource); err != nil {
			log.Fatalf("Failed to add resource %s: %v", resource.Name, err)
		}
	}

	// Start server
	log.Printf("Starting %s v%s", serverName, serverVersion)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func addToolHandler(s *server.MCPServer, tool mcp.Tool) error {
	switch tool.Name {
	case "setup_repositories":
		s.AddTool(tool, mcpServer.HandleSetupRepositories)
	case "validate_config":
		s.AddTool(tool, mcpServer.HandleValidateConfig)
	case "backup_restore_config":
		s.AddTool(tool, mcpServer.HandleBackupRestoreConfig)
	case "check_pull_requests":
		s.AddTool(tool, mcpServer.HandleCheckPullRequests)
	case "merge_pull_requests":
		s.AddTool(tool, mcpServer.HandleMergePullRequests)
	case "watch_repositories":
		s.AddTool(tool, mcpServer.HandleWatchRepositories)
	case "get_repository_stats":
		s.AddTool(tool, mcpServer.HandleGetRepositoryStats)
	case "test_notifications":
		s.AddTool(tool, mcpServer.HandleTestNotifications)
	case "lint_scripts":
		s.AddTool(tool, mcpServer.HandleLintScripts)
	case "check_dependencies":
		s.AddTool(tool, mcpServer.HandleCheckDependencies)
	case "install_dependencies":
		s.AddTool(tool, mcpServer.HandleInstallDependencies)
	default:
		return fmt.Errorf("unknown tool: %s", tool.Name)
	}

	return nil
}

func addResourceHandler(s *server.MCPServer, resource *mcp.Resource) error {
	switch resource.URI {
	case "config://current":
		s.AddResource(*resource, mcpServer.HandleConfigResource)
	case "stats://repositories":
		s.AddResource(*resource, mcpServer.HandleRepositoryStatsResource)
	case "makefile://targets":
		s.AddResource(*resource, mcpServer.HandleMakefileTargetsResource)
	case "env://status":
		s.AddResource(*resource, mcpServer.HandleEnvironmentStatusResource)
	default:
		return fmt.Errorf("unknown resource URI: %s", resource.URI)
	}

	return nil
}
