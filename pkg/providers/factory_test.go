package providers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
)

func TestNewFactory(t *testing.T) {
	cfg := &config.Config{}
	factory := NewFactory(cfg)

	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.config)
}

func TestFactory_CreateProviders(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectError   bool
		expectedCount int
		expectedTypes []string
	}{
		{
			name: "GitHub provider only",
			config: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{Token: "github_token"},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
				Repositories: map[string][]config.Repository{
					"github": {{Name: "owner/repo"}},
				},
			},
			expectError:   false,
			expectedCount: 1,
			expectedTypes: []string{"github"},
		},
		{
			name: "GitLab provider only",
			config: &config.Config{
				Auth: config.Auth{
					GitLab: config.GitLabAuth{
						Token: "gitlab_token",
						URL:   "https://gitlab.example.com",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
				Repositories: map[string][]config.Repository{
					"gitlab": {{Name: "owner/repo"}},
				},
			},
			expectError:   false,
			expectedCount: 1,
			expectedTypes: []string{"gitlab"},
		},
		{
			name: "Bitbucket provider only",
			config: &config.Config{
				Auth: config.Auth{
					Bitbucket: config.BitbucketAuth{
						Username:    "bitbucket_user",
						AppPassword: "bitbucket_password",
						Workspace:   "workspace",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
				Repositories: map[string][]config.Repository{
					"bitbucket": {{Name: "owner/repo"}},
				},
			},
			expectError:   false,
			expectedCount: 1,
			expectedTypes: []string{"bitbucket"},
		},
		{
			name: "all providers configured",
			config: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{Token: "github_token"},
					GitLab: config.GitLabAuth{
						Token: "gitlab_token",
						URL:   "https://gitlab.example.com",
					},
					Bitbucket: config.BitbucketAuth{
						Username:    "bitbucket_user",
						AppPassword: "bitbucket_password",
						Workspace:   "workspace",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
				Repositories: map[string][]config.Repository{
					"github": {{Name: "owner/github-repo"}},
					"gitlab": {{Name: "owner/gitlab-repo"}},
					"bitbucket": {{Name: "owner/bitbucket-repo"}},
				},
			},
			expectError:   false,
			expectedCount: 3,
			expectedTypes: []string{"github", "gitlab", "bitbucket"},
		},
		{
			name: "no providers configured",
			config: &config.Config{
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError:   true,
			expectedCount: 0,
		},
		{
			name: "incomplete Bitbucket config",
			config: &config.Config{
				Auth: config.Auth{
					Bitbucket: config.BitbucketAuth{
						Username: "bitbucket_user",
						// Missing AppPassword
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.config)
			providers, err := factory.CreateProviders()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, providers)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, providers)
				assert.Len(t, providers, tt.expectedCount)

				for _, providerType := range tt.expectedTypes {
					assert.Contains(t, providers, providerType)
					assert.NotNil(t, providers[providerType])
				}
			}
		})
	}
}

func TestFactory_createGitHubProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid GitHub config",
			config: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{Token: "github_token"},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty GitHub token",
			config: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{Token: ""},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: true,
			errorMsg:    "GitHub token is required",
		},
		{
			name: "environment variable token",
			config: &config.Config{
				Auth: config.Auth{
					GitHub: config.GitHubAuth{Token: "$GITHUB_TOKEN"},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment for env var test
			if tt.name == "environment variable token" {
				os.Setenv("GITHUB_TOKEN", "env_github_token")
				defer os.Unsetenv("GITHUB_TOKEN")
			}

			factory := NewFactory(tt.config)
			provider, err := factory.createGitHubProvider()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestFactory_createGitLabProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid GitLab config",
			config: &config.Config{
				Auth: config.Auth{
					GitLab: config.GitLabAuth{
						Token: "gitlab_token",
						URL:   "https://gitlab.example.com",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty GitLab token",
			config: &config.Config{
				Auth: config.Auth{
					GitLab: config.GitLabAuth{
						Token: "",
						URL:   "https://gitlab.example.com",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: true,
			errorMsg:    "GitLab token is required",
		},
		{
			name: "environment variable config",
			config: &config.Config{
				Auth: config.Auth{
					GitLab: config.GitLabAuth{
						Token: "$GITLAB_TOKEN",
						URL:   "$GITLAB_URL",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment for env var test
			if tt.name == "environment variable config" {
				os.Setenv("GITLAB_TOKEN", "env_gitlab_token")
				os.Setenv("GITLAB_URL", "https://gitlab.env.com")
				defer func() {
					os.Unsetenv("GITLAB_TOKEN")
					os.Unsetenv("GITLAB_URL")
				}()
			}

			factory := NewFactory(tt.config)
			provider, err := factory.createGitLabProvider()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestFactory_createBitbucketProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid Bitbucket config",
			config: &config.Config{
				Auth: config.Auth{
					Bitbucket: config.BitbucketAuth{
						Username:    "bitbucket_user",
						AppPassword: "bitbucket_password",
						Workspace:   "workspace",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing username",
			config: &config.Config{
				Auth: config.Auth{
					Bitbucket: config.BitbucketAuth{
						Username:    "",
						AppPassword: "bitbucket_password",
						Workspace:   "workspace",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: true,
			errorMsg:    "bitbucket username and app password are required",
		},
		{
			name: "missing app password",
			config: &config.Config{
				Auth: config.Auth{
					Bitbucket: config.BitbucketAuth{
						Username:    "bitbucket_user",
						AppPassword: "",
						Workspace:   "workspace",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: true,
			errorMsg:    "bitbucket username and app password are required",
		},
		{
			name: "environment variable config",
			config: &config.Config{
				Auth: config.Auth{
					Bitbucket: config.BitbucketAuth{
						Username:    "$BITBUCKET_USERNAME",
						AppPassword: "$BITBUCKET_APP_PASSWORD",
						Workspace:   "$BITBUCKET_WORKSPACE",
					},
				},
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 5.0,
						Burst:             10,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment for env var test
			if tt.name == "environment variable config" {
				os.Setenv("BITBUCKET_USERNAME", "env_username")
				os.Setenv("BITBUCKET_APP_PASSWORD", "env_password")
				os.Setenv("BITBUCKET_WORKSPACE", "env_workspace")
				defer func() {
					os.Unsetenv("BITBUCKET_USERNAME")
					os.Unsetenv("BITBUCKET_APP_PASSWORD")
					os.Unsetenv("BITBUCKET_WORKSPACE")
				}()
			}

			factory := NewFactory(tt.config)
			provider, err := factory.createBitbucketProvider()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestCreateProvider(t *testing.T) {
	repositories := []common.Repository{
		{FullName: "owner/repo1"},
		{FullName: "owner/repo2"},
	}

	tests := []struct {
		name         string
		providerType string
		setupEnv     func()
		cleanupEnv   func()
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "GitHub provider with GITHUB_TOKEN",
			providerType: "github",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "github_token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
			},
			expectError: false,
		},
		{
			name:         "GitHub provider with GH_TOKEN",
			providerType: "github",
			setupEnv: func() {
				os.Setenv("GH_TOKEN", "gh_token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GH_TOKEN")
			},
			expectError: false,
		},
		{
			name:         "GitHub provider without token",
			providerType: "github",
			setupEnv:     func() {},
			cleanupEnv:   func() {},
			expectError:  true,
			errorMsg:     "GitHub token not found",
		},
		{
			name:         "GitLab provider with token",
			providerType: "gitlab",
			setupEnv: func() {
				os.Setenv("GITLAB_TOKEN", "gitlab_token")
				os.Setenv("GITLAB_URL", "https://gitlab.example.com")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITLAB_TOKEN")
				os.Unsetenv("GITLAB_URL")
			},
			expectError: false,
		},
		{
			name:         "GitLab provider without token",
			providerType: "gitlab",
			setupEnv:     func() {},
			cleanupEnv:   func() {},
			expectError:  true,
			errorMsg:     "GitLab token not found",
		},
		{
			name:         "Bitbucket provider with credentials",
			providerType: "bitbucket",
			setupEnv: func() {
				os.Setenv("BITBUCKET_USERNAME", "username")
				os.Setenv("BITBUCKET_APP_PASSWORD", "password")
				os.Setenv("BITBUCKET_WORKSPACE", "workspace")
			},
			cleanupEnv: func() {
				os.Unsetenv("BITBUCKET_USERNAME")
				os.Unsetenv("BITBUCKET_APP_PASSWORD")
				os.Unsetenv("BITBUCKET_WORKSPACE")
			},
			expectError: false,
		},
		{
			name:         "Bitbucket provider without credentials",
			providerType: "bitbucket",
			setupEnv:     func() {},
			cleanupEnv:   func() {},
			expectError:  true,
			errorMsg:     "bitbucket credentials not found",
		},
		{
			name:         "unsupported provider type",
			providerType: "unsupported",
			setupEnv:     func() {},
			cleanupEnv:   func() {},
			expectError:  true,
			errorMsg:     "unsupported provider type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			provider, err := CreateProvider(tt.providerType, repositories)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestFactory_resolveEnvVar(t *testing.T) {
	factory := NewFactory(&config.Config{})

	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		setEnv   bool
		expected string
	}{
		{
			name:     "regular value",
			input:    "regular_value",
			expected: "regular_value",
		},
		{
			name:     "empty value",
			input:    "",
			expected: "",
		},
		{
			name:     "environment variable with value",
			input:    "$TEST_VAR",
			envVar:   "TEST_VAR",
			envValue: "env_value",
			setEnv:   true,
			expected: "env_value",
		},
		{
			name:     "environment variable not set",
			input:    "$TEST_VAR_NOT_SET",
			envVar:   "TEST_VAR_NOT_SET",
			setEnv:   false,
			expected: "",
		},
		{
			name:     "dollar sign only",
			input:    "$",
			expected: "$",
		},
		{
			name:     "dollar sign in middle",
			input:    "value$with$dollar",
			expected: "value$with$dollar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := factory.resolveEnvVar(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFactory_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("create all providers with environment variables", func(t *testing.T) {
		// Set up environment variables
		os.Setenv("GITHUB_TOKEN", "github_token")
		os.Setenv("GITLAB_TOKEN", "gitlab_token")
		os.Setenv("GITLAB_URL", "https://gitlab.example.com")
		os.Setenv("BITBUCKET_USERNAME", "username")
		os.Setenv("BITBUCKET_APP_PASSWORD", "password")
		os.Setenv("BITBUCKET_WORKSPACE", "workspace")

		defer func() {
			os.Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("GITLAB_TOKEN")
			os.Unsetenv("GITLAB_URL")
			os.Unsetenv("BITBUCKET_USERNAME")
			os.Unsetenv("BITBUCKET_APP_PASSWORD")
			os.Unsetenv("BITBUCKET_WORKSPACE")
		}()

		cfg := &config.Config{
			Auth: config.Auth{
				GitHub: config.GitHubAuth{Token: "$GITHUB_TOKEN"},
				GitLab: config.GitLabAuth{
					Token: "$GITLAB_TOKEN",
					URL:   "$GITLAB_URL",
				},
				Bitbucket: config.BitbucketAuth{
					Username:    "$BITBUCKET_USERNAME",
					AppPassword: "$BITBUCKET_APP_PASSWORD",
					Workspace:   "$BITBUCKET_WORKSPACE",
				},
			},
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
			Repositories: map[string][]config.Repository{
				"github": {{Name: "owner/github-repo"}},
				"gitlab": {{Name: "owner/gitlab-repo"}},
				"bitbucket": {{Name: "owner/bitbucket-repo"}},
			},
		}

		factory := NewFactory(cfg)
		providers, err := factory.CreateProviders()

		require.NoError(t, err)
		assert.Len(t, providers, 3)
		assert.Contains(t, providers, "github")
		assert.Contains(t, providers, "gitlab")
		assert.Contains(t, providers, "bitbucket")
	})
}

// Benchmark tests
func BenchmarkFactory_CreateProviders(b *testing.B) {
	cfg := &config.Config{
		Auth: config.Auth{
			GitHub: config.GitHubAuth{Token: "github_token"},
			GitLab: config.GitLabAuth{
				Token: "gitlab_token",
				URL:   "https://gitlab.example.com",
			},
			Bitbucket: config.BitbucketAuth{
				Username:    "username",
				AppPassword: "password",
				Workspace:   "workspace",
			},
		},
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 5.0,
				Burst:             10,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory := NewFactory(cfg)
		_, err := factory.CreateProviders()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFactory_resolveEnvVar(b *testing.B) {
	factory := NewFactory(&config.Config{})
	os.Setenv("BENCHMARK_VAR", "benchmark_value")
	defer os.Unsetenv("BENCHMARK_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory.resolveEnvVar("$BENCHMARK_VAR")
	}
}
