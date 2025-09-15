package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/time/rate"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

const ProviderName = "github"

// Provider implements the common.Provider interface for GitHub
type Provider struct {
	client          *github.Client
	token           string
	rateLimiter     *rate.Limiter
	behaviorManager *utils.BehaviorManager
	logger          *utils.Logger
}

// Config contains GitHub provider configuration
type Config struct {
	Token          string
	BaseURL        string
	RateLimit      float64
	RateBurst      int
	BehaviorConfig *config.Config
}

// NewProvider creates a new GitHub provider
func NewProvider(config Config) (*Provider, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	// Create HTTP client with authentication
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			MaxIdleConnsPerHost: 10,
		},
		Timeout: 30 * time.Second,
	}

	// Create GitHub client
	client := github.NewClient(httpClient).WithAuthToken(config.Token)

	// Set custom base URL if provided (for GitHub Enterprise)
	if config.BaseURL != "" {
		updatedClient, err := client.WithEnterpriseURLs(config.BaseURL, config.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set GitHub base URL: %w", err)
		}
		client = updatedClient
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

// Authenticate validates the GitHub token
func (p *Provider) Authenticate(ctx context.Context) error {
	return p.executeWithBehavior(ctx, "authenticate", func() error {
		user, _, err := p.client.Users.Get(ctx, "")
		if err != nil {
			return common.NewProviderError(ProviderName, common.ErrorTypeAuth,
				"failed to authenticate with GitHub", err)
		}

		p.logger.Infof("Authenticated as GitHub user: %s", user.GetLogin())
		return nil
	})
}

// ListRepositories lists all accessible repositories
func (p *Provider) ListRepositories(ctx context.Context) ([]common.Repository, error) {
	return executeWithBehaviorAndResult(ctx, p, "list_repositories", func() ([]common.Repository, error) {
		var allRepos []common.Repository
		opts := &github.RepositoryListByAuthenticatedUserOptions{
			ListOptions: github.ListOptions{PerPage: 100},
			Sort:        "updated",
			Direction:   "desc",
		}

		for {
			repos, resp, err := p.client.Repositories.ListByAuthenticatedUser(ctx, opts)
			if err != nil {
				return nil, common.NewProviderError(ProviderName, common.ErrorTypeNetwork,
					"failed to list repositories", err)
			}

			for _, repo := range repos {
				allRepos = append(allRepos, p.convertRepository(repo))
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
		repo, _, err := p.client.Repositories.Get(ctx, owner, name)
		if err != nil {
			return nil, p.handleGitHubError("failed to get repository", err)
		}

		convertedRepo := p.convertRepository(repo)
		return &convertedRepo, nil
	})
}

// ListPullRequests lists pull requests for a repository
func (p *Provider) ListPullRequests(ctx context.Context, repo common.Repository, opts common.ListPROptions) ([]common.PullRequest, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	ghOpts := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       string(opts.State),
		Sort:        opts.Sort,
		Direction:   opts.Direction,
		Base:        opts.Base,
		Head:        opts.Head,
	}

	var allPRs []common.PullRequest

	for {
		var prs []*github.PullRequest
		var resp *github.Response
		var err error

		err = p.executeWithBehavior(ctx, "list_pull_requests_page", func() error {
			prs, resp, err = p.client.PullRequests.List(ctx, owner, name, ghOpts)
			return err
		})
		if err != nil {
			return nil, p.handleGitHubError("failed to list pull requests", err)
		}

		for _, pr := range prs {
			convertedPR := p.convertPullRequest(pr)

			// Filter by age if specified
			if !opts.Since.IsZero() && convertedPR.CreatedAt.Before(opts.Since) {
				continue
			}

			allPRs = append(allPRs, convertedPR)
		}

		if resp.NextPage == 0 {
			break
		}
		ghOpts.Page = resp.NextPage
	}

	p.logger.Debugf("Found %d pull requests for %s", len(allPRs), repo.FullName)
	return allPRs, nil
}

// GetPullRequest gets a specific pull request
func (p *Provider) GetPullRequest(ctx context.Context, repo common.Repository, number int) (*common.PullRequest, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_pull_request", func() (*common.PullRequest, error) {
		pr, _, err := p.client.PullRequests.Get(ctx, owner, name, number)
		if err != nil {
			return nil, p.handleGitHubError("failed to get pull request", err)
		}

		convertedPR := p.convertPullRequest(pr)
		return &convertedPR, nil
	})
}

