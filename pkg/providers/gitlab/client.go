package gitlab

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/time/rate"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

const ProviderName = "gitlab"

// Provider implements the common.Provider interface for GitLab
type Provider struct {
	client          *gitlab.Client
	token           string
	rateLimiter     *rate.Limiter
	behaviorManager *utils.BehaviorManager
	logger          *utils.Logger
}

// Config contains GitLab provider configuration
type Config struct {
	Token          string
	BaseURL        string
	RateLimit      float64
	RateBurst      int
	BehaviorConfig *config.Config
}

// NewProvider creates a new GitLab provider
func NewProvider(config Config) (*Provider, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("GitLab token is required")
	}

	// Create GitLab client
	var client *gitlab.Client
	var err error

	if config.BaseURL != "" {
		client, err = gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.BaseURL))
	} else {
		client, err = gitlab.NewClient(config.Token)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Configure rate limiting
	rateLimit := config.RateLimit
	if rateLimit <= 0 {
		rateLimit = 5.0 // Default to 5 requests per second
	}

	rateBurst := config.RateBurst
	if rateBurst <= 0 {
		rateBurst = 10 // Default burst of 10
	}

	provider := &Provider{
		client:      client,
		token:       config.Token,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), rateBurst),
		logger:      utils.GetGlobalLogger().WithProvider(ProviderName),
	}

	// Initialize behavior manager if configuration is provided
	if config.BehaviorConfig != nil {
		provider.behaviorManager = utils.NewBehaviorManager(config.BehaviorConfig)
	}

	return provider, nil
}

// GetProviderName returns the provider name
func (p *Provider) GetProviderName() string {
	return ProviderName
}

// executeWithBehavior wraps API calls with rate limiting and retry logic
func (p *Provider) executeWithBehavior(ctx context.Context, operation string, fn func() error) error {
	if p.behaviorManager != nil {
		return p.behaviorManager.ExecuteWithBehavior(ctx, ProviderName, operation, fn)
	}
	// Fallback to manual rate limiting if no behavior manager
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return fn()
}

// executeWithBehaviorAndResult wraps API calls with rate limiting and retry logic, returning a result
func executeWithBehaviorAndResult[T any](ctx context.Context, p *Provider, operation string, fn func() (T, error)) (T, error) {
	if p.behaviorManager != nil {
		return utils.ExecuteWithBehaviorAndResult(ctx, p.behaviorManager, ProviderName, operation, fn)
	}
	// Fallback to manual rate limiting if no behavior manager
	var zero T
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return zero, err
	}
	return fn()
}

// Authenticate validates the GitLab token
func (p *Provider) Authenticate(ctx context.Context) error {
	return p.executeWithBehavior(ctx, "authenticate", func() error {
		user, _, err := p.client.Users.CurrentUser()
		if err != nil {
			return common.NewProviderError(ProviderName, common.ErrorTypeAuth,
				"failed to authenticate with GitLab", err)
		}

		p.logger.Infof("Authenticated as GitLab user: %s", user.Username)
		return nil
	})
}

// ListRepositories lists all accessible repositories
func (p *Provider) ListRepositories(ctx context.Context) ([]common.Repository, error) {
	return executeWithBehaviorAndResult(ctx, p, "list_repositories", func() ([]common.Repository, error) {
		var allRepos []common.Repository
		opts := &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{PerPage: 100},
			OrderBy:     gitlab.Ptr("updated_at"),
			Sort:        gitlab.Ptr("desc"),
			Membership:  gitlab.Ptr(true),
		}

		for {
			projects, resp, err := p.client.Projects.ListProjects(opts)
			if err != nil {
				return nil, common.NewProviderError(ProviderName, common.ErrorTypeNetwork,
					"failed to list repositories", err)
			}

			for _, project := range projects {
				allRepos = append(allRepos, p.convertProject(project))
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage

			// For pagination, we need to handle rate limiting for each page
			if resp.NextPage != 0 {
				if err := p.executeWithBehavior(ctx, "list_repositories_page", func() error {
					return nil // Just for rate limiting
				}); err != nil {
					return nil, err
				}
			}
		}

		p.logger.Infof("Found %d repositories", len(allRepos))
		return allRepos, nil
	})
}

// GetRepository gets a specific repository
func (p *Provider) GetRepository(ctx context.Context, owner, name string) (*common.Repository, error) {
	return executeWithBehaviorAndResult(ctx, p, "get_repository", func() (*common.Repository, error) {
		projectPath := fmt.Sprintf("%s/%s", owner, name)
		project, _, err := p.client.Projects.GetProject(projectPath, nil)
		if err != nil {
			return nil, p.handleGitLabError("failed to get repository", err)
		}

		convertedRepo := p.convertProject(project)
		return &convertedRepo, nil
	})
}

