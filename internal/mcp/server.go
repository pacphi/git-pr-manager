package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cphillipson/multi-gitter-pr-automation/internal/cli/commands"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// MCPServer implements the Model Context Protocol server using mark3labs/mcp-go
type MCPServer struct {
	server *server.MCPServer
	logger *utils.Logger
	cfg    *config.Config
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(logger *utils.Logger) *MCPServer {
	// Create the mark3labs MCP server
	mcpServer := server.NewMCPServer(
		"Git PR MCP Server",
		"1.0.0",
	)

	return &MCPServer{
		server: mcpServer,
		logger: logger,
	}
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server")

	// Load configuration if it exists
	loader := config.NewLoader()
	if cfg, err := loader.Load(""); err == nil {
		s.cfg = cfg
	}

	// Register tools and resources
	s.registerTools()
	s.registerResources()

	// Start the server using stdio
	return server.ServeStdio(s.server)
}

// registerTools registers all available tools with the MCP server
func (s *MCPServer) registerTools() {
	// Setup repositories tool
	setupTool := mcp.NewTool("setup_repositories",
		mcp.WithDescription("Run interactive setup wizard to configure repositories"),
	)
	s.server.AddTool(setupTool, s.handleSetupTool)

	// Validate configuration tool
	validateTool := mcp.NewTool("validate_configuration",
		mcp.WithDescription("Validate configuration file and connectivity"),
	)
	s.server.AddTool(validateTool, s.handleValidateTool)

	// Check pull requests tool
	checkTool := mcp.NewTool("check_pull_requests",
		mcp.WithDescription("Check PR status across all repositories"),
	)
	s.server.AddTool(checkTool, s.handleCheckTool)

	// Merge pull requests tool
	mergeTool := mcp.NewTool("merge_pull_requests",
		mcp.WithDescription("Merge ready pull requests"),
	)
	s.server.AddTool(mergeTool, s.handleMergeTool)

	// Watch repositories tool
	watchTool := mcp.NewTool("watch_repositories",
		mcp.WithDescription("Monitor PR status continuously"),
	)
	s.server.AddTool(watchTool, s.handleWatchTool)

	// Get repository statistics tool
	statsTool := mcp.NewTool("get_repository_statistics",
		mcp.WithDescription("Get detailed repository statistics"),
	)
	s.server.AddTool(statsTool, s.handleStatsTool)
}

// registerResources registers all available resources with the MCP server
func (s *MCPServer) registerResources() {
	// Current configuration resource
	configResource := mcp.NewResource(
		"config://current",
		"Current Configuration",
	)
	s.server.AddResource(configResource, s.handleConfigResource)

	// Repository statistics resource
	statsResource := mcp.NewResource(
		"stats://repositories",
		"Repository Statistics",
	)
	s.server.AddResource(statsResource, s.handleStatsResource)

	// Environment status resource
	envResource := mcp.NewResource(
		"env://status",
		"Environment Status",
	)
	s.server.AddResource(envResource, s.handleEnvResource)

	// Command help resource
	helpResource := mcp.NewResource(
		"help://commands",
		"Command Help",
	)
	s.server.AddResource(helpResource, s.handleHelpResource)
}

// Tool handlers - these match the mark3labs/mcp-go tool handler signature
func (s *MCPServer) handleSetupTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	result, err := s.executeSetup(args)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(result), nil
}

func (s *MCPServer) handleValidateTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	result, err := s.executeValidate(args)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(result), nil
}

func (s *MCPServer) handleCheckTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	result, err := s.executeCheck(args)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(result), nil
}

func (s *MCPServer) handleMergeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	result, err := s.executeMerge(args)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(result), nil
}

func (s *MCPServer) handleWatchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	result, err := s.executeWatch(args)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(result), nil
}

func (s *MCPServer) handleStatsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	result, err := s.executeStats(args)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(result), nil
}

