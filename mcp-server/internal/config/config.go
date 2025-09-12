// Package config provides configuration loading and validation functionality for the MCP server.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"github.com/cphillipson/multi-gitter-pr-automation/mcp-server/internal/types"
)

// Manager handles configuration loading and validation
type Manager struct {
	configPath string
	config     *types.Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	if configPath == "" {
		configPath = "config.yaml"
	}
	return &Manager{
		configPath: configPath,
	}
}

// Load reads and parses the configuration file
func (m *Manager) Load() error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", m.configPath)
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	m.config = &config
	return nil
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *types.Config {
	return m.config
}

// Validate checks the configuration for common issues
func (m *Manager) Validate() error {
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if we have at least one repository configured
	totalRepos := 0
	for provider, repos := range m.config.Repositories {
		totalRepos += len(repos)

		// Validate each repository
		for _, repo := range repos {
			if repo.Name == "" {
				return fmt.Errorf("repository name is required for provider %s", provider)
			}
			if repo.URL == "" {
				return fmt.Errorf("repository URL is required for %s", repo.Name)
			}
		}
	}

	if totalRepos == 0 {
		return fmt.Errorf("no repositories configured")
	}

	return nil
}

// GetRepositoryStats returns statistics about configured repositories
func (m *Manager) GetRepositoryStats() *types.RepositoryStats {
	if m.config == nil {
		return &types.RepositoryStats{}
	}

	stats := &types.RepositoryStats{}

	if githubRepos, ok := m.config.Repositories["github"]; ok {
		stats.GitHub = len(githubRepos)
	}
	if gitlabRepos, ok := m.config.Repositories["gitlab"]; ok {
		stats.GitLab = len(gitlabRepos)
	}
	if bitbucketRepos, ok := m.config.Repositories["bitbucket"]; ok {
		stats.Bitbucket = len(bitbucketRepos)
	}

	stats.Total = stats.GitHub + stats.GitLab + stats.Bitbucket

	return stats
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	absPath, _ := filepath.Abs(m.configPath)
	return absPath
}