// ListPullRequests lists merge requests for a repository
func (p *Provider) ListPullRequests(ctx context.Context, repo common.Repository, opts common.ListPROptions) ([]common.PullRequest, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	projectPath := fmt.Sprintf("%s/%s", owner, name)

	glOpts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions:  gitlab.ListOptions{PerPage: 100},
		State:        gitlab.Ptr(string(opts.State)),
		OrderBy:      gitlab.Ptr(opts.Sort),
		Sort:         gitlab.Ptr(opts.Direction),
		TargetBranch: gitlab.Ptr(opts.Base),
		SourceBranch: gitlab.Ptr(opts.Head),
	}

	if !opts.Since.IsZero() {
		glOpts.CreatedAfter = &opts.Since
	}

	var allMRs []common.PullRequest

	for {
		var mrs []*gitlab.BasicMergeRequest
		var resp *gitlab.Response
		var err error

		err = p.executeWithBehavior(ctx, "list_pull_requests_page", func() error {
			mrs, resp, err = p.client.MergeRequests.ListProjectMergeRequests(projectPath, glOpts)
			return err
		})
		if err != nil {
			return nil, p.handleGitLabError("failed to list merge requests", err)
		}

		for _, mr := range mrs {
			convertedMR := p.convertBasicMergeRequest(mr)
			allMRs = append(allMRs, convertedMR)
		}

		if resp.NextPage == 0 {
			break
		}
		glOpts.Page = resp.NextPage
	}

	p.logger.Debugf("Found %d merge requests for %s", len(allMRs), repo.FullName)
	return allMRs, nil
}

// GetPullRequest gets a specific merge request
func (p *Provider) GetPullRequest(ctx context.Context, repo common.Repository, number int) (*common.PullRequest, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_pull_request", func() (*common.PullRequest, error) {
		projectPath := fmt.Sprintf("%s/%s", owner, name)
		mr, _, err := p.client.MergeRequests.GetMergeRequest(projectPath, number, nil)
		if err != nil {
			return nil, p.handleGitLabError("failed to get merge request", err)
		}

		convertedMR := p.convertMergeRequest(mr)
		return &convertedMR, nil
	})
}

// MergePullRequest merges a merge request
func (p *Provider) MergePullRequest(ctx context.Context, repo common.Repository, pr common.PullRequest, opts common.MergeOptions) error {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return fmt.Errorf("invalid repository name: %w", err)
	}

	return p.executeWithBehavior(ctx, "merge_pull_request", func() error {
		projectPath := fmt.Sprintf("%s/%s", owner, name)

		acceptOpts := &gitlab.AcceptMergeRequestOptions{
			MergeCommitMessage:       gitlab.Ptr(opts.CommitTitle),
			SHA:                      gitlab.Ptr(opts.SHA),
			ShouldRemoveSourceBranch: gitlab.Ptr(opts.DeleteBranch),
		}

		// Convert merge method
		switch opts.Method {
		case common.MergeMethodMerge:
			// Default GitLab behavior is merge
		case common.MergeMethodSquash:
			acceptOpts.Squash = gitlab.Ptr(true)
		case common.MergeMethodRebase:
			// GitLab doesn't have direct rebase merge, use merge
			p.logger.Warn("Rebase merge not directly supported in GitLab, using merge")
		}

		_, _, err = p.client.MergeRequests.AcceptMergeRequest(projectPath, pr.Number, acceptOpts)
		if err != nil {
			return p.handleGitLabError("failed to merge request", err)
		}

		p.logger.Infof("Successfully merged MR #%d in %s", pr.Number, repo.FullName)
		return nil
	})
}

// GetPRStatus gets the status of a merge request
func (p *Provider) GetPRStatus(ctx context.Context, repo common.Repository, pr common.PullRequest) (*common.PRStatus, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_pr_status", func() (*common.PRStatus, error) {
		projectPath := fmt.Sprintf("%s/%s", owner, name)

		// Get commit status
		statuses, _, err := p.client.Commits.GetCommitStatuses(projectPath, pr.HeadSHA, nil)
		if err != nil {
			return nil, p.handleGitLabError("failed to get MR status", err)
		}

		// Determine overall state
		state := common.PRStatusSuccess
		if len(statuses) > 0 {
			// GitLab returns statuses in most recent first order
			state = common.PRStatusState(statuses[0].Status)
		}

		return &common.PRStatus{
			State:       state,
			Description: "Combined status",
			TargetURL:   "",
			Context:     "gitlab/combined-status",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	})
}

// GetChecks gets the checks for a merge request (GitLab pipelines)
func (p *Provider) GetChecks(ctx context.Context, repo common.Repository, pr common.PullRequest) ([]common.Check, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_checks", func() ([]common.Check, error) {
		projectPath := fmt.Sprintf("%s/%s", owner, name)

		// Get pipelines for the commit
		pipelines, _, err := p.client.Pipelines.ListProjectPipelines(projectPath, &gitlab.ListProjectPipelinesOptions{
			SHA: gitlab.Ptr(pr.HeadSHA),
		})
		if err != nil {
			return nil, p.handleGitLabError("failed to get pipelines", err)
		}

		checks := make([]common.Check, 0, len(pipelines))
		for _, pipeline := range pipelines {
			check := common.Check{
				ID:         strconv.Itoa(pipeline.ID),
				Name:       fmt.Sprintf("Pipeline #%d", pipeline.ID),
				Status:     common.CheckStatus(pipeline.Status),
				Conclusion: pipeline.Status,
				DetailsURL: pipeline.WebURL,
				Summary:    fmt.Sprintf("Pipeline %s", pipeline.Status),
			}

			if pipeline.CreatedAt != nil {
				check.StartedAt = pipeline.CreatedAt
			}
			if pipeline.UpdatedAt != nil {
				check.CompletedAt = pipeline.UpdatedAt
			}

			checks = append(checks, check)
		}

		return checks, nil
	})
}

// GetRateLimit gets the current rate limit information
func (p *Provider) GetRateLimit(ctx context.Context) (*common.RateLimit, error) {
	// GitLab doesn't expose rate limit information in the same way as GitHub
	// Return a default response indicating no specific limits
	return &common.RateLimit{
		Limit:     5000,
		Remaining: 4999,
		ResetTime: time.Now().Add(time.Hour),
	}, nil
}
