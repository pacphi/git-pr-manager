package merge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/pr"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// Executor handles PR merging operations
type Executor struct {
	providers map[string]common.Provider
	config    *config.Config
	logger    *utils.Logger
}

// NewExecutor creates a new merge executor
func NewExecutor(providers map[string]common.Provider, cfg *config.Config) *Executor {
	return &Executor{
		providers: providers,
		config:    cfg,
		logger:    utils.GetGlobalLogger(),
	}
}

// MergeResult contains the result of a merge operation
type MergeResult struct {
	Provider    string    `json:"provider"`
	Repository  string    `json:"repository"`
	PullRequest int       `json:"pull_request"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	MergeMethod string    `json:"merge_method"`
	MergedAt    time.Time `json:"merged_at"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
	Success     bool      `json:"success"`
	Error       error     `json:"error,omitempty"`
	Skipped     bool      `json:"skipped"`
	Reason      string    `json:"reason,omitempty"`
}

// MergeOptions contains options for merge operations
type MergeOptions struct {
	DryRun          bool
	Force           bool
	DeleteBranches  bool
	CustomMessage   string
	RequireApproval bool
}

// MergePRs merges all ready PRs from the provided results with sequential processing per repository
func (e *Executor) MergePRs(ctx context.Context, results []pr.ProcessResult, opts MergeOptions) ([]MergeResult, error) {
	e.logger.Info("Starting PR merge operations")

	// Pre-allocate mergeResults with estimated capacity
	totalEstimatedPRs := 0
	for _, result := range results {
		totalEstimatedPRs += len(result.PullRequests)
	}
	mergeResults := make([]MergeResult, 0, totalEstimatedPRs)
	resultsChan := make(chan MergeResult, 100) // Buffer for results

	// Group merge operations by repository to ensure sequential processing per repo
	repositoryTasks := make(map[string][]func(context.Context) error)

	for _, result := range results {
		if result.Error != nil {
			e.logger.WithError(result.Error).Warnf("Skipping repository %s due to processing error", result.Repository.FullName)
			continue
		}

		provider, exists := e.providers[result.Provider]
		if !exists {
			e.logger.Warnf("Provider %s not available, skipping", result.Provider)
			continue
		}

		repoKey := result.Repository.FullName

		for _, processedPR := range result.PullRequests {
			if processedPR.Error != nil {
				e.logger.WithError(processedPR.Error).Warnf("Skipping PR #%d due to processing error", processedPR.PullRequest.Number)
				continue
			}

			if processedPR.Skipped {
				resultsChan <- MergeResult{
					Provider:    result.Provider,
					Repository:  result.Repository.FullName,
					PullRequest: processedPR.PullRequest.Number,
					Title:       processedPR.PullRequest.Title,
					Author:      processedPR.PullRequest.Author.Login,
					Skipped:     true,
					Reason:      processedPR.Reason,
				}
				continue
			}

			if !processedPR.Ready && !opts.Force {
				resultsChan <- MergeResult{
					Provider:    result.Provider,
					Repository:  result.Repository.FullName,
					PullRequest: processedPR.PullRequest.Number,
					Title:       processedPR.PullRequest.Title,
					Author:      processedPR.PullRequest.Author.Login,
					Skipped:     true,
					Reason:      processedPR.Reason,
				}
				continue
			}

			// Create merge task and group by repository
			// Capture variables for closure
			providerCopy := provider
			repoRef := result.Repository
			prRef := processedPR.PullRequest
			task := func(ctx context.Context) error {
				mergeResult := e.mergePR(ctx, providerCopy, repoRef, prRef, opts)
				resultsChan <- mergeResult
				return nil // Don't fail the entire batch on individual merge failures
			}

			repositoryTasks[repoKey] = append(repositoryTasks[repoKey], task)
		}
	}

	// Create parallel tasks for each repository (repositories run in parallel, PRs within each repo run sequentially)
	repoTasks := make([]func(context.Context) error, 0, len(repositoryTasks))
	for repoName, repoMergeTasks := range repositoryTasks {
		// Capture variables for closure
		repoTasksCopy := repoMergeTasks
		repoNameCopy := repoName

		repoTask := func(ctx context.Context) error {
			e.logger.Infof("Processing %d PRs sequentially for repository %s", len(repoTasksCopy), repoNameCopy)

			for i, task := range repoTasksCopy {
				if err := task(ctx); err != nil {
					e.logger.WithError(err).Errorf("Failed to execute merge task %d for repository %s", i, repoNameCopy)
				}

				// Add delay between merges within the same repository (except for the last one)
				// Skip delay in dry-run mode since no actual merges are happening
				if i < len(repoTasksCopy)-1 && e.config.Behavior.MergeDelay > 0 && !opts.DryRun {
					e.logger.Debugf("Waiting %v before next merge in repository %s", e.config.Behavior.MergeDelay, repoNameCopy)
					time.Sleep(e.config.Behavior.MergeDelay)
				}
			}
			return nil
		}

		repoTasks = append(repoTasks, repoTask)
	}

	// Execute repository tasks in parallel (each repo processes its PRs sequentially)
	if len(repoTasks) > 0 {
		executor := utils.NewParallelExecutor(e.config.Behavior.Concurrency)
		if err := executor.Execute(ctx, repoTasks); err != nil {
			e.logger.WithError(err).Error("Error in repository merge execution")
			// Continue to collect results even if some tasks failed
		}
	}

	// Close results channel and collect all results
	close(resultsChan)
	for result := range resultsChan {
		mergeResults = append(mergeResults, result)
	}

	// Log summary
	e.logMergeSummary(mergeResults, opts.DryRun)

	return mergeResults, nil
}

