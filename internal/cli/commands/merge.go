package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cphillipson/multi-gitter-pr-automation/internal/executor"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/merge"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/notifications"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/pr"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// MergeFlags contains flags for the merge command
type MergeFlags struct {
	Providers       []string
	Repositories    []string
	MaxAge          string
	DryRun          bool
	Force           bool
	DeleteBranches  bool
	CustomMessage   string
	RequireApproval bool
	Confirm         bool
}

// NewMergeCommand creates the merge command
func NewMergeCommand() *cobra.Command {
	var flags MergeFlags

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge ready pull requests",
		Long: `Merge pull requests that are ready across all configured repositories.

This command first checks the status of all PRs, then merges those that meet
the configured criteria and are ready for merging.`,
		Example: `  # Merge all ready PRs (dry run by default)
  git-pr-cli merge --dry-run

  # Actually merge ready PRs
  git-pr-cli merge

  # Force merge (skip readiness checks)
  git-pr-cli merge --force

  # Delete branches after merge
  git-pr-cli merge --delete-branches

  # Merge specific repositories
  git-pr-cli merge --repos owner/repo1,owner/repo2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMergeCommand(cmd.Context(), flags)
		},
	}

	// Add flags
	cmd.Flags().StringSliceVarP(&flags.Providers, "providers", "p", nil, "Comma-separated list of providers to check (github,gitlab,bitbucket)")
	cmd.Flags().StringSliceVarP(&flags.Repositories, "repos", "r", nil, "Comma-separated list of repositories to check")
	cmd.Flags().StringVar(&flags.MaxAge, "max-age", "", "Maximum age of PRs to consider (e.g., 7d, 24h)")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Show what would be merged without actually merging")
	cmd.Flags().BoolVar(&flags.Force, "force", false, "Force merge even if PR is not ready")
	cmd.Flags().BoolVar(&flags.DeleteBranches, "delete-branches", false, "Delete branches after successful merge")
	cmd.Flags().StringVar(&flags.CustomMessage, "message", "", "Custom commit message for merge")
	cmd.Flags().BoolVar(&flags.RequireApproval, "require-approval", false, "Require manual approval before merging")
	cmd.Flags().BoolVarP(&flags.Confirm, "confirm", "y", false, "Confirm merge operation without prompting")

	return cmd
}

func runMergeCommand(ctx context.Context, flags MergeFlags) error {
	logger := utils.GetGlobalLogger()
	logger.Info("Starting PR merge operation")

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return HandleConfigError(err, "merge")
	}

	// Create executor
	exec, err := executor.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	defer func() {
		if closeErr := exec.Close(); closeErr != nil {
			logger.WithError(closeErr).Warn("Failed to close executor")
		}
	}()

	// Parse max age
	var maxAge time.Duration
	if flags.MaxAge != "" {
		maxAge, err = utils.ParseDuration(flags.MaxAge)
		if err != nil {
			return fmt.Errorf("invalid max-age: %w", err)
		}
	}

	// Set up processing options
	processOpts := pr.ProcessOptions{
		DryRun:        false, // Always process for real to get accurate status
		Providers:     flags.Providers,
		Repositories:  flags.Repositories,
		MaxAge:        maxAge,
		RequireChecks: true, // Always require checks for merge operations
	}

	// Process PRs to find ready ones
	logger.Info("Checking PR status...")
	results, err := exec.ProcessPRs(ctx, processOpts)
	if err != nil {
		return fmt.Errorf("failed to process PRs: %w", err)
	}

	// Count ready PRs
	readyCount := 0
	for _, result := range results {
		for _, pr := range result.PullRequests {
			if pr.Ready && !pr.Skipped && pr.Error == nil {
				readyCount++
			}
		}
	}

	if readyCount == 0 {
		logger.Info("No PRs ready for merging")
		return nil
	}

	logger.Infof("Found %d PRs ready for merging", readyCount)

	// Confirm merge operation if not in dry-run mode and not auto-confirmed
	if !flags.DryRun && !flags.Confirm {
		if !confirmMerge(readyCount) {
			logger.Info("Merge operation cancelled")
			return nil
		}
	}

	// Validate mergeability
	if !flags.Force {
		logger.Info("Validating mergeability...")
		if err := exec.ValidateMergeability(ctx, results); err != nil {
			return fmt.Errorf("merge validation failed: %w", err)
		}
	}

	// Set up merge options
	mergeOpts := merge.MergeOptions{
		DryRun:          flags.DryRun,
		Force:           flags.Force,
		DeleteBranches:  flags.DeleteBranches,
		CustomMessage:   flags.CustomMessage,
		RequireApproval: flags.RequireApproval,
	}

	// Perform merges
	logger.Info("Starting merge operations...")
	mergeResults, err := exec.MergePRs(ctx, results, mergeOpts)
	if err != nil {
		return fmt.Errorf("merge operations failed: %w", err)
	}

	// Send notifications if configured
	notificationManager, err := notifications.NewManager(cfg)
	if err != nil {
		logger.WithError(err).Warn("Failed to create notification manager")
	} else if notificationManager.HasNotifiers() {
		logger.Info("Sending merge result notifications...")
		if err := notificationManager.SendMergeResults(ctx, mergeResults); err != nil {
			logger.WithError(err).Warn("Failed to send merge result notifications")
		} else {
			logger.Info("Merge result notifications sent successfully")
		}
	}

	// Output results
	outputMergeResults(mergeResults, flags.DryRun)
	return nil
}

func confirmMerge(count int) bool {
	fmt.Printf("\nAbout to merge %d pull request(s). Continue? [y/N]: ", count)
	var response string
	_, _ = fmt.Scanln(&response) // Ignore read errors, treat as "no"
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}

func outputMergeResults(results []merge.MergeResult, dryRun bool) {
	successful := 0
	skipped := 0
	failed := 0

	action := "merged"
	if dryRun {
		action = "would merge"
	}

	fmt.Printf("\nMerge Results\n")
	fmt.Printf("=============\n\n")

	for _, result := range results {
		if result.Skipped {
			skipped++
			fmt.Printf("⏭️  SKIPPED: PR #%d in %s - %s\n", result.PullRequest, result.Repository, result.Reason)
		} else if result.Success {
			successful++
			if dryRun {
				fmt.Printf("✅ WOULD MERGE: PR #%d in %s using %s\n", result.PullRequest, result.Repository, result.MergeMethod)
			} else {
				fmt.Printf("✅ MERGED: PR #%d in %s using %s\n", result.PullRequest, result.Repository, result.MergeMethod)
			}
		} else {
			failed++
			fmt.Printf("❌ FAILED: PR #%d in %s - %s\n", result.PullRequest, result.Repository, result.Error.Error())
		}
	}

	fmt.Printf("\nSummary: %s %d PRs, skipped %d, failed %d\n", action, successful, skipped, failed)
}
