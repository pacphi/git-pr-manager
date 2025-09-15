package gitlab

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
			name: "valid config with token",
			config: Config{
				Token:     "gitlab_token",
				BaseURL:   "https://gitlab.example.com",
				RateLimit: 5.0,
				RateBurst: 10,
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
			name: "valid config without custom BaseURL",
			config: Config{
				Token:     "gitlab_token",
				RateLimit: 5.0,
				RateBurst: 10,
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
			name: "empty token",
			config: Config{
				Token:     "",
				BaseURL:   "https://gitlab.example.com",
				RateLimit: 5.0,
				RateBurst: 10,
			},
			expectError: true,
			errorMsg:    "GitLab token is required",
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
				assert.Equal(t, tt.config.Token, provider.token)
				assert.NotNil(t, provider.client)
				assert.NotNil(t, provider.rateLimiter)
				assert.NotNil(t, provider.logger)
			}
		})
	}
}

func TestProvider_GetProviderName(t *testing.T) {
	config := Config{
		Token:     "test_token",
		RateLimit: 5.0,
		RateBurst: 10,
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
	assert.Equal(t, "gitlab", provider.GetProviderName())
}

func TestProvider_ListPullRequests(t *testing.T) {
	// Create a mock GitLab server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/merge_requests")

		// Mock response for merge requests
		response := `[
			{
				"id": 123,
				"iid": 1,
				"title": "Test MR",
				"description": "Test description",
				"state": "opened",
				"source_branch": "feature-branch",
				"target_branch": "main",
				"author": {
					"id": 1,
					"username": "test_user",
					"name": "Test User"
				},
				"created_at": "2024-01-15T10:00:00Z",
				"updated_at": "2024-01-15T11:00:00Z",
				"web_url": "https://gitlab.example.com/owner/repo/-/merge_requests/1",
				"head_pipeline": {
					"id": 456,
					"status": "success"
				},
				"sha": "abc123def456",
				"labels": ["bug", "urgent"],
				"draft": false,
				"work_in_progress": false
			}
		]`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{
		FullName: "owner/repo",
	}

	opts := common.ListPROptions{
		State: "opened",
	}

	prs, err := provider.ListPullRequests(context.Background(), repo, opts)
	assert.NoError(t, err)
	assert.Len(t, prs, 1)

	pr := prs[0]
	assert.Equal(t, 1, pr.Number)
	assert.Equal(t, "Test MR", pr.Title)
	assert.Equal(t, "Test description", pr.Body)
	assert.Equal(t, "opened", pr.State)
	assert.Equal(t, "feature-branch", pr.HeadBranch)
	assert.Equal(t, "main", pr.BaseBranch)
	assert.Equal(t, "test_user", pr.Author.Login)
	assert.Equal(t, "Test User", pr.Author.Name)
	assert.Equal(t, "abc123def456", pr.HeadSHA)
	assert.Contains(t, pr.Labels, "bug")
	assert.Contains(t, pr.Labels, "urgent")
	assert.False(t, pr.Draft)
	assert.NotNil(t, pr.CreatedAt)
	assert.NotNil(t, pr.UpdatedAt)
}

func TestProvider_ListPullRequests_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{FullName: "owner/repo"}
	opts := common.ListPROptions{State: "opened"}

	prs, err := provider.ListPullRequests(context.Background(), repo, opts)
	assert.NoError(t, err)
	assert.Len(t, prs, 0)
}

func TestProvider_ListPullRequests_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer server.Close()

	config := Config{
		Token:     "invalid_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{FullName: "owner/repo"}
	opts := common.ListPROptions{State: "opened"}

	prs, err := provider.ListPullRequests(context.Background(), repo, opts)
	assert.Error(t, err)
	assert.Nil(t, prs)
}

func TestProvider_GetPullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/merge_requests/1")

		response := `{
			"id": 123,
			"iid": 1,
			"title": "Test MR",
			"description": "Test description",
			"state": "opened",
			"source_branch": "feature-branch",
			"target_branch": "main",
			"author": {
				"id": 1,
				"username": "test_user",
				"name": "Test User"
			},
			"created_at": "2024-01-15T10:00:00Z",
			"updated_at": "2024-01-15T11:00:00Z",
			"web_url": "https://gitlab.example.com/owner/repo/-/merge_requests/1",
			"sha": "abc123def456",
			"labels": ["feature"],
			"draft": false
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{FullName: "owner/repo"}

	pr, err := provider.GetPullRequest(context.Background(), repo, 1)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 1, pr.Number)
	assert.Equal(t, "Test MR", pr.Title)
	assert.Equal(t, "test_user", pr.Author.Login)
}

func TestProvider_GetPullRequest_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not Found"}`))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{FullName: "owner/repo"}

	pr, err := provider.GetPullRequest(context.Background(), repo, 999)
	assert.Error(t, err)
	assert.Nil(t, pr)
}

