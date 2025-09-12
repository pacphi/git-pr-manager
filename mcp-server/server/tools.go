package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/cphillipson/multi-gitter-pr-automation/mcp-server/internal/config"
	"github.com/cphillipson/multi-gitter-pr-automation/mcp-server/internal/executor"
)

// Helper functions for parameter access

// handleJSONOutput handles JSON output formatting for commands
func handleJSONOutput(exec *executor.Executor, makeTarget string) (*mcp.CallToolResult, error) {
	if err := os.Setenv("OUTPUT_FORMAT", "json"); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to set OUTPUT_FORMAT: %v", err)), nil
	}
	result := exec.ExecuteMake(makeTarget)
	
	// Try to parse JSON output
	var jsonResult interface{}
	if result.Success && result.Output != "" {
		if err := json.Unmarshal([]byte(result.Output), &jsonResult); err == nil {
			response := map[string]interface{}{
				"success": true,
				"data":    jsonResult,
			}
			return mcp.NewToolResultText(formatResponse(response)), nil
		}
	}
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	return mcp.NewToolResultText(formatResponse(response)), nil
}

func getStringParam(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

func getFloatParam(args map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := args[key].(float64); ok {
		return val
	}
	if val, ok := args[key].(int); ok {
		return float64(val)
	}
	return defaultValue
}

func requireStringParam(args map[string]interface{}, key string) (string, error) {
	if val, ok := args[key].(string); ok && val != "" {
		return val, nil
	}
	return "", fmt.Errorf("required parameter '%s' is missing or empty", key)
}

// CreateTools creates all MCP tools for the PR automation server
func CreateTools() []mcp.Tool {
	return []mcp.Tool{
		// Setup and Configuration Tools
		createSetupRepositoriesTool(),
		createValidateConfigTool(),
		createBackupRestoreConfigTool(),
		
		// PR Management Tools
		createCheckPullRequestsTool(),
		createMergePullRequestsTool(),
		createWatchRepositoriesTool(),
		
		// Repository Tools
		createGetRepositoryStatsTool(),
		createTestNotificationsTool(),
		createLintScriptsTool(),
		
		// Utility Tools
		createCheckDependenciesTool(),
		createInstallDependenciesTool(),
	}
}

// Setup and Configuration Tools

func createSetupRepositoriesTool() mcp.Tool {
	return mcp.NewTool("setup_repositories",
		mcp.WithDescription("Run the interactive setup wizard to configure repositories automatically"),
		mcp.WithString("mode", mcp.Description("Setup mode: 'full', 'wizard', 'preview', or 'additive'")),
	)
}

func createValidateConfigTool() mcp.Tool {
	return mcp.NewTool("validate_config",
		mcp.WithDescription("Validate the configuration file and check for common issues"),
		mcp.WithString("config_file", mcp.Description("Path to config file (default: config.yaml)")),
	)
}

func createBackupRestoreConfigTool() mcp.Tool {
	return mcp.NewTool("backup_restore_config",
		mcp.WithDescription("Backup or restore configuration files"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: 'backup' or 'restore'")),
		mcp.WithString("config_file", mcp.Description("Path to config file (default: config.yaml)")),
	)
}

// PR Management Tools

func createCheckPullRequestsTool() mcp.Tool {
	return mcp.NewTool("check_pull_requests",
		mcp.WithDescription("Check pull request status across all configured repositories"),
		mcp.WithString("provider", mcp.Description("Filter by provider: github, gitlab, bitbucket")),
		mcp.WithString("output_format", mcp.Description("Output format: 'text' or 'json'")),
		mcp.WithString("config_file", mcp.Description("Path to config file (default: config.yaml)")),
	)
}

func createMergePullRequestsTool() mcp.Tool {
	return mcp.NewTool("merge_pull_requests",
		mcp.WithDescription("Merge ready pull requests across configured repositories"),
		mcp.WithString("mode", mcp.Description("Mode: 'merge', 'dry-run', or 'force'")),
		mcp.WithString("config_file", mcp.Description("Path to config file (default: config.yaml)")),
	)
}

func createWatchRepositoriesTool() mcp.Tool {
	return mcp.NewTool("watch_repositories",
		mcp.WithDescription("Continuously monitor PR status with periodic refresh"),
		mcp.WithNumber("interval", mcp.Description("Refresh interval in seconds (default: 30)")),
		mcp.WithString("config_file", mcp.Description("Path to config file (default: config.yaml)")),
	)
}

// Repository Tools

func createGetRepositoryStatsTool() mcp.Tool {
	return mcp.NewTool("get_repository_stats",
		mcp.WithDescription("Get statistics about configured repositories"),
		mcp.WithString("config_file", mcp.Description("Path to config file (default: config.yaml)")),
	)
}

func createTestNotificationsTool() mcp.Tool {
	return mcp.NewTool("test_notifications",
		mcp.WithDescription("Test Slack and email notification configuration"),
	)
}

func createLintScriptsTool() mcp.Tool {
	return mcp.NewTool("lint_scripts",
		mcp.WithDescription("Lint shell scripts using shellcheck"),
	)
}

// Utility Tools

func createCheckDependenciesTool() mcp.Tool {
	return mcp.NewTool("check_dependencies",
		mcp.WithDescription("Check if required dependencies (yq, jq, curl, gh) are installed"),
	)
}

func createInstallDependenciesTool() mcp.Tool {
	return mcp.NewTool("install_dependencies",
		mcp.WithDescription("Install required dependencies automatically"),
		mcp.WithString("platform", mcp.Description("Target platform: macos, linux (auto-detected if not specified)")),
	)
}

// Tool Handlers

// HandleSetupRepositories runs the interactive setup wizard to configure repositories
func HandleSetupRepositories(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	
	// Determine which setup command to run
	args := request.Params.Arguments
	mode := getStringParam(args, "mode", "")
	
	var makeTarget string
	switch mode {
	case "full":
		makeTarget = "setup-full"
	case "preview":
		makeTarget = "wizard-preview"
	case "additive":
		makeTarget = "wizard-additive"
	default:
		makeTarget = "setup-wizard"
	}
	
	result := exec.ExecuteMake(makeTarget)
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
		"action":  makeTarget,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleValidateConfig validates the configuration file and checks for common issues
func HandleValidateConfig(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	configFile := getStringParam(args, "config_file", "config.yaml")
	
	// Use config manager to validate
	configMgr := config.NewManager(configFile)
	if err := configMgr.Load(); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"config_file": configFile,
		}
		return mcp.NewToolResultText(formatResponse(response)), err
	}
	
	if err := configMgr.Validate(); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"config_file": configFile,
		}
		return mcp.NewToolResultText(formatResponse(response)), err
	}
	
	stats := configMgr.GetRepositoryStats()
	
	response := map[string]interface{}{
		"success":     true,
		"message":     "Configuration is valid",
		"config_file": configMgr.GetConfigPath(),
		"stats":       stats,
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleBackupRestoreConfig backs up or restores configuration files
func HandleBackupRestoreConfig(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	action, err := requireStringParam(args, "action")
	if err != nil {
		return mcp.NewToolResultError("action parameter is required"), err
	}
	
	exec := executor.NewExecutor("..")
	
	var makeTarget string
	switch action {
	case "backup":
		makeTarget = "backup-config"
	case "restore":
		makeTarget = "restore-config"
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid action: %s. Must be 'backup' or 'restore'", action)), nil
	}
	
	result := exec.ExecuteMake(makeTarget)
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
		"action":  action,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleCheckPullRequests checks pull request status across all configured repositories
func HandleCheckPullRequests(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	args := request.Params.Arguments
	
	// Handle config file override
	if configFile := getStringParam(args, "config_file", ""); configFile != "" {
		if err := os.Setenv("CONFIG_FILE", configFile); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set CONFIG_FILE: %v", err)), nil
		}
	}
	
	outputFormat := getStringParam(args, "output_format", "")
	
	// Handle JSON output
	if outputFormat == "json" {
		return handleJSONOutput(exec, "check-prs-json")
	}
	
	// Handle provider filtering
	var makeTarget string
	if provider := getStringParam(args, "provider", ""); provider != "" {
		switch strings.ToLower(provider) {
		case "github":
			makeTarget = "check-github"
		case "gitlab":
			makeTarget = "check-gitlab"
		case "bitbucket":
			makeTarget = "check-bitbucket"
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid provider: %s. Must be github, gitlab, or bitbucket", provider)), nil
		}
	} else {
		makeTarget = "check-prs"
	}
	
	result := exec.ExecuteMake(makeTarget)
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	if provider := getStringParam(args, "provider", ""); provider != "" {
		response["provider"] = provider
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleMergePullRequests merges ready pull requests across configured repositories
func HandleMergePullRequests(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	args := request.Params.Arguments
	
	// Handle config file override
	if configFile := getStringParam(args, "config_file", ""); configFile != "" {
		if err := os.Setenv("CONFIG_FILE", configFile); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set CONFIG_FILE: %v", err)), nil
		}
	}
	
	mode := getStringParam(args, "mode", "")
	
	var makeTarget string
	var response map[string]interface{}
	
	switch mode {
	case "dry-run":
		if err := os.Setenv("DRY_RUN", "true"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set DRY_RUN: %v", err)), nil
		}
		makeTarget = "dry-run"
		result := exec.ExecuteMake(makeTarget)
		
		response = map[string]interface{}{
			"success": result.Success,
			"output":  result.Output,
			"dry_run": true,
		}
		if result.Error != "" {
			response["error"] = result.Error
		}
	case "force":
		if err := os.Setenv("FORCE", "true"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set FORCE: %v", err)), nil
		}
		makeTarget = "force-merge"
		result := exec.ExecuteMake(makeTarget)
		
		response = map[string]interface{}{
			"success": result.Success,
			"output":  result.Output,
			"forced":  true,
		}
		if result.Error != "" {
			response["error"] = result.Error
		}
	default:
		makeTarget = "merge-prs"
		result := exec.ExecuteMake(makeTarget)
		
		response = map[string]interface{}{
			"success": result.Success,
			"output":  result.Output,
		}
		if result.Error != "" {
			response["error"] = result.Error
		}
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleWatchRepositories continuously monitors PR status (not supported in MCP)
func HandleWatchRepositories(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	interval := getFloatParam(args, "interval", 30)
	
	response := map[string]interface{}{
		"success": false,
		"error":   "Continuous watching is not supported in MCP request/response model",
		"suggestion": "Use check_pull_requests repeatedly or run 'make watch' in terminal",
		"interval": interval,
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleGetRepositoryStats gets statistics about configured repositories
func HandleGetRepositoryStats(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	configFile := getStringParam(args, "config_file", "config.yaml")
	
	configMgr := config.NewManager(configFile)
	if err := configMgr.Load(); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		return mcp.NewToolResultText(formatResponse(response)), err
	}
	
	stats := configMgr.GetRepositoryStats()
	
	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
		"config_file": configMgr.GetConfigPath(),
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleTestNotifications tests Slack and email notification configuration
func HandleTestNotifications(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	result := exec.ExecuteMake("test-notifications")
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleLintScripts lints shell scripts using shellcheck
func HandleLintScripts(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	result := exec.ExecuteMake("lint")
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleCheckDependencies checks if required dependencies are installed
func HandleCheckDependencies(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	result := exec.CheckDependencies()
	
	// Also check environment variables
	missingEnvVars := exec.ValidateEnvironment()
	
	response := map[string]interface{}{
		"success":        result.Success,
		"output":         result.Output,
		"missing_env_vars": missingEnvVars,
		"env_vars_ok":    len(missingEnvVars) == 0,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// HandleInstallDependencies installs required dependencies automatically
func HandleInstallDependencies(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exec := executor.NewExecutor("..")
	args := request.Params.Arguments
	
	var makeTarget string
	if platform := getStringParam(args, "platform", ""); platform != "" {
		switch strings.ToLower(platform) {
		case "macos":
			makeTarget = "install-macos"
		case "linux":
			makeTarget = "install-linux"
		default:
			return mcp.NewToolResultError(fmt.Sprintf("unsupported platform: %s. Must be 'macos' or 'linux'", platform)), nil
		}
	} else {
		makeTarget = "install"
	}
	
	result := exec.ExecuteMake(makeTarget)
	
	response := map[string]interface{}{
		"success": result.Success,
		"output":  result.Output,
	}
	if result.Error != "" {
		response["error"] = result.Error
	}
	if platform := getStringParam(args, "platform", ""); platform != "" {
		response["platform"] = platform
	}
	
	return mcp.NewToolResultText(formatResponse(response)), nil
}

// Helper function to format responses as JSON
func formatResponse(response map[string]interface{}) string {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting response: %v", err)
	}
	return string(data)
}