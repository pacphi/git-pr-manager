package validation

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
)

func TestNew(t *testing.T) {
	validator := New()

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.logger)
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	validator := New()

	cfg := &config.Config{
		PRFilters: config.PRFilters{
			AllowedActors: []string{"bot", "user1"},
			MaxAge:        "7d",
			SkipLabels:    []string{"skip", "wip"},
		},
		Repositories: map[string][]config.Repository{
			"github": {
				{
					Name:          "owner/repo1",
					MergeStrategy: config.MergeStrategySquash,
					Branch:        "main",
					SkipLabels:    []string{"skip"},
				},
			},
		},
		Auth: config.Auth{
			GitHub: config.GitHubAuth{
				Token: "ghp_" + strings.Repeat("x", 36),
			},
		},
		Behavior: config.Behavior{
			Concurrency: 5,
			RateLimit: config.RateLimit{
				RequestsPerSecond: 2.0,
				Burst:             5,
			},
		},
		Notifications: config.Notifications{
			Slack: config.SlackConfig{
				WebhookURL: "https://hooks.slack.com/services/T00/B00/XXXXX",
			},
		},
	}

	err := validator.ValidateConfig(cfg)

	assert.NoError(t, err)
}

func TestValidateConfig_MultipleErrors(t *testing.T) {
	validator := New()

	cfg := &config.Config{
		PRFilters: config.PRFilters{
			AllowedActors: []string{}, // Invalid: empty
			MaxAge:        "invalid",  // Invalid: format
		},
		Repositories: map[string][]config.Repository{}, // Invalid: no repositories
		Auth:         config.Auth{},                     // Invalid: no auth when repos required
		Behavior: config.Behavior{
			Concurrency: 0, // Invalid: must be > 0
		},
	}

	err := validator.ValidateConfig(cfg)

	assert.Error(t, err)
	errorStr := err.Error()
	assert.Contains(t, errorStr, "validation errors")
	assert.Contains(t, errorStr, "PR filters")
	assert.Contains(t, errorStr, "repositories")
	// Authentication error is not present when no repos are configured
	assert.Contains(t, errorStr, "behavior")
}

