package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
)

// handleGitHubError converts GitHub API errors to common provider errors
func (p *Provider) handleGitHubError(message string, err error) error {
	if err == nil {
		return nil
	}

	// Handle GitHub-specific errors
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		return p.convertGitHubError(message, ghErr)
	}

	// Handle HTTP errors
	if strings.Contains(err.Error(), "context canceled") {
		return common.NewProviderError(ProviderName, common.ErrorTypeNetwork, message, err)
	}

	if strings.Contains(err.Error(), "timeout") {
		return common.NewProviderError(ProviderName, common.ErrorTypeNetwork, message, err)
	}

	return common.NewProviderError(ProviderName, common.ErrorTypeUnknown, message, err)
}

// convertGitHubError converts GitHub ErrorResponse to provider error
func (p *Provider) convertGitHubError(message string, ghErr *github.ErrorResponse) error {
	var errorType common.ErrorType

	switch ghErr.Response.StatusCode {
	case http.StatusUnauthorized:
		errorType = common.ErrorTypeAuth
	case http.StatusForbidden:
		errorType = common.ErrorTypePermission
	case http.StatusNotFound:
		errorType = common.ErrorTypeNotFound
	case http.StatusConflict:
		errorType = common.ErrorTypeConflict
	case http.StatusUnprocessableEntity:
		errorType = common.ErrorTypeValidation
	case http.StatusTooManyRequests:
		errorType = common.ErrorTypeRateLimit
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		errorType = common.ErrorTypeNetwork
	default:
		errorType = common.ErrorTypeUnknown
	}

	providerError := common.NewProviderError(ProviderName, errorType, message, ghErr)
	providerError.StatusCode = ghErr.Response.StatusCode

	// Add retry after header if present for rate limiting
	if errorType == common.ErrorTypeRateLimit {
		if retryAfter := ghErr.Response.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				retryDuration := time.Duration(seconds) * time.Second
				providerError.RetryAfter = &retryDuration
			}
		}
	}

	// Add error details from GitHub response
	if len(ghErr.Errors) > 0 {
		errorDetails := make([]string, len(ghErr.Errors))
		for i, err := range ghErr.Errors {
			errorDetails[i] = err.Message
		}
		providerError.Message = fmt.Sprintf("%s: %s", message, strings.Join(errorDetails, ", "))
	} else if ghErr.Message != "" {
		providerError.Message = fmt.Sprintf("%s: %s", message, ghErr.Message)
	}

	return providerError
}

// deleteBranch deletes a branch from the repository
func (p *Provider) deleteBranch(ctx context.Context, owner, repo, branch string) error {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	_, err := p.client.Git.DeleteRef(ctx, owner, repo, "heads/"+branch)
	if err != nil {
		return p.handleGitHubError("failed to delete branch", err)
	}

	p.logger.Debugf("Deleted branch %s in %s/%s", branch, owner, repo)
	return nil
}

// GetRepositoryTopics gets the topics for a repository
func (p *Provider) GetRepositoryTopics(ctx context.Context, owner, name string) ([]string, error) {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	topics, _, err := p.client.Repositories.ListAllTopics(ctx, owner, name)
	if err != nil {
		return nil, p.handleGitHubError("failed to get repository topics", err)
	}

	return topics, nil
}

// GetPullRequestReviews gets the reviews for a pull request
func (p *Provider) GetPullRequestReviews(ctx context.Context, owner, name string, number int) ([]*github.PullRequestReview, error) {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	reviews, _, err := p.client.PullRequests.ListReviews(ctx, owner, name, number, nil)
	if err != nil {
		return nil, p.handleGitHubError("failed to get pull request reviews", err)
	}

	return reviews, nil
}

// GetPullRequestFiles gets the files changed in a pull request
func (p *Provider) GetPullRequestFiles(ctx context.Context, owner, name string, number int) ([]*github.CommitFile, error) {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	files, _, err := p.client.PullRequests.ListFiles(ctx, owner, name, number, nil)
	if err != nil {
		return nil, p.handleGitHubError("failed to get pull request files", err)
	}

	return files, nil
}

// GetPullRequestCommits gets the commits in a pull request
func (p *Provider) GetPullRequestCommits(ctx context.Context, owner, name string, number int) ([]*github.RepositoryCommit, error) {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	commits, _, err := p.client.PullRequests.ListCommits(ctx, owner, name, number, nil)
	if err != nil {
		return nil, p.handleGitHubError("failed to get pull request commits", err)
	}

	return commits, nil
}

// IsPullRequestReady checks if a pull request is ready to be merged
func (p *Provider) IsPullRequestReady(ctx context.Context, repo common.Repository, pr common.PullRequest, requireChecks bool) (bool, string, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return false, "invalid repository name", err
	}

	// Check if PR is open
	if !pr.IsOpen() {
		return false, "pull request is not open", nil
	}

	// Check if PR is a draft
	if pr.IsDraft() {
		return false, "pull request is a draft", nil
	}

	// Check if PR is mergeable
	if pr.Mergeable != nil && !*pr.Mergeable {
		return false, "pull request has merge conflicts", nil
	}

	// Check status checks if required
	if requireChecks {
		status, err := p.GetPRStatus(ctx, repo, pr)
		if err != nil {
			return false, "failed to get PR status", err
		}

		if !status.IsSuccessful() {
			return false, fmt.Sprintf("status checks not passing: %s", status.State), nil
		}

		// Also check individual check runs
		checks, err := p.GetChecks(ctx, repo, pr)
		if err != nil {
			return false, "failed to get checks", err
		}

		for _, check := range checks {
			if !check.IsSuccessful() && !check.IsFailed() {
				return false, fmt.Sprintf("check '%s' is not complete", check.Name), nil
			}
			if check.IsFailed() {
				return false, fmt.Sprintf("check '%s' failed", check.Name), nil
			}
		}
	}

	// Unused variables to satisfy compiler
	_ = owner
	_ = name

	return true, "ready to merge", nil
}

// GetBranchProtection gets the branch protection rules for a branch
func (p *Provider) GetBranchProtection(ctx context.Context, owner, name, branch string) (*github.Protection, error) {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	protection, _, err := p.client.Repositories.GetBranchProtection(ctx, owner, name, branch)
	if err != nil {
		// Branch protection might not exist, which is not an error
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, p.handleGitHubError("failed to get branch protection", err)
	}

	return protection, nil
}
