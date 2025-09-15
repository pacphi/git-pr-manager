package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				Username:    "test_user",
				AppPassword: "test_password",
				Workspace:   "test_workspace",
				RateLimit:   5.0,
				RateBurst:   10,
				BehaviorConfig: &config.Config{
					Behavior: config.Behavior{
						RateLimit: config.RateLimit{
							RequestsPerSecond: 5.0,
							Burst:             10,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty username",
			config: Config{
				Username:    "",
				AppPassword: "test_password",
				Workspace:   "test_workspace",
				RateLimit:   5.0,
				RateBurst:   10,
			},
			expectError: true,
			errorMsg:    "username and app password are required",
		},
		{
			name: "empty app password",
			config: Config{
				Username:    "test_user",
				AppPassword: "",
				Workspace:   "test_workspace",
				RateLimit:   5.0,
				RateBurst:   10,
			},
			expectError: true,
			errorMsg:    "username and app password are required",
		},
		{
			name: "empty workspace",
			config: Config{
				Username:    "test_user",
				AppPassword: "test_password",
				Workspace:   "",
				RateLimit:   5.0,
				RateBurst:   10,
			},
			expectError: true,
			errorMsg:    "workspace is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.config.Username, provider.username)
				assert.Equal(t, tt.config.AppPassword, provider.appPassword)
				assert.Equal(t, tt.config.Workspace, provider.workspace)
				assert.NotNil(t, provider.httpClient)
				assert.NotNil(t, provider.rateLimiter)
				assert.NotNil(t, provider.logger)
			}
		})
	}
}

func TestProvider_GetProviderName(t *testing.T) {
	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)

	assert.Equal(t, ProviderName, provider.GetProviderName())
	assert.Equal(t, "bitbucket", provider.GetProviderName())
}

func TestProvider_ListPullRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/pullrequests")

		// Check basic auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test_user", username)
		assert.Equal(t, "test_password", password)

		response := `{
			"values": [
				{
					"id": 1,
					"title": "Test PR",
					"description": "Test description",
					"state": "OPEN",
					"source": {
						"branch": {
							"name": "feature-branch"
						},
						"commit": {
							"hash": "abc123def456"
						}
					},
					"destination": {
						"branch": {
							"name": "main"
						}
					},
					"author": {
						"username": "test_user",
						"display_name": "Test User"
					},
					"created_on": "2024-01-15T10:00:00.000000+00:00",
					"updated_on": "2024-01-15T11:00:00.000000+00:00",
					"links": {
						"html": {
							"href": "https://bitbucket.org/workspace/repo/pull-requests/1"
						}
					}
				}
			],
			"page": 1,
			"pagelen": 50,
			"size": 1
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()


	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Set the test server URL
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{
		FullName: "workspace/repo",
	}

	opts := common.ListPROptions{
		State: "open",
	}

	prs, err := provider.ListPullRequests(context.Background(), repo, opts)
	assert.NoError(t, err)
	assert.Len(t, prs, 1)

	pr := prs[0]
	assert.Equal(t, 1, pr.Number)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, "Test description", pr.Body)
	assert.Equal(t, "OPEN", pr.State)
	assert.Equal(t, "feature-branch", pr.HeadBranch)
	assert.Equal(t, "main", pr.BaseBranch)
	assert.Equal(t, "test_user", pr.Author.Login)
	assert.Equal(t, "Test User", pr.Author.Name)
	assert.Equal(t, "abc123def456", pr.HeadSHA)
	assert.NotNil(t, pr.CreatedAt)
	assert.NotNil(t, pr.UpdatedAt)
}

func TestProvider_ListPullRequests_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"values": [],
			"page": 1,
			"pagelen": 50,
			"size": 0
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}
	opts := common.ListPROptions{State: "open"}

	prs, err := provider.ListPullRequests(context.Background(), repo, opts)
	assert.NoError(t, err)
	assert.Len(t, prs, 0)
}

func TestProvider_ListPullRequests_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Authentication required"}}`))
	}))
	defer server.Close()

	config := Config{
		Username:    "invalid_user",
		AppPassword: "invalid_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}
	opts := common.ListPROptions{State: "open"}

	prs, err := provider.ListPullRequests(context.Background(), repo, opts)
	assert.Error(t, err)
	assert.Nil(t, prs)
}

