package pr

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
)

// MockProvider implements common.Provider for testing
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Authenticate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) GetProviderName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) ListRepositories(ctx context.Context) ([]common.Repository, error) {
	args := m.Called(ctx)
	return args.Get(0).([]common.Repository), args.Error(1)
}

func (m *MockProvider) GetRepository(ctx context.Context, owner, name string) (*common.Repository, error) {
	args := m.Called(ctx, owner, name)
	return args.Get(0).(*common.Repository), args.Error(1)
}

func (m *MockProvider) ListPullRequests(ctx context.Context, repo common.Repository, opts common.ListPROptions) ([]common.PullRequest, error) {
	args := m.Called(ctx, repo, opts)
	return args.Get(0).([]common.PullRequest), args.Error(1)
}

func (m *MockProvider) GetPullRequest(ctx context.Context, repo common.Repository, number int) (*common.PullRequest, error) {
	args := m.Called(ctx, repo, number)
	return args.Get(0).(*common.PullRequest), args.Error(1)
}

func (m *MockProvider) MergePullRequest(ctx context.Context, repo common.Repository, pr common.PullRequest, opts common.MergeOptions) error {
	args := m.Called(ctx, repo, pr, opts)
	return args.Error(0)
}

func (m *MockProvider) GetPRStatus(ctx context.Context, repo common.Repository, pr common.PullRequest) (*common.PRStatus, error) {
	args := m.Called(ctx, repo, pr)
	return args.Get(0).(*common.PRStatus), args.Error(1)
}

func (m *MockProvider) GetChecks(ctx context.Context, repo common.Repository, pr common.PullRequest) ([]common.Check, error) {
	args := m.Called(ctx, repo, pr)
	return args.Get(0).([]common.Check), args.Error(1)
}

func (m *MockProvider) GetRateLimit(ctx context.Context) (*common.RateLimit, error) {
	args := m.Called(ctx)
	return args.Get(0).(*common.RateLimit), args.Error(1)
}

// Test helper functions
func createTestConfig() *config.Config {
	return &config.Config{
		Behavior: config.Behavior{
			Concurrency: 3,
		},
		PRFilters: config.PRFilters{
			AllowedActors: []string{"test-user", "admin"},
			SkipLabels:    []string{"skip", "do-not-merge"},
			MaxAge:        "30d",
		},
		Repositories: map[string][]config.Repository{
			"github": {
				{
					Name:          "owner/test-repo",
					SkipLabels:    []string{"wip"},
					RequireChecks: true,
				},
				{
					Name:          "owner/another-repo",
					RequireChecks: false,
				},
			},
		},
	}
}

func createTestRepository(name string) common.Repository {
	return common.Repository{
		ID:       "repo-123",
		Name:     "test-repo",
		FullName: name,
		Provider: "github",
		Owner: common.User{
			Login: "owner",
			ID:    "owner-123",
		},
		DefaultBranch: "main",
		CreatedAt:     time.Now().Add(-365 * 24 * time.Hour),
		UpdatedAt:     time.Now().Add(-1 * time.Hour),
	}
}

func createTestPullRequest(number int, title string, state common.PRState) common.PullRequest {
	mergeable := true
	return common.PullRequest{
		ID:         fmt.Sprintf("pr-%d", number),
		Number:     number,
		Title:      title,
		Body:       fmt.Sprintf("Test PR body for %d", number),
		State:      state,
		HeadSHA:    fmt.Sprintf("sha%d", number),
		HeadBranch: fmt.Sprintf("feature/test-%d", number),
		BaseBranch: "main",
		Author: common.User{
			Login: "test-user",
			ID:    "user-123",
		},
		Mergeable: &mergeable,
		Draft:     false,
		Locked:    false,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
		Labels:    []common.Label{},
	}
}

func createTestPRStatus(state common.PRStatusState) *common.PRStatus {
	return &common.PRStatus{
		State:       state,
		Description: "All checks passed",
		Context:     "continuous-integration",
		TargetURL:   "https://example.com/build/123",
		CreatedAt:   time.Now().Add(-30 * time.Minute),
		UpdatedAt:   time.Now().Add(-5 * time.Minute),
	}
}

