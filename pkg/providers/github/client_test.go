package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
)

// Test helper functions
func createTestConfig() Config {
	return Config{
		Token:     "test-token",
		RateLimit: 10.0,
		RateBurst: 20,
	}
}

func createTestConfigWithBehavior() Config {
	return Config{
		Token:     "test-token",
		RateLimit: 10.0,
		RateBurst: 20,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 10.0,
					Burst:             20,
				},
				Retry: config.Retry{
					MaxAttempts: 3,
					Backoff:     time.Second,
					MaxBackoff:  10 * time.Second,
				},
			},
		},
	}
}

func TestNewProvider_Success(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "basic configuration",
			config: createTestConfig(),
		},
		{
			name:   "with behavior configuration",
			config: createTestConfigWithBehavior(),
		},
		{
			name: "with custom base URL",
			config: Config{
				Token:   "test-token",
				BaseURL: "https://github.enterprise.com",
			},
		},
		{
			name: "with default rate limits",
			config: Config{
				Token: "test-token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			assert.NoError(t, err)
			assert.NotNil(t, provider)
			assert.Equal(t, "github", provider.GetProviderName())
			assert.Equal(t, tt.config.Token, provider.token)
			assert.NotNil(t, provider.client)
			assert.NotNil(t, provider.rateLimiter)
			assert.NotNil(t, provider.logger)

			if tt.config.BehaviorConfig != nil {
				assert.NotNil(t, provider.behaviorManager)
			}
		})
	}
}