func TestProvider_GetPullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/pullrequests/1")

		response := `{
			"id": 1,
			"title": "Test PR",
			"description": "Test description",
			"state": "OPEN",
			"source": {
				"branch": {
					"name": "feature-branch"
				},
				"commit": {
					"hash": "abc123def456"
				}
			},
			"destination": {
				"branch": {
					"name": "main"
				}
			},
			"author": {
				"username": "test_user",
				"display_name": "Test User"
			},
			"created_on": "2024-01-15T10:00:00.000000+00:00",
			"updated_on": "2024-01-15T11:00:00.000000+00:00",
			"links": {
				"html": {
					"href": "https://bitbucket.org/workspace/repo/pull-requests/1"
				}
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}

	pr, err := provider.GetPullRequest(context.Background(), repo, 1)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 1, pr.Number)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, "test_user", pr.Author.Login)
}

func TestProvider_GetPullRequest_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": {"message": "Pull request not found"}}`))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}

	pr, err := provider.GetPullRequest(context.Background(), repo, 999)
	assert.Error(t, err)
	assert.Nil(t, pr)
}

func TestProvider_MergePullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/pullrequests/1/merge")

		response := `{
			"id": 1,
			"state": "MERGED",
			"merge_commit": {
				"hash": "def456abc789"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}
	pr := common.PullRequest{Number: 1}
	mergeOpts := common.MergeOptions{
		Method:        common.MergeMethodMerge,
		CommitTitle:   "Test merge",
		CommitMessage: "Merging test PR",
	}

	err = provider.MergePullRequest(context.Background(), repo, pr, mergeOpts)
	assert.NoError(t, err)
}

func TestProvider_MergePullRequest_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": {"message": "There are merge conflicts"}}`))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}
	pr := common.PullRequest{Number: 1}
	mergeOpts := common.MergeOptions{
		Method:        common.MergeMethodMerge,
		CommitTitle:   "Test merge",
		CommitMessage: "Merging test PR",
	}

	err = provider.MergePullRequest(context.Background(), repo, pr, mergeOpts)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "conflict")
}