func createTestChecks(passing bool) []common.Check {
	var conclusion string
	if passing {
		conclusion = "success"
	} else {
		conclusion = "failure"
	}

	now := time.Now()
	startedAt := now.Add(-30 * time.Minute)
	completedAt := now.Add(-5 * time.Minute)

	return []common.Check{
		{
			ID:          "check-1",
			Name:        "CI",
			Status:      common.CheckStatusCompleted,
			Conclusion:  conclusion,
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		},
		{
			ID:          "check-2",
			Name:        "Tests",
			Status:      common.CheckStatusCompleted,
			Conclusion:  conclusion,
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		},
	}
}

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name      string
		providers map[string]common.Provider
		config    *config.Config
	}{
		{
			name: "creates processor with valid configuration",
			providers: map[string]common.Provider{
				"github": &MockProvider{},
			},
			config: createTestConfig(),
		},
		{
			name:      "creates processor with empty providers",
			providers: map[string]common.Provider{},
			config:    createTestConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(tt.providers, tt.config)

			assert.NotNil(t, processor)
			assert.Equal(t, tt.providers, processor.providers)
			assert.Equal(t, tt.config, processor.config)
			assert.NotNil(t, processor.logger)
			assert.NotNil(t, processor.executor)
		})
	}
}

func TestProcessor_ProcessAllPRs_Success(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	repo1 := createTestRepository("owner/test-repo")
	repo2 := createTestRepository("owner/another-repo")

	mockProvider.On("GetRepository", mock.Anything, "owner", "test-repo").Return(&repo1, nil)
	mockProvider.On("GetRepository", mock.Anything, "owner", "another-repo").Return(&repo2, nil)

	prs := []common.PullRequest{
		createTestPullRequest(123, "Feature: Add functionality", common.PRStateOpen),
		createTestPullRequest(124, "Fix: Bug fix", common.PRStateOpen),
	}
	mockProvider.On("ListPullRequests", mock.Anything, repo1, mock.Anything).Return(prs, nil)
	mockProvider.On("ListPullRequests", mock.Anything, repo2, mock.Anything).Return([]common.PullRequest{}, nil)

	status := createTestPRStatus(common.PRStatusSuccess)
	checks := createTestChecks(true)

	for _, pr := range prs {
		mockProvider.On("GetPRStatus", mock.Anything, repo1, pr).Return(status, nil)
		mockProvider.On("GetChecks", mock.Anything, repo1, pr).Return(checks, nil)
	}

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	opts := ProcessOptions{
		Providers: []string{"github"},
	}

	results, err := processor.ProcessAllPRs(context.Background(), opts)

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// First repository should have 2 PRs
	result1 := results[0]
	assert.Equal(t, "github", result1.Provider)
	assert.Equal(t, "owner/test-repo", result1.Repository.FullName)
	assert.Len(t, result1.PullRequests, 2)
	assert.NoError(t, result1.Error)

	// Both PRs should be ready since they pass all checks
	for _, processedPR := range result1.PullRequests {
		assert.True(t, processedPR.Ready)
		assert.False(t, processedPR.Skipped)
		assert.NoError(t, processedPR.Error)
	}

	// Second repository should have 0 PRs
	result2 := results[1]
	assert.Equal(t, "github", result2.Provider)
	assert.Equal(t, "owner/another-repo", result2.Repository.FullName)
	assert.Len(t, result2.PullRequests, 0)
	assert.NoError(t, result2.Error)

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessAllPRs_ProviderFilter(t *testing.T) {
	mockGitHub := &MockProvider{}
	mockGitLab := &MockProvider{}

	mockGitHub.On("GetProviderName").Return("github").Maybe()
	mockGitLab.On("GetProviderName").Return("gitlab").Maybe()

	// Only GitHub should be called (but it has 2 repositories in the config)
	repo1 := createTestRepository("owner/test-repo")
	repo2 := createTestRepository("owner/another-repo")
	mockGitHub.On("GetRepository", mock.Anything, "owner", "test-repo").Return(&repo1, nil)
	mockGitHub.On("GetRepository", mock.Anything, "owner", "another-repo").Return(&repo2, nil)
	mockGitHub.On("ListPullRequests", mock.Anything, repo1, mock.Anything).Return([]common.PullRequest{}, nil)
	mockGitHub.On("ListPullRequests", mock.Anything, repo2, mock.Anything).Return([]common.PullRequest{}, nil)

	cfg := createTestConfig()
	cfg.Repositories["gitlab"] = []config.Repository{
		{Name: "owner/gitlab-repo"},
	}

	processor := NewProcessor(map[string]common.Provider{
		"github": mockGitHub,
		"gitlab": mockGitLab,
	}, cfg)

	opts := ProcessOptions{
		Providers: []string{"github"},
	}

	results, err := processor.ProcessAllPRs(context.Background(), opts)

	assert.NoError(t, err)
	assert.Len(t, results, 2) // 2 GitHub repositories processed
	assert.Equal(t, "github", results[0].Provider)
	assert.Equal(t, "github", results[1].Provider)

	mockGitHub.AssertExpectations(t)
	mockGitLab.AssertNotCalled(t, "GetRepository")
}

func TestProcessor_ProcessAllPRs_RepositoryFilter(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github").Maybe()

	// Only test-repo should be processed
	repo := createTestRepository("owner/test-repo")
	mockProvider.On("GetRepository", mock.Anything, "owner", "test-repo").Return(&repo, nil)
	mockProvider.On("ListPullRequests", mock.Anything, repo, mock.Anything).Return([]common.PullRequest{}, nil)

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	opts := ProcessOptions{
		Repositories: []string{"test-repo"},
	}

	results, err := processor.ProcessAllPRs(context.Background(), opts)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "owner/test-repo", results[0].Repository.FullName)

	mockProvider.AssertExpectations(t)
	mockProvider.AssertNotCalled(t, "GetRepository", mock.Anything, "owner", "another-repo")
}

