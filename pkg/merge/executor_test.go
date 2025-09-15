package merge

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/pr"
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
		Repositories: map[string][]config.Repository{
			"github": {
				{
					Name:          "owner/test-repo",
					MergeStrategy: config.MergeStrategySquash,
				},
			},
		},
	}
}

func createTestPullRequest(number int, title string) common.PullRequest {
	return common.PullRequest{
		ID:         fmt.Sprintf("pr-%d", number),
		Number:     number,
		Title:      title,
		Body:       fmt.Sprintf("Test PR body for %d", number),
		HeadSHA:    fmt.Sprintf("sha%d", number),
		HeadBranch: fmt.Sprintf("feature/test-%d", number),
		Author: common.User{
			Login: "test-user",
			ID:    "user-123",
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}
}

func createTestRepository() common.Repository {
	return common.Repository{
		ID:       "repo-123",
		Name:     "test-repo",
		FullName: "owner/test-repo",
		Provider: "github",
		Owner: common.User{
			Login: "owner",
			ID:    "owner-123",
		},
	}
}

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name      string
		providers map[string]common.Provider
		config    *config.Config
		wantErr   bool
	}{
		{
			name: "creates executor with valid configuration",
			providers: map[string]common.Provider{
				"github": &MockProvider{},
			},
			config:  createTestConfig(),
			wantErr: false,
		},
		{
			name:      "creates executor with empty providers",
			providers: map[string]common.Provider{},
			config:    createTestConfig(),
			wantErr:   false,
		},
		{
			name: "creates executor with nil config",
			providers: map[string]common.Provider{
				"github": &MockProvider{},
			},
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.providers, tt.config)

			if tt.wantErr {
				assert.Nil(t, executor)
			} else {
				assert.NotNil(t, executor)
				assert.Equal(t, tt.providers, executor.providers)
				assert.Equal(t, tt.config, executor.config)
				assert.NotNil(t, executor.logger)
			}
		})
	}
}

func TestMergeExecutor_MergePRs_SuccessfulMerge(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")
	mockProvider.On("MergePullRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       true,
					Status: pr.PRStatus{
						Ready: true,
					},
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: false})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 1)

	result := mergeResults[0]
	assert.Equal(t, "github", result.Provider)
	assert.Equal(t, "owner/test-repo", result.Repository)
	assert.Equal(t, 123, result.PullRequest)
	assert.Equal(t, "Feature: Add new functionality", result.Title)
	assert.Equal(t, "test-user", result.Author)
	assert.Equal(t, "squash", result.MergeMethod)
	assert.True(t, result.Success)
	assert.False(t, result.Skipped)
	assert.NoError(t, result.Error)

	mockProvider.AssertExpectations(t)
}

func TestMergeExecutor_MergePRs_DryRun(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       true,
					Status: pr.PRStatus{
						Ready: true,
					},
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: true})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 1)

	result := mergeResults[0]
	assert.True(t, result.Success)
	assert.False(t, result.Skipped)
	assert.Equal(t, "dry run - would merge", result.Reason)
	assert.NoError(t, result.Error)

	// No merge call should be made in dry run
	mockProvider.AssertNotCalled(t, "MergePullRequest")
}

func TestMergeExecutor_MergePRs_MergeFailure(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")
	mockProvider.On("MergePullRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("merge conflict"))

	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       true,
					Status: pr.PRStatus{
						Ready: true,
					},
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: false})

	assert.NoError(t, err) // Individual failures don't cause overall error
	assert.Len(t, mergeResults, 1)

	result := mergeResults[0]
	assert.False(t, result.Success)
	assert.False(t, result.Skipped)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "merge conflict")

	mockProvider.AssertExpectations(t)
}

