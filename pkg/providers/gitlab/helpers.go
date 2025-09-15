package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.com/gitlab-org/api/client-go"

	"github.com/pacphi/git-pr-manager/pkg/providers/common"
)

// handleGitLabError converts GitLab API errors to common provider errors
func (p *Provider) handleGitLabError(message string, err error) error {
	if err == nil {
		return nil
	}

	// Handle GitLab-specific errors
	if glErr, ok := err.(*gitlab.ErrorResponse); ok {
		return p.convertGitLabError(message, glErr)
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

// convertGitLabError converts GitLab ErrorResponse to provider error
func (p *Provider) convertGitLabError(message string, glErr *gitlab.ErrorResponse) error {
	var errorType common.ErrorType

	switch glErr.Response.StatusCode {
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

	providerError := common.NewProviderError(ProviderName, errorType, message, glErr)
	providerError.StatusCode = glErr.Response.StatusCode

	// Add retry after header if present for rate limiting
	if errorType == common.ErrorTypeRateLimit {
		if retryAfter := glErr.Response.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				retryDuration := time.Duration(seconds) * time.Second
				providerError.RetryAfter = &retryDuration
			}
		}
	}

	// Add error message from GitLab response
	if glErr.Message != "" {
		providerError.Message = fmt.Sprintf("%s: %s", message, glErr.Message)
	}

	return providerError
}

// convertProject converts a GitLab project to common repository format
func (p *Provider) convertProject(project *gitlab.Project) common.Repository {
	var visibility common.Visibility
	switch project.Visibility {
	case gitlab.PrivateVisibility:
		visibility = common.VisibilityPrivate
	case gitlab.InternalVisibility:
		visibility = common.VisibilityInternal
	default:
		visibility = common.VisibilityPublic
	}

	var language string
	if project.DefaultBranch != "" {
		// Try to get the main language from project statistics
		// This is a simplified approach as GitLab's language detection is complex
		language = "Unknown"
	}

	return common.Repository{
		ID:            strconv.Itoa(project.ID),
		Name:          project.Name,
		FullName:      project.PathWithNamespace,
		Description:   project.Description,
		Language:      language,
		Visibility:    visibility,
		DefaultBranch: project.DefaultBranch,
		CreatedAt: func() time.Time {
			if project.CreatedAt != nil {
				return *project.CreatedAt
			}
			return time.Time{}
		}(),
		UpdatedAt: func() time.Time {
			if project.LastActivityAt != nil {
				return *project.LastActivityAt
			}
			return time.Time{}
		}(),
		PushedAt: func() time.Time {
			if project.LastActivityAt != nil {
				return *project.LastActivityAt
			}
			return time.Time{}
		}(),
		StarCount:  project.StarCount,
		ForkCount:  project.ForksCount,
		IsArchived: project.Archived,
		IsDisabled: false, // GitLab doesn't have disabled concept
		IsFork:     project.ForkedFromProject != nil,
		IsPrivate:  project.Visibility == gitlab.PrivateVisibility,
		HasIssues:  project.IssuesAccessLevel != "disabled",
		HasWiki:    project.WikiAccessLevel != "disabled",
		Owner: func() common.User {
			if project.Owner != nil {
				return common.User{
					ID:    strconv.Itoa(project.Owner.ID),
					Login: project.Owner.Username,
					Name:  project.Owner.Name,
					Email: project.Owner.Email,
					Type:  "User",
				}
			}
			return common.User{
				ID:    "0",
				Login: "unknown",
				Name:  "Unknown",
				Email: "",
				Type:  "User",
			}
		}(),
		CloneURL: project.HTTPURLToRepo,
		SSHURL:   project.SSHURLToRepo,
		WebURL:   project.WebURL,
		Topics:   project.Topics,
	}
}

// convertBasicMergeRequest converts a GitLab basic merge request to common pull request format
func (p *Provider) convertBasicMergeRequest(mr *gitlab.BasicMergeRequest) common.PullRequest {
	// Convert BasicMergeRequest to the common format
	// Note: BasicMergeRequest has limited fields compared to full MergeRequest
	var state common.PRState
	switch mr.State {
	case "merged":
		state = common.PRStateMerged
	case "closed":
		state = common.PRStateClosed
	default:
		state = common.PRStateOpen
	}

	return common.PullRequest{
		ID:     strconv.Itoa(mr.IID),
		Number: mr.IID,
		Title:  mr.Title,
		Body:   mr.Description,
		State:  state,
		Author: func() common.User {
			if mr.Author != nil {
				return common.User{
					ID:    strconv.Itoa(mr.Author.ID),
					Login: mr.Author.Username,
					Name:  mr.Author.Name,
					Email: "",
					Type:  "User",
				}
			}
			return common.User{
				ID:    "0",
				Login: "unknown",
				Name:  "Unknown",
				Email: "",
				Type:  "User",
			}
		}(),
		HeadSHA: mr.SHA,
		Labels:  p.convertLabelsFromStrings(mr.Labels),
		CreatedAt: func() time.Time {
			if mr.CreatedAt != nil {
				return *mr.CreatedAt
			}
			return time.Time{}
		}(),
		UpdatedAt: func() time.Time {
			if mr.UpdatedAt != nil {
				return *mr.UpdatedAt
			}
			return time.Time{}
		}(),
		// BasicMergeRequest has limited fields, so many will be empty
		BaseBranch: mr.TargetBranch,
		HeadBranch: mr.SourceBranch,
		URL:        mr.WebURL,
	}
}

// convertMergeRequest converts a GitLab merge request to common pull request format
func (p *Provider) convertMergeRequest(mr *gitlab.MergeRequest) common.PullRequest {
	var state common.PRState
	switch mr.State {
	case "merged":
		state = common.PRStateMerged
	case "closed":
		state = common.PRStateClosed
	default:
		state = common.PRStateOpen
	}

	var mergeable *bool
	switch mr.DetailedMergeStatus {
	case "mergeable":
		t := true
		mergeable = &t
	case "not_mergeable":
		f := false
		mergeable = &f
	}

	// Convert labels
	labels := make([]common.Label, 0, len(mr.Labels))
	for _, label := range mr.Labels {
		labels = append(labels, common.Label{
			Name:        label,
			Description: "",
			Color:       "",
		})
	}

	return common.PullRequest{
		ID:        strconv.Itoa(mr.IID),
		Number:    mr.IID,
		Title:     mr.Title,
		Body:      mr.Description,
		State:     state,
		Draft:     mr.Draft,
		Mergeable: mergeable,
		Locked:    false, // GitLab doesn't have locked concept
		CreatedAt: func() time.Time {
			if mr.CreatedAt != nil {
				return *mr.CreatedAt
			}
			return time.Time{}
		}(),
		UpdatedAt: func() time.Time {
			if mr.UpdatedAt != nil {
				return *mr.UpdatedAt
			}
			return time.Time{}
		}(),
		MergedAt:   mr.MergedAt,
		ClosedAt:   mr.ClosedAt,
		HeadBranch: mr.SourceBranch,
		BaseBranch: mr.TargetBranch,
		HeadSHA:    mr.SHA,
		Author: func() common.User {
			if mr.Author != nil {
				return common.User{
					ID:    strconv.Itoa(mr.Author.ID),
					Login: mr.Author.Username,
					Name:  mr.Author.Name,
					Email: "",
					Type:  "User",
				}
			}
			return common.User{
				ID:    "0",
				Login: "unknown",
				Name:  "Unknown",
				Email: "",
				Type:  "User",
			}
		}(),
		Labels:   labels,
		URL:      mr.WebURL,
		DiffURL:  fmt.Sprintf("%s.diff", mr.WebURL),
		PatchURL: fmt.Sprintf("%s.patch", mr.WebURL),
	}
}

// IsMergeRequestReady checks if a merge request is ready to be merged
func (p *Provider) IsMergeRequestReady(ctx context.Context, repo common.Repository, mr common.PullRequest, requireChecks bool) (bool, string, error) {
	owner, name, err := common.ParseRepository(repo.FullName)
	if err != nil {
		return false, "invalid repository name", err
	}

	// Check if MR is open
	if !mr.IsOpen() {
		return false, "merge request is not open", nil
	}

	// Check if MR is a draft
	if mr.IsDraft() {
		return false, "merge request is a draft", nil
	}

	// Check if MR is mergeable
	if mr.Mergeable != nil && !*mr.Mergeable {
		return false, "merge request has conflicts", nil
	}

	// Check pipelines if required
	if requireChecks {
		status, err := p.GetPRStatus(ctx, repo, mr)
		if err != nil {
			return false, "failed to get MR status", err
		}

		if !status.IsSuccessful() {
			return false, fmt.Sprintf("pipelines not passing: %s", status.State), nil
		}

		// Also check individual pipelines
		checks, err := p.GetChecks(ctx, repo, mr)
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

// convertLabelsFromStrings converts string labels to common.Label
func (p *Provider) convertLabelsFromStrings(labels []string) []common.Label {
	result := make([]common.Label, 0, len(labels))
	for _, label := range labels {
		result = append(result, common.Label{
			Name:        label,
			Description: "",
			Color:       "",
		})
	}
	return result
}
