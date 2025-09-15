package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	assert.NotNil(t, loader)
	assert.NotNil(t, loader.validator)
}

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(t *testing.T) string
		setupEnv    func(t *testing.T)
		cleanupEnv  func(t *testing.T)
		expectError bool
		validate    func(t *testing.T, config *Config)
	}{
		{
			name: "invalid config file - missing required fields",
			setupFile: func(t *testing.T) string {
				content := `
repositories:
  github:
    - name: "repo1"
`
				return createTempConfigFile(t, content)
			},
			expectError: true,
		},
		{
			name: "config file does not exist",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/config.yaml"
			},
			expectError: true,
		},
		{
			name: "invalid yaml format",
			setupFile: func(t *testing.T) string {
				content := `
pr_filters:
  allowed_actors:
    - "user1
# Missing closing quote - invalid YAML
`
				return createTempConfigFile(t, content)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv(t)
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv(t)
			}

			configPath := tt.setupFile(t)
			if configPath != "/nonexistent/config.yaml" {
				defer os.Remove(configPath)
			}

			loader := NewLoader()
			config, err := loader.Load(configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				if err != nil {
					t.Logf("Error: %v", err)
				}
				assert.NoError(t, err)
				assert.NotNil(t, config)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

func TestLoader_Save(t *testing.T) {
	loader := NewLoader()

	config := &Config{
		PRFilters: PRFilters{
			AllowedActors: []string{"user1"},
		},
		Repositories: map[string][]Repository{
			"github": {
				{Name: "repo1", MergeStrategy: MergeStrategySquash},
			},
		},
		Auth: Auth{
			GitHub: GitHubAuth{Token: "test_token"},
		},
	}

	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := loader.Save(config, configPath)
	assert.NoError(t, err)

	// Verify file exists
	assert.FileExists(t, configPath)

	// Verify file contains expected content
	data, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "user1")
	assert.Contains(t, string(data), "test_token")
}

func TestLoader_BackupConfig(t *testing.T) {
	loader := NewLoader()

	t.Run("successful backup", func(t *testing.T) {
		// Create original config file
		content := `
pr_filters:
  allowed_actors:
    - "user1"
repositories:
  github:
    - name: "repo1"
auth:
  github:
    token: "test_token"
`
		configPath := createTempConfigFile(t, content)
		defer os.Remove(configPath)

		backupPath, err := loader.BackupConfig(configPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, backupPath)
		defer os.Remove(backupPath)

		// Verify backup file exists and has same content
		assert.FileExists(t, backupPath)

		originalData, err := os.ReadFile(configPath)
		assert.NoError(t, err)

		backupData, err := os.ReadFile(backupPath)
		assert.NoError(t, err)

		assert.Equal(t, originalData, backupData)
	})

	t.Run("backup non-existent file", func(t *testing.T) {
		backupPath, err := loader.BackupConfig("/nonexistent/config.yaml")
		assert.Error(t, err)
		assert.Empty(t, backupPath)
		assert.Contains(t, err.Error(), "config file does not exist")
	})
}

func TestLoader_findConfigFile(t *testing.T) {
	loader := NewLoader()

	t.Run("find config in current directory", func(t *testing.T) {
		// Create config file in current directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer func() {
			os.Chdir(originalWd)
		}()

		configPath := filepath.Join(tmpDir, "config.yaml")
		err = os.WriteFile(configPath, []byte("test"), 0644)
		require.NoError(t, err)

		found := loader.findConfigFile()
		assert.Equal(t, "config.yaml", found)
	})

	t.Run("no config file found", func(t *testing.T) {
		// Change to temp directory without config files
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer func() {
			os.Chdir(originalWd)
		}()

		found := loader.findConfigFile()
		assert.Empty(t, found)
	})
}

