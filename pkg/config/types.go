package config

import (
	"time"
)

// Config represents the complete configuration for the application
type Config struct {
	PRFilters     PRFilters               `yaml:"pr_filters" validate:"required"`
	Repositories  map[string][]Repository `yaml:"repositories" validate:"required"`
	Auth          Auth                    `yaml:"auth" validate:"required"`
	Notifications Notifications           `yaml:"notifications"`
	Behavior      Behavior                `yaml:"behavior"`
}

// PRFilters defines filtering criteria for pull requests
type PRFilters struct {
	AllowedActors []string `yaml:"allowed_actors" validate:"required,min=1"`
	SkipLabels    []string `yaml:"skip_labels"`
	MaxAge        string   `yaml:"max_age,omitempty"`
}

// Repository represents a single repository configuration
type Repository struct {
	Name           string        `yaml:"name" validate:"required"`
	AutoMerge      bool          `yaml:"auto_merge"`
	MergeStrategy  MergeStrategy `yaml:"merge_strategy,omitempty"`
	SkipLabels     []string      `yaml:"skip_labels,omitempty"`
	Branch         string        `yaml:"branch,omitempty"`
	RequireChecks  bool          `yaml:"require_checks"`
	MinApprovals   int           `yaml:"min_approvals,omitempty"`
	DeleteBranches bool          `yaml:"delete_branches"`
}

// MergeStrategy defines how PRs should be merged
type MergeStrategy string

const (
	MergeStrategyMerge  MergeStrategy = "merge"
	MergeStrategySquash MergeStrategy = "squash"
	MergeStrategyRebase MergeStrategy = "rebase"
)

// Auth contains authentication configuration for all providers
type Auth struct {
	GitHub    GitHubAuth    `yaml:"github"`
	GitLab    GitLabAuth    `yaml:"gitlab"`
	Bitbucket BitbucketAuth `yaml:"bitbucket"`
}

// GitHubAuth contains GitHub-specific authentication
type GitHubAuth struct {
	Token string `yaml:"token" envconfig:"GITHUB_TOKEN" validate:"required"`
}

// GitLabAuth contains GitLab-specific authentication
type GitLabAuth struct {
	Token string `yaml:"token" envconfig:"GITLAB_TOKEN"`
	URL   string `yaml:"url" envconfig:"GITLAB_URL"`
}

// BitbucketAuth contains Bitbucket-specific authentication
type BitbucketAuth struct {
	Username    string `yaml:"username" envconfig:"BITBUCKET_USERNAME"`
	AppPassword string `yaml:"app_password" envconfig:"BITBUCKET_APP_PASSWORD"`
	Workspace   string `yaml:"workspace" envconfig:"BITBUCKET_WORKSPACE"`
}

// Notifications contains notification configuration
type Notifications struct {
	Slack SlackConfig `yaml:"slack"`
	Email EmailConfig `yaml:"email"`
}

// SlackConfig contains Slack notification settings
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url" envconfig:"SLACK_WEBHOOK_URL"`
	Channel    string `yaml:"channel"`
	Enabled    bool   `yaml:"enabled"`
}

// EmailConfig contains email notification settings
type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host" envconfig:"SMTP_HOST"`
	SMTPPort     int      `yaml:"smtp_port" envconfig:"SMTP_PORT"`
	SMTPUsername string   `yaml:"smtp_username" envconfig:"SMTP_USERNAME"`
	SMTPPassword string   `yaml:"smtp_password" envconfig:"SMTP_PASSWORD"`
	From         string   `yaml:"from" envconfig:"EMAIL_FROM"`
	To           []string `yaml:"to"`
	Enabled      bool     `yaml:"enabled"`
}

// Behavior contains behavioral configuration
type Behavior struct {
	RateLimit       RateLimit     `yaml:"rate_limit"`
	Retry           Retry         `yaml:"retry"`
	Concurrency     int           `yaml:"concurrency"`
	DryRun          bool          `yaml:"dry_run"`
	WatchInterval   string        `yaml:"watch_interval"`
	RequireApproval bool          `yaml:"require_approval"`
	MergeDelay      time.Duration `yaml:"merge_delay"`
	DeleteBranches  bool          `yaml:"delete_branches"`
}

// RateLimit contains rate limiting configuration
type RateLimit struct {
	RequestsPerSecond float64       `yaml:"requests_per_second"`
	Burst             int           `yaml:"burst"`
	Timeout           time.Duration `yaml:"timeout"`
}

// Retry contains retry configuration
type Retry struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Backoff     time.Duration `yaml:"backoff"`
	MaxBackoff  time.Duration `yaml:"max_backoff"`
}

// Provider represents a supported Git provider
type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderGitLab    Provider = "gitlab"
	ProviderBitbucket Provider = "bitbucket"
)

// IsValid checks if the merge strategy is valid
func (ms MergeStrategy) IsValid() bool {
	switch ms {
	case MergeStrategyMerge, MergeStrategySquash, MergeStrategyRebase:
		return true
	default:
		return false
	}
}

// IsValid checks if the provider is valid
func (p Provider) IsValid() bool {
	switch p {
	case ProviderGitHub, ProviderGitLab, ProviderBitbucket:
		return true
	default:
		return false
	}
}
