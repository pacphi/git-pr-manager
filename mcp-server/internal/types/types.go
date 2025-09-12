// Package types defines the data structures used throughout the MCP server.
package types

// Config represents the YAML configuration structure
type Config struct {
	Config       GlobalConfig               `yaml:"config"`
	Repositories map[string][]Repository    `yaml:"repositories"`
	Auth         map[string]interface{}     `yaml:"auth"`
	Notifications NotificationConfig        `yaml:"notifications"`
}

// GlobalConfig holds global configuration settings
type GlobalConfig struct {
	DefaultMergeStrategy string    `yaml:"default_merge_strategy"`
	AutoMerge            AutoMerge `yaml:"auto_merge"`
	PRFilters            PRFilters `yaml:"pr_filters"`
}

// AutoMerge configuration
type AutoMerge struct {
	Enabled         bool `yaml:"enabled"`
	WaitForChecks   bool `yaml:"wait_for_checks"`
	RequireApproval bool `yaml:"require_approval"`
}

// PRFilters configuration
type PRFilters struct {
	AllowedActors []string `yaml:"allowed_actors"`
	SkipLabels    []string `yaml:"skip_labels"`
}

// Repository represents a single repository configuration
type Repository struct {
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	Provider      string `yaml:"provider"`
	AuthType      string `yaml:"auth_type"`
	MergeStrategy string `yaml:"merge_strategy"`
	AutoMerge     bool   `yaml:"auto_merge"`
}

// NotificationConfig holds notification settings
type NotificationConfig struct {
	Slack SlackConfig `yaml:"slack"`
	Email EmailConfig `yaml:"email"`
}

// SlackConfig holds Slack notification settings
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Enabled    bool   `yaml:"enabled"`
}

// EmailConfig holds email notification settings
type EmailConfig struct {
	SMTPServer string `yaml:"smtp_server"`
	SMTPPort   int    `yaml:"smtp_port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Recipient  string `yaml:"recipient"`
	Enabled    bool   `yaml:"enabled"`
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
}

// RepositoryStats represents statistics for repositories
type RepositoryStats struct {
	GitHub    int `json:"github"`
	GitLab    int `json:"gitlab"`
	Bitbucket int `json:"bitbucket"`
	Total     int `json:"total"`
}