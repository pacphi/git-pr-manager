package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/notifications"
	"github.com/pacphi/git-pr-manager/pkg/pr"
	"github.com/pacphi/git-pr-manager/pkg/providers"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// StatsFlags contains flags for the stats command
type StatsFlags struct {
	Output        string
	Detailed      bool
	Period        string
	Format        string
	Sort          string
	Top           int
	IncludeClosed bool
	Provider      string
}

// RepositoryStats contains statistics for a repository
type RepositoryStats struct {
	Repository common.Repository `json:"repository"`
	TotalPRs   int               `json:"total_prs"`
	ReadyPRs   int               `json:"ready_prs"`
	SkippedPRs int               `json:"skipped_prs"`
	OpenPRs    int               `json:"open_prs"`
	DraftPRs   int               `json:"draft_prs"`
	OldestPR   *time.Time        `json:"oldest_pr,omitempty"`
	NewestPR   *time.Time        `json:"newest_pr,omitempty"`
	TopActors  []ActorStats      `json:"top_actors"`
	TopLabels  []string          `json:"top_labels"`
}

// ActorStats contains statistics for a PR actor
type ActorStats struct {
	Actor string `json:"actor"`
	Count int    `json:"count"`
}

// GlobalStats contains global statistics across all repositories
type GlobalStats struct {
	TotalRepositories int                      `json:"total_repositories"`
	TotalPRs          int                      `json:"total_prs"`
	ReadyPRs          int                      `json:"ready_prs"`
	SkippedPRs        int                      `json:"skipped_prs"`
	ByProvider        map[string]ProviderStats `json:"by_provider"`
	ByLanguage        map[string]int           `json:"by_language"`
	TopActors         []ActorStats             `json:"top_actors"`
	TopLabels         []string                 `json:"top_labels"`
	Repositories      []RepositoryStats        `json:"repositories,omitempty"`
	NotificationStats NotificationStats        `json:"notification_stats"`
	GeneratedAt       time.Time                `json:"generated_at"`
}

// NotificationStats contains statistics about the notification system
type NotificationStats struct {
	ConfiguredNotifiers int  `json:"configured_notifiers"`
	SlackEnabled        bool `json:"slack_enabled"`
	EmailEnabled        bool `json:"email_enabled"`
}

// ProviderStats contains statistics for a provider
type ProviderStats struct {
	Repositories int `json:"repositories"`
	TotalPRs     int `json:"total_prs"`
	ReadyPRs     int `json:"ready_prs"`
	SkippedPRs   int `json:"skipped_prs"`
}

