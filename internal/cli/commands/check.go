package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pacphi/git-pr-manager/internal/executor"
	"github.com/pacphi/git-pr-manager/pkg/notifications"
	"github.com/pacphi/git-pr-manager/pkg/pr"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// CheckFlags contains flags for the check command
type CheckFlags struct {
	Providers     []string
	Repositories  []string
	MaxAge        string
	Output        string
	RequireChecks bool
	ShowSkipped   bool
	ShowDetails   bool
	ShowStatus    bool
	ReadyOnly     bool
}

// NewCheckCommand creates the check command
func NewCheckCommand() *cobra.Command {
	var flags CheckFlags

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check pull request status across repositories",
		Long: `Check the status of pull requests across all configured repositories.

This command scans all configured repositories for open pull requests from trusted
actors and reports their status, including readiness for merging.`,
		Example: `  # Check all repositories
  git-pr-cli check

  # Check specific providers
  git-pr-cli check --providers github,gitlab

  # Check specific repositories
  git-pr-cli check --repos owner/repo1,owner/repo2

  # Output as JSON
  git-pr-cli check --output json

  # Show skipped PRs
  git-pr-cli check --show-skipped`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckCommand(cmd.Context(), flags)
		},
	}

	// Add flags
	cmd.Flags().StringSliceVarP(&flags.Providers, "providers", "p", nil, "Comma-separated list of providers to check (github,gitlab,bitbucket)")
	cmd.Flags().StringSliceVarP(&flags.Repositories, "repos", "r", nil, "Comma-separated list of repositories to check")
	cmd.Flags().StringVar(&flags.MaxAge, "max-age", "", "Maximum age of PRs to consider (e.g., 7d, 24h)")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "table", "Output format (table, json, yaml, csv, summary)")
	cmd.Flags().BoolVar(&flags.RequireChecks, "require-checks", false, "Require all status checks to pass")
	cmd.Flags().BoolVar(&flags.ShowSkipped, "show-skipped", false, "Show skipped PRs in output")
	cmd.Flags().BoolVar(&flags.ShowDetails, "show-details", false, "Show detailed information about each PR")
	cmd.Flags().BoolVar(&flags.ShowStatus, "show-status", false, "Show detailed status information")
	cmd.Flags().BoolVar(&flags.ReadyOnly, "ready-only", false, "Show only PRs that are ready to merge")

	return cmd
}

func runCheckCommand(ctx context.Context, flags CheckFlags) error {
	logger := utils.GetGlobalLogger()
	logger.Info("Starting PR check operation")

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return HandleConfigError(err, "check")
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
	opts := pr.ProcessOptions{
		DryRun:        false,
		Providers:     flags.Providers,
		Repositories:  flags.Repositories,
		MaxAge:        maxAge,
		RequireChecks: flags.RequireChecks,
	}

	// Process PRs
	results, err := exec.ProcessPRs(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to process PRs: %w", err)
	}

	// Send PR summary notifications if configured
	notificationManager, err := notifications.NewManager(cfg)
	if err != nil {
		logger.WithError(err).Warn("Failed to create notification manager")
	} else if notificationManager.HasNotifiers() {
		// Calculate summary statistics
		totalPRs, readyPRs := calculatePRStats(results)
		repositories := extractRepositoriesFromResults(results)

		logger.Info("Sending PR summary notifications...")
		if err := notificationManager.SendPRSummary(ctx, repositories, totalPRs, readyPRs); err != nil {
			logger.WithError(err).Warn("Failed to send PR summary notifications")
		} else {
			logger.Info("PR summary notifications sent successfully")
		}
	}

	// Output results
	return outputCheckResults(results, flags)
}

// filterResults applies filtering logic based on flags and returns filtered results
func filterResults(results []pr.ProcessResult, flags CheckFlags) []pr.ProcessResult {
	var filteredResults []pr.ProcessResult

	for _, result := range results {
		if result.Error != nil {
			// Always include error results
			filteredResults = append(filteredResults, result)
			continue
		}

		if len(result.PullRequests) == 0 {
			if !flags.ReadyOnly {
				// Include repos with no PRs only if not filtering for ready-only
				filteredResults = append(filteredResults, result)
			}
			continue
		}

		// Filter PRs within this result
		var filteredPRs []pr.ProcessedPR
		for _, pr := range result.PullRequests {
			if pr.Skipped && !flags.ShowSkipped {
				continue
			}
			if flags.ReadyOnly && !pr.Ready {
				continue
			}
			filteredPRs = append(filteredPRs, pr)
		}

		// Only include the result if it has displayable PRs
		if len(filteredPRs) > 0 || !flags.ReadyOnly {
			newResult := result
			newResult.PullRequests = filteredPRs
			filteredResults = append(filteredResults, newResult)
		}
	}

	return filteredResults
}