func TestProvider_ListRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/repositories")

		response := `{
			"values": [
				{
					"uuid": "{123-456-789}",
					"name": "test-repo",
					"full_name": "workspace/test-repo",
					"description": "Test repository",
					"language": "Go",
					"is_private": true,
					"created_on": "2024-01-15T10:00:00.000000+00:00",
					"updated_on": "2024-01-15T11:00:00.000000+00:00",
					"links": {
						"html": {
							"href": "https://bitbucket.org/workspace/test-repo"
						}
					}
				},
				{
					"uuid": "{789-456-123}",
					"name": "another-repo",
					"full_name": "workspace/another-repo",
					"description": "Another repository",
					"language": "Python",
					"is_private": false,
					"created_on": "2024-01-10T10:00:00.000000+00:00",
					"updated_on": "2024-01-14T11:00:00.000000+00:00",
					"links": {
						"html": {
							"href": "https://bitbucket.org/workspace/another-repo"
						}
					}
				}
			],
			"page": 1,
			"pagelen": 50,
			"size": 2
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repos, err := provider.ListRepositories(context.Background())
	assert.NoError(t, err)
	assert.Len(t, repos, 2)

	repo1 := repos[0]
	assert.Equal(t, "test-repo", repo1.Name)
	assert.Equal(t, "workspace/test-repo", repo1.FullName)
	assert.Equal(t, "Test repository", repo1.Description)
	assert.Equal(t, "Go", repo1.Language)
	assert.True(t, repo1.IsPrivate)

	repo2 := repos[1]
	assert.Equal(t, "another-repo", repo2.Name)
	assert.Equal(t, "workspace/another-repo", repo2.FullName)
	assert.False(t, repo2.IsPrivate)
}

func TestProvider_RateLimiting(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	// Configure with very strict rate limiting for testing
	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   1.0, // 1 request per second
		RateBurst:   1,   // Burst of 1
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 1.0,
					Burst:             1,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}

	// First request should go through immediately
	start := time.Now()
	_, _ = provider.GetPullRequest(context.Background(), repo, 1)
	firstDuration := time.Since(start)

	// Second request should be rate limited
	start = time.Now()
	_, _ = provider.GetPullRequest(context.Background(), repo, 2)
	secondDuration := time.Since(start)

	// The second request should take longer due to rate limiting
	assert.Greater(t, secondDuration, firstDuration)
	assert.Greater(t, secondDuration, 500*time.Millisecond) // Should be rate limited
}

func TestProvider_BasicAuthValidation(t *testing.T) {
	expectedUsername := "test_user"
	expectedPassword := "test_password"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok, "Basic auth should be present")
		assert.Equal(t, expectedUsername, username)
		assert.Equal(t, expectedPassword, password)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"values": []}`))
	}))
	defer server.Close()

	config := Config{
		Username:    expectedUsername,
		AppPassword: expectedPassword,
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}
	opts := common.ListPROptions{}

	_, err = provider.ListPullRequests(context.Background(), repo, opts)
	assert.NoError(t, err)
}

func TestProvider_InvalidRepository(t *testing.T) {
	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)

	// Test with invalid repository format
	invalidRepo := common.Repository{FullName: "invalid-repo-format"}
	opts := common.ListPROptions{}

	_, err = provider.ListPullRequests(context.Background(), invalidRepo, opts)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "invalid")
}

func TestProvider_ContextCancellation(t *testing.T) {
	// Server with a delay to test context cancellation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	provider.httpClient.SetBaseURL(server.URL)

	// Create a context that will timeout quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	repo := common.Repository{FullName: "workspace/repo"}

	_, err = provider.GetPullRequest(ctx, repo, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestProvider_EdgeCases(t *testing.T) {
	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   5.0,
		RateBurst:   10,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 5.0,
					Burst:             10,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)

	t.Run("empty repository name", func(t *testing.T) {
		repo := common.Repository{FullName: ""}
		_, err := provider.GetPullRequest(context.Background(), repo, 1)
		assert.Error(t, err)
	})

	t.Run("zero PR number", func(t *testing.T) {
		repo := common.Repository{FullName: "workspace/repo"}
		_, err := provider.GetPullRequest(context.Background(), repo, 0)
		assert.Error(t, err)
	})

	t.Run("negative PR number", func(t *testing.T) {
		repo := common.Repository{FullName: "workspace/repo"}
		_, err := provider.GetPullRequest(context.Background(), repo, -1)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkProvider_ListPullRequests(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"values": [
				{
					"id": 1,
					"title": "Test PR",
					"state": "OPEN",
					"author": {"username": "test_user"},
					"created_on": "2024-01-15T10:00:00.000000+00:00"
				}
			]
		}`))
	}))
	defer server.Close()

	config := Config{
		Username:    "test_user",
		AppPassword: "test_password",
		Workspace:   "test_workspace",
		RateLimit:   1000.0, // High rate limit for benchmarking
		RateBurst:   100,
		BehaviorConfig: &config.Config{
			Behavior: config.Behavior{
				RateLimit: config.RateLimit{
					RequestsPerSecond: 1000.0,
					Burst:             100,
				},
			},
		},
	}

	provider, err := NewProvider(config)
	require.NoError(b, err)
	provider.httpClient.SetBaseURL(server.URL)

	repo := common.Repository{FullName: "workspace/repo"}
	opts := common.ListPROptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.ListPullRequests(context.Background(), repo, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
