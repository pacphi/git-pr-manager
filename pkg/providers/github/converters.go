package github

import (
	"strconv"
	"time"

	"github.com/google/go-github/v57/github"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
)

// convertRepository converts a GitHub repository to common.Repository
func (p *Provider) convertRepository(repo *github.Repository) common.Repository {
	var pushedAt time.Time
	if repo.PushedAt != nil {
		pushedAt = repo.PushedAt.Time
	}

	visibility := common.VisibilityPublic
	if repo.GetPrivate() {
		visibility = common.VisibilityPrivate
	}

	return common.Repository{
		ID:            strconv.FormatInt(repo.GetID(), 10),
		Name:          repo.GetName(),
		FullName:      repo.GetFullName(),
		Description:   repo.GetDescription(),
		Language:      repo.GetLanguage(),
		Visibility:    visibility,
		DefaultBranch: repo.GetDefaultBranch(),
		CreatedAt:     repo.GetCreatedAt().Time,
		UpdatedAt:     repo.GetUpdatedAt().Time,
		PushedAt:      pushedAt,
		StarCount:     repo.GetStargazersCount(),
		ForkCount:     repo.GetForksCount(),
		IsArchived:    repo.GetArchived(),
		IsDisabled:    repo.GetDisabled(),
		IsFork:        repo.GetFork(),
		IsPrivate:     repo.GetPrivate(),
		HasIssues:     repo.GetHasIssues(),
		HasWiki:       repo.GetHasWiki(),
		Owner: common.User{
			ID:    strconv.FormatInt(repo.GetOwner().GetID(), 10),
			Login: repo.GetOwner().GetLogin(),
			Name:  repo.GetOwner().GetName(),
			Email: repo.GetOwner().GetEmail(),
			Type:  repo.GetOwner().GetType(),
		},
		CloneURL: repo.GetCloneURL(),
		SSHURL:   repo.GetSSHURL(),
		WebURL:   repo.GetHTMLURL(),
		Topics:   repo.Topics,
	}
}

// convertPullRequest converts a GitHub pull request to common.PullRequest
func (p *Provider) convertPullRequest(pr *github.PullRequest) common.PullRequest {
	var mergedAt, closedAt *time.Time
	if pr.MergedAt != nil {
		mergedAt = &pr.MergedAt.Time
	}
	if pr.ClosedAt != nil {
		closedAt = &pr.ClosedAt.Time
	}

	// Convert state
	var state common.PRState
	switch pr.GetState() {
	case "open":
		state = common.PRStateOpen
	case "closed":
		if pr.GetMerged() {
			state = common.PRStateMerged
		} else {
			state = common.PRStateClosed
		}
	default:
		state = common.PRStateOpen
	}

	// Convert assignees
	assignees := make([]common.User, 0, len(pr.Assignees))
	for _, assignee := range pr.Assignees {
		assignees = append(assignees, p.convertUser(assignee))
	}

	// Convert requested reviewers
	reviewers := make([]common.User, 0, len(pr.RequestedReviewers))
	for _, reviewer := range pr.RequestedReviewers {
		reviewers = append(reviewers, p.convertUser(reviewer))
	}

	// Convert labels
	labels := make([]common.Label, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		labels = append(labels, p.convertLabel(label))
	}

	// Convert milestone
	var milestone *common.Milestone
	if pr.Milestone != nil {
		milestone = p.convertMilestone(pr.Milestone)
	}

	commonPR := common.PullRequest{
		ID:             strconv.FormatInt(pr.GetID(), 10),
		Number:         pr.GetNumber(),
		Title:          pr.GetTitle(),
		Body:           pr.GetBody(),
		State:          state,
		Author:         p.convertUser(pr.GetUser()),
		Assignees:      assignees,
		Reviewers:      reviewers,
		Labels:         labels,
		Milestone:      milestone,
		BaseBranch:     pr.GetBase().GetRef(),
		HeadBranch:     pr.GetHead().GetRef(),
		HeadSHA:        pr.GetHead().GetSHA(),
		URL:            pr.GetHTMLURL(),
		DiffURL:        pr.GetDiffURL(),
		PatchURL:       pr.GetPatchURL(),
		MergeCommitSHA: pr.GetMergeCommitSHA(),
		MergedAt:       mergedAt,
		ClosedAt:       closedAt,
		CreatedAt:      pr.GetCreatedAt().Time,
		UpdatedAt:      pr.GetUpdatedAt().Time,
		Draft:          pr.GetDraft(),
		Locked:         pr.GetLocked(),
		Comments:       pr.GetComments(),
		Commits:        pr.GetCommits(),
		Additions:      pr.GetAdditions(),
		Deletions:      pr.GetDeletions(),
		ChangedFiles:   pr.GetChangedFiles(),
		Metadata: map[string]interface{}{
			"mergeable":             pr.GetMergeable(),
			"mergeable_state":       pr.GetMergeableState(),
			"rebaseable":            pr.GetRebaseable(),
			"review_comments":       pr.GetReviewComments(),
			"maintainer_can_modify": pr.GetMaintainerCanModify(),
		},
	}

	// Set mergeable status if available
	if pr.Mergeable != nil {
		commonPR.Mergeable = pr.Mergeable
	}

	return commonPR
}

// convertUser converts a GitHub user to common.User
func (p *Provider) convertUser(user *github.User) common.User {
	if user == nil {
		return common.User{}
	}

	return common.User{
		ID:        strconv.FormatInt(user.GetID(), 10),
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		AvatarURL: user.GetAvatarURL(),
		URL:       user.GetHTMLURL(),
		Type:      user.GetType(),
	}
}

// convertLabel converts a GitHub label to common.Label
func (p *Provider) convertLabel(label *github.Label) common.Label {
	if label == nil {
		return common.Label{}
	}

	return common.Label{
		ID:          strconv.FormatInt(label.GetID(), 10),
		Name:        label.GetName(),
		Description: label.GetDescription(),
		Color:       label.GetColor(),
	}
}

// convertMilestone converts a GitHub milestone to common.Milestone
func (p *Provider) convertMilestone(milestone *github.Milestone) *common.Milestone {
	if milestone == nil {
		return nil
	}

	var dueOn, closedAt *time.Time
	if milestone.DueOn != nil {
		dueOn = &milestone.DueOn.Time
	}
	if milestone.ClosedAt != nil {
		closedAt = &milestone.ClosedAt.Time
	}

	return &common.Milestone{
		ID:          strconv.FormatInt(milestone.GetID(), 10),
		Title:       milestone.GetTitle(),
		Description: milestone.GetDescription(),
		State:       milestone.GetState(),
		DueOn:       dueOn,
		ClosedAt:    closedAt,
		CreatedAt:   milestone.GetCreatedAt().Time,
		UpdatedAt:   milestone.GetUpdatedAt().Time,
	}
}
