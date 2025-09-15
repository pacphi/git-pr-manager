package providers

import (
	"fmt"
	"os"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/bitbucket"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/github"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/gitlab"
)

// Factory creates provider instances based on configuration
type Factory struct {
	config *config.Config
}

// NewFactory creates a new provider factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateProviders creates all configured providers
func (f *Factory) CreateProviders() (map[string]common.Provider, error) {
	providers := make(map[string]common.Provider)

	// Create GitHub provider if configured
	if f.config.Auth.GitHub.Token != "" {
		provider, err := f.createGitHubProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub provider: %w", err)
		}
		providers["github"] = provider
	}

	// Create GitLab provider if configured
	if f.config.Auth.GitLab.Token != "" {
		provider, err := f.createGitLabProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab provider: %w", err)
		}
		providers["gitlab"] = provider
	}

	// Create Bitbucket provider if configured
	if f.config.Auth.Bitbucket.Username != "" && f.config.Auth.Bitbucket.AppPassword != "" {
		provider, err := f.createBitbucketProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to create Bitbucket provider: %w", err)
		}
		providers["bitbucket"] = provider
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	return providers, nil
}

// createGitHubProvider creates a GitHub provider instance
func (f *Factory) createGitHubProvider() (common.Provider, error) {
	token := f.resolveEnvVar(f.config.Auth.GitHub.Token)
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	config := github.Config{
		Token:     token,
		BaseURL:   "", // Use default GitHub.com
		RateLimit: f.config.Behavior.RateLimit.RequestsPerSecond,
		RateBurst: f.config.Behavior.RateLimit.Burst,
	}

	return github.NewProvider(config)
}

// createGitLabProvider creates a GitLab provider instance
func (f *Factory) createGitLabProvider() (common.Provider, error) {
	token := f.resolveEnvVar(f.config.Auth.GitLab.Token)
	if token == "" {
		return nil, fmt.Errorf("GitLab token is required")
	}

	config := gitlab.Config{
		Token:     token,
		BaseURL:   f.resolveEnvVar(f.config.Auth.GitLab.URL),
		RateLimit: f.config.Behavior.RateLimit.RequestsPerSecond,
		RateBurst: f.config.Behavior.RateLimit.Burst,
	}

	return gitlab.NewProvider(config)
}

// createBitbucketProvider creates a Bitbucket provider instance
func (f *Factory) createBitbucketProvider() (common.Provider, error) {
	username := f.resolveEnvVar(f.config.Auth.Bitbucket.Username)
	appPassword := f.resolveEnvVar(f.config.Auth.Bitbucket.AppPassword)

	if username == "" || appPassword == "" {
		return nil, fmt.Errorf("bitbucket username and app password are required")
	}

	config := bitbucket.Config{
		Username:    username,
		AppPassword: appPassword,
		Workspace:   f.resolveEnvVar(f.config.Auth.Bitbucket.Workspace),
		RateLimit:   f.config.Behavior.RateLimit.RequestsPerSecond,
		RateBurst:   f.config.Behavior.RateLimit.Burst,
	}

	return bitbucket.NewProvider(config)
}

// CreateProvider creates a single provider by type with given repositories
func CreateProvider(providerType string, repositories []common.Repository) (common.Provider, error) {
	// Create a minimal config for the provider type
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 10, // Default rate limit
				Burst:             20, // Default burst
			},
		},
	}

	// Set authentication from environment variables
	switch providerType {
	case "github":
		cfg.Auth.GitHub.Token = os.Getenv("GITHUB_TOKEN")
		if cfg.Auth.GitHub.Token == "" {
			cfg.Auth.GitHub.Token = os.Getenv("GH_TOKEN")
		}
		if cfg.Auth.GitHub.Token == "" {
			return nil, fmt.Errorf("GitHub token not found in environment (GITHUB_TOKEN or GH_TOKEN)")
		}
	case "gitlab":
		cfg.Auth.GitLab.Token = os.Getenv("GITLAB_TOKEN")
		cfg.Auth.GitLab.URL = os.Getenv("GITLAB_URL")
		if cfg.Auth.GitLab.Token == "" {
			return nil, fmt.Errorf("GitLab token not found in environment (GITLAB_TOKEN)")
		}
	case "bitbucket":
		cfg.Auth.Bitbucket.Username = os.Getenv("BITBUCKET_USERNAME")
		cfg.Auth.Bitbucket.AppPassword = os.Getenv("BITBUCKET_APP_PASSWORD")
		cfg.Auth.Bitbucket.Workspace = os.Getenv("BITBUCKET_WORKSPACE")
		if cfg.Auth.Bitbucket.Username == "" || cfg.Auth.Bitbucket.AppPassword == "" {
			return nil, fmt.Errorf("bitbucket credentials not found in environment (BITBUCKET_USERNAME, BITBUCKET_APP_PASSWORD)")
		}
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	factory := NewFactory(cfg)

	switch providerType {
	case "github":
		return factory.createGitHubProvider()
	case "gitlab":
		return factory.createGitLabProvider()
	case "bitbucket":
		return factory.createBitbucketProvider()
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// resolveEnvVar resolves environment variable references in config values
func (f *Factory) resolveEnvVar(value string) string {
	if value == "" {
		return ""
	}

	// If value starts with $, treat it as an environment variable
	if len(value) > 1 && value[0] == '$' {
		envVar := value[1:]
		return os.Getenv(envVar)
	}

	return value
}