func TestLoader_setDefaults(t *testing.T) {
	loader := NewLoader()

	t.Run("set repository defaults", func(t *testing.T) {
		config := &Config{
			Repositories: map[string][]Repository{
				"github": {
					{Name: "repo1"}, // No merge strategy or branch set
					{Name: "repo2", MergeStrategy: MergeStrategyRebase, Branch: "develop"}, // Has values set
				},
			},
		}

		loader.setDefaults(config)

		assert.Equal(t, MergeStrategySquash, config.Repositories["github"][0].MergeStrategy)
		assert.Equal(t, "main", config.Repositories["github"][0].Branch)
		assert.Equal(t, MergeStrategyRebase, config.Repositories["github"][1].MergeStrategy)
		assert.Equal(t, "develop", config.Repositories["github"][1].Branch)
	})

	t.Run("set behavior defaults", func(t *testing.T) {
		config := &Config{}

		loader.setDefaults(config)

		assert.Equal(t, 5.0, config.Behavior.RateLimit.RequestsPerSecond)
		assert.Equal(t, 10, config.Behavior.RateLimit.Burst)
		assert.Equal(t, 30*time.Second, config.Behavior.RateLimit.Timeout)
		assert.Equal(t, 3, config.Behavior.Retry.MaxAttempts)
		assert.Equal(t, 1*time.Second, config.Behavior.Retry.Backoff)
		assert.Equal(t, 30*time.Second, config.Behavior.Retry.MaxBackoff)
		assert.Equal(t, 5, config.Behavior.Concurrency)
		assert.Equal(t, "30s", config.Behavior.WatchInterval)
	})

	t.Run("set gitlab url default", func(t *testing.T) {
		config := &Config{
			Auth: Auth{
				GitLab: GitLabAuth{Token: "test_token"},
			},
		}

		loader.setDefaults(config)

		assert.Equal(t, "https://gitlab.com", config.Auth.GitLab.URL)
	})

	t.Run("set email smtp port default", func(t *testing.T) {
		config := &Config{
			Notifications: Notifications{
				Email: EmailConfig{
					SMTPHost: "smtp.example.com",
				},
			},
		}

		loader.setDefaults(config)

		assert.Equal(t, 587, config.Notifications.Email.SMTPPort)
	})
}