func TestMergeExecutor_MergePRs_SkippedPRs(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	pr2 := createTestPullRequest(124, "Feature: Another feature")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       false,
					Skipped:     true,
					Reason:      "not ready",
				},
				{
					PullRequest: pr2,
					Ready:       false,
					Reason:      "missing approvals",
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: false})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 2)

	// Both PRs should be skipped
	for _, result := range mergeResults {
		assert.True(t, result.Skipped)
		assert.False(t, result.Success)
	}

	// No merge calls should be made
	mockProvider.AssertNotCalled(t, "MergePullRequest")
}

func TestMergeExecutor_MergePRs_ForceMode(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")
	mockProvider.On("MergePullRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       false,
					Reason:      "missing approvals",
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{
		DryRun: false,
		Force:  true,
	})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 1)

	result := mergeResults[0]
	assert.True(t, result.Success)
	assert.False(t, result.Skipped)

	mockProvider.AssertExpectations(t)
}

func TestMergeExecutor_MergePRs_ConcurrentMerging(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	// Set up expectations for concurrent calls
	for i := 0; i < 5; i++ {
		mockProvider.On("MergePullRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()
	}

	config := createTestConfig()
	config.Behavior.Concurrency = 3
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	repo := createTestRepository()
	var processedPRs []pr.ProcessedPR

	// Create 5 ready PRs
	for i := 1; i <= 5; i++ {
		pullRequest := createTestPullRequest(120+i, fmt.Sprintf("Feature: Test %d", i))
		processedPRs = append(processedPRs, pr.ProcessedPR{
			PullRequest: pullRequest,
			Ready:       true,
			Status: pr.PRStatus{
				Ready: true,
			},
		})
	}

	results := []pr.ProcessResult{
		{
			Provider:     "github",
			Repository:   repo,
			PullRequests: processedPRs,
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: false})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 5)

	// Verify all PRs were processed
	for _, result := range mergeResults {
		assert.True(t, result.Success)
		assert.False(t, result.Skipped)
	}

	mockProvider.AssertExpectations(t)
}

func TestMergeExecutor_MergePRs_ProviderNotFound(t *testing.T) {
	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       true,
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: false})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 0) // No results since provider not found
}

func TestMergeExecutor_MergePRs_RepositoryConfigNotFound(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	// Change repository name so config won't be found
	config.Repositories["github"][0].Name = "different/repo"

	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	pr1 := createTestPullRequest(123, "Feature: Add new functionality")
	repo := createTestRepository()

	results := []pr.ProcessResult{
		{
			Provider:   "github",
			Repository: repo,
			PullRequests: []pr.ProcessedPR{
				{
					PullRequest: pr1,
					Ready:       true,
				},
			},
		},
	}

	mergeResults, err := executor.MergePRs(context.Background(), results, MergeOptions{DryRun: false})

	assert.NoError(t, err)
	assert.Len(t, mergeResults, 1)

	result := mergeResults[0]
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "repository configuration not found")
}

