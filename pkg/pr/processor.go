package pr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// Processor handles PR discovery and filtering
type Processor struct {
	providers map[string]common.Provider
	config    *config.Config
	logger    *utils.Logger
	executor  *utils.ParallelExecutor
}

// NewProcessor creates a new PR processor
func NewProcessor(providers map[string]common.Provider, cfg *config.Config) *Processor {
	return &Processor{
		providers: providers,
		config:    cfg,
		logger:    utils.GetGlobalLogger(),
		executor:  utils.NewParallelExecutor(cfg.Behavior.Concurrency),
	}
}

// ProcessResult contains the result of processing a repository
type ProcessResult struct {
	Provider     string            `json:"provider"`
	Repository   common.Repository `json:"repository"`
	PullRequests []ProcessedPR     `json:"pull_requests"`
	Error        error             `json:"error,omitempty"`
}

// ProcessedPR contains a PR and its processing status
type ProcessedPR struct {
	PullRequest common.PullRequest `json:"pull_request"`
	Status      PRStatus           `json:"status"`
	Reason      string             `json:"reason"`
	Ready       bool               `json:"ready"`
	Skipped     bool               `json:"skipped"`
	Error       error              `json:"error,omitempty"`
}

// PRStatus contains detailed status information about a PR
type PRStatus struct {
	State     common.PRStatusState `json:"state"`
	Checks    []common.Check       `json:"checks"`
	Ready     bool                 `json:"ready"`
	Reason    string               `json:"reason"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// ProcessOptions contains options for processing PRs
type ProcessOptions struct {
	DryRun        bool
	Providers     []string
	Repositories  []string
	MaxAge        time.Duration
	RequireChecks bool
	SkipLabels    []string
	IncludeClosed bool
}

// ProcessAllPRs processes PRs across all configured repositories
func (p *Processor) ProcessAllPRs(ctx context.Context, opts ProcessOptions) ([]ProcessResult, error) {
	p.logger.Info("Starting PR processing across all repositories")

	var allTasks []func(context.Context) (ProcessResult, error)

	// Create tasks for each provider and repository
	for providerName, repos := range p.config.Repositories {
		// Skip if provider not in filter
		if len(opts.Providers) > 0 && !contains(opts.Providers, providerName) {
			continue
		}

		provider, exists := p.providers[providerName]
		if !exists {
			p.logger.Warnf("Provider %s not available, skipping", providerName)
			continue
		}

		for _, repoConfig := range repos {
			// Skip if repository not in filter
			if len(opts.Repositories) > 0 && !containsRepoFilter(opts.Repositories, repoConfig.Name) {
				continue
			}

			// Capture variables for closure
			providerName := providerName
			repoConfig := repoConfig
			provider := provider

			task := func(ctx context.Context) (ProcessResult, error) {
				return p.processRepository(ctx, provider, providerName, repoConfig, opts)
			}

			allTasks = append(allTasks, task)
		}
	}

	if len(allTasks) == 0 {
		return nil, fmt.Errorf("no repositories to process")
	}

	p.logger.Infof("Processing %d repositories", len(allTasks))

	// Execute tasks in parallel using a simple approach for now
	results := make([]ProcessResult, 0, len(allTasks))
	for _, task := range allTasks {
		result, err := task(ctx)
		if err != nil {
			return nil, fmt.Errorf("task execution failed: %w", err)
		}
		results = append(results, result)
	}

	// Log summary
	p.logProcessingSummary(results)

	return results, nil
}

// processRepository processes PRs for a single repository
func (p *Processor) processRepository(ctx context.Context, provider common.Provider, providerName string, repoConfig config.Repository, opts ProcessOptions) (ProcessResult, error) {
	logger := p.logger.WithProvider(providerName).WithRepo(repoConfig.Name)
	logger.Debug("Processing repository")

	result := ProcessResult{
		Provider: providerName,
	}

	// Get repository information
	owner, name, err := common.ParseRepository(repoConfig.Name)
	if err != nil {
		result.Error = fmt.Errorf("invalid repository name: %w", err)
		return result, nil
	}

	repo, err := provider.GetRepository(ctx, owner, name)
	if err != nil {
		result.Error = fmt.Errorf("failed to get repository: %w", err)
		return result, nil
	}

	result.Repository = *repo

	// List pull requests
	prOpts := common.ListPROptions{
		State:   common.PRStateOpen,
		PerPage: 100,
	}

	if opts.MaxAge > 0 {
		prOpts.Since = time.Now().Add(-opts.MaxAge)
	}

	prs, err := provider.ListPullRequests(ctx, *repo, prOpts)
	if err != nil {
		result.Error = fmt.Errorf("failed to list pull requests: %w", err)
		return result, nil
	}

	logger.Debugf("Found %d pull requests", len(prs))

	// Process each PR
	for _, pr := range prs {
		processedPR := p.processPR(ctx, provider, *repo, pr, repoConfig, opts)
		result.PullRequests = append(result.PullRequests, processedPR)
	}

	return result, nil
}

// processPR processes a single pull request
func (p *Processor) processPR(ctx context.Context, provider common.Provider, repo common.Repository, pr common.PullRequest, repoConfig config.Repository, opts ProcessOptions) ProcessedPR {
	logger := p.logger.WithProvider(provider.GetProviderName()).WithRepo(repo.FullName).WithPR(pr.Number)

	processed := ProcessedPR{
		PullRequest: pr,
		Status: PRStatus{
			UpdatedAt: time.Now(),
		},
	}

	// Check if PR author is allowed
	if !p.isAuthorAllowed(pr.Author) {
		processed.Skipped = true
		processed.Reason = fmt.Sprintf("author '%s' not in allowed actors", pr.Author.Login)
		logger.Debug(processed.Reason)
		return processed
	}

	// Check skip labels (global and repository-specific)
	allSkipLabels := append(p.config.PRFilters.SkipLabels, repoConfig.SkipLabels...)
	if opts.SkipLabels != nil {
		allSkipLabels = append(allSkipLabels, opts.SkipLabels...)
	}

	if pr.HasAnyLabel(allSkipLabels) {
		processed.Skipped = true
		processed.Reason = "PR has skip labels"
		logger.Debug(processed.Reason)
		return processed
	}

	// Check PR age if configured
	if p.config.PRFilters.MaxAge != "" {
		maxAge, err := utils.ParseDuration(p.config.PRFilters.MaxAge)
		if err == nil && pr.IsOld(maxAge) {
			processed.Skipped = true
			processed.Reason = fmt.Sprintf("PR is older than %s", maxAge)
			logger.Debug(processed.Reason)
			return processed
		}
	}

	// Get PR status and checks
	status, err := provider.GetPRStatus(ctx, repo, pr)
	if err != nil {
		processed.Error = fmt.Errorf("failed to get PR status: %w", err)
		logger.WithError(err).Error("Failed to get PR status")
		return processed
	}

	checks, err := provider.GetChecks(ctx, repo, pr)
	if err != nil {
		processed.Error = fmt.Errorf("failed to get PR checks: %w", err)
		logger.WithError(err).Error("Failed to get PR checks")
		return processed
	}

	processed.Status = PRStatus{
		State:     status.State,
		Checks:    checks,
		UpdatedAt: time.Now(),
	}

	// Check if PR is ready to merge
	ready, reason := p.isPRReady(pr, status, checks, repoConfig, opts)
	processed.Ready = ready
	processed.Status.Ready = ready
	processed.Status.Reason = reason
	processed.Reason = reason

	if ready {
		logger.Info("PR is ready for merge")
	} else {
		logger.Debugf("PR not ready: %s", reason)
	}

	return processed
}

// isAuthorAllowed checks if the PR author is in the allowed actors list
func (p *Processor) isAuthorAllowed(author common.User) bool {
	for _, allowedActor := range p.config.PRFilters.AllowedActors {
		if strings.EqualFold(author.Login, allowedActor) {
			return true
		}
	}
	return false
}

// isPRReady checks if a PR is ready to be merged
func (p *Processor) isPRReady(pr common.PullRequest, status *common.PRStatus, checks []common.Check, repoConfig config.Repository, opts ProcessOptions) (bool, string) {
	// Check if PR is open
	if !pr.IsOpen() {
		return false, "PR is not open"
	}

	// Check if PR is a draft
	if pr.IsDraft() {
		return false, "PR is a draft"
	}

	// Check if PR is mergeable
	if pr.Mergeable != nil && !*pr.Mergeable {
		return false, "PR has merge conflicts"
	}

	// Check if PR is locked
	if pr.Locked {
		return false, "PR is locked"
	}

	// Check status if required
	requireChecks := opts.RequireChecks || repoConfig.RequireChecks
	if requireChecks {
		if !status.IsSuccessful() {
			return false, fmt.Sprintf("status checks not passing: %s", status.State)
		}

		// Check individual checks
		for _, check := range checks {
			if !check.IsCompleted() {
				return false, fmt.Sprintf("check '%s' is still running", check.Name)
			}
			if check.IsFailed() {
				return false, fmt.Sprintf("check '%s' failed", check.Name)
			}
		}
	}

	// All checks passed
	return true, "ready to merge"
}

// logProcessingSummary logs a summary of processing results
func (p *Processor) logProcessingSummary(results []ProcessResult) {
	totalRepos := len(results)
	totalPRs := 0
	readyPRs := 0
	skippedPRs := 0
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
			continue
		}

		totalPRs += len(result.PullRequests)
		for _, pr := range result.PullRequests {
			if pr.Error != nil {
				errorCount++
			} else if pr.Skipped {
				skippedPRs++
			} else if pr.Ready {
				readyPRs++
			}
		}
	}

	p.logger.WithFields(map[string]interface{}{
		"total_repositories": totalRepos,
		"total_prs":          totalPRs,
		"ready_prs":          readyPRs,
		"skipped_prs":        skippedPRs,
		"errors":             errorCount,
	}).Info("PR processing completed")
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func containsRepoFilter(filters []string, repoName string) bool {
	for _, filter := range filters {
		if strings.Contains(strings.ToLower(repoName), strings.ToLower(filter)) {
			return true
		}
	}
	return false
}