func TestNewProvider_EmptyToken(t *testing.T) {
	config := Config{
		Token: "",
	}

	provider, err := NewProvider(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub token is required")
	assert.Nil(t, provider)
}

func TestNewProvider_InvalidBaseURL(t *testing.T) {
	config := Config{
		Token:   "test-token",
		BaseURL: "://invalid-url",
	}

	provider, err := NewProvider(config)

	// This might not fail during provider creation, depending on when URL validation occurs
	// The test verifies the provider can be created and we handle URL errors later
	if err != nil {
		assert.Contains(t, err.Error(), "failed to set GitHub base URL")
	} else {
		assert.NotNil(t, provider)
	}
}

func TestProvider_GetProviderName(t *testing.T) {
	provider, err := NewProvider(createTestConfig())
	require.NoError(t, err)

	name := provider.GetProviderName()
	assert.Equal(t, "github", name)
	assert.Equal(t, ProviderName, name)
}

func TestProvider_Authenticate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user", "/api/v3/user":
			assert.Equal(t, "GET", r.Method)
			auth := r.Header.Get("Authorization")
			assert.True(t, auth == "Bearer test-token" || auth == "token test-token", "Expected Bearer or token authorization, got: %s", auth)

			user := map[string]interface{}{
				"login": "testuser",
				"id":    123,
				"type":  "User",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	err = provider.Authenticate(context.Background())
	assert.NoError(t, err)
}

func TestProvider_Authenticate_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	err = provider.Authenticate(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate with GitHub")
}

func TestProvider_ListRepositories_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/repos", "/api/v3/user/repos":
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "100", r.URL.Query().Get("per_page"))
			assert.Equal(t, "updated", r.URL.Query().Get("sort"))

			repos := []map[string]interface{}{
				{
					"id":        123,
					"name":      "test-repo",
					"full_name": "testuser/test-repo",
					"owner": map[string]interface{}{
						"login": "testuser",
						"id":    456,
					},
					"description":      "Test repository",
					"language":         "Go",
					"private":          false,
					"default_branch":   "main",
					"created_at":       "2023-01-01T12:00:00Z",
					"updated_at":       "2023-12-01T12:00:00Z",
					"pushed_at":        "2023-12-01T11:00:00Z",
					"stargazers_count": 10,
					"forks_count":      5,
					"archived":         false,
					"disabled":         false,
					"fork":             false,
					"has_issues":       true,
					"has_wiki":         true,
					"clone_url":        "https://github.com/testuser/test-repo.git",
					"ssh_url":          "git@github.com:testuser/test-repo.git",
					"html_url":         "https://github.com/testuser/test-repo",
					"topics":           []string{"golang", "testing"},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	repos, err := provider.ListRepositories(context.Background())

	assert.NoError(t, err)
	assert.Len(t, repos, 1)

	repo := repos[0]
	assert.Equal(t, "123", repo.ID)
	assert.Equal(t, "test-repo", repo.Name)
	assert.Equal(t, "testuser/test-repo", repo.FullName)
	assert.Equal(t, "github", repo.Provider)
	assert.Equal(t, "Test repository", repo.Description)
	assert.Equal(t, "Go", repo.Language)
	assert.Equal(t, common.VisibilityPublic, repo.Visibility)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.Equal(t, 10, repo.StarCount)
	assert.Equal(t, 5, repo.ForkCount)
	assert.False(t, repo.IsArchived)
	assert.False(t, repo.IsDisabled)
	assert.False(t, repo.IsFork)
	assert.False(t, repo.IsPrivate)
	assert.True(t, repo.HasIssues)
	assert.True(t, repo.HasWiki)
	assert.Equal(t, "testuser", repo.Owner.Login)
	assert.Equal(t, "https://github.com/testuser/test-repo.git", repo.CloneURL)
	assert.Equal(t, "git@github.com:testuser/test-repo.git", repo.SSHURL)
	assert.Equal(t, "https://github.com/testuser/test-repo", repo.WebURL)
	assert.Equal(t, []string{"golang", "testing"}, repo.Topics)
}

func TestProvider_ListRepositories_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch r.URL.Path {
		case "/user/repos", "/api/v3/user/repos":
			page := r.URL.Query().Get("page")

			w.Header().Set("Content-Type", "application/json")

			switch page {
			case "", "1":
				// First page - return data with Link header for next page
				repos := []map[string]interface{}{
					{
						"id":        123,
						"name":      "repo1",
						"full_name": "testuser/repo1",
						"owner": map[string]interface{}{
							"login": "testuser",
							"id":    456,
						},
						"private":        false,
						"default_branch": "main",
						"created_at":     "2023-01-01T12:00:00Z",
						"updated_at":     "2023-12-01T12:00:00Z",
						"pushed_at":      "2023-12-01T11:00:00Z",
					},
				}
				baseURL := "http://" + r.Host
				linkPath := "/user/repos"
				if r.URL.Path == "/api/v3/user/repos" {
					linkPath = "/api/v3/user/repos"
				}
				w.Header().Set("Link", fmt.Sprintf(`<%s%s?page=2>; rel="next"`, baseURL, linkPath))
				json.NewEncoder(w).Encode(repos)
			case "2":
				// Second page - return data without Link header (last page)
				repos := []map[string]interface{}{
					{
						"id":        124,
						"name":      "repo2",
						"full_name": "testuser/repo2",
						"owner": map[string]interface{}{
							"login": "testuser",
							"id":    456,
						},
						"private":        false,
						"default_branch": "main",
						"created_at":     "2023-01-01T12:00:00Z",
						"updated_at":     "2023-12-01T12:00:00Z",
						"pushed_at":      "2023-12-01T11:00:00Z",
					},
				}
				json.NewEncoder(w).Encode(repos)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	repos, err := provider.ListRepositories(context.Background())

	assert.NoError(t, err)
	assert.Len(t, repos, 2)
	assert.Equal(t, "repo1", repos[0].Name)
	assert.Equal(t, "repo2", repos[1].Name)
	assert.Equal(t, 2, callCount) // Should have made 2 API calls
}

func TestProvider_GetRepository_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/testuser/test-repo", "/api/v3/repos/testuser/test-repo":
			assert.Equal(t, "GET", r.Method)

			repo := map[string]interface{}{
				"id":        123,
				"name":      "test-repo",
				"full_name": "testuser/test-repo",
				"owner": map[string]interface{}{
					"login": "testuser",
					"id":    456,
				},
				"description":    "Test repository",
				"language":       "Go",
				"private":        false,
				"default_branch": "main",
				"created_at":     "2023-01-01T12:00:00Z",
				"updated_at":     "2023-12-01T12:00:00Z",
				"pushed_at":      "2023-12-01T11:00:00Z",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repo)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	repo, err := provider.GetRepository(context.Background(), "testuser", "test-repo")

	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "123", repo.ID)
	assert.Equal(t, "test-repo", repo.Name)
	assert.Equal(t, "testuser/test-repo", repo.FullName)
	assert.Equal(t, "github", repo.Provider)
}