func TestLoader_validateBusinessRules(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config passes validation",
			config: &Config{
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1", MergeStrategy: MergeStrategySquash},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
			},
			expectError: false,
		},
		{
			name: "invalid merge strategy fails validation",
			config: &Config{
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1", MergeStrategy: "invalid"},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
			},
			expectError: true,
			errorMsg:    "invalid merge strategy",
		},
		{
			name: "no authentication fails validation",
			config: &Config{
				Repositories: map[string][]Repository{
					"github": {
						{Name: "repo1", MergeStrategy: MergeStrategySquash},
					},
				},
				Auth: Auth{},
			},
			expectError: true,
			errorMsg:    "at least one provider must have authentication configured",
		},
		{
			name: "repositories without matching auth fails",
			config: &Config{
				Repositories: map[string][]Repository{
					"gitlab": {
						{Name: "repo1", MergeStrategy: MergeStrategySquash},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
			},
			expectError: true,
			errorMsg:    "GitLab repositories configured but no GitLab token provided",
		},
		{
			name: "slack enabled without webhook fails",
			config: &Config{
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
				Notifications: Notifications{
					Slack: SlackConfig{Enabled: true},
				},
			},
			expectError: true,
			errorMsg:    "slack notifications enabled but no webhook URL provided",
		},
		{
			name: "email enabled without smtp host fails",
			config: &Config{
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
				Notifications: Notifications{
					Email: EmailConfig{Enabled: true},
				},
			},
			expectError: true,
			errorMsg:    "email notifications enabled but no SMTP host provided",
		},
		{
			name: "email enabled without from address fails",
			config: &Config{
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
				Notifications: Notifications{
					Email: EmailConfig{
						Enabled:  true,
						SMTPHost: "smtp.example.com",
					},
				},
			},
			expectError: true,
			errorMsg:    "email notifications enabled but no 'from' address provided",
		},
		{
			name: "email enabled without to addresses fails",
			config: &Config{
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
				Notifications: Notifications{
					Email: EmailConfig{
						Enabled:  true,
						SMTPHost: "smtp.example.com",
						From:     "noreply@example.com",
					},
				},
			},
			expectError: true,
			errorMsg:    "email notifications enabled but no 'to' addresses provided",
		},
		{
			name: "unsupported provider fails",
			config: &Config{
				Repositories: map[string][]Repository{
					"unsupported": {
						{Name: "repo1", MergeStrategy: MergeStrategySquash},
					},
				},
				Auth: Auth{
					GitHub: GitHubAuth{Token: "test_token"},
				},
			},
			expectError: true,
			errorMsg:    "unsupported provider",
		},
		{
			name: "bitbucket with incomplete auth fails",
			config: &Config{
				Repositories: map[string][]Repository{
					"bitbucket": {
						{Name: "repo1", MergeStrategy: MergeStrategySquash},
					},
				},
				Auth: Auth{
					Bitbucket: BitbucketAuth{Username: "user"}, // Missing app password
				},
			},
			expectError: true,
			errorMsg:    "at least one provider must have authentication configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.validateBusinessRules(tt.config)
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

func TestLoader_processEnvVars(t *testing.T) {
	loader := NewLoader()

	t.Run("process github env vars", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "env_github_token")
		defer os.Unsetenv("GITHUB_TOKEN")

		config := &Config{}
		err := loader.processEnvVars(config)
		assert.NoError(t, err)
		assert.Equal(t, "env_github_token", config.Auth.GitHub.Token)
	})

	t.Run("process gitlab env vars", func(t *testing.T) {
		os.Setenv("GITLAB_TOKEN", "env_gitlab_token")
		os.Setenv("GITLAB_URL", "https://gitlab.example.com")
		defer func() {
			os.Unsetenv("GITLAB_TOKEN")
			os.Unsetenv("GITLAB_URL")
		}()

		config := &Config{}
		err := loader.processEnvVars(config)
		assert.NoError(t, err)
		assert.Equal(t, "env_gitlab_token", config.Auth.GitLab.Token)
		assert.Equal(t, "https://gitlab.example.com", config.Auth.GitLab.URL)
	})

	t.Run("process slack env vars", func(t *testing.T) {
		os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/env")
		defer os.Unsetenv("SLACK_WEBHOOK_URL")

		config := &Config{}
		err := loader.processEnvVars(config)
		assert.NoError(t, err)
		assert.Equal(t, "https://hooks.slack.com/env", config.Notifications.Slack.WebhookURL)
	})
}

func TestLoadConfigFromPath(t *testing.T) {
	// Test that the function exists and handles non-existent files
	config, err := LoadConfigFromPath("/nonexistent/path")
	assert.Error(t, err)
	assert.Nil(t, config)
}

// Helper function to create temporary config files for testing
func createTempConfigFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

// Benchmark tests
func BenchmarkLoader_Load(b *testing.B) {
	content := `
pr_filters:
  allowed_actors:
    - "user1"
    - "user2"
repositories:
  github:
    - name: "repo1"
      merge_strategy: "squash"
auth:
  github:
    token: "test_token"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		b.Fatal(err)
	}
	tmpFile.Close()

	loader := NewLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.Load(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoader_setDefaults(b *testing.B) {
	config := &Config{
		Repositories: map[string][]Repository{
			"github": {
				{Name: "repo1"},
				{Name: "repo2"},
			},
		},
	}

	loader := NewLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset config for each iteration
		testConfig := *config
		for provider, repos := range config.Repositories {
			testConfig.Repositories[provider] = make([]Repository, len(repos))
			copy(testConfig.Repositories[provider], repos)
		}
		loader.setDefaults(&testConfig)
	}
}