func outputCheckResults(results []pr.ProcessResult, flags CheckFlags) error {
	switch strings.ToLower(flags.Output) {
	case "json":
		return outputCheckResultsJSON(results, flags)
	case "yaml":
		return outputCheckResultsYAML(results, flags)
	case "csv":
		return outputCheckResultsCSV(results, flags)
	case "summary":
		return outputCheckResultsSummary(results, flags)
	case "table":
		fallthrough
	default:
		return outputCheckResultsTable(results, flags)
	}
}

func outputCheckResultsJSON(results []pr.ProcessResult, flags CheckFlags) error {
	filteredResults := filterResults(results, flags)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(filteredResults)
}

func outputCheckResultsSummary(results []pr.ProcessResult, flags CheckFlags) error {
	logger := utils.GetGlobalLogger().WithComponent("check-summary")
	totalRepos := 0
	totalPRs := 0
	readyPRs := 0
	skippedPRs := 0
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
			continue
		}

		totalRepos++
		totalPRs += len(result.PullRequests)

		for _, pr := range result.PullRequests {
			if pr.Skipped && !flags.ShowSkipped {
				continue
			}
			if flags.ReadyOnly && !pr.Ready {
				continue
			}

			if pr.Error != nil {
				errorCount++
			} else if pr.Skipped {
				skippedPRs++
			} else if pr.Ready {
				readyPRs++
			}
		}
	}

	// Output to stdout for CLI usage
	fmt.Printf("PR Check Summary\n")
	fmt.Printf("================\n")
	fmt.Printf("Repositories: %d\n", totalRepos)
	fmt.Printf("Total PRs: %d\n", totalPRs)
	fmt.Printf("Ready PRs: %d\n", readyPRs)

	if flags.ShowSkipped {
		fmt.Printf("Skipped PRs: %d\n", skippedPRs)
	}

	if errorCount > 0 {
		fmt.Printf("Errors: %d\n", errorCount)
	}

	// Log structured summary for observability
	logger.WithFields(map[string]interface{}{
		"total_repositories": totalRepos,
		"total_prs":          totalPRs,
		"ready_prs":          readyPRs,
		"skipped_prs":        skippedPRs,
		"error_count":        errorCount,
	}).Info("PR check summary completed")

	return nil
}