func TestProvider_MergePullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/merge_requests/1/merge")

		response := `{
			"id": 123,
			"iid": 1,
			"state": "merged",
			"merge_commit_sha": "def456abc789",
			"merged_at": "2024-01-15T12:00:00Z"
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{FullName: "owner/repo"}
	pr := common.PullRequest{Number: 1}
	mergeOpts := common.MergeOptions{
		Method:        common.MergeMethodMerge,
		CommitTitle:   "Test merge",
		CommitMessage: "Merging test MR",
	}

	err = provider.MergePullRequest(context.Background(), repo, pr, mergeOpts)
	assert.NoError(t, err)
}

func TestProvider_MergePullRequest_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Merge conflicts"}`))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repo := common.Repository{FullName: "owner/repo"}
	pr := common.PullRequest{Number: 1}
	mergeOpts := common.MergeOptions{
		Method:        common.MergeMethodMerge,
		CommitTitle:   "Test merge",
		CommitMessage: "Merging test MR",
	}

	err = provider.MergePullRequest(context.Background(), repo, pr, mergeOpts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merge conflicts")
}

func TestProvider_ListRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/projects")

		response := `[
			{
				"id": 123,
				"name": "test-repo",
				"path_with_namespace": "owner/test-repo",
				"description": "Test repository",
				"default_branch": "main",
				"visibility": "private",
				"web_url": "https://gitlab.example.com/owner/test-repo",
				"created_at": "2024-01-15T10:00:00Z",
				"last_activity_at": "2024-01-15T11:00:00Z"
			},
			{
				"id": 456,
				"name": "another-repo",
				"path_with_namespace": "owner/another-repo",
				"description": "Another repository",
				"default_branch": "master",
				"visibility": "public",
				"web_url": "https://gitlab.example.com/owner/another-repo",
				"created_at": "2024-01-10T10:00:00Z",
				"last_activity_at": "2024-01-14T11:00:00Z"
			}
		]`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	repos, err := provider.ListRepositories(context.Background())
	assert.NoError(t, err)
	assert.Len(t, repos, 2)

	repo1 := repos[0]
	assert.Equal(t, "test-repo", repo1.Name)
	assert.Equal(t, "owner/test-repo", repo1.FullName)
	assert.Equal(t, "Test repository", repo1.Description)
	assert.Equal(t, "main", repo1.DefaultBranch)
	assert.True(t, repo1.IsPrivate)

	repo2 := repos[1]
	assert.Equal(t, "another-repo", repo2.Name)
	assert.Equal(t, "owner/another-repo", repo2.FullName)
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
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 1.0, // 1 request per second
		RateBurst: 1,   // Burst of 1
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

	repo := common.Repository{FullName: "owner/repo"}

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

func TestProvider_InvalidRepository(t *testing.T) {
	config := Config{
		Token:     "test_token",
		RateLimit: 5.0,
		RateBurst: 10,
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
	opts := common.ListPROptions{State: "opened"}

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
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 5.0,
		RateBurst: 10,
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

	// Create a context that will timeout quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	repo := common.Repository{FullName: "owner/repo"}

	_, err = provider.GetPullRequest(ctx, repo, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestProvider_EdgeCases(t *testing.T) {
	config := Config{
		Token:     "test_token",
		RateLimit: 5.0,
		RateBurst: 10,
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
		repo := common.Repository{FullName: "owner/repo"}
		_, err := provider.GetPullRequest(context.Background(), repo, 0)
		assert.Error(t, err)
	})

	t.Run("negative PR number", func(t *testing.T) {
		repo := common.Repository{FullName: "owner/repo"}
		_, err := provider.GetPullRequest(context.Background(), repo, -1)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkProvider_ListPullRequests(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": 123,
				"iid": 1,
				"title": "Test MR",
				"state": "opened",
				"author": {"username": "test_user"},
				"created_at": "2024-01-15T10:00:00Z"
			}
		]`))
	}))
	defer server.Close()

	config := Config{
		Token:     "test_token",
		BaseURL:   server.URL,
		RateLimit: 1000.0, // High rate limit for benchmarking
		RateBurst: 100,
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

	repo := common.Repository{FullName: "owner/repo"}
	opts := common.ListPROptions{State: "opened"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.ListPullRequests(context.Background(), repo, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
