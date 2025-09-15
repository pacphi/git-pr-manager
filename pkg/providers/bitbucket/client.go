package bitbucket

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

const (
	ProviderName = "bitbucket"
	BaseURL      = "https://api.bitbucket.org/2.0"
)

// Provider implements the common.Provider interface for Bitbucket
type Provider struct {
	httpClient      *utils.HTTPClient
	username        string
	appPassword     string
	workspace       string
	rateLimiter     *rate.Limiter
	behaviorManager *utils.BehaviorManager
	logger          *utils.Logger
}

// Config contains Bitbucket provider configuration
type Config struct {
	Username       string
	AppPassword    string
	Workspace      string
	RateLimit      float64
	RateBurst      int
	BehaviorConfig *config.Config
}

// BitbucketRepository represents a Bitbucket repository response
type BitbucketRepository struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	IsPrivate   bool      `json:"is_private"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
	Size        int       `json:"size"`
	MainBranch  struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
	Owner struct {
		UUID        string `json:"uuid"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Type        string `json:"type"`
	} `json:"owner"`
	Links struct {
		Clone []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

// BitbucketPullRequest represents a Bitbucket pull request response
type BitbucketPullRequest struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	CreatedOn   time.Time  `json:"created_on"`
	UpdatedOn   time.Time  `json:"updated_on"`
	MergedOn    *time.Time `json:"merge_commit,omitempty"`
	ClosedOn    *time.Time `json:"closed_on,omitempty"`
	Author      struct {
		UUID        string `json:"uuid"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Type        string `json:"type"`
	} `json:"author"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
		Commit struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
	Links struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Diff struct {
			Href string `json:"href"`
		} `json:"diff"`
	} `json:"links"`
}