func TestProcessor_ProcessAllPRs_NoRepositories(t *testing.T) {
	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{}, config)

	opts := ProcessOptions{
		Providers: []string{"nonexistent"},
	}

	results, err := processor.ProcessAllPRs(context.Background(), opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no repositories to process")
	assert.Nil(t, results)
}

func TestProcessor_ProcessPR_AuthorNotAllowed(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	pr.Author.Login = "unauthorized-user"

	repo := createTestRepository("owner/test-repo")
	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{})

	assert.True(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Contains(t, processed.Reason, "not in allowed actors")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_SkipLabels(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{}, config)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	pr.Labels = []common.Label{
		{Name: "skip", Color: "red"},
		{Name: "feature", Color: "blue"},
	}

	repo := createTestRepository("owner/test-repo")
	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{})

	assert.True(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Contains(t, processed.Reason, "skip labels")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_MaxAge(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{}, config)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	pr.CreatedAt = time.Now().Add(-40 * 24 * time.Hour) // 40 days old

	repo := createTestRepository("owner/test-repo")
	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{})

	assert.True(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Contains(t, processed.Reason, "older than")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_StatusCheckFailure(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	status := createTestPRStatus(common.PRStatusFailure)
	checks := createTestChecks(false)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	repo := createTestRepository("owner/test-repo")

	mockProvider.On("GetPRStatus", mock.Anything, repo, pr).Return(status, nil)
	mockProvider.On("GetChecks", mock.Anything, repo, pr).Return(checks, nil)

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{
		RequireChecks: true,
	})

	assert.False(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Contains(t, processed.Reason, "status checks not passing")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_ReadyToMerge(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	status := createTestPRStatus(common.PRStatusSuccess)
	checks := createTestChecks(true)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	repo := createTestRepository("owner/test-repo")

	mockProvider.On("GetPRStatus", mock.Anything, repo, pr).Return(status, nil)
	mockProvider.On("GetChecks", mock.Anything, repo, pr).Return(checks, nil)

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{
		RequireChecks: true,
	})

	assert.False(t, processed.Skipped)
	assert.True(t, processed.Ready)
	assert.Equal(t, "ready to merge", processed.Reason)

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_Draft(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	status := createTestPRStatus(common.PRStatusSuccess)
	checks := createTestChecks(true)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	pr.Draft = true
	repo := createTestRepository("owner/test-repo")

	mockProvider.On("GetPRStatus", mock.Anything, repo, pr).Return(status, nil)
	mockProvider.On("GetChecks", mock.Anything, repo, pr).Return(checks, nil)

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{})

	assert.False(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Contains(t, processed.Reason, "draft")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_MergeConflicts(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	status := createTestPRStatus(common.PRStatusSuccess)
	checks := createTestChecks(true)

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	mergeable := false
	pr.Mergeable = &mergeable
	repo := createTestRepository("owner/test-repo")

	mockProvider.On("GetPRStatus", mock.Anything, repo, pr).Return(status, nil)
	mockProvider.On("GetChecks", mock.Anything, repo, pr).Return(checks, nil)

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{})

	assert.False(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Contains(t, processed.Reason, "merge conflicts")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessPR_ErrorGettingStatus(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	repo := createTestRepository("owner/test-repo")

	mockProvider.On("GetPRStatus", mock.Anything, repo, pr).Return((*common.PRStatus)(nil), errors.New("API error"))

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	processed := processor.processPR(context.Background(), mockProvider, repo, pr, repoConfig, ProcessOptions{})

	assert.False(t, processed.Skipped)
	assert.False(t, processed.Ready)
	assert.Error(t, processed.Error)
	assert.Contains(t, processed.Error.Error(), "failed to get PR status")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessRepository_InvalidName(t *testing.T) {
	cfg := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{}, cfg)

	repoConfig := config.Repository{
		Name: "invalid-repo-name", // Missing owner/repo format
	}

	result, err := processor.processRepository(context.Background(), &MockProvider{}, "github", repoConfig, ProcessOptions{})

	assert.NoError(t, err) // processRepository doesn't return errors, puts them in result
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "invalid repository name")
}

func TestProcessor_ProcessRepository_GetRepositoryError(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github").Maybe()
	mockProvider.On("GetRepository", mock.Anything, "owner", "test-repo").Return((*common.Repository)(nil), errors.New("not found"))

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	result, err := processor.processRepository(context.Background(), mockProvider, "github", repoConfig, ProcessOptions{})

	assert.NoError(t, err)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to get repository")

	mockProvider.AssertExpectations(t)
}

func TestProcessor_ProcessRepository_ListPRsError(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github").Maybe()

	repo := createTestRepository("owner/test-repo")
	mockProvider.On("GetRepository", mock.Anything, "owner", "test-repo").Return(&repo, nil)
	mockProvider.On("ListPullRequests", mock.Anything, repo, mock.Anything).Return([]common.PullRequest(nil), errors.New("API error"))

	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repoConfig := config.Repositories["github"][0]

	result, err := processor.processRepository(context.Background(), mockProvider, "github", repoConfig, ProcessOptions{})

	assert.NoError(t, err)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to list pull requests")

	mockProvider.AssertExpectations(t)
}

func TestIsAuthorAllowed(t *testing.T) {
	config := createTestConfig()
	processor := NewProcessor(map[string]common.Provider{}, config)

	tests := []struct {
		name     string
		author   common.User
		expected bool
	}{
		{
			name:     "allowed user",
			author:   common.User{Login: "test-user"},
			expected: true,
		},
		{
			name:     "admin user",
			author:   common.User{Login: "admin"},
			expected: true,
		},
		{
			name:     "case insensitive match",
			author:   common.User{Login: "TEST-USER"},
			expected: true,
		},
		{
			name:     "not allowed user",
			author:   common.User{Login: "hacker"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.isAuthorAllowed(tt.author)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessOptions(t *testing.T) {
	opts := ProcessOptions{
		DryRun:        true,
		Providers:     []string{"github", "gitlab"},
		Repositories:  []string{"owner/repo1", "owner/repo2"},
		MaxAge:        24 * time.Hour,
		RequireChecks: true,
		SkipLabels:    []string{"wip", "draft"},
		IncludeClosed: false,
	}

	assert.True(t, opts.DryRun)
	assert.Equal(t, []string{"github", "gitlab"}, opts.Providers)
	assert.Equal(t, []string{"owner/repo1", "owner/repo2"}, opts.Repositories)
	assert.Equal(t, 24*time.Hour, opts.MaxAge)
	assert.True(t, opts.RequireChecks)
	assert.Equal(t, []string{"wip", "draft"}, opts.SkipLabels)
	assert.False(t, opts.IncludeClosed)
}

func TestProcessedPR(t *testing.T) {
	pr := createTestPullRequest(123, "Test PR", common.PRStateOpen)
	status := PRStatus{
		State:     common.PRStatusSuccess,
		Ready:     true,
		Reason:    "ready to merge",
		UpdatedAt: time.Now(),
	}

	processed := ProcessedPR{
		PullRequest: pr,
		Status:      status,
		Reason:      "ready to merge",
		Ready:       true,
		Skipped:     false,
		Error:       nil,
	}

	assert.Equal(t, pr, processed.PullRequest)
	assert.Equal(t, status, processed.Status)
	assert.Equal(t, "ready to merge", processed.Reason)
	assert.True(t, processed.Ready)
	assert.False(t, processed.Skipped)
	assert.NoError(t, processed.Error)
}

func TestContains(t *testing.T) {
	slice := []string{"github", "gitlab", "bitbucket"}

	tests := []struct {
		name     string
		item     string
		expected bool
	}{
		{
			name:     "contains item",
			item:     "github",
			expected: true,
		},
		{
			name:     "case insensitive",
			item:     "GITHUB",
			expected: true,
		},
		{
			name:     "does not contain",
			item:     "azure",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsRepoFilter(t *testing.T) {
	filters := []string{"test", "demo", "example"}

	tests := []struct {
		name     string
		repoName string
		expected bool
	}{
		{
			name:     "matches test",
			repoName: "owner/test-repo",
			expected: true,
		},
		{
			name:     "matches demo",
			repoName: "org/demo-app",
			expected: true,
		},
		{
			name:     "case insensitive",
			repoName: "owner/TEST-REPO",
			expected: true,
		},
		{
			name:     "no match",
			repoName: "owner/production-app",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsRepoFilter(filters, tt.repoName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