// NewStatsCommand creates the stats command
func NewStatsCommand() *cobra.Command {
	var flags StatsFlags

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show repository and PR statistics",
		Long: `Display comprehensive statistics about configured repositories and pull requests.

Shows information about:
- Repository counts by provider and language
- PR counts and status distribution
- Top contributors and labels
- Age and activity metrics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.Output, "output", "o", "table", "output format (table, json) - DEPRECATED, use --format")
	cmd.Flags().BoolVarP(&flags.Detailed, "detailed", "d", false, "show detailed per-repository stats")
	cmd.Flags().StringVar(&flags.Period, "period", "30d", "time period for analysis (1d, 7d, 30d, 90d)")
	cmd.Flags().StringVar(&flags.Format, "format", "table", "output format (table, json, yaml, csv, text)")
	cmd.Flags().StringVar(&flags.Sort, "sort", "prs", "sort repositories by field (prs, ready, name, age)")
	cmd.Flags().IntVar(&flags.Top, "top", 0, "show only top N repositories (0 = all)")
	cmd.Flags().BoolVar(&flags.IncludeClosed, "include-closed", false, "include closed PRs in statistics")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", "filter by provider (github, gitlab, bitbucket)")

	return cmd
}

// runStats generates and displays repository statistics
func runStats(ctx context.Context, flags StatsFlags) error {
	// Validate flags
	if flags.Top < 0 {
		fmt.Fprintln(os.Stderr, "--top must be non-negative (0 to show all repositories)")
		os.Exit(1)
	}

	logger := utils.GetGlobalLogger()

	logger.Info("Gathering repository statistics...")

	// Parse period
	period, err := parsePeriod(flags.Period)
	if err != nil {
		return fmt.Errorf("invalid period: %w", err)
	}

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return HandleConfigError(err, "stats")
	}

	// Create providers (filtered by provider if specified)
	factory := providers.NewFactory(cfg)
	providerMap, err := factory.CreateProviders()
	if err != nil {
		return fmt.Errorf("failed to create providers: %w", err)
	}

	// Filter providers if needed
	if flags.Provider != "" {
		filteredMap := make(map[string]common.Provider)
		if provider, exists := providerMap[flags.Provider]; exists {
			filteredMap[flags.Provider] = provider
			providerMap = filteredMap
		} else {
			return fmt.Errorf("provider %s not found or not configured", flags.Provider)
		}
	}

	// Create processor
	processor := pr.NewProcessor(providerMap, cfg)

	// Process all repositories to get PR data
	processOpts := pr.ProcessOptions{
		DryRun:        true, // We only want to analyze, not modify
		RequireChecks: false,
		MaxAge:        period,
		IncludeClosed: flags.IncludeClosed,
	}

	results, err := processor.ProcessAllPRs(ctx, processOpts)
	if err != nil {
		return fmt.Errorf("failed to process repositories: %w", err)
	}

	// Generate statistics
	stats := generateStats(results, flags, cfg)

	// Determine output format (support both --output and --format for backwards compatibility)
	format := flags.Format
	if format == "table" && flags.Output != "table" {
		format = flags.Output
	}

	// Output statistics
	return outputStats(stats, format, flags)
}

// generateStats generates statistics from processing results
func generateStats(results []pr.ProcessResult, flags StatsFlags, cfg *config.Config) *GlobalStats {
	stats := &GlobalStats{
		ByProvider:   make(map[string]ProviderStats),
		ByLanguage:   make(map[string]int),
		GeneratedAt:  time.Now(),
		Repositories: []RepositoryStats{},
	}

	actorCounts := make(map[string]int)
	labelCounts := make(map[string]int)

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		stats.TotalRepositories++

		// Update language stats
		if result.Repository.Language != "" {
			stats.ByLanguage[result.Repository.Language]++
		}

		// Process repository-specific stats
		repoStats := RepositoryStats{
			Repository: result.Repository,
			TopActors:  []ActorStats{},
			TopLabels:  []string{},
		}

		repoActorCounts := make(map[string]int)
		repoLabelCounts := make(map[string]int)

		for _, processedPR := range result.PullRequests {
			stats.TotalPRs++
			repoStats.TotalPRs++

			if processedPR.Error != nil {
				continue
			}

			if processedPR.Skipped {
				stats.SkippedPRs++
				repoStats.SkippedPRs++
			} else if processedPR.Ready {
				stats.ReadyPRs++
				repoStats.ReadyPRs++
			}

			pr := processedPR.PullRequest

			// Count open and draft PRs
			if pr.IsOpen() {
				repoStats.OpenPRs++
				if pr.IsDraft() {
					repoStats.DraftPRs++
				}
			}

			// Track PR ages
			if repoStats.OldestPR == nil || pr.CreatedAt.Before(*repoStats.OldestPR) {
				repoStats.OldestPR = &pr.CreatedAt
			}
			if repoStats.NewestPR == nil || pr.CreatedAt.After(*repoStats.NewestPR) {
				repoStats.NewestPR = &pr.CreatedAt
			}

			// Count actors
			actor := pr.Author.Login
			actorCounts[actor]++
			repoActorCounts[actor]++

			// Count labels
			for _, label := range pr.Labels {
				labelCounts[label.Name]++
				repoLabelCounts[label.Name]++
			}
		}

		// Generate top actors for repository
		repoStats.TopActors = getTopActors(repoActorCounts, 5)

		// Generate top labels for repository
		repoStats.TopLabels = getTopLabels(repoLabelCounts, 5)

		// Update provider stats
		providerStats := stats.ByProvider[result.Provider]
		providerStats.Repositories++
		providerStats.TotalPRs += repoStats.TotalPRs
		providerStats.ReadyPRs += repoStats.ReadyPRs
		providerStats.SkippedPRs += repoStats.SkippedPRs
		stats.ByProvider[result.Provider] = providerStats

		if flags.Detailed {
			stats.Repositories = append(stats.Repositories, repoStats)
		}
	}

	// Generate global top actors and labels
	stats.TopActors = getTopActors(actorCounts, 10)
	stats.TopLabels = getTopLabels(labelCounts, 10)

	// Sort and limit repositories if requested
	if flags.Detailed {
		sortRepositories(stats.Repositories, flags.Sort)
		if flags.Top > 0 && flags.Top < len(stats.Repositories) {
			stats.Repositories = stats.Repositories[:flags.Top]
		}
	}

	// Add notification statistics
	notificationManager, err := notifications.NewManager(cfg)
	if err == nil {
		stats.NotificationStats = NotificationStats{
			ConfiguredNotifiers: notificationManager.GetNotifierCount(),
			SlackEnabled:        cfg.Notifications.Slack.Enabled,
			EmailEnabled:        cfg.Notifications.Email.Enabled,
		}
	}

	return stats
}

// getTopActors returns the top N actors by PR count
func getTopActors(actorCounts map[string]int, n int) []ActorStats {
	type actorCount struct {
		actor string
		count int
	}

	actors := make([]actorCount, 0, len(actorCounts))
	for actor, count := range actorCounts {
		actors = append(actors, actorCount{actor: actor, count: count})
	}

	sort.Slice(actors, func(i, j int) bool {
		return actors[i].count > actors[j].count
	})

	result := make([]ActorStats, 0, min(n, len(actors)))
	for i, actor := range actors {
		if i >= n {
			break
		}
		result = append(result, ActorStats{
			Actor: actor.actor,
			Count: actor.count,
		})
	}

	return result
}

// getTopLabels returns the top N labels by usage count
func getTopLabels(labelCounts map[string]int, n int) []string {
	type labelCount struct {
		label string
		count int
	}

	labels := make([]labelCount, 0, len(labelCounts))
	for label, count := range labelCounts {
		labels = append(labels, labelCount{label: label, count: count})
	}

	sort.Slice(labels, func(i, j int) bool {
		return labels[i].count > labels[j].count
	})

	result := make([]string, 0, min(n, len(labels)))
	for i, label := range labels {
		if i >= n {
			break
		}
		result = append(result, fmt.Sprintf("%s (%d)", label.label, label.count))
	}

	return result
}

// outputStatsJSON outputs statistics in JSON format
func outputStatsJSON(stats *GlobalStats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// outputStatsTable outputs statistics in table format
func outputStatsTable(stats *GlobalStats, detailed bool) error {
	fmt.Printf("=== Git PR Automation Statistics ===\n")
	fmt.Printf("Generated at: %s\n\n", stats.GeneratedAt.Format("2006-01-02 15:04:05 MST"))

	// Global summary
	fmt.Printf("=== Global Summary ===\n")
	fmt.Printf("Repositories: %d\n", stats.TotalRepositories)
	fmt.Printf("Total PRs: %d\n", stats.TotalPRs)
	fmt.Printf("Ready to merge: %d (%.1f%%)\n", stats.ReadyPRs, percentage(stats.ReadyPRs, stats.TotalPRs))
	fmt.Printf("Skipped: %d (%.1f%%)\n", stats.SkippedPRs, percentage(stats.SkippedPRs, stats.TotalPRs))
	fmt.Println()

	// Provider breakdown
	if len(stats.ByProvider) > 0 {
		fmt.Printf("=== By Provider ===\n")
		for provider, providerStats := range stats.ByProvider {
			fmt.Printf("%s:\n", provider)
			fmt.Printf("  Repositories: %d\n", providerStats.Repositories)
			fmt.Printf("  Total PRs: %d\n", providerStats.TotalPRs)
			fmt.Printf("  Ready: %d (%.1f%%)\n", providerStats.ReadyPRs, percentage(providerStats.ReadyPRs, providerStats.TotalPRs))
			fmt.Printf("  Skipped: %d (%.1f%%)\n", providerStats.SkippedPRs, percentage(providerStats.SkippedPRs, providerStats.TotalPRs))
		}
		fmt.Println()
	}

	// Language breakdown
	if len(stats.ByLanguage) > 0 {
		fmt.Printf("=== By Language ===\n")

		// Sort languages by repository count
		type langCount struct {
			lang  string
			count int
		}

		var langs []langCount
		for lang, count := range stats.ByLanguage {
			langs = append(langs, langCount{lang: lang, count: count})
		}

		sort.Slice(langs, func(i, j int) bool {
			return langs[i].count > langs[j].count
		})

		for _, lang := range langs {
			fmt.Printf("  %s: %d repositories\n", lang.lang, lang.count)
		}
		fmt.Println()
	}

	// Top actors
	if len(stats.TopActors) > 0 {
		fmt.Printf("=== Top Contributors ===\n")
		for i, actor := range stats.TopActors {
			if i >= 5 { // Limit to top 5 in table format
				break
			}
			fmt.Printf("  %d. %s (%d PRs)\n", i+1, actor.Actor, actor.Count)
		}
		fmt.Println()
	}

	// Top labels
	if len(stats.TopLabels) > 0 {
		fmt.Printf("=== Most Common Labels ===\n")
		for i, label := range stats.TopLabels {
			if i >= 5 { // Limit to top 5 in table format
				break
			}
			fmt.Printf("  %d. %s\n", i+1, label)
		}
		fmt.Println()
	}

	// Detailed repository stats
	if detailed && len(stats.Repositories) > 0 {
		fmt.Printf("=== Repository Details ===\n")

		for _, repo := range stats.Repositories {
			fmt.Printf("\n%s (%s):\n", repo.Repository.FullName, repo.Repository.Language)
			fmt.Printf("  PRs: %d total, %d ready, %d skipped, %d open, %d draft\n",
				repo.TotalPRs, repo.ReadyPRs, repo.SkippedPRs, repo.OpenPRs, repo.DraftPRs)

			if repo.OldestPR != nil && repo.NewestPR != nil {
				fmt.Printf("  PR age range: %s to %s\n",
					repo.OldestPR.Format("2006-01-02"),
					repo.NewestPR.Format("2006-01-02"))
			}

			if len(repo.TopActors) > 0 {
				var actors []string
				for _, actor := range repo.TopActors {
					actors = append(actors, fmt.Sprintf("%s (%d)", actor.Actor, actor.Count))
				}
				fmt.Printf("  Top actors: %s\n", strings.Join(actors, ", "))
			}

			if len(repo.TopLabels) > 0 {
				fmt.Printf("  Top labels: %s\n", strings.Join(repo.TopLabels, ", "))
			}
		}
	}

	return nil
}

// parsePeriod parses a period string like "30d" into a time.Duration
func parsePeriod(period string) (time.Duration, error) {
	switch period {
	case "1d":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	case "90d":
		return 90 * 24 * time.Hour, nil
	default:
		// Try parsing as duration
		return time.ParseDuration(period)
	}
}

// sortRepositories sorts repositories based on the specified field
func sortRepositories(repos []RepositoryStats, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Repository.FullName < repos[j].Repository.FullName
		})
	case "ready":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].ReadyPRs > repos[j].ReadyPRs
		})
	case "age":
		sort.Slice(repos, func(i, j int) bool {
			if repos[i].OldestPR == nil && repos[j].OldestPR == nil {
				return false
			}
			if repos[i].OldestPR == nil {
				return false
			}
			if repos[j].OldestPR == nil {
				return true
			}
			return repos[i].OldestPR.Before(*repos[j].OldestPR)
		})
	case "prs":
		fallthrough
	default:
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].TotalPRs > repos[j].TotalPRs
		})
	}
}

// outputStats outputs statistics in the specified format
func outputStats(stats *GlobalStats, format string, flags StatsFlags) error {
	switch strings.ToLower(format) {
	case "json":
		return outputStatsJSON(stats)
	case "yaml":
		return outputStatsYAML(stats)
	case "csv":
		return outputStatsCSV(stats, flags)
	case "text":
		return outputStatsText(stats, flags)
	case "table":
		fallthrough
	default:
		return outputStatsTable(stats, flags.Detailed)
	}
}

// outputStatsYAML outputs statistics in YAML format
func outputStatsYAML(stats *GlobalStats) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer func() {
		if err := encoder.Close(); err != nil {
			utils.GetGlobalLogger().WithError(err).Error("Failed to close encoder")
		}
	}()
	encoder.SetIndent(2)
	return encoder.Encode(stats)
}

// outputStatsCSV outputs statistics in CSV format
func outputStatsCSV(stats *GlobalStats, _ StatsFlags) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header for repository stats
	header := []string{"Repository", "Provider", "Language", "Total PRs", "Ready PRs", "Skipped PRs", "Open PRs", "Draft PRs"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write repository data
	for _, repo := range stats.Repositories {
		row := []string{
			repo.Repository.FullName,
			repo.Repository.Provider,
			repo.Repository.Language,
			fmt.Sprintf("%d", repo.TotalPRs),
			fmt.Sprintf("%d", repo.ReadyPRs),
			fmt.Sprintf("%d", repo.SkippedPRs),
			fmt.Sprintf("%d", repo.OpenPRs),
			fmt.Sprintf("%d", repo.DraftPRs),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// outputStatsText outputs statistics in plain text format
func outputStatsText(stats *GlobalStats, flags StatsFlags) error {
	fmt.Printf("Git PR Statistics\n")
	fmt.Printf("Generated: %s\n", stats.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Repositories: %d\n", stats.TotalRepositories)
	fmt.Printf("Total PRs: %d\n", stats.TotalPRs)
	fmt.Printf("Ready PRs: %d\n", stats.ReadyPRs)
	fmt.Printf("Skipped PRs: %d\n", stats.SkippedPRs)

	if len(stats.ByProvider) > 0 {
		fmt.Printf("\nBy Provider:\n")
		for provider, providerStats := range stats.ByProvider {
			fmt.Printf("  %s: %d repos, %d PRs\n", provider, providerStats.Repositories, providerStats.TotalPRs)
		}
	}

	if len(stats.TopActors) > 0 {
		fmt.Printf("\nTop Contributors:\n")
		limit := flags.Top
		if limit == 0 || limit > len(stats.TopActors) {
			limit = len(stats.TopActors)
		}
		for i := 0; i < limit && i < len(stats.TopActors); i++ {
			actor := stats.TopActors[i]
			fmt.Printf("  %s: %d PRs\n", actor.Actor, actor.Count)
		}
	}

	return nil
}

// percentage calculates percentage with safe division
func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