func TestDetermineMergeMethod(t *testing.T) {
	executor := NewExecutor(map[string]common.Provider{}, createTestConfig())

	tests := []struct {
		name           string
		repoConfig     config.Repository
		expectedMethod common.MergeMethod
	}{
		{
			name: "merge strategy",
			repoConfig: config.Repository{
				MergeStrategy: config.MergeStrategyMerge,
			},
			expectedMethod: common.MergeMethodMerge,
		},
		{
			name: "squash strategy",
			repoConfig: config.Repository{
				MergeStrategy: config.MergeStrategySquash,
			},
			expectedMethod: common.MergeMethodSquash,
		},
		{
			name: "rebase strategy",
			repoConfig: config.Repository{
				MergeStrategy: config.MergeStrategyRebase,
			},
			expectedMethod: common.MergeMethodRebase,
		},
		{
			name: "empty strategy defaults to squash",
			repoConfig: config.Repository{
				MergeStrategy: "",
			},
			expectedMethod: common.MergeMethodSquash,
		},
		{
			name: "invalid strategy defaults to squash",
			repoConfig: config.Repository{
				MergeStrategy: "invalid",
			},
			expectedMethod: common.MergeMethodSquash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := executor.determineMergeMethod(tt.repoConfig)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	executor := NewExecutor(map[string]common.Provider{}, createTestConfig())

	pr := createTestPullRequest(123, "Feature: Add cool stuff")

	tests := []struct {
		name          string
		method        common.MergeMethod
		customMessage string
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "squash merge with PR number",
			method:        common.MergeMethodSquash,
			customMessage: "",
			expectedTitle: "Feature: Add cool stuff (#123)",
			expectedBody:  "Test PR body for 123",
		},
		{
			name:          "merge commit",
			method:        common.MergeMethodMerge,
			customMessage: "",
			expectedTitle: "Merge pull request #123 from feature/test-123",
			expectedBody:  "Feature: Add cool stuff",
		},
		{
			name:          "rebase merge",
			method:        common.MergeMethodRebase,
			customMessage: "",
			expectedTitle: "Feature: Add cool stuff",
			expectedBody:  "",
		},
		{
			name:          "custom message",
			method:        common.MergeMethodSquash,
			customMessage: "Custom commit message",
			expectedTitle: "Custom commit message",
			expectedBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body := executor.generateCommitMessage(pr, tt.method, tt.customMessage)
			assert.Equal(t, tt.expectedTitle, title)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestValidateMergeability(t *testing.T) {
	mockProvider := &MockProvider{}
	mockProvider.On("GetProviderName").Return("github")

	config := createTestConfig()
	executor := NewExecutor(map[string]common.Provider{
		"github": mockProvider,
	}, config)

	tests := []struct {
		name    string
		results []pr.ProcessResult
		wantErr bool
	}{
		{
			name: "valid results",
			results: []pr.ProcessResult{
				{
					Provider:   "github",
					Repository: createTestRepository(),
					PullRequests: []pr.ProcessedPR{
						{
							PullRequest: createTestPullRequest(123, "Test"),
							Ready:       true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "results with errors",
			results: []pr.ProcessResult{
				{
					Provider: "github",
					Error:    errors.New("API error"),
				},
			},
			wantErr: false, // Errors in results are ignored during validation
		},
		{
			name: "provider not available",
			results: []pr.ProcessResult{
				{
					Provider:   "gitlab",
					Repository: createTestRepository(),
					PullRequests: []pr.ProcessedPR{
						{
							PullRequest: createTestPullRequest(123, "Test"),
							Ready:       true,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ValidateMergeability(context.Background(), tt.results)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeOptions(t *testing.T) {
	opts := MergeOptions{
		DryRun:          true,
		Force:           true,
		DeleteBranches:  true,
		CustomMessage:   "Custom message",
		RequireApproval: true,
	}

	assert.True(t, opts.DryRun)
	assert.True(t, opts.Force)
	assert.True(t, opts.DeleteBranches)
	assert.Equal(t, "Custom message", opts.CustomMessage)
	assert.True(t, opts.RequireApproval)
}

func TestMergeResult(t *testing.T) {
	now := time.Now()
	result := MergeResult{
		Provider:    "github",
		Repository:  "owner/repo",
		PullRequest: 123,
		Title:       "Test PR",
		Author:      "test-user",
		MergeMethod: "squash",
		MergedAt:    now,
		CommitSHA:   "abc123",
		Success:     true,
		Error:       nil,
		Skipped:     false,
		Reason:      "successfully merged",
	}

	assert.Equal(t, "github", result.Provider)
	assert.Equal(t, "owner/repo", result.Repository)
	assert.Equal(t, 123, result.PullRequest)
	assert.Equal(t, "Test PR", result.Title)
	assert.Equal(t, "test-user", result.Author)
	assert.Equal(t, "squash", result.MergeMethod)
	assert.Equal(t, now, result.MergedAt)
	assert.Equal(t, "abc123", result.CommitSHA)
	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.False(t, result.Skipped)
	assert.Equal(t, "successfully merged", result.Reason)
}