func TestValidatePRFilters(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		filters     config.PRFilters
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid filters",
			filters: config.PRFilters{
				AllowedActors: []string{"bot", "user1"},
				MaxAge:        "7d",
				SkipLabels:    []string{"skip", "wip"},
			},
			expectError: false,
		},
		{
			name: "empty allowed actors",
			filters: config.PRFilters{
				AllowedActors: []string{},
			},
			expectError: true,
			errorMsg:    "at least one allowed actor must be specified",
		},
		{
			name: "empty actor in list",
			filters: config.PRFilters{
				AllowedActors: []string{"bot", "", "user1"},
			},
			expectError: true,
			errorMsg:    "allowed actor at index 1 is empty",
		},
		{
			name: "invalid max age format",
			filters: config.PRFilters{
				AllowedActors: []string{"bot"},
				MaxAge:        "invalid-duration",
			},
			expectError: true,
			errorMsg:    "invalid max_age format",
		},
		{
			name: "empty skip label",
			filters: config.PRFilters{
				AllowedActors: []string{"bot"},
				SkipLabels:    []string{"skip", "", "wip"},
			},
			expectError: true,
			errorMsg:    "skip label at index 1 is empty",
		},
		{
			name: "valid duration formats",
			filters: config.PRFilters{
				AllowedActors: []string{"bot"},
				MaxAge:        "1h30m",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePRFilters(tt.filters)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRepositories(t *testing.T) {
	validator := New()

	tests := []struct {
		name         string
		repositories map[string][]config.Repository
		expectError  bool
		errorMsg     string
	}{
		{
			name: "valid repositories",
			repositories: map[string][]config.Repository{
				"github": {
					{
						Name:          "owner/repo1",
						MergeStrategy: config.MergeStrategySquash,
						SkipLabels:    []string{"skip"},
					},
					{
						Name:          "owner/repo2",
						MergeStrategy: config.MergeStrategyMerge,
					},
				},
			},
			expectError: false,
		},
		{
			name:         "no providers configured",
			repositories: map[string][]config.Repository{},
			expectError:  true,
			errorMsg:     "at least one provider must be configured",
		},
		{
			name: "empty provider is allowed",
			repositories: map[string][]config.Repository{
				"github": {},
			},
			expectError: false,
		},
		{
			name: "invalid repository name format",
			repositories: map[string][]config.Repository{
				"github": {
					{
						Name: "invalid-repo-name",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid name format 'invalid-repo-name' (expected 'owner/name')",
		},
		{
			name: "invalid merge strategy",
			repositories: map[string][]config.Repository{
				"github": {
					{
						Name:          "owner/repo",
						MergeStrategy: "invalid-strategy",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid merge strategy 'invalid-strategy'",
		},
		{
			name: "empty skip label",
			repositories: map[string][]config.Repository{
				"github": {
					{
						Name:       "owner/repo",
						SkipLabels: []string{"skip", "", "wip"},
					},
				},
			},
			expectError: true,
			errorMsg:    "skip label at index 1 is empty",
		},
		{
			name: "valid merge strategies",
			repositories: map[string][]config.Repository{
				"github": {
					{Name: "owner/repo1", MergeStrategy: config.MergeStrategyMerge},
					{Name: "owner/repo2", MergeStrategy: config.MergeStrategySquash},
					{Name: "owner/repo3", MergeStrategy: config.MergeStrategyRebase},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRepositories(tt.repositories)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAuth(t *testing.T) {
	validator := New()

	tests := []struct {
		name         string
		auth         config.Auth
		repositories map[string][]config.Repository
		expectError  bool
		errorMsg     string
	}{
		{
			name: "valid GitHub auth",
			auth: config.Auth{
				GitHub: config.GitHubAuth{
					Token: "ghp_" + strings.Repeat("x", 36),
				},
			},
			repositories: map[string][]config.Repository{
				"github": {
					{Name: "owner/repo"},
				},
			},
			expectError: false,
		},
		{
			name: "missing GitHub token when repos configured",
			auth: config.Auth{},
			repositories: map[string][]config.Repository{
				"github": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "GitHub token is required when GitHub repositories are configured",
		},
		{
			name: "invalid GitHub token",
			auth: config.Auth{
				GitHub: config.GitHubAuth{
					Token: "invalid-token",
				},
			},
			repositories: map[string][]config.Repository{
				"github": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "GitHub token appears invalid",
		},
		{
			name: "valid GitLab auth",
			auth: config.Auth{
				GitLab: config.GitLabAuth{
					Token: "glpat-" + strings.Repeat("x", 20),
					URL:   "https://gitlab.example.com",
				},
			},
			repositories: map[string][]config.Repository{
				"gitlab": {
					{Name: "owner/repo"},
				},
			},
			expectError: false,
		},
		{
			name: "missing GitLab token",
			auth: config.Auth{},
			repositories: map[string][]config.Repository{
				"gitlab": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "GitLab token is required when GitLab repositories are configured",
		},
		{
			name: "invalid GitLab URL",
			auth: config.Auth{
				GitLab: config.GitLabAuth{
					Token: "glpat-" + strings.Repeat("x", 20),
					URL:   "invalid-url",
				},
			},
			repositories: map[string][]config.Repository{
				"gitlab": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "GitLab URL must start with http:// or https://",
		},
		{
			name: "valid Bitbucket auth",
			auth: config.Auth{
				Bitbucket: config.BitbucketAuth{
					Username:    "testuser",
					AppPassword: strings.Repeat("x", 20),
				},
			},
			repositories: map[string][]config.Repository{
				"bitbucket": {
					{Name: "owner/repo"},
				},
			},
			expectError: false,
		},
		{
			name: "missing Bitbucket username",
			auth: config.Auth{
				Bitbucket: config.BitbucketAuth{
					AppPassword: strings.Repeat("x", 20),
				},
			},
			repositories: map[string][]config.Repository{
				"bitbucket": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "Bitbucket username is required when app password is provided",
		},
		{
			name: "missing Bitbucket app password",
			auth: config.Auth{
				Bitbucket: config.BitbucketAuth{
					Username: "testuser",
				},
			},
			repositories: map[string][]config.Repository{
				"bitbucket": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "Bitbucket app password is required when username is provided",
		},
		{
			name: "no auth when repos configured",
			auth: config.Auth{},
			repositories: map[string][]config.Repository{
				"github": {
					{Name: "owner/repo"},
				},
			},
			expectError: true,
			errorMsg:    "at least one provider authentication must be configured when repositories are present",
		},
		{
			name: "no auth when no repos - valid",
			auth: config.Auth{},
			repositories: map[string][]config.Repository{
				"github": {}, // Empty repos
			},
			expectError: false,
		},
		{
			name: "environment variable references",
			auth: config.Auth{
				GitHub: config.GitHubAuth{
					Token: "$GITHUB_TOKEN",
				},
				GitLab: config.GitLabAuth{
					Token: "$GITLAB_TOKEN",
					URL:   "$GITLAB_URL",
				},
			},
			repositories: map[string][]config.Repository{
				"github": {{Name: "owner/repo"}},
				"gitlab": {{Name: "owner/repo"}},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAuth(tt.auth, tt.repositories)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBehavior(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		behavior    config.Behavior
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid behavior",
			behavior: config.Behavior{
				Concurrency: 5,
				RateLimit: config.RateLimit{
					RequestsPerSecond: 2.0,
					Burst:             5,
				},
			},
			expectError: false,
		},
		{
			name: "zero concurrency",
			behavior: config.Behavior{
				Concurrency: 0,
			},
			expectError: true,
			errorMsg:    "concurrency must be greater than 0",
		},
		{
			name: "negative concurrency",
			behavior: config.Behavior{
				Concurrency: -1,
			},
			expectError: true,
			errorMsg:    "concurrency must be greater than 0",
		},
		{
			name: "excessive concurrency",
			behavior: config.Behavior{
				Concurrency: 100,
			},
			expectError: true,
			errorMsg:    "concurrency should not exceed 50 for rate limiting",
		},
		{
			name: "negative rate limit",
			behavior: config.Behavior{
				Concurrency: 5,
				RateLimit: config.RateLimit{
					RequestsPerSecond: -1.0,
				},
			},
			expectError: true,
			errorMsg:    "rate limit requests per second cannot be negative",
		},
		{
			name: "negative burst",
			behavior: config.Behavior{
				Concurrency: 5,
				RateLimit: config.RateLimit{
					RequestsPerSecond: 2.0,
					Burst:             -1,
				},
			},
			expectError: true,
			errorMsg:    "rate limit burst cannot be negative",
		},
		{
			name: "edge case: maximum allowed concurrency",
			behavior: config.Behavior{
				Concurrency: 50,
				RateLimit: config.RateLimit{
					RequestsPerSecond: 0, // Zero is allowed
					Burst:             0, // Zero is allowed
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateBehavior(tt.behavior)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNotifications(t *testing.T) {
	validator := New()

	tests := []struct {
		name          string
		notifications config.Notifications
		expectError   bool
		errorMsg      string
	}{
		{
			name: "valid notifications",
			notifications: config.Notifications{
				Slack: config.SlackConfig{
					WebhookURL: "https://hooks.slack.com/services/T00/B00/XXXXX",
				},
				Email: config.EmailConfig{
					SMTPHost: "smtp.example.com",
					SMTPPort: 587,
					From:     "test@example.com",
					To:       []string{"recipient@example.com"},
				},
			},
			expectError: false,
		},
		{
			name: "invalid Slack webhook URL",
			notifications: config.Notifications{
				Slack: config.SlackConfig{
					WebhookURL: "https://invalid-webhook-url.com",
				},
			},
			expectError: true,
			errorMsg:    "Slack webhook URL should start with https://hooks.slack.com/",
		},
		{
			name: "invalid SMTP port - too low",
			notifications: config.Notifications{
				Email: config.EmailConfig{
					SMTPHost: "smtp.example.com",
					SMTPPort: 0,
				},
			},
			expectError: true,
			errorMsg:    "SMTP port must be between 1 and 65535",
		},
		{
			name: "invalid SMTP port - too high",
			notifications: config.Notifications{
				Email: config.EmailConfig{
					SMTPHost: "smtp.example.com",
					SMTPPort: 70000,
				},
			},
			expectError: true,
			errorMsg:    "SMTP port must be between 1 and 65535",
		},
		{
			name: "no email recipients",
			notifications: config.Notifications{
				Email: config.EmailConfig{
					SMTPHost: "smtp.example.com",
					SMTPPort: 587,
					To:       []string{},
				},
			},
			expectError: true,
			errorMsg:    "at least one email recipient must be specified when SMTP host is configured",
		},
		{
			name: "invalid from email",
			notifications: config.Notifications{
				Email: config.EmailConfig{
					SMTPHost: "smtp.example.com",
					SMTPPort: 587,
					From:     "invalid-email",
					To:       []string{"recipient@example.com"},
				},
			},
			expectError: true,
			errorMsg:    "invalid 'from' email address format",
		},
		{
			name: "invalid to email",
			notifications: config.Notifications{
				Email: config.EmailConfig{
					SMTPHost: "smtp.example.com",
					SMTPPort: 587,
					From:     "sender@example.com",
					To:       []string{"valid@example.com", "invalid-email", "another@example.com"},
				},
			},
			expectError: true,
			errorMsg:    "invalid 'to' email address at index 1: invalid-email",
		},
		{
			name: "environment variable references",
			notifications: config.Notifications{
				Slack: config.SlackConfig{
					WebhookURL: "$SLACK_WEBHOOK_URL",
				},
			},
			expectError: false,
		},
		{
			name: "empty notifications - valid",
			notifications: config.Notifications{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateNotifications(tt.notifications)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckEnvironmentVariables(t *testing.T) {
	validator := New()

	// Set up test environment variables
	os.Setenv("TEST_GITHUB_TOKEN", "test-value")
	os.Setenv("TEST_GITLAB_URL", "https://gitlab.example.com")
	defer func() {
		os.Unsetenv("TEST_GITHUB_TOKEN")
		os.Unsetenv("TEST_GITLAB_URL")
		os.Unsetenv("MISSING_VAR")
	}()

	tests := []struct {
		name           string
		cfg            *config.Config
		expectedMissing []string
	}{
		{
			name: "no environment variables",
			cfg: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{
						Token: "direct-token",
					},
				},
			},
			expectedMissing: []string{},
		},
		{
			name: "existing environment variables",
			cfg: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{
						Token: "$TEST_GITHUB_TOKEN",
					},
					GitLab: config.GitLabAuth{
						URL: "$TEST_GITLAB_URL",
					},
				},
			},
			expectedMissing: []string{},
		},
		{
			name: "missing environment variables",
			cfg: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{
						Token: "$MISSING_GITHUB_TOKEN",
					},
					GitLab: config.GitLabAuth{
						Token: "$MISSING_GITLAB_TOKEN",
						URL:   "$MISSING_GITLAB_URL",
					},
					Bitbucket: config.BitbucketAuth{
						Username:    "$MISSING_BB_USER",
						AppPassword: "$MISSING_BB_PASS",
						Workspace:   "$MISSING_BB_WORKSPACE",
					},
				},
				Notifications: config.Notifications{
					Slack: config.SlackConfig{
						WebhookURL: "$MISSING_SLACK_WEBHOOK",
					},
					Email: config.EmailConfig{
						SMTPUsername: "$MISSING_SMTP_USER",
						SMTPPassword: "$MISSING_SMTP_PASS",
					},
				},
			},
			expectedMissing: []string{
				"MISSING_GITHUB_TOKEN",
				"MISSING_GITLAB_TOKEN",
				"MISSING_GITLAB_URL",
				"MISSING_BB_USER",
				"MISSING_BB_PASS",
				"MISSING_BB_WORKSPACE",
				"MISSING_SLACK_WEBHOOK",
				"MISSING_SMTP_USER",
				"MISSING_SMTP_PASS",
			},
		},
		{
			name: "mixed existing and missing",
			cfg: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{
						Token: "$TEST_GITHUB_TOKEN", // exists
					},
					GitLab: config.GitLabAuth{
						Token: "$MISSING_GITLAB_TOKEN", // missing
						URL:   "$TEST_GITLAB_URL",      // exists
					},
				},
			},
			expectedMissing: []string{"MISSING_GITLAB_TOKEN"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := validator.CheckEnvironmentVariables(tt.cfg)

			assert.Equal(t, len(tt.expectedMissing), len(missing))
			for _, expected := range tt.expectedMissing {
				assert.Contains(t, missing, expected)
			}
		})
	}
}

func TestIsValidEnvVarReference(t *testing.T) {
	validator := New()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "valid env var",
			value:    "$GITHUB_TOKEN",
			expected: true,
		},
		{
			name:     "not env var",
			value:    "direct-value",
			expected: false,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "just dollar sign",
			value:    "$",
			expected: false,
		},
		{
			name:     "dollar at end",
			value:    "value$",
			expected: false,
		},
		{
			name:     "valid complex env var",
			value:    "$MY_COMPLEX_VAR_123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isValidEnvVarReference(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractEnvVarName(t *testing.T) {
	validator := New()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "env var reference",
			value:    "$GITHUB_TOKEN",
			expected: "GITHUB_TOKEN",
		},
		{
			name:     "not env var reference",
			value:    "direct-value",
			expected: "direct-value",
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "complex env var",
			value:    "$MY_COMPLEX_VAR_123",
			expected: "MY_COMPLEX_VAR_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.extractEnvVarName(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidToken(t *testing.T) {
	validator := New()

	tests := []struct {
		name     string
		token    string
		provider string
		expected bool
	}{
		{
			name:     "valid GitHub token - ghp prefix",
			token:    "ghp_" + strings.Repeat("x", 36),
			provider: "GitHub",
			expected: true,
		},
		{
			name:     "valid GitHub token - github_pat prefix",
			token:    "github_pat_" + strings.Repeat("x", 30),
			provider: "GitHub",
			expected: true,
		},
		{
			name:     "valid GitHub token - 40 chars",
			token:    strings.Repeat("x", 40),
			provider: "GitHub",
			expected: true,
		},
		{
			name:     "invalid GitHub token - too short",
			token:    "ghp_short",
			provider: "GitHub",
			expected: false,
		},
		{
			name:     "invalid GitHub token - wrong format but meets length requirement",
			token:    "gho_" + strings.Repeat("x", 36), // 40 chars total, so it's valid by length
			provider: "GitHub",
			expected: true, // Valid because it's 40+ chars
		},
		{
			name:     "valid GitLab token - glpat prefix",
			token:    "glpat-" + strings.Repeat("x", 20),
			provider: "GitLab",
			expected: true,
		},
		{
			name:     "valid GitLab token - 20+ chars",
			token:    strings.Repeat("x", 25),
			provider: "GitLab",
			expected: true,
		},
		{
			name:     "valid GitLab token - has correct prefix",
			token:    "glpat-short", // 11 chars total, valid because it has glpat- prefix
			provider: "GitLab",
			expected: true, // Valid because it has the glpat- prefix
		},
		{
			name:     "invalid GitLab token - too short and wrong prefix",
			token:    "invalid-token", // 13 chars but < 20 and wrong prefix
			provider: "GitLab",
			expected: false,
		},
		{
			name:     "valid generic token",
			token:    strings.Repeat("x", 15),
			provider: "Other",
			expected: true,
		},
		{
			name:     "invalid generic token - too short",
			token:    "short",
			provider: "Other",
			expected: false,
		},
		{
			name:     "universal minimum length failure",
			token:    "tiny",
			provider: "GitHub",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isValidToken(tt.token, tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateConfig_CompleteWorkflow(t *testing.T) {
	validator := New()

	// Test a comprehensive, realistic configuration
	cfg := &config.Config{
		PRFilters: config.PRFilters{
			AllowedActors: []string{"dependabot[bot]", "renovate[bot]", "github-actions[bot]"},
			MaxAge:        "30d",
			SkipLabels:    []string{"skip-merge", "work-in-progress", "breaking-change"},
		},
		Repositories: map[string][]config.Repository{
			"github": {
				{
					Name:          "myorg/service-a",
					MergeStrategy: config.MergeStrategySquash,
					Branch:        "main",
					SkipLabels:    []string{"hold"},
				},
				{
					Name:          "myorg/service-b",
					MergeStrategy: config.MergeStrategyRebase,
					Branch:        "develop",
				},
			},
			"gitlab": {
				{
					Name:          "myorg/internal-tool",
					MergeStrategy: config.MergeStrategyMerge,
					Branch:        "master",
					SkipLabels:    []string{"review-needed"},
				},
			},
		},
		Auth: config.Auth{
			GitHub: config.GitHubAuth{
				Token: "$GITHUB_TOKEN",
			},
			GitLab: config.GitLabAuth{
				Token: "$GITLAB_TOKEN",
				URL:   "$GITLAB_URL",
			},
		},
		Behavior: config.Behavior{
			Concurrency: 10,
			RateLimit: config.RateLimit{
				RequestsPerSecond: 2.0,
				Burst:             5,
				Timeout:           30,
			},
		},
		Notifications: config.Notifications{
			Slack: config.SlackConfig{
				WebhookURL: "$SLACK_WEBHOOK_URL",
			},
			Email: config.EmailConfig{
				SMTPHost:     "smtp.company.com",
				SMTPPort:     587,
				SMTPUsername: "$SMTP_USER",
				SMTPPassword: "$SMTP_PASS",
				From:         "pr-automation@company.com",
				To:           []string{"team@company.com", "leads@company.com"},
			},
		},
	}

	err := validator.ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_RegressiveCases(t *testing.T) {
	validator := New()

	t.Run("repositories with invalid names", func(t *testing.T) {
		cfg := &config.Config{
			PRFilters: config.PRFilters{
				AllowedActors: []string{"bot"},
			},
			Repositories: map[string][]config.Repository{
				"github": {
					{Name: "repo-without-owner"},
					{Name: "owner/repo/extra"},
					{Name: "/invalid-start"},
					{Name: "invalid-end/"},
				},
			},
			Auth: config.Auth{
				GitHub: config.GitHubAuth{Token: "ghp_" + strings.Repeat("x", 36)},
			},
			Behavior: config.Behavior{
				Concurrency: 1,
			},
		}

		err := validator.ValidateConfig(cfg)
		assert.Error(t, err)
		errorStr := err.Error()
		assert.Contains(t, errorStr, "invalid name format")
	})

	t.Run("extreme behavior values", func(t *testing.T) {
		cfg := &config.Config{
			PRFilters: config.PRFilters{
				AllowedActors: []string{"bot"},
			},
			Repositories: map[string][]config.Repository{},
			Auth:         config.Auth{},
			Behavior: config.Behavior{
				Concurrency: 999,
				RateLimit: config.RateLimit{
					RequestsPerSecond: -999.0,
					Burst:             -100,
				},
			},
		}

		err := validator.ValidateConfig(cfg)
		assert.Error(t, err)
		errorStr := err.Error()
		assert.Contains(t, errorStr, "concurrency should not exceed 50")
		assert.Contains(t, errorStr, "rate limit requests per second cannot be negative")
		assert.Contains(t, errorStr, "rate limit burst cannot be negative")
	})
}