// Resource handlers - these match the mark3labs/mcp-go resource handler signature
func (s *MCPServer) handleConfigResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	content, err := s.readConfigResource()
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "application/json",
		Text:     content,
	}}, nil
}

func (s *MCPServer) handleStatsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	content, err := s.readStatsResource()
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "application/json",
		Text:     content,
	}}, nil
}

func (s *MCPServer) handleEnvResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	content, err := s.readEnvResource()
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "application/json",
		Text:     content,
	}}, nil
}

func (s *MCPServer) handleHelpResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	content, err := s.readHelpResource()
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "text/plain",
		Text:     content,
	}}, nil
}

// Helper methods for executing tools and reading resources

func (s *MCPServer) executeSetup(args map[string]interface{}) (string, error) {
	// Create a setup command and execute it
	cmd := commands.NewSetupCommand()
	if force, ok := args["force"].(bool); ok && force {
		if err := cmd.Flags().Set("force", "true"); err != nil {
			return "", fmt.Errorf("failed to set force flag: %w", err)
		}
	}

	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) executeValidate(args map[string]interface{}) (string, error) {
	cmd := commands.NewValidateCommand()

	if checkAuth, ok := args["check_auth"].(bool); ok && checkAuth {
		if err := cmd.Flags().Set("check-auth", "true"); err != nil {
			return "", fmt.Errorf("failed to set check-auth flag: %w", err)
		}
	}
	if checkRepos, ok := args["check_repos"].(bool); ok && checkRepos {
		if err := cmd.Flags().Set("check-repos", "true"); err != nil {
			return "", fmt.Errorf("failed to set check-repos flag: %w", err)
		}
	}
	if verbose, ok := args["verbose"].(bool); ok && verbose {
		if err := cmd.Flags().Set("verbose", "true"); err != nil {
			return "", fmt.Errorf("failed to set verbose flag: %w", err)
		}
	}

	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) executeCheck(args map[string]interface{}) (string, error) {
	cmd := commands.NewCheckCommand()

	if provider, ok := args["provider"].(string); ok {
		if err := cmd.Flags().Set("provider", provider); err != nil {
			return "", fmt.Errorf("failed to set provider flag: %w", err)
		}
	}
	if repos, ok := args["repos"].(string); ok {
		if err := cmd.Flags().Set("repos", repos); err != nil {
			return "", fmt.Errorf("failed to set repos flag: %w", err)
		}
	}
	if format, ok := args["format"].(string); ok {
		if err := cmd.Flags().Set("format", format); err != nil {
			return "", fmt.Errorf("failed to set format flag: %w", err)
		}
	}
	if showDetails, ok := args["show_details"].(bool); ok && showDetails {
		if err := cmd.Flags().Set("show-details", "true"); err != nil {
			return "", fmt.Errorf("failed to set show-details flag: %w", err)
		}
	}

	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) executeMerge(args map[string]interface{}) (string, error) {
	cmd := commands.NewMergeCommand()

	if provider, ok := args["provider"].(string); ok {
		if err := cmd.Flags().Set("provider", provider); err != nil {
			return "", fmt.Errorf("failed to set provider flag: %w", err)
		}
	}
	if repos, ok := args["repos"].(string); ok {
		if err := cmd.Flags().Set("repos", repos); err != nil {
			return "", fmt.Errorf("failed to set repos flag: %w", err)
		}
	}
	if dryRun, ok := args["dry_run"].(bool); ok && dryRun {
		if err := cmd.Flags().Set("dry-run", "true"); err != nil {
			return "", fmt.Errorf("failed to set dry-run flag: %w", err)
		}
	}
	if force, ok := args["force"].(bool); ok && force {
		if err := cmd.Flags().Set("force", "true"); err != nil {
			return "", fmt.Errorf("failed to set force flag: %w", err)
		}
	}

	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) executeWatch(args map[string]interface{}) (string, error) {
	cmd := commands.NewWatchCommand()

	if interval, ok := args["interval"].(string); ok {
		if err := cmd.Flags().Set("interval", interval); err != nil {
			return "", fmt.Errorf("failed to set interval flag: %w", err)
		}
	}
	if maxIter, ok := args["max_iterations"].(float64); ok {
		if err := cmd.Flags().Set("max-iterations", fmt.Sprintf("%.0f", maxIter)); err != nil {
			return "", fmt.Errorf("failed to set max-iterations flag: %w", err)
		}
	}
	if provider, ok := args["provider"].(string); ok {
		if err := cmd.Flags().Set("provider", provider); err != nil {
			return "", fmt.Errorf("failed to set provider flag: %w", err)
		}
	}
	if repos, ok := args["repos"].(string); ok {
		if err := cmd.Flags().Set("repos", repos); err != nil {
			return "", fmt.Errorf("failed to set repos flag: %w", err)
		}
	}

	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) executeStats(args map[string]interface{}) (string, error) {
	cmd := commands.NewStatsCommand()

	if provider, ok := args["provider"].(string); ok {
		if err := cmd.Flags().Set("provider", provider); err != nil {
			return "", fmt.Errorf("failed to set provider flag: %w", err)
		}
	}
	if format, ok := args["format"].(string); ok {
		if err := cmd.Flags().Set("format", format); err != nil {
			return "", fmt.Errorf("failed to set format flag: %w", err)
		}
	}
	if period, ok := args["period"].(string); ok {
		if err := cmd.Flags().Set("period", period); err != nil {
			return "", fmt.Errorf("failed to set period flag: %w", err)
		}
	}
	if sort, ok := args["sort"].(string); ok {
		if err := cmd.Flags().Set("sort", sort); err != nil {
			return "", fmt.Errorf("failed to set sort flag: %w", err)
		}
	}
	if top, ok := args["top"].(float64); ok {
		if err := cmd.Flags().Set("top", fmt.Sprintf("%.0f", top)); err != nil {
			return "", fmt.Errorf("failed to set top flag: %w", err)
		}
	}

	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) executeCommand(cmd *cobra.Command, args []string) (string, error) {
	// Capture command output
	var output strings.Builder
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute command
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, output.String())
	}

	return output.String(), nil
}