// MergePullRequest merges a pull request
func (p *Provider) MergePullRequest(ctx context.Context, repo common.Repository, pr common.PullRequest, opts common.MergeOptions) error {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return fmt.Errorf("invalid repository name: %w", err)
	}

	return p.executeWithBehavior(ctx, "merge_pull_request", func() error {
		// Convert merge method
		var mergeMethod string
		switch opts.Method {
		case common.MergeMethodMerge:
			mergeMethod = "merge"
		case common.MergeMethodSquash:
			mergeMethod = "squash"
		case common.MergeMethodRebase:
			mergeMethod = "rebase"
		default:
			mergeMethod = "squash" // Default to squash
		}

		mergeOpts := &github.PullRequestOptions{
			CommitTitle: opts.CommitTitle,
			SHA:         opts.SHA,
		}

		_, _, err = p.client.PullRequests.Merge(ctx, owner, name, pr.Number, mergeMethod, mergeOpts)
		if err != nil {
			return p.handleGitHubError("failed to merge pull request", err)
		}

		// Delete branch if requested
		if opts.DeleteBranch {
			if err := p.deleteBranch(ctx, owner, name, pr.HeadBranch); err != nil {
				p.logger.Warnf("Failed to delete branch %s: %v", pr.HeadBranch, err)
			}
		}

		p.logger.Infof("Successfully merged PR #%d in %s", pr.Number, repo.FullName)
		return nil
	})
}

// GetPRStatus gets the status of a pull request
func (p *Provider) GetPRStatus(ctx context.Context, repo common.Repository, pr common.PullRequest) (*common.PRStatus, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_pr_status", func() (*common.PRStatus, error) {
		// Get combined status for the PR's head SHA
		status, _, err := p.client.Repositories.GetCombinedStatus(ctx, owner, name, pr.HeadSHA, nil)
		if err != nil {
			return nil, p.handleGitHubError("failed to get PR status", err)
		}

		return &common.PRStatus{
			State:       common.PRStatusState(status.GetState()),
			Description: "Combined status",
			TargetURL:   "", // status.GetURL() not available in this version
			Context:     "github/combined-status",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	})
}

// GetChecks gets the checks for a pull request
func (p *Provider) GetChecks(ctx context.Context, repo common.Repository, pr common.PullRequest) ([]common.Check, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_checks", func() ([]common.Check, error) {
		// Get check runs for the PR's head SHA
		checkRuns, _, err := p.client.Checks.ListCheckRunsForRef(ctx, owner, name, pr.HeadSHA, &github.ListCheckRunsOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		})
		if err != nil {
			return nil, p.handleGitHubError("failed to get checks", err)
		}

		checks := make([]common.Check, 0, len(checkRuns.CheckRuns))
		for _, checkRun := range checkRuns.CheckRuns {
			check := common.Check{
				ID:         strconv.FormatInt(checkRun.GetID(), 10),
				Name:       checkRun.GetName(),
				Status:     common.CheckStatus(checkRun.GetStatus()),
				Conclusion: checkRun.GetConclusion(),
				DetailsURL: checkRun.GetDetailsURL(),
				Summary:    checkRun.GetOutput().GetSummary(),
				Text:       checkRun.GetOutput().GetText(),
			}

			if checkRun.StartedAt != nil {
				check.StartedAt = &checkRun.StartedAt.Time
			}
			if checkRun.CompletedAt != nil {
				check.CompletedAt = &checkRun.CompletedAt.Time
			}

			checks = append(checks, check)
		}

		return checks, nil
	})
}

// GetRateLimit gets the current rate limit information
func (p *Provider) GetRateLimit(ctx context.Context) (*common.RateLimit, error) {
	return executeWithBehaviorAndResult(ctx, p, "get_rate_limit", func() (*common.RateLimit, error) {
		rateLimit, _, err := p.client.RateLimit.Get(ctx)
		if err != nil {
			return nil, p.handleGitHubError("failed to get rate limit", err)
		}

		return &common.RateLimit{
			Limit:     rateLimit.Core.Limit,
			Remaining: rateLimit.Core.Remaining,
			ResetTime: rateLimit.Core.Reset.Time,
		}, nil
	})
}
