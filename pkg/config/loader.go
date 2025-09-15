package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ConfigValidationError represents a user-facing configuration validation error
type ConfigValidationError struct {
	Message string
}

func (e *ConfigValidationError) Error() string {
	return e.Message
}

// Loader handles configuration loading and validation
type Loader struct {
	validator *validator.Validate
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		validator: validator.New(),
	}
}

// Load loads configuration from file and environment variables
func (l *Loader) Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = l.findConfigFile()
	}

	if configPath == "" {
		return nil, fmt.Errorf("no configuration file found")
	}

	// Only read the config file if viper hasn't already loaded it
	if viper.ConfigFileUsed() != configPath {
		// Load config file
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}
	}

	// Allow environment variables to override config
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Fix viper.Unmarshal issues by manually extracting problematic fields
	config.PRFilters.AllowedActors = viper.GetStringSlice("pr_filters.allowed_actors")
	config.PRFilters.SkipLabels = viper.GetStringSlice("pr_filters.skip_labels")
	config.PRFilters.MaxAge = viper.GetString("pr_filters.max_age")

	// Process environment variables for auth
	if err := l.processEnvVars(&config); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	// Set defaults
	l.setDefaults(&config)

	// Validate configuration with helpful error messages
	if err := l.validateWithHelpfulErrors(&config); err != nil {
		return nil, err
	}

	// Additional business logic validation
	if err := l.validateBusinessRules(&config); err != nil {
		return nil, fmt.Errorf("config business rules validation failed: %w", err)
	}

	return &config, nil
}

