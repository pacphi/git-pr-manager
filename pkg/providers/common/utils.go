package common

import (
	"fmt"
	"strings"
	"time"
)

// IsBot checks if a user is a bot based on common patterns
func (u *User) IsBot() bool {
	if u.Type == "bot" || u.Type == "Bot" {
		return true
	}

	lowerLogin := strings.ToLower(u.Login)
	botPatterns := []string{
		"[bot]",
		"bot",
		"dependabot",
		"renovate",
		"greenkeeper",
		"snyk",
		"github-actions",
		"codecov",
	}

	for _, pattern := range botPatterns {
		if strings.Contains(lowerLogin, pattern) {
			return true
		}
	}

	return false
}

// HasLabel checks if a pull request has a specific label
func (pr *PullRequest) HasLabel(labelName string) bool {
	for _, label := range pr.Labels {
		if strings.EqualFold(label.Name, labelName) {
			return true
		}
	}
	return false
}

// HasAnyLabel checks if a pull request has any of the specified labels
func (pr *PullRequest) HasAnyLabel(labelNames []string) bool {
	for _, labelName := range labelNames {
		if pr.HasLabel(labelName) {
			return true
		}
	}
	return false
}

// IsOpen checks if the pull request is open
func (pr *PullRequest) IsOpen() bool {
	return pr.State == PRStateOpen
}

// IsClosed checks if the pull request is closed
func (pr *PullRequest) IsClosed() bool {
	return pr.State == PRStateClosed
}

// IsMerged checks if the pull request is merged
func (pr *PullRequest) IsMerged() bool {
	return pr.State == PRStateMerged || pr.MergedAt != nil
}

// IsDraft checks if the pull request is a draft
func (pr *PullRequest) IsDraft() bool {
	return pr.Draft
}

// Age returns the age of the pull request
func (pr *PullRequest) Age() time.Duration {
	return time.Since(pr.CreatedAt)
}

// IsOld checks if the pull request is older than the specified duration
func (pr *PullRequest) IsOld(maxAge time.Duration) bool {
	return pr.Age() > maxAge
}

// IsSuccessful checks if the status is successful
func (s *PRStatus) IsSuccessful() bool {
	return s.State == PRStatusSuccess
}

// IsPending checks if the status is pending
func (s *PRStatus) IsPending() bool {
	return s.State == PRStatusPending
}

// IsError checks if the status is in error or failure state
func (s *PRStatus) IsError() bool {
	return s.State == PRStatusError || s.State == PRStatusFailure
}

// IsCompleted checks if the check is completed
func (c *Check) IsCompleted() bool {
	return c.Status == CheckStatusCompleted
}

// IsSuccessful checks if the check is completed successfully
func (c *Check) IsSuccessful() bool {
	return c.IsCompleted() && (c.Conclusion == "success" || c.Conclusion == "neutral")
}

// IsFailed checks if the check has failed
func (c *Check) IsFailed() bool {
	return c.IsCompleted() && (c.Conclusion == "failure" || c.Conclusion == "cancelled" || c.Conclusion == "timed_out")
}

// IsRateLimited checks if we're currently rate limited
func (rl *RateLimit) IsRateLimited() bool {
	return rl.Remaining == 0 && time.Now().Before(rl.ResetTime)
}

// TimeToReset returns the duration until the rate limit resets
func (rl *RateLimit) TimeToReset() time.Duration {
	if time.Now().After(rl.ResetTime) {
		return 0
	}
	return time.Until(rl.ResetTime)
}

// ShouldRetry determines if an error is retryable
func (e *ProviderError) ShouldRetry() bool {
	switch e.Type {
	case ErrorTypeRateLimit, ErrorTypeNetwork:
		return true
	case ErrorTypeAuth, ErrorTypeNotFound, ErrorTypePermission, ErrorTypeValidation:
		return false
	default:
		return false
	}
}


// ParseRepository parses a repository identifier (owner/name) from a string
func ParseRepository(repoStr string) (owner, name string, err error) {
	parts := strings.Split(repoStr, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format, expected 'owner/name', got: %s", repoStr)
	}
	return parts[0], parts[1], nil
}