// mergePR merges a single pull request
func (e *Executor) mergePR(ctx context.Context, provider common.Provider, repo common.Repository, pr common.PullRequest, opts MergeOptions) MergeResult {
	logger := e.logger.WithProvider(provider.GetProviderName()).WithRepo(repo.FullName).WithPR(pr.Number)

	result := MergeResult{
		Provider:    provider.GetProviderName(),
		Repository:  repo.FullName,
		PullRequest: pr.Number,
		Title:       pr.Title,
		Author:      pr.Author.Login,
	}

	// Get repository configuration
	repoConfigs := e.config.Repositories[provider.GetProviderName()]
	var repoConfig *config.Repository
	for _, cfg := range repoConfigs {
		if cfg.Name == repo.FullName {
			repoConfig = &cfg
			break
		}
	}

	if repoConfig == nil {
		result.Error = fmt.Errorf("repository configuration not found")
		return result
	}

	// Determine merge method
	mergeMethod := e.determineMergeMethod(*repoConfig)
	result.MergeMethod = string(mergeMethod)

	// Prepare merge options
	mergeOpts := common.MergeOptions{
		Method:       mergeMethod,
		SHA:          pr.HeadSHA,
		DeleteBranch: opts.DeleteBranches,
	}

	// Set commit title and message
	mergeOpts.CommitTitle, mergeOpts.CommitMessage = e.generateCommitMessage(pr, mergeMethod, opts.CustomMessage)

	if opts.DryRun {
		logger.Infof("[DRY RUN] Would merge PR with method %s", mergeMethod)
		result.Success = true
		result.Reason = "dry run - would merge"
		return result
	}

	// Perform the actual merge
	logger.Infof("Merging PR with method %s", mergeMethod)
	err := provider.MergePullRequest(ctx, repo, pr, mergeOpts)
	if err != nil {
		result.Error = fmt.Errorf("merge failed: %w", err)
		logger.WithError(err).Error("Failed to merge PR")
		return result
	}

	result.Success = true
	result.MergedAt = time.Now()
	result.Reason = "successfully merged"

	logger.Infof("Successfully merged PR #%d", pr.Number)
	return result
}

// determineMergeMethod determines the merge method to use for a repository
func (e *Executor) determineMergeMethod(repoConfig config.Repository) common.MergeMethod {
	if repoConfig.MergeStrategy != "" {
		switch repoConfig.MergeStrategy {
		case config.MergeStrategyMerge:
			return common.MergeMethodMerge
		case config.MergeStrategySquash:
			return common.MergeMethodSquash
		case config.MergeStrategyRebase:
			return common.MergeMethodRebase
		}
	}

	// Default to squash
	return common.MergeMethodSquash
}

// generateCommitMessage generates commit title and message for the merge
func (e *Executor) generateCommitMessage(pr common.PullRequest, method common.MergeMethod, customMessage string) (string, string) {
	if customMessage != "" {
		return customMessage, ""
	}

	title := pr.Title
	message := ""

	switch method {
	case common.MergeMethodSquash:
		// For squash merges, include PR number
		if !strings.Contains(title, fmt.Sprintf("#%d", pr.Number)) {
			title = fmt.Sprintf("%s (#%d)", title, pr.Number)
		}

		// Include PR body if it's not too long
		if pr.Body != "" && len(pr.Body) < 500 {
			message = pr.Body
		}

	case common.MergeMethodMerge:
		// For merge commits, use default GitHub format
		title = fmt.Sprintf("Merge pull request #%d from %s", pr.Number, pr.HeadBranch)
		message = pr.Title

	case common.MergeMethodRebase:
		// For rebase, keep original title
		// No additional message needed
	}

	return title, message
}

// logMergeSummary logs a summary of merge operations
func (e *Executor) logMergeSummary(results []MergeResult, dryRun bool) {
	total := len(results)
	successful := 0
	skipped := 0
	failed := 0

	for _, result := range results {
		if result.Skipped {
			skipped++
		} else if result.Success {
			successful++
		} else {
			failed++
		}
	}

	action := "merged"
	if dryRun {
		action = "would merge"
	}

	e.logger.WithFields(map[string]interface{}{
		"total":      total,
		"successful": successful,
		"skipped":    skipped,
		"failed":     failed,
		"dry_run":    dryRun,
	}).Infof("Merge operations completed: %s %d PRs", action, successful)

	// Log individual failures
	for _, result := range results {
		if result.Error != nil {
			e.logger.WithError(result.Error).Errorf("Failed to merge PR #%d in %s", result.PullRequest, result.Repository)
		}
	}
}

// ValidateMergeability validates that PRs can be merged before attempting
func (e *Executor) ValidateMergeability(ctx context.Context, results []pr.ProcessResult) error {
	var errors []string

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		provider, exists := e.providers[result.Provider]
		if !exists {
			errors = append(errors, fmt.Sprintf("provider %s not available", result.Provider))
			continue
		}

		for _, processedPR := range result.PullRequests {
			if !processedPR.Ready || processedPR.Skipped || processedPR.Error != nil {
				continue
			}

			// Additional validation can be added here
			// For example, checking branch protection rules, required reviews, etc.
			_ = provider // Use provider for additional checks if needed
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}