// Save saves configuration to file
func (l *Loader) Save(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// BackupConfig creates a backup of the current config file with timestamp
func (l *Loader) BackupConfig(configPath string) (string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Use timestamp for consistent backup naming
	backupPath := fmt.Sprintf("%s.backup.%d", configPath, time.Now().Unix())

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// findConfigFile searches for config file in common locations
func (l *Loader) findConfigFile() string {
	locations := []string{
		"config.yaml",
		"config.yml",
		"~/.config/git-pr/config.yaml",
		"~/.git-pr.yaml",
		"/etc/git-pr/config.yaml",
	}

	for _, location := range locations {
		if strings.HasPrefix(location, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			location = filepath.Join(home, strings.TrimPrefix(location, "~/"))
		}

		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	return ""
}

// processEnvVars processes environment variables for auth sections
func (l *Loader) processEnvVars(config *Config) error {
	if err := envconfig.Process("github", &config.Auth.GitHub); err != nil {
		return fmt.Errorf("failed to process GitHub env vars: %w", err)
	}

	if err := envconfig.Process("gitlab", &config.Auth.GitLab); err != nil {
		return fmt.Errorf("failed to process GitLab env vars: %w", err)
	}

	if err := envconfig.Process("bitbucket", &config.Auth.Bitbucket); err != nil {
		return fmt.Errorf("failed to process Bitbucket env vars: %w", err)
	}

	if err := envconfig.Process("slack", &config.Notifications.Slack); err != nil {
		return fmt.Errorf("failed to process Slack env vars: %w", err)
	}

	if err := envconfig.Process("email", &config.Notifications.Email); err != nil {
		return fmt.Errorf("failed to process Email env vars: %w", err)
	}

	return nil
}

// LoadConfigFromPath loads configuration from a specific file path
func LoadConfigFromPath(configPath string) (*Config, error) {
	loader := NewLoader()
	return loader.Load(configPath)
}

// setDefaults sets default values for configuration using environment utilities
func (l *Loader) setDefaults(config *Config) {
	// Set default merge strategy for repositories that don't have one
	for provider, repos := range config.Repositories {
		for idx := range repos {
			if config.Repositories[provider][idx].MergeStrategy == "" {
				config.Repositories[provider][idx].MergeStrategy = MergeStrategySquash
			}
			if config.Repositories[provider][idx].Branch == "" {
				config.Repositories[provider][idx].Branch = getEnvOrDefault("DEFAULT_BRANCH", "main")
			}
		}
	}

	// Set default behavior values using environment variables
	if config.Behavior.RateLimit.RequestsPerSecond == 0 {
		config.Behavior.RateLimit.RequestsPerSecond = float64(getEnvIntOrDefault("RATE_LIMIT_RPS", 5))
	}
	if config.Behavior.RateLimit.Burst == 0 {
		config.Behavior.RateLimit.Burst = getEnvIntOrDefault("RATE_LIMIT_BURST", 10)
	}
	if config.Behavior.RateLimit.Timeout == 0 {
		config.Behavior.RateLimit.Timeout = getEnvDurationOrDefault("RATE_LIMIT_TIMEOUT", 30*time.Second)
	}

	if config.Behavior.Retry.MaxAttempts == 0 {
		config.Behavior.Retry.MaxAttempts = getEnvIntOrDefault("RETRY_MAX_ATTEMPTS", 3)
	}
	if config.Behavior.Retry.Backoff == 0 {
		config.Behavior.Retry.Backoff = getEnvDurationOrDefault("RETRY_BACKOFF", 1*time.Second)
	}
	if config.Behavior.Retry.MaxBackoff == 0 {
		config.Behavior.Retry.MaxBackoff = getEnvDurationOrDefault("RETRY_MAX_BACKOFF", 30*time.Second)
	}

	if config.Behavior.Concurrency == 0 {
		config.Behavior.Concurrency = getEnvIntOrDefault("CONCURRENCY", 5)
	}

	if config.Behavior.WatchInterval == "" {
		config.Behavior.WatchInterval = getEnvOrDefault("WATCH_INTERVAL", "30s")
	}

	// Set default GitLab URL if not provided
	if config.Auth.GitLab.Token != "" && config.Auth.GitLab.URL == "" {
		config.Auth.GitLab.URL = getEnvOrDefault("GITLAB_URL", "https://gitlab.com")
	}

	// Set default SMTP port if not provided
	if config.Notifications.Email.SMTPHost != "" && config.Notifications.Email.SMTPPort == 0 {
		config.Notifications.Email.SMTPPort = getEnvIntOrDefault("SMTP_PORT", 587)
	}
}

// validateBusinessRules validates business-specific configuration rules
func (l *Loader) validateBusinessRules(config *Config) error {
	// Validate merge strategies
	for provider, repos := range config.Repositories {
		for _, repo := range repos {
			if !repo.MergeStrategy.IsValid() {
				return fmt.Errorf("invalid merge strategy '%s' for repository %s in provider %s",
					repo.MergeStrategy, repo.Name, provider)
			}
		}
	}

	// Validate that at least one provider has authentication configured
	hasAuth := config.Auth.GitHub.Token != ""
	if config.Auth.GitLab.Token != "" {
		hasAuth = true
	}
	if config.Auth.Bitbucket.Username != "" && config.Auth.Bitbucket.AppPassword != "" {
		hasAuth = true
	}

	if !hasAuth {
		return fmt.Errorf("at least one provider must have authentication configured")
	}

	// Validate that repositories exist only for providers with authentication
	for provider, repos := range config.Repositories {
		if len(repos) == 0 {
			continue // Skip providers with no repositories
		}

		switch Provider(provider) {
		case ProviderGitHub:
			if isEmptyOrTemplate(config.Auth.GitHub.Token, "${GITHUB_TOKEN}") {
				return fmt.Errorf("GitHub repositories configured but no GitHub token provided")
			}
		case ProviderGitLab:
			if isEmptyOrTemplate(config.Auth.GitLab.Token, "${GITLAB_TOKEN}") {
				return fmt.Errorf("GitLab repositories configured but no GitLab token provided")
			}
		case ProviderBitbucket:
			username := config.Auth.Bitbucket.Username
			password := config.Auth.Bitbucket.AppPassword
			if isEmptyOrTemplate(username, "${BITBUCKET_USERNAME}") || isEmptyOrTemplate(password, "${BITBUCKET_APP_PASSWORD}") {
				return fmt.Errorf("bitbucket repositories configured but incomplete bitbucket authentication")
			}
		default:
			return fmt.Errorf("unsupported provider: %s", provider)
		}
	}

	// Validate notification configuration
	if config.Notifications.Slack.Enabled && config.Notifications.Slack.WebhookURL == "" {
		return fmt.Errorf("slack notifications enabled but no webhook URL provided")
	}

	if config.Notifications.Email.Enabled {
		if config.Notifications.Email.SMTPHost == "" {
			return fmt.Errorf("email notifications enabled but no SMTP host provided")
		}
		if config.Notifications.Email.From == "" {
			return fmt.Errorf("email notifications enabled but no 'from' address provided")
		}
		if len(config.Notifications.Email.To) == 0 {
			return fmt.Errorf("email notifications enabled but no 'to' addresses provided")
		}
	}

	return nil
}

// isEmptyOrTemplate checks if a value is empty or still contains an unexpanded environment variable template
func isEmptyOrTemplate(value, template string) bool {
	return value == "" || value == template
}

// validateWithHelpfulErrors performs validation with user-friendly error messages
func (l *Loader) validateWithHelpfulErrors(config *Config) error {
	// Check for missing required fields with helpful messages
	var missingEnvVars []string
	var missingConfigFields []string

	// Check PR filters
	if len(config.PRFilters.AllowedActors) == 0 {
		missingConfigFields = append(missingConfigFields, "pr_filters.allowed_actors (must specify at least one trusted actor like 'dependabot[bot]')")
	}

	// Check repositories configuration
	if len(config.Repositories) == 0 {
		missingConfigFields = append(missingConfigFields, "repositories (must configure at least one repository)")
	}

	// Check authentication based on configured providers (only if they have repositories)
	for provider, repos := range config.Repositories {
		if len(repos) == 0 {
			continue // Skip providers with no repositories
		}

		switch Provider(provider) {
		case ProviderGitHub:
			if isEmptyOrTemplate(config.Auth.GitHub.Token, "${GITHUB_TOKEN}") {
				missingEnvVars = append(missingEnvVars, "GITHUB_TOKEN (required for GitHub repositories)")
			}
		case ProviderGitLab:
			if isEmptyOrTemplate(config.Auth.GitLab.Token, "${GITLAB_TOKEN}") {
				missingEnvVars = append(missingEnvVars, "GITLAB_TOKEN (required for GitLab repositories)")
			}
		case ProviderBitbucket:
			username := config.Auth.Bitbucket.Username
			password := config.Auth.Bitbucket.AppPassword
			if isEmptyOrTemplate(username, "${BITBUCKET_USERNAME}") || isEmptyOrTemplate(password, "${BITBUCKET_APP_PASSWORD}") {
				if isEmptyOrTemplate(username, "${BITBUCKET_USERNAME}") {
					missingEnvVars = append(missingEnvVars, "BITBUCKET_USERNAME (required for Bitbucket repositories)")
				}
				if isEmptyOrTemplate(password, "${BITBUCKET_APP_PASSWORD}") {
					missingEnvVars = append(missingEnvVars, "BITBUCKET_APP_PASSWORD (required for Bitbucket repositories)")
				}
			}
		}
	}

	// Build helpful error message
	if len(missingEnvVars) > 0 || len(missingConfigFields) > 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString("Configuration validation failed:\n")

		if len(missingConfigFields) > 0 {
			errorMsg.WriteString("\nMissing or invalid configuration fields:\n")
			for _, field := range missingConfigFields {
				errorMsg.WriteString(fmt.Sprintf("  - %s\n", field))
			}
		}

		if len(missingEnvVars) > 0 {
			errorMsg.WriteString("\nMissing environment variables:\n")
			for _, envVar := range missingEnvVars {
				errorMsg.WriteString(fmt.Sprintf("  - %s\n", envVar))
			}
			errorMsg.WriteString("\nTo fix this, set the required environment variables:\n")
			for _, envVar := range missingEnvVars {
				envName := strings.Split(envVar, " ")[0]
				errorMsg.WriteString(fmt.Sprintf("  export %s=\"your-token-here\"\n", envName))
			}
		}

		return &ConfigValidationError{Message: errorMsg.String()}
	}

	// Run standard struct validation for other fields
	if err := l.validator.Struct(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}