func (s *MCPServer) readConfigResource() (string, error) {
	if s.cfg != nil {
		data, err := json.MarshalIndent(s.cfg, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Try to read config file directly
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, ".config", "git-pr-cli", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "No configuration found", nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *MCPServer) readStatsResource() (string, error) {
	// Execute stats command and return JSON output
	cmd := commands.NewStatsCommand()
	if err := cmd.Flags().Set("format", "json"); err != nil {
		return "", fmt.Errorf("failed to set format flag: %w", err)
	}
	return s.executeCommand(cmd, []string{})
}

func (s *MCPServer) readEnvResource() (string, error) {
	envStatus := map[string]interface{}{
		"github_token":           os.Getenv("GITHUB_TOKEN") != "",
		"gitlab_token":           os.Getenv("GITLAB_TOKEN") != "",
		"bitbucket_username":     os.Getenv("BITBUCKET_USERNAME") != "",
		"bitbucket_app_password": os.Getenv("BITBUCKET_APP_PASSWORD") != "",
		"config_path":            filepath.Join(os.Getenv("HOME"), ".config", "git-pr-cli", "config.yaml"),
	}

	data, err := json.MarshalIndent(envStatus, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *MCPServer) readHelpResource() (string, error) {
	help := `Git PR CLI Commands:

setup    - Interactive setup wizard for repositories
check    - Check PR status across repositories
merge    - Merge ready pull requests
watch    - Monitor PR status continuously
stats    - Get repository statistics
validate - Validate configuration and connectivity
test     - Test notification setup

Use --help with any command for detailed usage information.

Examples:
  git-pr-cli check --format json
  git-pr-cli merge --dry-run
  git-pr-cli watch --interval 5m
  git-pr-cli stats --period 7d --format csv
`
	return help, nil
}
