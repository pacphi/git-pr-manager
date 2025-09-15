package common

import (
	"context"
	"fmt"
	"time"
)

// Provider defines the interface that all Git providers must implement
type Provider interface {
	// Authentication and setup
	Authenticate(ctx context.Context) error
	GetProviderName() string

	// Repository operations
	ListRepositories(ctx context.Context) ([]Repository, error)
	GetRepository(ctx context.Context, owner, name string) (*Repository, error)

	// Pull request operations
	ListPullRequests(ctx context.Context, repo Repository, opts ListPROptions) ([]PullRequest, error)
	GetPullRequest(ctx context.Context, repo Repository, number int) (*PullRequest, error)
	MergePullRequest(ctx context.Context, repo Repository, pr PullRequest, opts MergeOptions) error

	// Status and check operations
	GetPRStatus(ctx context.Context, repo Repository, pr PullRequest) (*PRStatus, error)
	GetChecks(ctx context.Context, repo Repository, pr PullRequest) ([]Check, error)

	// Utility operations
	GetRateLimit(ctx context.Context) (*RateLimit, error)
}

// Visibility represents repository visibility
type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityInternal Visibility = "internal"
)

// Repository represents a repository in any Git provider
type Repository struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	FullName      string     `json:"full_name"`
	Provider      string     `json:"provider"`
	Description   string     `json:"description"`
	Language      string     `json:"language"`
	Visibility    Visibility `json:"visibility"`
	DefaultBranch string     `json:"default_branch"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	PushedAt      time.Time  `json:"pushed_at"`
	StarCount     int        `json:"star_count"`
	ForkCount     int        `json:"fork_count"`
	IsArchived    bool       `json:"is_archived"`
	IsDisabled    bool       `json:"is_disabled"`
	IsFork        bool       `json:"is_fork"`
	IsPrivate     bool       `json:"is_private"`
	HasIssues     bool       `json:"has_issues"`
	HasWiki       bool       `json:"has_wiki"`
	Owner         User       `json:"owner"`
	CloneURL      string     `json:"clone_url"`
	SSHURL        string     `json:"ssh_url"`
	WebURL        string     `json:"web_url"`
	Topics        []string   `json:"topics"`
}

// PullRequest represents a pull request in any Git provider
type PullRequest struct {
	ID             string                 `json:"id"`
	Number         int                    `json:"number"`
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	State          PRState                `json:"state"`
	Author         User                   `json:"author"`
	Assignees      []User                 `json:"assignees"`
	Reviewers      []User                 `json:"reviewers"`
	Labels         []Label                `json:"labels"`
	Milestone      *Milestone             `json:"milestone,omitempty"`
	BaseBranch     string                 `json:"base_branch"`
	HeadBranch     string                 `json:"head_branch"`
	HeadSHA        string                 `json:"head_sha"`
	URL            string                 `json:"url"`
	DiffURL        string                 `json:"diff_url"`
	PatchURL       string                 `json:"patch_url"`
	Mergeable      *bool                  `json:"mergeable,omitempty"`
	MergeCommitSHA string                 `json:"merge_commit_sha,omitempty"`
	MergedAt       *time.Time             `json:"merged_at,omitempty"`
	ClosedAt       *time.Time             `json:"closed_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Draft          bool                   `json:"draft"`
	Locked         bool                   `json:"locked"`
	Comments       int                    `json:"comments"`
	Commits        int                    `json:"commits"`
	Additions      int                    `json:"additions"`
	Deletions      int                    `json:"deletions"`
	ChangedFiles   int                    `json:"changed_files"`
	StatusChecks   []Check                `json:"status_checks,omitempty"`
	PassedChecks   int                    `json:"passed_checks"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"` // Provider-specific metadata
}

// User represents a user in any Git provider
type User struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	Type      string `json:"type"`
}

// Label represents a label in any Git provider
type Label struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Milestone represents a milestone in any Git provider
type Milestone struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	DueOn       *time.Time `json:"due_on,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PRStatus represents the overall status of a pull request
type PRStatus struct {
	State       PRStatusState `json:"state"`
	Description string        `json:"description"`
	TargetURL   string        `json:"target_url,omitempty"`
	Context     string        `json:"context"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Check represents a status check on a pull request
type Check struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Status      CheckStatus `json:"status"`
	Conclusion  string      `json:"conclusion,omitempty"`
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	DetailsURL  string      `json:"details_url,omitempty"`
	Summary     string      `json:"summary,omitempty"`
	Text        string      `json:"text,omitempty"`
}

// RateLimit represents API rate limiting information
type RateLimit struct {
	Limit      int       `json:"limit"`
	Remaining  int       `json:"remaining"`
	ResetTime  time.Time `json:"reset_time"`
	RetryAfter *int      `json:"retry_after,omitempty"`
}

// PRState represents the state of a pull request
type PRState string

const (
	PRStateOpen   PRState = "open"
	PRStateClosed PRState = "closed"
	PRStateMerged PRState = "merged"
)

// PRStatusState represents the state of a pull request status
type PRStatusState string

const (
	PRStatusPending PRStatusState = "pending"
	PRStatusSuccess PRStatusState = "success"
	PRStatusError   PRStatusState = "error"
	PRStatusFailure PRStatusState = "failure"
)

// CheckStatus represents the status of a check
type CheckStatus string

const (
	CheckStatusQueued     CheckStatus = "queued"
	CheckStatusInProgress CheckStatus = "in_progress"
	CheckStatusCompleted  CheckStatus = "completed"
)

// ListPROptions contains options for listing pull requests
type ListPROptions struct {
	State     PRState   `json:"state,omitempty"`
	Base      string    `json:"base,omitempty"`
	Head      string    `json:"head,omitempty"`
	Sort      string    `json:"sort,omitempty"`
	Direction string    `json:"direction,omitempty"`
	Page      int       `json:"page,omitempty"`
	PerPage   int       `json:"per_page,omitempty"`
	Since     time.Time `json:"since,omitempty"`
}

// MergeOptions contains options for merging pull requests
type MergeOptions struct {
	Method        MergeMethod `json:"method"`
	CommitTitle   string      `json:"commit_title,omitempty"`
	CommitMessage string      `json:"commit_message,omitempty"`
	SHA           string      `json:"sha,omitempty"`
	DeleteBranch  bool        `json:"delete_branch"`
}

// MergeMethod represents different merge methods
type MergeMethod string

const (
	MergeMethodMerge  MergeMethod = "merge"
	MergeMethodSquash MergeMethod = "squash"
	MergeMethodRebase MergeMethod = "rebase"
)

// Error types for common provider errors
type ErrorType string

const (
	ErrorTypeAuth       ErrorType = "authentication"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeRateLimit  ErrorType = "rate_limit"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeConflict   ErrorType = "conflict"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeUnknown    ErrorType = "unknown"
)

// ProviderError represents a provider-specific error
type ProviderError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Provider   string                 `json:"provider"`
	StatusCode int                    `json:"status_code,omitempty"`
	RetryAfter *time.Duration         `json:"retry_after,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Original   error                  `json:"-"`
}

func (e *ProviderError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s provider error (%s): %s (original: %v)", e.Provider, e.Type, e.Message, e.Original)
	}
	return fmt.Sprintf("%s provider error (%s): %s", e.Provider, e.Type, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Original
}

// NewProviderError creates a new provider error
func NewProviderError(provider string, errorType ErrorType, message string, original error) *ProviderError {
	return &ProviderError{
		Type:     errorType,
		Message:  message,
		Provider: provider,
		Original: original,
	}
}