func outputCheckResultsTable(results []pr.ProcessResult, flags CheckFlags) error {
	// Determine column headers based on flags
	headers := []string{"PROVIDER", "REPOSITORY", "PR", "TITLE", "STATUS"}
	if flags.ShowStatus {
		headers = append(headers, "CHECKS", "MERGEABLE")
	}
	if flags.ShowDetails {
		headers = append(headers, "AUTHOR", "CREATED", "UPDATED")
	}
	headers = append(headers, "REASON")

	// Calculate column widths
	widths := []int{15, 30, 8, 25, 10}
	if flags.ShowStatus {
		widths = append(widths, 8, 10)
	}
	if flags.ShowDetails {
		widths = append(widths, 15, 12, 12)
	}
	widths = append(widths, 30)

	// Print headers
	fmt.Printf("\n")
	for i, header := range headers {
		fmt.Printf("%-*s ", widths[i], header)
	}
	fmt.Printf("\n")

	totalWidth := 0
	for _, w := range widths {
		totalWidth += w + 1
	}
	fmt.Printf("%s\n", strings.Repeat("=", totalWidth))

	for _, result := range results {
		if result.Error != nil {
			// Print error row
			fmt.Printf("%-*s %-*s %-*s %-*s %-*s",
				widths[0], result.Provider,
				widths[1], utils.Truncate(result.Repository.FullName, widths[1], "..."),
				widths[2], "-",
				widths[3], "-",
				widths[4], "ERROR")

			colIndex := 5
			if flags.ShowStatus {
				fmt.Printf(" %-*s %-*s", widths[colIndex], "-", widths[colIndex+1], "-")
				colIndex += 2
			}
			if flags.ShowDetails {
				fmt.Printf(" %-*s %-*s %-*s", widths[colIndex], "-", widths[colIndex+1], "-", widths[colIndex+2], "-")
			}
			fmt.Printf(" %s\n", result.Error.Error())
			continue
		}

		if len(result.PullRequests) == 0 {
			if !flags.ReadyOnly {
				// Print no PRs row
				fmt.Printf("%-*s %-*s %-*s %-*s %-*s",
					widths[0], result.Provider,
					widths[1], utils.Truncate(result.Repository.FullName, widths[1], "..."),
					widths[2], "-",
					widths[3], "No PRs",
					widths[4], "-")

				colIndex := 5
				if flags.ShowStatus {
					fmt.Printf(" %-*s %-*s", widths[colIndex], "-", widths[colIndex+1], "-")
					colIndex += 2
				}
				if flags.ShowDetails {
					fmt.Printf(" %-*s %-*s %-*s", widths[colIndex], "-", widths[colIndex+1], "-", widths[colIndex+2], "-")
				}
				fmt.Printf(" %s\n", "")
			}
			continue
		}

		// Check if any PRs will be displayed after filtering
		hasDisplayablePRs := false
		for _, pr := range result.PullRequests {
			if pr.Skipped && !flags.ShowSkipped {
				continue
			}
			if flags.ReadyOnly && !pr.Ready {
				continue
			}
			hasDisplayablePRs = true
			break
		}

		if !hasDisplayablePRs && flags.ReadyOnly {
			// Skip this repository entirely if no ready PRs
			continue
		}

		for _, pr := range result.PullRequests {
			if pr.Skipped && !flags.ShowSkipped {
				continue
			}
			if flags.ReadyOnly && !pr.Ready {
				continue
			}

			if pr.Error != nil {
				fmt.Printf("%-*s %-*s %-*d %-*s %-*s",
					widths[0], result.Provider,
					widths[1], utils.Truncate(result.Repository.FullName, widths[1], "..."),
					widths[2], pr.PullRequest.Number,
					widths[3], utils.Truncate(pr.PullRequest.Title, widths[3], "..."),
					widths[4], "ERROR")

				colIndex := 5
				if flags.ShowStatus {
					fmt.Printf(" %-*s %-*s", widths[colIndex], "-", widths[colIndex+1], "-")
					colIndex += 2
				}
				if flags.ShowDetails {
					fmt.Printf(" %-*s %-*s %-*s", widths[colIndex], "-", widths[colIndex+1], "-", widths[colIndex+2], "-")
				}
				fmt.Printf(" %s\n", pr.Error.Error())
				continue
			}

			status := "NOT READY"
			if pr.Skipped {
				status = "SKIPPED"
			} else if pr.Ready {
				status = "READY"
			}

			fmt.Printf("%-*s %-*s %-*d %-*s %-*s",
				widths[0], result.Provider,
				widths[1], utils.Truncate(result.Repository.FullName, widths[1], "..."),
				widths[2], pr.PullRequest.Number,
				widths[3], utils.Truncate(pr.PullRequest.Title, widths[3], "..."),
				widths[4], status)

			colIndex := 5
			if flags.ShowStatus {
				checks := "N/A"
				mergeable := "N/A"
				if pr.PullRequest.StatusChecks != nil {
					if len(pr.PullRequest.StatusChecks) > 0 {
						checks = fmt.Sprintf("%d/%d", pr.PullRequest.PassedChecks, len(pr.PullRequest.StatusChecks))
					} else {
						checks = "NONE"
					}
				}
				if pr.PullRequest.Mergeable != nil {
					if *pr.PullRequest.Mergeable {
						mergeable = "YES"
					} else {
						mergeable = "NO"
					}
				}
				fmt.Printf(" %-*s %-*s", widths[colIndex], checks, widths[colIndex+1], mergeable)
				colIndex += 2
			}

			if flags.ShowDetails {
				author := utils.Truncate(pr.PullRequest.Author.Login, widths[colIndex], "...")
				created := utils.TimeAgo(pr.PullRequest.CreatedAt)
				updated := utils.TimeAgo(pr.PullRequest.UpdatedAt)
				fmt.Printf(" %-*s %-*s %-*s", widths[colIndex], author, widths[colIndex+1], created, widths[colIndex+2], updated)
			}

			fmt.Printf(" %s\n", pr.Reason)
		}
	}

	fmt.Printf("\n")
	return nil
}

