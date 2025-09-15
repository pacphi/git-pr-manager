package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/pr"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// WatchFlags contains flags for the watch command
type WatchFlags struct {
	Interval      time.Duration
	ShowReady     bool
	ShowSkipped   bool
	ShowAll       bool
	Output        string
	ClearScreen   bool
	MaxAge        time.Duration
	RequireChecks bool
	MaxIterations int
	Provider      string
	Repos         string
}

// NewWatchCommand creates the watch command
func NewWatchCommand() *cobra.Command {
	var flags WatchFlags

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Continuously monitor pull request status",
		Long: `Continuously monitor pull request status across all configured repositories.

The watch mode refreshes the status periodically and displays real-time information
about pull requests that are ready to merge, skipped, or have errors.

Use Ctrl+C to stop monitoring.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(cmd.Context(), flags)
		},
	}

	cmd.Flags().DurationVarP(&flags.Interval, "interval", "i", 5*time.Minute, "refresh interval")
	cmd.Flags().BoolVar(&flags.ShowReady, "show-ready", true, "show ready PRs")
	cmd.Flags().BoolVar(&flags.ShowSkipped, "show-skipped", false, "show skipped PRs")
	cmd.Flags().BoolVar(&flags.ShowAll, "show-all", false, "show all PRs regardless of status")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "table", "output format (table, json)")
	cmd.Flags().BoolVar(&flags.ClearScreen, "clear", true, "clear screen between updates")
	cmd.Flags().DurationVar(&flags.MaxAge, "max-age", 0, "maximum age of PRs to include")
	cmd.Flags().BoolVar(&flags.RequireChecks, "require-checks", false, "require status checks to pass")
	cmd.Flags().IntVar(&flags.MaxIterations, "max-iterations", 0, "maximum number of check iterations (0 = infinite)")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", "filter by provider (github, gitlab, bitbucket)")
	cmd.Flags().StringVar(&flags.Repos, "repos", "", "filter repositories by pattern")

	return cmd
}

// runWatch runs the watch command
func runWatch(ctx context.Context, flags WatchFlags) error {
	// Validate flags
	if flags.MaxIterations < 0 {
		fmt.Fprintln(os.Stderr, "--max-iterations must be non-negative (0 for unlimited iterations)")
		os.Exit(1)
	}

	logger := utils.GetGlobalLogger()

	logger.Info("Starting PR monitoring...")
	logger.Infof("Refresh interval: %s", flags.Interval)
	logger.Info("Press Ctrl+C to stop monitoring")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create cancellable context
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signals
	go func() {
		<-sigChan
		logger.Info("Shutting down monitoring...")
		cancel()
	}()

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return HandleConfigError(err, "watch")
	}

	// Create providers
	factory := providers.NewFactory(cfg)
	providerMap, err := factory.CreateProviders()
	if err != nil {
		return fmt.Errorf("failed to create providers: %w", err)
	}

	// Filter providers if specified
	if flags.Provider != "" {
		filteredMap := make(map[string]common.Provider)
		for name, provider := range providerMap {
			if name == flags.Provider {
				filteredMap[name] = provider
			}
		}
		if len(filteredMap) == 0 {
			return fmt.Errorf("provider %s not found or not configured", flags.Provider)
		}
		providerMap = filteredMap
	}

	// Create processor
	processor := pr.NewProcessor(providerMap, cfg)

	// Start monitoring loop
	ticker := time.NewTicker(flags.Interval)
	defer ticker.Stop()

	// Track iterations if max-iterations is specified
	var iteration int
	maxIterations := flags.MaxIterations

	// Initial run
	iteration++
	if err := performCheck(watchCtx, processor, flags); err != nil {
		logger.WithError(err).Error("Initial check failed")
	}

	// Check if we should stop after the first iteration
	if maxIterations > 0 && iteration >= maxIterations {
		logger.Infof("Completed %d iteration(s), stopping", iteration)
		return nil
	}

	// Monitor loop
	for {
		select {
		case <-watchCtx.Done():
			logger.Info("Monitoring stopped")
			return nil

		case <-ticker.C:
			iteration++
			if err := performCheck(watchCtx, processor, flags); err != nil {
				logger.WithError(err).Error("Check failed")
			}

			// Check if we've reached the maximum iterations
			if maxIterations > 0 && iteration >= maxIterations {
				logger.Infof("Completed %d iteration(s), stopping", iteration)
				return nil
			}
		}
	}
}

// performCheck performs a single check iteration
func performCheck(ctx context.Context, processor *pr.Processor, flags WatchFlags) error {
	// Clear screen if requested
	if flags.ClearScreen {
		clearScreen()
	}

	// Show timestamp
	fmt.Printf("=== Git PR Automation Monitor ===\n")
	fmt.Printf("Last updated: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Refresh interval: %s\n\n", flags.Interval)

	// Process repositories
	processOpts := pr.ProcessOptions{
		DryRun:        true, // Watch mode is always dry-run
		RequireChecks: flags.RequireChecks,
	}

	if flags.MaxAge > 0 {
		processOpts.MaxAge = flags.MaxAge
	}

	// Add repository filtering if specified
	if flags.Repos != "" {
		processOpts.Repositories = parseRepoFilter(flags.Repos)
	}

	results, err := processor.ProcessAllPRs(ctx, processOpts)
	if err != nil {
		return fmt.Errorf("failed to process PRs: %w", err)
	}

	// Display results based on output format
	switch flags.Output {
	case "json":
		return displayWatchResultsJSON(results, flags)
	case "table":
		return displayWatchResultsTable(results, flags)
	default:
		return fmt.Errorf("unsupported output format: %s", flags.Output)
	}
}

// displayWatchResultsTable displays results in table format
func displayWatchResultsTable(results []pr.ProcessResult, flags WatchFlags) error {
	totalRepos := len(results)
	totalPRs := 0
	readyPRs := 0
	skippedPRs := 0
	errorCount := 0

	// Count totals
	for _, result := range results {
		if result.Error != nil {
			errorCount++
			continue
		}

		for _, processedPR := range result.PullRequests {
			totalPRs++

			if processedPR.Error != nil {
				errorCount++
			} else if processedPR.Skipped {
				skippedPRs++
			} else if processedPR.Ready {
				readyPRs++
			}
		}
	}

	// Display summary
	fmt.Printf("=== Summary ===\n")
	fmt.Printf("Repositories: %d\n", totalRepos)
	fmt.Printf("Total PRs: %d\n", totalPRs)
	fmt.Printf("Ready to merge: %d\n", readyPRs)
	fmt.Printf("Skipped: %d\n", skippedPRs)
	if errorCount > 0 {
		fmt.Printf("Errors: %d\n", errorCount)
	}
	fmt.Println()

	// Display PRs based on flags
	if flags.ShowAll {
		displayAllPRs(results)
	} else {
		if flags.ShowReady && readyPRs > 0 {
			displayReadyPRs(results)
		}

		if flags.ShowSkipped && skippedPRs > 0 {
			displaySkippedPRs(results)
		}

		// Always show errors
		if errorCount > 0 {
			displayErrorPRs(results)
		}
	}

	// Show next refresh time
	fmt.Printf("\nNext refresh in %s (Press Ctrl+C to stop)\n", flags.Interval)

	return nil
}

// displayReadyPRs displays PRs that are ready to merge
func displayReadyPRs(results []pr.ProcessResult) {
	fmt.Printf("=== Ready to Merge ===\n")

	count := 0
	for _, result := range results {
		if result.Error != nil {
			continue
		}

		for _, processedPR := range result.PullRequests {
			if processedPR.Ready && processedPR.Error == nil {
				count++
				pr := processedPR.PullRequest
				age := time.Since(pr.CreatedAt)

				fmt.Printf("✅ %s #%d: %s\n",
					result.Repository.FullName,
					pr.Number,
					pr.Title)
				fmt.Printf("   Author: %s | Age: %s | URL: %s\n",
					pr.Author.Login,
					formatDuration(age),
					pr.URL)
			}
		}
	}

	if count == 0 {
		fmt.Printf("No PRs ready to merge.\n")
	}
	fmt.Println()
}

// displaySkippedPRs displays PRs that were skipped
func displaySkippedPRs(results []pr.ProcessResult) {
	fmt.Printf("=== Skipped PRs ===\n")

	count := 0
	for _, result := range results {
		if result.Error != nil {
			continue
		}

		for _, processedPR := range result.PullRequests {
			if processedPR.Skipped && processedPR.Error == nil {
				count++
				pr := processedPR.PullRequest

				fmt.Printf("⏭️ %s #%d: %s\n",
					result.Repository.FullName,
					pr.Number,
					pr.Title)
				fmt.Printf("   Reason: %s | Author: %s | URL: %s\n",
					processedPR.Reason,
					pr.Author.Login,
					pr.URL)
			}
		}
	}

	if count == 0 {
		fmt.Printf("No PRs skipped.\n")
	}
	fmt.Println()
}

// displayErrorPRs displays PRs that had errors
func displayErrorPRs(results []pr.ProcessResult) {
	fmt.Printf("=== Errors ===\n")

	count := 0
	for _, result := range results {
		if result.Error != nil {
			count++
			fmt.Printf("❌ Repository %s: %s\n",
				result.Repository.FullName,
				result.Error.Error())
			continue
		}

		for _, processedPR := range result.PullRequests {
			if processedPR.Error != nil {
				count++
				pr := processedPR.PullRequest
				fmt.Printf("❌ %s #%d: %s\n",
					result.Repository.FullName,
					pr.Number,
					processedPR.Error.Error())
			}
		}
	}

	if count == 0 {
		fmt.Printf("No errors.\n")
	}
	fmt.Println()
}

// displayAllPRs displays all PRs regardless of status
func displayAllPRs(results []pr.ProcessResult) {
	fmt.Printf("=== All Pull Requests ===\n")

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ Repository %s: %s\n",
				result.Repository.FullName,
				result.Error.Error())
			continue
		}

		if len(result.PullRequests) == 0 {
			continue
		}

		fmt.Printf("\n%s (%d PRs):\n", result.Repository.FullName, len(result.PullRequests))

		for _, processedPR := range result.PullRequests {
			pr := processedPR.PullRequest
			age := time.Since(pr.CreatedAt)

			var status string
			var statusText string

			if processedPR.Error != nil {
				status = "❌"
				statusText = fmt.Sprintf("error: %s", processedPR.Error.Error())
			} else if processedPR.Skipped {
				status = "⏭️"
				statusText = fmt.Sprintf("skipped: %s", processedPR.Reason)
			} else if processedPR.Ready {
				status = "✅"
				statusText = "ready to merge"
			} else {
				status = "⏸️"
				statusText = processedPR.Reason
			}

			fmt.Printf("  %s #%d: %s\n",
				status,
				pr.Number,
				pr.Title)
			fmt.Printf("     Status: %s\n", statusText)
			fmt.Printf("     Author: %s | Age: %s\n",
				pr.Author.Login,
				formatDuration(age))
		}
	}
	fmt.Println()
}

// displayWatchResultsJSON displays results in JSON format
func displayWatchResultsJSON(results []pr.ProcessResult, flags WatchFlags) error {
	// Filter results based on flags
	filteredResults := make([]pr.ProcessResult, 0, len(results))

	for _, result := range results {
		if result.Error != nil {
			// Always include errors
			filteredResults = append(filteredResults, result)
			continue
		}

		filteredPRs := make([]pr.ProcessedPR, 0, len(result.PullRequests))

		for _, processedPR := range result.PullRequests {
			include := false

			if flags.ShowAll {
				include = true
			} else {
				if processedPR.Error != nil {
					include = true // Always show errors
				} else if processedPR.Ready && flags.ShowReady {
					include = true
				} else if processedPR.Skipped && flags.ShowSkipped {
					include = true
				}
			}

			if include {
				filteredPRs = append(filteredPRs, processedPR)
			}
		}

		if len(filteredPRs) > 0 {
			filteredResult := result
			filteredResult.PullRequests = filteredPRs
			filteredResults = append(filteredResults, filteredResult)
		}
	}

	return outputWatchResultsJSON(filteredResults)
}

// outputWatchResultsJSON outputs watch results in JSON format
func outputWatchResultsJSON(results []pr.ProcessResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
}

// parseRepoFilter parses a repository filter string into a slice of repository patterns
func parseRepoFilter(repoFilter string) []string {
	if repoFilter == "" {
		return nil
	}

	// Support comma-separated list
	repos := strings.Split(repoFilter, ",")
	for i, repo := range repos {
		repos[i] = strings.TrimSpace(repo)
	}

	return repos
}
