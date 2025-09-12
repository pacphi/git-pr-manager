// Package server provides the MCP server implementation with tools and resources for PR automation.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/cphillipson/multi-gitter-pr-automation/mcp-server/internal/config"
)

// CreateResources creates all MCP resources for the PR automation server
func CreateResources() []*mcp.Resource {
	return []*mcp.Resource{
		createConfigResource(),
		createRepositoryStatsResource(),
		createMakefileTargetsResource(),
		createEnvironmentStatusResource(),
	}
}

// Resource creation functions

func createConfigResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "config://current",
		Name:        "Current Configuration",
		Description: "The current YAML configuration file with repository settings",
		MIMEType:    "text/yaml",
	}
}

func createRepositoryStatsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "stats://repositories",
		Name:        "Repository Statistics",
		Description: "Statistics about configured repositories by provider",
		MIMEType:    "application/json",
	}
}

func createMakefileTargetsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "makefile://targets",
		Name:        "Available Make Targets",
		Description: "List of all available Makefile targets and their descriptions",
		MIMEType:    "application/json",
	}
}

func createEnvironmentStatusResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "env://status",
		Name:        "Environment Status",
		Description: "Status of required environment variables and dependencies",
		MIMEType:    "application/json",
	}
}

// Resource handlers with correct signatures

// HandleConfigResource handles requests for the current YAML configuration
func HandleConfigResource(_ context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
	// Parse URI to get config file path
	configFile := "config.yaml"
	
	// Try to read the config file
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Try config.sample if config.yaml doesn't exist
		if _, err := os.Stat("config.sample"); err == nil {
			configFile = "config.sample"
		} else {
			return nil, fmt.Errorf("no configuration file found (config.yaml or config.sample)")
		}
	}
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}
	
	// Return as TextResourceContents
	content := map[string]interface{}{
		"uri":      request.Params.URI,
		"mimeType": "text/yaml",
		"text":     string(data),
	}
	return []interface{}{content}, nil
}

// HandleRepositoryStatsResource handles requests for repository statistics by provider
func HandleRepositoryStatsResource(_ context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
	configFile := "config.yaml"
	
	configMgr := config.NewManager(configFile)
	if err := configMgr.Load(); err != nil {
		// Try to get stats from config.sample if main config doesn't exist
		configMgr = config.NewManager("config.sample")
		if err := configMgr.Load(); err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	}
	
	stats := configMgr.GetRepositoryStats()
	
	// Add additional metadata
	result := map[string]interface{}{
		"stats":       stats,
		"config_file": configMgr.GetConfigPath(),
		"timestamp":   fmt.Sprintf("%d", os.Getpid()), // Simple timestamp
	}
	
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats: %w", err)
	}
	
	// Return as TextResourceContents
	content := map[string]interface{}{
		"uri":      request.Params.URI,
		"mimeType": "application/json",
		"text":     string(data),
	}
	return []interface{}{content}, nil
}

// HandleMakefileTargetsResource handles requests for available Makefile targets
func HandleMakefileTargetsResource(_ context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
	// Read and parse Makefile to extract targets and descriptions
	makefilePath := filepath.Clean(filepath.Join("..", "Makefile"))
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Makefile: %w", err)
	}
	
	targets := parseMakefileTargets(string(data))
	
	result := map[string]interface{}{
		"targets":      targets,
		"total_count":  len(targets),
		"makefile":     makefilePath,
	}
	
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal targets: %w", err)
	}
	
	// Return as TextResourceContents
	content := map[string]interface{}{
		"uri":      request.Params.URI,
		"mimeType": "application/json",
		"text":     string(jsonData),
	}
	return []interface{}{content}, nil
}

// HandleEnvironmentStatusResource handles requests for environment variable and dependency status
func HandleEnvironmentStatusResource(_ context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
	// Check environment variables
	requiredEnvVars := []string{
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"BITBUCKET_USERNAME", 
		"BITBUCKET_APP_PASSWORD",
	}
	
	optionalEnvVars := []string{
		"BITBUCKET_WORKSPACE",
		"SLACK_WEBHOOK_URL",
		"EMAIL_USERNAME",
		"EMAIL_PASSWORD",
		"EMAIL_RECIPIENT",
	}
	
	envStatus := make(map[string]map[string]interface{})
	
	// Check required env vars
	envStatus["required"] = make(map[string]interface{})
	allRequiredSet := true
	for _, envVar := range requiredEnvVars {
		value := os.Getenv(envVar)
		isSet := value != ""
		if !isSet {
			allRequiredSet = false
		}
		envStatus["required"][envVar] = map[string]interface{}{
			"set":    isSet,
			"masked": maskValue(value),
		}
	}
	
	// Check optional env vars
	envStatus["optional"] = make(map[string]interface{})
	for _, envVar := range optionalEnvVars {
		value := os.Getenv(envVar)
		envStatus["optional"][envVar] = map[string]interface{}{
			"set":    value != "",
			"masked": maskValue(value),
		}
	}
	
	// Check dependencies
	dependencies := []string{"yq", "jq", "curl", "gh"}
	depStatus := make(map[string]bool)
	allDepsInstalled := true
	
	for _, dep := range dependencies {
		// Simple check - this is a basic implementation
		_, err := os.Stat(fmt.Sprintf("/usr/local/bin/%s", dep))
		if err != nil {
			_, err = os.Stat(fmt.Sprintf("/usr/bin/%s", dep))
		}
		installed := err == nil
		depStatus[dep] = installed
		if !installed && dep != "gh" { // gh is optional
			allDepsInstalled = false
		}
	}
	
	result := map[string]interface{}{
		"environment_variables": envStatus,
		"dependencies":         depStatus,
		"status": map[string]interface{}{
			"all_required_env_vars_set": allRequiredSet,
			"all_dependencies_installed": allDepsInstalled,
			"ready_for_operation":       allRequiredSet && allDepsInstalled,
		},
	}
	
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal environment status: %w", err)
	}
	
	// Return as TextResourceContents
	content := map[string]interface{}{
		"uri":      request.Params.URI,
		"mimeType": "application/json",
		"text":     string(jsonData),
	}
	return []interface{}{content}, nil
}

// Helper functions

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + strings.Repeat("*", len(value)-4)
}

func parseMakefileTargets(content string) map[string]string {
	targets := make(map[string]string)
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		// Look for lines with "## " comments (help descriptions)
		if strings.Contains(line, ": ##") {
			parts := strings.SplitN(line, ": ##", 2)
			if len(parts) == 2 {
				target := strings.TrimSpace(parts[0])
				description := strings.TrimSpace(parts[1])
				targets[target] = description
			}
		}
	}
	
	return targets
}