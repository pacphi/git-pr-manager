package bitbucket

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pacphi/git-pr-manager/pkg/providers/common"
)

// handleBitbucketError converts Bitbucket API errors to common provider errors
func (p *Provider) handleBitbucketError(message string, err error) error {
	if err == nil {
		return nil
	}

	// Handle HTTP errors
	if strings.Contains(err.Error(), "context canceled") {
		return common.NewProviderError(ProviderName, common.ErrorTypeNetwork, message, err)
	}

	if strings.Contains(err.Error(), "timeout") {
		return common.NewProviderError(ProviderName, common.ErrorTypeNetwork, message, err)
	}

	// Try to extract HTTP status from error message
	errorType := common.ErrorTypeUnknown
	if strings.Contains(err.Error(), "401") {
		errorType = common.ErrorTypeAuth
	} else if strings.Contains(err.Error(), "403") {
		errorType = common.ErrorTypePermission
	} else if strings.Contains(err.Error(), "404") {
		errorType = common.ErrorTypeNotFound
	} else if strings.Contains(err.Error(), "409") {
		errorType = common.ErrorTypeConflict
	} else if strings.Contains(err.Error(), "422") {
		errorType = common.ErrorTypeValidation
	} else if strings.Contains(err.Error(), "429") {
		errorType = common.ErrorTypeRateLimit
	} else if strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "502") ||
		strings.Contains(err.Error(), "503") || strings.Contains(err.Error(), "504") {
		errorType = common.ErrorTypeNetwork
	}

	providerError := common.NewProviderError(ProviderName, errorType, message, err)
	return providerError
}

// convertRepository converts a Bitbucket repository to common repository format
func (p *Provider) convertRepository(repo *BitbucketRepository) common.Repository {
	visibility := common.VisibilityPublic
	if repo.IsPrivate {
		visibility = common.VisibilityPrivate
	}

	// Get clone URLs
	var cloneURL, sshURL string
	for _, link := range repo.Links.Clone {
		switch link.Name {
		case "https":
			cloneURL = link.Href
		case "ssh":
			sshURL = link.Href
		}
	}

	return common.Repository{
		ID:            repo.UUID,
		Name:          repo.Name,
		FullName:      repo.FullName,
		Description:   repo.Description,
		Language:      repo.Language,
		Visibility:    visibility,
		DefaultBranch: repo.MainBranch.Name,
		CreatedAt:     repo.CreatedOn,
		UpdatedAt:     repo.UpdatedOn,
		PushedAt:      repo.UpdatedOn,
		StarCount:     0,     // Bitbucket doesn't expose star count in repository API
		ForkCount:     0,     // Bitbucket doesn't expose fork count in repository API
		IsArchived:    false, // Bitbucket doesn't have archived concept in API
		IsDisabled:    false,
		IsFork:        false, // Would need additional API call to determine
		IsPrivate:     repo.IsPrivate,
		HasIssues:     true, // Bitbucket has issues by default
		HasWiki:       true, // Bitbucket has wiki by default
		Owner: common.User{
			ID:    repo.Owner.UUID,
			Login: repo.Owner.Username,
			Name:  repo.Owner.DisplayName,
			Email: "",
			Type:  repo.Owner.Type,
		},
		CloneURL: cloneURL,
		SSHURL:   sshURL,
		WebURL:   repo.Links.HTML.Href,
		Topics:   []string{}, // Bitbucket doesn't have topics
	}
}

// convertPullRequest converts a Bitbucket pull request to common pull request format
func (p *Provider) convertPullRequest(pr *BitbucketPullRequest) common.PullRequest {
	state := common.PRStateOpen
	switch strings.ToLower(pr.State) {
	case "merged":
		state = common.PRStateMerged
	case "declined":
		state = common.PRStateClosed
	case "superseded":
		state = common.PRStateClosed
	}

	// Bitbucket doesn't provide mergeable status in PR list, assume mergeable if open
	var mergeable *bool
	if state == common.PRStateOpen {
		t := true
		mergeable = &t
	}

	return common.PullRequest{
		ID:         strconv.Itoa(pr.ID),
		Number:     pr.ID,
		Title:      pr.Title,
		Body:       pr.Description,
		State:      state,
		Draft:      false, // Bitbucket doesn't have draft PRs in the same way
		Mergeable:  mergeable,
		Locked:     false,
		CreatedAt:  pr.CreatedOn,
		UpdatedAt:  pr.UpdatedOn,
		MergedAt:   pr.MergedOn,
		ClosedAt:   pr.ClosedOn,
		HeadBranch: pr.Source.Branch.Name,
		BaseBranch: pr.Destination.Branch.Name,
		HeadSHA:    pr.Source.Commit.Hash,
		Author: common.User{
			ID:    pr.Author.UUID,
			Login: pr.Author.Username,
			Name:  pr.Author.DisplayName,
			Email: "",
			Type:  pr.Author.Type,
		},
		Labels:   []common.Label{}, // Bitbucket doesn't have PR labels
		URL:      pr.Links.HTML.Href,
		DiffURL:  pr.Links.Diff.Href,
		PatchURL: pr.Links.Diff.Href + "?format=patch",
	}
}

// deleteBranch deletes a branch from the repository
func (p *Provider) deleteBranch(ctx context.Context, owner, repo, branch string) error {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	url := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", owner, repo, branch)
	err := p.httpClient.Delete(ctx, url, nil)
	if err != nil {
		return p.handleBitbucketError("failed to delete branch", err)
	}

	p.logger.Debugf("Deleted branch %s in %s/%s", branch, owner, repo)
	return nil
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

	// Check if PR is mergeable
	if pr.Mergeable != nil && !*pr.Mergeable {
		return false, "pull request has conflicts", nil
	}

	// Check pipelines if required
	if requireChecks {
		status, err := p.GetPRStatus(ctx, repo, pr)
		if err != nil {
			return false, "failed to get PR status", err
		}

		if !status.IsSuccessful() {
			return false, fmt.Sprintf("pipelines not passing: %s", status.State), nil
		}

		// Also check individual pipelines
		checks, err := p.GetChecks(ctx, repo, pr)
		if err != nil {
			return false, "failed to get checks", err
		}

		for _, check := range checks {
			if !check.IsSuccessful() && !check.IsFailed() {
				return false, fmt.Sprintf("pipeline '%s' is not complete", check.Name), nil
			}
			if check.IsFailed() {
				return false, fmt.Sprintf("pipeline '%s' failed", check.Name), nil
			}
		}
	}

	// Unused variables to satisfy compiler
	_ = owner
	_ = name

	return true, "ready to merge", nil
}