func outputCheckResultsYAML(results []pr.ProcessResult, flags CheckFlags) error {
	filteredResults := filterResults(results, flags)
	encoder := yaml.NewEncoder(os.Stdout)
	defer func() {
		if err := encoder.Close(); err != nil {
			utils.GetGlobalLogger().WithError(err).Error("Failed to close encoder")
		}
	}()
	encoder.SetIndent(2)
	return encoder.Encode(filteredResults)
}

func outputCheckResultsCSV(results []pr.ProcessResult, flags CheckFlags) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"Provider", "Repository", "PR Number", "Title", "Author", "Age", "Status", "Ready", "Reason"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, result := range results {
		if result.Error != nil {
			row := []string{
				result.Provider,
				result.Repository.FullName,
				"-",
				"-",
				"-",
				"-",
				"ERROR",
				"false",
				result.Error.Error(),
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
			continue
		}

		if len(result.PullRequests) == 0 {
			if !flags.ReadyOnly {
				row := []string{
					result.Provider,
					result.Repository.FullName,
					"-",
					"No PRs",
					"-",
					"-",
					"-",
					"false",
					"",
				}
				if err := writer.Write(row); err != nil {
					return fmt.Errorf("failed to write CSV row: %w", err)
				}
			}
			continue
		}

		// Check if any PRs will be displayed after filtering
		hasDisplayablePRs := false
		for _, pr := range result.PullRequests {
			if pr.Skipped && !flags.ShowSkipped {
				continue
			}
			if flags.ReadyOnly && !pr.Ready {
				continue
			}
			hasDisplayablePRs = true
			break
		}

		if !hasDisplayablePRs && flags.ReadyOnly {
			// Skip this repository entirely if no ready PRs
			continue
		}

		for _, pr := range result.PullRequests {
			if pr.Skipped && !flags.ShowSkipped {
				continue
			}
			if flags.ReadyOnly && !pr.Ready {
				continue
			}

			if pr.Error != nil {
				row := []string{
					result.Provider,
					result.Repository.FullName,
					fmt.Sprintf("%d", pr.PullRequest.Number),
					pr.PullRequest.Title,
					pr.PullRequest.Author.Login,
					utils.TimeAgo(pr.PullRequest.CreatedAt),
					"ERROR",
					"false",
					pr.Error.Error(),
				}
				if err := writer.Write(row); err != nil {
					return fmt.Errorf("failed to write CSV row: %w", err)
				}
				continue
			}

			status := "NOT READY"
			ready := "false"
			if pr.Skipped {
				status = "SKIPPED"
			} else if pr.Ready {
				status = "READY"
				ready = "true"
			}

			row := []string{
				result.Provider,
				result.Repository.FullName,
				fmt.Sprintf("%d", pr.PullRequest.Number),
				pr.PullRequest.Title,
				pr.PullRequest.Author.Login,
				utils.TimeAgo(pr.PullRequest.CreatedAt),
				status,
				ready,
				pr.Reason,
			}

			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	return nil
}

// calculatePRStats calculates total and ready PR counts from results
func calculatePRStats(results []pr.ProcessResult) (totalPRs, readyPRs int) {
	for _, result := range results {
		if result.Error != nil {
			continue
		}
		totalPRs += len(result.PullRequests)
		for _, pr := range result.PullRequests {
			if pr.Ready && !pr.Skipped && pr.Error == nil {
				readyPRs++
			}
		}
	}
	return totalPRs, readyPRs
}

// extractRepositoriesFromResults extracts repository information from process results
func extractRepositoriesFromResults(results []pr.ProcessResult) []common.Repository {
	repositories := make([]common.Repository, 0, len(results))
	for _, result := range results {
		if result.Error != nil {
			continue
		}
		// Create a basic repository structure from the result
		repository := common.Repository{
			FullName:  result.Repository.FullName,
			Provider:  result.Provider,
			IsPrivate: false, // We don't have this info in process results
		}
		repositories = append(repositories, repository)
	}
	return repositories
}