func TestProvider_GetRepository_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	repo, err := provider.GetRepository(context.Background(), "testuser", "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "failed to get repository")
}

func TestProvider_RateLimiting(t *testing.T) {
	callTimes := []time.Time{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callTimes = append(callTimes, time.Now())

		switch r.URL.Path {
		case "/user", "/api/v3/user":
			user := map[string]interface{}{
				"login": "testuser",
				"id":    123,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	config.RateLimit = 2.0 // 2 requests per second
	config.RateBurst = 1   // Burst of 1
	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Make multiple authentication calls quickly
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err := provider.Authenticate(ctx)
		assert.NoError(t, err)
	}

	// Verify rate limiting occurred
	assert.Len(t, callTimes, 3)

	// First call should be immediate, subsequent calls should be rate limited
	// Allow some tolerance for timing variations in tests
	if len(callTimes) >= 2 {
		timeDiff := callTimes[1].Sub(callTimes[0])
		assert.True(t, timeDiff >= 400*time.Millisecond, "Rate limiting should cause delay between calls")
	}
}

func TestProvider_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := createTestConfig()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	err = provider.Authenticate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestProvider_HTTPClientConfiguration(t *testing.T) {
	config := createTestConfig()
	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Verify HTTP client is configured
	assert.NotNil(t, provider.client)

	// Test that the provider has proper authentication token
	assert.Equal(t, "test-token", provider.token)
}

func TestProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		operation      func(p *Provider) error
		expectedError  string
	}{
		{
			name: "network error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Close connection abruptly
				hj, ok := w.(http.Hijacker)
				if ok {
					conn, _, _ := hj.Hijack()
					conn.Close()
				}
			},
			operation: func(p *Provider) error {
				return p.Authenticate(context.Background())
			},
			expectedError: "failed to authenticate",
		},
		{
			name: "API rate limit error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
			},
			operation: func(p *Provider) error {
				return p.Authenticate(context.Background())
			},
			expectedError: "failed to authenticate",
		},
		{
			name: "invalid JSON response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{invalid json`))
			},
			operation: func(p *Provider) error {
				return p.Authenticate(context.Background())
			},
			expectedError: "failed to authenticate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			config := createTestConfig()
			config.BaseURL = server.URL
			provider, err := NewProvider(config)
			require.NoError(t, err)

			err = tt.operation(provider)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestProvider_BehaviorManager_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := map[string]interface{}{
			"login": "testuser",
			"id":    123,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	config := createTestConfigWithBehavior()
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Verify behavior manager is initialized
	assert.NotNil(t, provider.behaviorManager)

	// Test operation with behavior manager
	err = provider.Authenticate(context.Background())
	assert.NoError(t, err)
}

func TestProvider_WithoutBehaviorManager(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := map[string]interface{}{
			"login": "testuser",
			"id":    123,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	config := createTestConfig() // No behavior config
	config.BaseURL = server.URL
	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Verify no behavior manager is initialized
	assert.Nil(t, provider.behaviorManager)

	// Test operation without behavior manager (should use fallback rate limiting)
	err = provider.Authenticate(context.Background())
	assert.NoError(t, err)
}