// BitbucketCommitStatus represents a Bitbucket commit status
type BitbucketCommitStatus struct {
	State       string    `json:"state"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// BitbucketUser represents the current user
type BitbucketUser struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

// NewProvider creates a new Bitbucket provider
func NewProvider(config Config) (*Provider, error) {
	if config.Username == "" || config.AppPassword == "" {
		return nil, fmt.Errorf("bitbucket username and app password are required")
	}

	// Configure rate limiting
	rateLimit := config.RateLimit
	if rateLimit <= 0 {
		rateLimit = 1.0 // Default to 1 request per second (Bitbucket is more restrictive)
	}

	rateBurst := config.RateBurst
	if rateBurst <= 0 {
		rateBurst = 5 // Default burst of 5
	}

	// Use the rate-limited HTTP client from utils
	restyClient := utils.RateLimitedHTTPClient(rateLimit, rateBurst)
	restyClient.SetBaseURL(BaseURL)
	restyClient.SetBasicAuth(config.Username, config.AppPassword)
	restyClient.SetHeader("Accept", "application/json")

	// Create HTTPClient with proper configuration
	httpClient := utils.NewHTTPClientFromConfig(utils.HTTPClientConfig{
		BaseURL: BaseURL,
		Timeout: 30 * time.Second,
		Headers: map[string]string{
			"Accept": "application/json",
		},
		RateLimiter: rate.NewLimiter(rate.Limit(rateLimit), rateBurst),
	})

	// Set basic auth
	httpClient.SetBasicAuth(config.Username, config.AppPassword)

	provider := &Provider{
		httpClient:  httpClient,
		username:    config.Username,
		appPassword: config.AppPassword,
		workspace:   config.Workspace,
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

// Authenticate validates the Bitbucket credentials
func (p *Provider) Authenticate(ctx context.Context) error {
	return p.executeWithBehavior(ctx, "authenticate", func() error {
		var user BitbucketUser
		err := p.httpClient.Get(ctx, "/user", &user)
		if err != nil {
			return common.NewProviderError(ProviderName, common.ErrorTypeAuth,
				"failed to authenticate with Bitbucket", err)
		}

		p.logger.Infof("Authenticated as Bitbucket user: %s", user.Username)
		return nil
	})
}

// ListRepositories lists all accessible repositories
func (p *Provider) ListRepositories(ctx context.Context) ([]common.Repository, error) {
	return executeWithBehaviorAndResult(ctx, p, "list_repositories", func() ([]common.Repository, error) {
		var allRepos []common.Repository
		var url string

		// If workspace is specified, list repositories in that workspace
		if p.workspace != "" {
			url = fmt.Sprintf("/repositories/%s", p.workspace)
		} else {
			url = fmt.Sprintf("/repositories/%s", p.username)
		}

		page := 1
		for {
			var response struct {
				Values []BitbucketRepository `json:"values"`
				Next   string                `json:"next"`
			}

			err := p.httpClient.Get(ctx, fmt.Sprintf("%s?page=%d&pagelen=50", url, page), &response)
			if err != nil {
				return nil, common.NewProviderError(ProviderName, common.ErrorTypeNetwork,
					"failed to list repositories", err)
			}

			for _, repo := range response.Values {
				allRepos = append(allRepos, p.convertRepository(&repo))
			}

			if response.Next == "" {
				break
			}
			page++

			// For pagination, add a small delay between pages
			if response.Next != "" {
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
		var repo BitbucketRepository
		url := fmt.Sprintf("/repositories/%s/%s", owner, name)
		err := p.httpClient.Get(ctx, url, &repo)
		if err != nil {
			return nil, p.handleBitbucketError("failed to get repository", err)
		}

		convertedRepo := p.convertRepository(&repo)
		return &convertedRepo, nil
	})
}

// ListPullRequests lists pull requests for a repository
func (p *Provider) ListPullRequests(ctx context.Context, repo common.Repository, opts common.ListPROptions) ([]common.PullRequest, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	var allPRs []common.PullRequest
	url := fmt.Sprintf("/repositories/%s/%s/pullrequests", owner, name)

	// Add query parameters
	params := []string{}
	if opts.State != "" {
		params = append(params, fmt.Sprintf("state=%s", strings.ToUpper(string(opts.State))))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	page := 1
	for {
		var response struct {
			Values []BitbucketPullRequest `json:"values"`
			Next   string                 `json:"next"`
		}

		pageURL := fmt.Sprintf("%s&page=%d&pagelen=50", url, page)
		if !strings.Contains(url, "?") {
			pageURL = fmt.Sprintf("%s?page=%d&pagelen=50", url, page)
		}

		err := p.executeWithBehavior(ctx, "list_pull_requests_page", func() error {
			return p.httpClient.Get(ctx, pageURL, &response)
		})
		if err != nil {
			return nil, p.handleBitbucketError("failed to list pull requests", err)
		}

		for _, pr := range response.Values {
			convertedPR := p.convertPullRequest(&pr)

			// Filter by age if specified
			if !opts.Since.IsZero() && convertedPR.CreatedAt.Before(opts.Since) {
				continue
			}

			allPRs = append(allPRs, convertedPR)
		}

		if response.Next == "" {
			break
		}
		page++
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
		var pr BitbucketPullRequest
		url := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", owner, name, number)
		err = p.httpClient.Get(ctx, url, &pr)
		if err != nil {
			return nil, p.handleBitbucketError("failed to get pull request", err)
		}

		convertedPR := p.convertPullRequest(&pr)
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
		mergeStrategy := "merge_commit"
		switch opts.Method {
		case common.MergeMethodMerge:
			mergeStrategy = "merge_commit"
		case common.MergeMethodSquash:
			mergeStrategy = "squash"
		case common.MergeMethodRebase:
			// Bitbucket doesn't support rebase merge, use squash
			mergeStrategy = "squash"
			p.logger.Warn("Rebase merge not supported in Bitbucket, using squash")
		}

		mergeData := map[string]interface{}{
			"type":           "pullrequest",
			"message":        opts.CommitTitle,
			"merge_strategy": mergeStrategy,
		}

		url := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge", owner, name, pr.Number)
		err = p.httpClient.Post(ctx, url, mergeData, nil)
		if err != nil {
			return p.handleBitbucketError("failed to merge pull request", err)
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
		// Get commit statuses
		var response struct {
			Values []BitbucketCommitStatus `json:"values"`
		}

		url := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses", owner, name, pr.HeadSHA)
		err = p.httpClient.Get(ctx, url, &response)
		if err != nil {
			return nil, p.handleBitbucketError("failed to get PR status", err)
		}

		// Determine overall state
		state := common.PRStatusSuccess
		if len(response.Values) > 0 {
			// Use the first status (most recent)
			bbState := response.Values[0].State
			switch strings.ToLower(bbState) {
			case "successful":
				state = common.PRStatusSuccess
			case "failed":
				state = common.PRStatusFailure
			case "inprogress":
				state = common.PRStatusPending
			default:
				state = common.PRStatusPending
			}
		}

		return &common.PRStatus{
			State:       state,
			Description: "Combined status",
			TargetURL:   "",
			Context:     "bitbucket/combined-status",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	})
}

// GetChecks gets the checks for a pull request (Bitbucket pipelines)
func (p *Provider) GetChecks(ctx context.Context, repo common.Repository, pr common.PullRequest) ([]common.Check, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	return executeWithBehaviorAndResult(ctx, p, "get_checks", func() ([]common.Check, error) {
		// Get pipelines for the commit
		var response struct {
			Values []struct {
				UUID  string `json:"uuid"`
				State struct {
					Name string `json:"name"`
				} `json:"state"`
				CreatedOn   time.Time  `json:"created_on"`
				CompletedOn *time.Time `json:"completed_on"`
			} `json:"values"`
		}

		url := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses/build", owner, name, pr.HeadSHA)
		err = p.httpClient.Get(ctx, url, &response)
		if err != nil {
			// Pipelines might not exist, return empty slice
			return []common.Check{}, err
		}

		checks := make([]common.Check, 0, len(response.Values))
		for i, pipeline := range response.Values {
			check := common.Check{
				ID:         pipeline.UUID,
				Name:       fmt.Sprintf("Pipeline #%d", i+1),
				Status:     common.CheckStatus(strings.ToLower(pipeline.State.Name)),
				Conclusion: pipeline.State.Name,
				Summary:    fmt.Sprintf("Pipeline %s", pipeline.State.Name),
				StartedAt:  &pipeline.CreatedOn,
			}

			if pipeline.CompletedOn != nil {
				check.CompletedAt = pipeline.CompletedOn
			}

			checks = append(checks, check)
		}

		return checks, nil
	})
}

// GetRateLimit gets the current rate limit information
func (p *Provider) GetRateLimit(ctx context.Context) (*common.RateLimit, error) {
	// Bitbucket doesn't expose rate limit information in headers
	// Return a conservative estimate
	return &common.RateLimit{
		Limit:     1000,
		Remaining: 999,
		ResetTime: time.Now().Add(time.Hour),
	}, nil
}
