package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeStrategy_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		strategy MergeStrategy
		expected bool
	}{
		{
			name:     "merge strategy is valid",
			strategy: MergeStrategyMerge,
			expected: true,
		},
		{
			name:     "squash strategy is valid",
			strategy: MergeStrategySquash,
			expected: true,
		},
		{
			name:     "rebase strategy is valid",
			strategy: MergeStrategyRebase,
			expected: true,
		},
		{
			name:     "empty strategy is invalid",
			strategy: "",
			expected: false,
		},
		{
			name:     "invalid strategy is invalid",
			strategy: "invalid",
			expected: false,
		},
		{
			name:     "capitalized strategy is invalid",
			strategy: "MERGE",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProvider_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected bool
	}{
		{
			name:     "github provider is valid",
			provider: ProviderGitHub,
			expected: true,
		},
		{
			name:     "gitlab provider is valid",
			provider: ProviderGitLab,
			expected: true,
		},
		{
			name:     "bitbucket provider is valid",
			provider: ProviderBitbucket,
			expected: true,
		},
		{
			name:     "empty provider is invalid",
			provider: "",
			expected: false,
		},
		{
			name:     "invalid provider is invalid",
			provider: "invalid",
			expected: false,
		},
		{
			name:     "capitalized provider is invalid",
			provider: "GITHUB",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_StructValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid minimal config",
			config: Config{
				PRFilters: PRFilters{
					AllowedActors: []string{"user1"},
				},
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1"},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "token"},
				},
			},
			valid: true,
		},
		{
			name: "config missing pr_filters",
			config: Config{
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1"},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "token"},
				},
			},
			valid: false,
		},
		{
			name: "config missing repositories",
			config: Config{
				PRFilters: PRFilters{
					AllowedActors: []string{"user1"},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "token"},
				},
			},
			valid: false,
		},
		{
			name: "config missing auth",
			config: Config{
				PRFilters: PRFilters{
					AllowedActors: []string{"user1"},
				},
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1"},
					},
				},
			},
			valid: false,
		},
		{
			name: "config with empty allowed_actors",
			config: Config{
				PRFilters: PRFilters{
					AllowedActors: []string{},
				},
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1"},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "token"},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			err := loader.validator.Struct(&tt.config)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRepository_Validation(t *testing.T) {
	tests := []struct {
		name       string
		repository Repository
		valid      bool
	}{
		{
			name: "valid repository with all fields",
			repository: Repository{
				Name:          "test-repo",
				AutoMerge:     true,
				MergeStrategy: MergeStrategySquash,
				SkipLabels:    []string{"wip", "draft"},
				Branch:        "main",
				RequireChecks: true,
				MinApprovals:  2,
			},
			valid: true,
		},
		{
			name: "valid minimal repository",
			repository: Repository{
				Name: "test-repo",
			},
			valid: true,
		},
		{
			name: "repository missing name",
			repository: Repository{
				AutoMerge:     true,
				MergeStrategy: MergeStrategySquash,
			},
			valid: false,
		},
		{
			name: "repository with empty name",
			repository: Repository{
				Name: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			err := loader.validator.Struct(&tt.repository)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPRFilters_Validation(t *testing.T) {
	tests := []struct {
		name      string
		prFilters PRFilters
		valid     bool
	}{
		{
			name: "valid pr filters with multiple actors",
			prFilters: PRFilters{
				AllowedActors: []string{"user1", "user2", "bot"},
				SkipLabels:    []string{"wip", "draft"},
				MaxAge:        "7d",
			},
			valid: true,
		},
		{
			name: "valid pr filters minimal",
			prFilters: PRFilters{
				AllowedActors: []string{"user1"},
			},
			valid: true,
		},
		{
			name: "invalid pr filters empty actors",
			prFilters: PRFilters{
				AllowedActors: []string{},
			},
			valid: false,
		},
		{
			name: "invalid pr filters nil actors",
			prFilters: PRFilters{
				AllowedActors: nil,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			err := loader.validator.Struct(&tt.prFilters)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuth_Structures(t *testing.T) {
	t.Run("GitHubAuth validation", func(t *testing.T) {
		tests := []struct {
			name  string
			auth  GitHubAuth
			valid bool
		}{
			{
				name:  "valid github auth",
				auth:  GitHubAuth{Token: "ghp_test_token"},
				valid: true,
			},
			{
				name:  "invalid github auth empty token",
				auth:  GitHubAuth{Token: ""},
				valid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loader := NewLoader()
				err := loader.validator.Struct(&tt.auth)
				if tt.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("GitLabAuth structure", func(t *testing.T) {
		auth := GitLabAuth{
			Token: "glpat_test_token",
			URL:   "https://gitlab.example.com",
		}
		assert.Equal(t, "glpat_test_token", auth.Token)
		assert.Equal(t, "https://gitlab.example.com", auth.URL)
	})

	t.Run("BitbucketAuth structure", func(t *testing.T) {
		auth := BitbucketAuth{
			Username:    "testuser",
			AppPassword: "test_password",
			Workspace:   "test_workspace",
		}
		assert.Equal(t, "testuser", auth.Username)
		assert.Equal(t, "test_password", auth.AppPassword)
		assert.Equal(t, "test_workspace", auth.Workspace)
	})
}

func TestNotificationConfigs(t *testing.T) {
	t.Run("SlackConfig structure", func(t *testing.T) {
		config := SlackConfig{
			WebhookURL: "https://hooks.slack.com/test",
			Channel:    "#general",
			Enabled:    true,
		}
		assert.Equal(t, "https://hooks.slack.com/test", config.WebhookURL)
		assert.Equal(t, "#general", config.Channel)
		assert.True(t, config.Enabled)
	})

	t.Run("EmailConfig structure", func(t *testing.T) {
		config := EmailConfig{
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUsername: "user@example.com",
			SMTPPassword: "password",
			From:         "noreply@example.com",
			To:           []string{"admin@example.com", "dev@example.com"},
			Enabled:      true,
		}
		assert.Equal(t, "smtp.example.com", config.SMTPHost)
		assert.Equal(t, 587, config.SMTPPort)
		assert.Equal(t, "user@example.com", config.SMTPUsername)
		assert.Equal(t, "password", config.SMTPPassword)
		assert.Equal(t, "noreply@example.com", config.From)
		assert.Equal(t, []string{"admin@example.com", "dev@example.com"}, config.To)
		assert.True(t, config.Enabled)
	})
}

func TestBehaviorConfig(t *testing.T) {
	t.Run("RateLimit structure", func(t *testing.T) {
		rateLimit := RateLimit{
			RequestsPerSecond: 5.0,
			Burst:             10,
			Timeout:           30000000000, // 30 seconds in nanoseconds
		}
		assert.Equal(t, 5.0, rateLimit.RequestsPerSecond)
		assert.Equal(t, 10, rateLimit.Burst)
		assert.Equal(t, int64(30000000000), int64(rateLimit.Timeout))
	})

	t.Run("Retry structure", func(t *testing.T) {
		retry := Retry{
			MaxAttempts: 3,
			Backoff:     1000000000,  // 1 second in nanoseconds
			MaxBackoff:  30000000000, // 30 seconds in nanoseconds
		}
		assert.Equal(t, 3, retry.MaxAttempts)
		assert.Equal(t, int64(1000000000), int64(retry.Backoff))
		assert.Equal(t, int64(30000000000), int64(retry.MaxBackoff))
	})

	t.Run("Behavior structure", func(t *testing.T) {
		behavior := Behavior{
			RateLimit: RateLimit{
				RequestsPerSecond: 5.0,
				Burst:             10,
			},
			Retry: Retry{
				MaxAttempts: 3,
			},
			Concurrency:     5,
			DryRun:          true,
			WatchInterval:   "30s",
			RequireApproval: true,
		}
		assert.Equal(t, 5.0, behavior.RateLimit.RequestsPerSecond)
		assert.Equal(t, 3, behavior.Retry.MaxAttempts)
		assert.Equal(t, 5, behavior.Concurrency)
		assert.True(t, behavior.DryRun)
		assert.Equal(t, "30s", behavior.WatchInterval)
		assert.True(t, behavior.RequireApproval)
	})
}

func TestConstants(t *testing.T) {
	t.Run("MergeStrategy constants", func(t *testing.T) {
		assert.Equal(t, MergeStrategy("merge"), MergeStrategyMerge)
		assert.Equal(t, MergeStrategy("squash"), MergeStrategySquash)
		assert.Equal(t, MergeStrategy("rebase"), MergeStrategyRebase)
	})

	t.Run("Provider constants", func(t *testing.T) {
		assert.Equal(t, Provider("github"), ProviderGitHub)
		assert.Equal(t, Provider("gitlab"), ProviderGitLab)
		assert.Equal(t, Provider("bitbucket"), ProviderBitbucket)
	})
}

// Benchmark tests for validation
func BenchmarkMergeStrategy_IsValid(b *testing.B) {
	strategy := MergeStrategySquash
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.IsValid()
	}
}

func BenchmarkProvider_IsValid(b *testing.B) {
	provider := ProviderGitHub
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.IsValid()
	}
}

func BenchmarkConfig_Validation(b *testing.B) {
	config := Config{
		PRFilters: PRFilters{
			AllowedActors: []string{"user1", "user2"},
			SkipLabels:    []string{"wip"},
			MaxAge:        "7d",
		},
		Repositories: map[string][]Repository{
			"github": {
				{
					Name:          "repo1",
					AutoMerge:     true,
					MergeStrategy: MergeStrategySquash,
				},
			},
		},
		Auth: Auth{
			GitHub: GitHubAuth{Token: "test_token"},
		},
	}

	loader := NewLoader()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.validator.Struct(&config)
	}
}
