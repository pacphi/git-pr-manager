package wizard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// Wizard provides interactive configuration setup
type Wizard struct {
	config      Config
	logger      *utils.Logger
	discoveries map[string][]common.Repository
}

// Config contains wizard configuration
type Config struct {
	ConfigPath  string
	Preview     bool
	Additive    bool
	Interactive bool
}

// ProviderDiscovery contains discovered repositories for a provider
type ProviderDiscovery struct {
	Provider     string
	Repositories []common.Repository
	Error        error
}

// FilterOptions contains repository filtering options
type FilterOptions struct {
	IncludePrivate  bool
	IncludePublic   bool
	IncludeForks    bool
	IncludeArchived bool
	MinStars        int
	MaxAge          int    // days
	NameFilter      string // substring match
	OwnerFilter     string // specific owner
	LanguageFilter  string // specific language
}

// New creates a new wizard instance
func New(config Config) *Wizard {
	return &Wizard{
		config:      config,
		logger:      utils.GetGlobalLogger().WithComponent("wizard"),
		discoveries: make(map[string][]common.Repository),
	}
}

// Run runs the interactive setup wizard
func (w *Wizard) Run(ctx context.Context) error {
	w.logger.Info("Starting interactive setup wizard...")

	// Step 1: Load existing config if additive
	var existingConfig *config.Config
	if w.config.Additive {
		if cfg, err := w.loadExistingConfig(); err == nil {
			existingConfig = cfg
			w.logger.Info("Loaded existing configuration for additive mode")
		} else {
			w.logger.Warnf("Could not load existing config: %v", err)
		}
	}

	// Step 2: Discover repositories
	if err := w.discoverRepositories(ctx); err != nil {
		return fmt.Errorf("repository discovery failed: %w", err)
	}

	// Step 3: Present discovered repositories and get user selection
	selectedRepos := w.selectRepositories()
	if len(selectedRepos) == 0 {
		w.logger.Info("No repositories selected, wizard cancelled")
		return nil
	}

	// Step 4: Configure options
	cfg := w.buildConfiguration(selectedRepos, existingConfig)

	// Step 5: Preview or save configuration
	if w.config.Preview {
		return w.previewConfiguration(cfg)
	}

	return w.saveConfiguration(cfg)
}

// Preview shows what would be configured without making changes
func (w *Wizard) Preview(ctx context.Context) error {
	w.logger.Info("Running wizard in preview mode...")

	if err := w.discoverRepositories(ctx); err != nil {
		return fmt.Errorf("repository discovery failed: %w", err)
	}

	// Show discovery results
	w.printDiscoveryResults()

	// Build sample configuration
	allRepos := w.getAllDiscoveredRepositories()
	cfg := w.buildConfiguration(allRepos, nil)

	return w.previewConfiguration(cfg)
}

// RunAdditive runs the wizard in additive mode
func (w *Wizard) RunAdditive(ctx context.Context) error {
	w.config.Additive = true
	return w.Run(ctx)
}

// discoverRepositories discovers repositories from all configured providers
func (w *Wizard) discoverRepositories(ctx context.Context) error {
	w.logger.Info("Discovering repositories from configured providers...")

	// Create temporary config for discovery
	tempConfig := &config.Config{
		Auth: config.Auth{
			GitHub: config.GitHubAuth{
				Token: "$GITHUB_TOKEN",
			},
			GitLab: config.GitLabAuth{
				Token: "$GITLAB_TOKEN",
				URL:   "$GITLAB_URL",
			},
			Bitbucket: config.BitbucketAuth{
				Username:    "$BITBUCKET_USERNAME",
				AppPassword: "$BITBUCKET_APP_PASSWORD",
				Workspace:   "$BITBUCKET_WORKSPACE",
			},
		},
		Behavior: config.Behavior{
			Concurrency: 5,
			RateLimit: config.RateLimit{
				RequestsPerSecond: 5.0,
				Burst:             10,
				Timeout:           30 * time.Second,
			},
		},
	}

	factory := providers.NewFactory(tempConfig)
	providerMap, err := factory.CreateProvidersForDiscovery()
	if err != nil {
		return fmt.Errorf("failed to create providers: %w", err)
	}

	// Discover repositories from each provider
	for name, provider := range providerMap {
		w.logger.Infof("Discovering repositories from %s...", name)

		repos, err := provider.ListRepositories(ctx)
		if err != nil {
			w.logger.Warnf("Failed to discover repositories from %s: %v", name, err)
			continue
		}

		w.discoveries[name] = repos
		w.logger.Infof("Discovered %d repositories from %s", len(repos), name)
	}

	if len(w.discoveries) == 0 {
		return fmt.Errorf("no repositories discovered from any provider")
	}

	return nil
}

// selectRepositories presents discovered repositories for user selection
func (w *Wizard) selectRepositories() map[string][]config.Repository {
	w.printDiscoveryResults()

	if !w.config.Interactive {
		// Non-interactive mode: select all repositories
		return w.getAllDiscoveredRepositories()
	}

	fmt.Println("\n=== Repository Selection ===")
	fmt.Println("Choose how to select repositories:")
	fmt.Println("1. Select all repositories")
	fmt.Println("2. Select by provider")
	fmt.Println("3. Select individually")
	fmt.Println("4. Apply filters and select")

	var choice int
	fmt.Print("Enter choice (1-4): ")
	_, _ = fmt.Scanln(&choice) // Ignore read errors

	switch choice {
	case 1:
		return w.selectAllRepositories()
	case 2:
		return w.selectByProvider()
	case 3:
		return w.selectIndividually()
	case 4:
		return w.selectWithFilters()
	default:
		fmt.Println("Invalid choice, selecting all repositories")
		return w.selectAllRepositories()
	}
}

// printDiscoveryResults prints the discovery results summary
func (w *Wizard) printDiscoveryResults() {
	fmt.Println("\n=== Repository Discovery Results ===")

	for provider, repos := range w.discoveries {
		fmt.Printf("%s: %d repositories\n", provider, len(repos))

		// Show sample repositories
		if len(repos) > 0 {
			fmt.Println("  Sample repositories:")
			count := len(repos)
			if count > 5 {
				count = 5
			}
			for i := 0; i < count; i++ {
				repo := repos[i]
				visibility := "public"
				if repo.IsPrivate {
					visibility = "private"
				}
				fmt.Printf("    - %s (%s, %s)\n", repo.FullName, repo.Language, visibility)
			}
			if len(repos) > 5 {
				fmt.Printf("    ... and %d more\n", len(repos)-5)
			}
		}
		fmt.Println()
	}
}

// selectAllRepositories selects all discovered repositories
func (w *Wizard) selectAllRepositories() map[string][]config.Repository {
	fmt.Println("Selecting all discovered repositories...")
	return w.getAllDiscoveredRepositories()
}

// selectByProvider allows selection by provider
func (w *Wizard) selectByProvider() map[string][]config.Repository {
	selected := make(map[string][]config.Repository)

	for provider, repos := range w.discoveries {
		fmt.Printf("\nInclude all repositories from %s? [Y/n]: ", provider)
		var response string
		_, _ = fmt.Scanln(&response) // Ignore read errors

		if response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			for _, repo := range repos {
				selected[provider] = append(selected[provider], w.convertRepository(repo))
			}
			fmt.Printf("Selected %d repositories from %s\n", len(repos), provider)
		}
	}

	return selected
}

// selectIndividually allows individual repository selection
func (w *Wizard) selectIndividually() map[string][]config.Repository {
	selected := make(map[string][]config.Repository)

	for provider, repos := range w.discoveries {
		fmt.Printf("\n=== %s Repositories ===\n", provider)

		for _, repo := range repos {
			visibility := "public"
			if repo.IsPrivate {
				visibility = "private"
			}

			fmt.Printf("Include %s (%s, %s, %d stars)? [y/N]: ",
				repo.FullName, repo.Language, visibility, repo.StarCount)

			var response string
			_, _ = fmt.Scanln(&response) // Ignore read errors

			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				selected[provider] = append(selected[provider], w.convertRepository(repo))
			}
		}
	}

	return selected
}

// selectWithFilters applies filters and presents filtered repositories
func (w *Wizard) selectWithFilters() map[string][]config.Repository {
	fmt.Println("\n=== Repository Filters ===")

	filters := FilterOptions{
		IncludePrivate:  true,
		IncludePublic:   true,
		IncludeForks:    false,
		IncludeArchived: false,
	}

	// Get filter preferences
	fmt.Print("Include private repositories? [Y/n]: ")
	var response string
	_, _ = fmt.Scanln(&response) // Ignore read errors
	filters.IncludePrivate = response == "" || strings.ToLower(response) == "y"

	fmt.Print("Include public repositories? [Y/n]: ")
	_, _ = fmt.Scanln(&response) // Ignore read errors
	filters.IncludePublic = response == "" || strings.ToLower(response) == "y"

	fmt.Print("Include forked repositories? [y/N]: ")
	_, _ = fmt.Scanln(&response) // Ignore read errors
	filters.IncludeForks = strings.ToLower(response) == "y"

	fmt.Print("Minimum stars (0 for no limit): ")
	_, _ = fmt.Scanln(&filters.MinStars) // Ignore read errors

	fmt.Print("Filter by language (empty for no filter): ")
	_, _ = fmt.Scanln(&filters.LanguageFilter) // Ignore read errors

	fmt.Print("Filter by name (empty for no filter): ")
	_, _ = fmt.Scanln(&filters.NameFilter) // Ignore read errors

	// Apply filters and select
	filtered := w.applyFilters(filters)
	return w.confirmFilteredSelection(filtered)
}

// Helper functions

// getAllDiscoveredRepositories returns all discovered repositories
func (w *Wizard) getAllDiscoveredRepositories() map[string][]config.Repository {
	selected := make(map[string][]config.Repository)

	for provider, repos := range w.discoveries {
		for _, repo := range repos {
			selected[provider] = append(selected[provider], w.convertRepository(repo))
		}
	}

	return selected
}

// convertRepository converts a common.Repository to config.Repository
func (w *Wizard) convertRepository(repo common.Repository) config.Repository {
	return config.Repository{
		Name:          repo.FullName,
		AutoMerge:     true,
		MergeStrategy: config.MergeStrategySquash,
		RequireChecks: true,
		SkipLabels:    []string{},
	}
}

// applyFilters applies the specified filters to discovered repositories
func (w *Wizard) applyFilters(filters FilterOptions) map[string][]common.Repository {
	filtered := make(map[string][]common.Repository)

	for provider, repos := range w.discoveries {
		for _, repo := range repos {
			if w.matchesFilters(repo, filters) {
				filtered[provider] = append(filtered[provider], repo)
			}
		}
	}

	return filtered
}

// matchesFilters checks if a repository matches the specified filters
func (w *Wizard) matchesFilters(repo common.Repository, filters FilterOptions) bool {
	// Visibility filter
	if repo.IsPrivate && !filters.IncludePrivate {
		return false
	}
	if !repo.IsPrivate && !filters.IncludePublic {
		return false
	}

	// Fork filter
	if repo.IsFork && !filters.IncludeForks {
		return false
	}

	// Archived filter
	if repo.IsArchived && !filters.IncludeArchived {
		return false
	}

	// Stars filter
	if repo.StarCount < filters.MinStars {
		return false
	}

	// Language filter
	if filters.LanguageFilter != "" && !strings.EqualFold(repo.Language, filters.LanguageFilter) {
		return false
	}

	// Name filter
	if filters.NameFilter != "" && !strings.Contains(strings.ToLower(repo.FullName), strings.ToLower(filters.NameFilter)) {
		return false
	}

	return true
}

// confirmFilteredSelection confirms the filtered selection
func (w *Wizard) confirmFilteredSelection(filtered map[string][]common.Repository) map[string][]config.Repository {
	fmt.Println("\n=== Filtered Results ===")

	totalCount := 0
	for provider, repos := range filtered {
		fmt.Printf("%s: %d repositories\n", provider, len(repos))
		totalCount += len(repos)
	}

	if totalCount == 0 {
		fmt.Println("No repositories match the specified filters.")
		return make(map[string][]config.Repository)
	}

	fmt.Printf("\nTotal: %d repositories\n", totalCount)
	fmt.Print("Proceed with filtered selection? [Y/n]: ")

	var response string
	_, _ = fmt.Scanln(&response) // Ignore read errors

	if response != "" && strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		return make(map[string][]config.Repository)
	}

	// Convert to config repositories
	selected := make(map[string][]config.Repository)
	for provider, repos := range filtered {
		for _, repo := range repos {
			selected[provider] = append(selected[provider], w.convertRepository(repo))
		}
	}

	return selected
}

// buildConfiguration builds the final configuration
func (w *Wizard) buildConfiguration(selected map[string][]config.Repository, existing *config.Config) *config.Config {
	cfg := &config.Config{
		PRFilters: config.PRFilters{
			AllowedActors: []string{
				"dependabot[bot]",
				"renovate[bot]",
				"github-actions[bot]",
			},
			SkipLabels: []string{
				"do-not-merge",
				"wip",
				"work-in-progress",
				"on-hold",
			},
			MaxAge: "30d",
		},
		Repositories: selected,
		Auth: config.Auth{
			GitHub: config.GitHubAuth{
				Token: "$GITHUB_TOKEN",
			},
			GitLab: config.GitLabAuth{
				Token: "$GITLAB_TOKEN",
				URL:   "$GITLAB_URL",
			},
			Bitbucket: config.BitbucketAuth{
				Username:    "$BITBUCKET_USERNAME",
				AppPassword: "$BITBUCKET_APP_PASSWORD",
				Workspace:   "$BITBUCKET_WORKSPACE",
			},
		},
		Notifications: config.Notifications{
			Slack: config.SlackConfig{
				WebhookURL: "$SLACK_WEBHOOK_URL",
				Channel:    "#git-pr-automation",
				Enabled:    true,
			},
		},
		Behavior: config.Behavior{
			Concurrency: 10,
			DryRun:      false,
			RateLimit: config.RateLimit{
				RequestsPerSecond: 5.0,
				Burst:             10,
				Timeout:           30 * time.Second,
			},
		},
	}

	// Merge with existing config if additive
	if existing != nil {
		cfg = w.mergeConfigurations(cfg, existing)
	}

	return cfg
}

// mergeConfigurations merges new configuration with existing configuration
func (w *Wizard) mergeConfigurations(newCfg, existing *config.Config) *config.Config {
	// Use existing configuration as base
	merged := *existing

	// Merge repositories
	if merged.Repositories == nil {
		merged.Repositories = make(map[string][]config.Repository)
	}

	for provider, repos := range newCfg.Repositories {
		// Add new repositories, avoiding duplicates
		existingRepos := merged.Repositories[provider]

		for _, newRepo := range repos {
			exists := false
			for _, existingRepo := range existingRepos {
				if existingRepo.Name == newRepo.Name {
					exists = true
					break
				}
			}
			if !exists {
				merged.Repositories[provider] = append(merged.Repositories[provider], newRepo)
			}
		}
	}

	return &merged
}

// loadExistingConfig loads existing configuration for additive mode
func (w *Wizard) loadExistingConfig() (*config.Config, error) {
	if _, err := os.Stat(w.config.ConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", w.config.ConfigPath)
	}

	loader := config.NewLoader()
	return loader.Load(w.config.ConfigPath)
}

// previewConfiguration shows the configuration that would be created
func (w *Wizard) previewConfiguration(cfg *config.Config) error {
	fmt.Println("\n=== Configuration Preview ===")

	// Count repositories
	totalRepos := 0
	for provider, repos := range cfg.Repositories {
		fmt.Printf("%s: %d repositories\n", provider, len(repos))
		totalRepos += len(repos)
	}

	fmt.Printf("Total repositories: %d\n", totalRepos)
	fmt.Printf("Allowed actors: %v\n", cfg.PRFilters.AllowedActors)
	fmt.Printf("Concurrency: %d\n", cfg.Behavior.Concurrency)

	fmt.Println("\nThis configuration would be written to:", w.config.ConfigPath)
	fmt.Println("Run the wizard without --preview to create the configuration file.")

	return nil
}

// saveConfiguration saves the configuration to file
func (w *Wizard) saveConfiguration(cfg *config.Config) error {
	// Use sophisticated backup from setup commands
	if _, err := os.Stat(w.config.ConfigPath); err == nil {
		if err := w.backupConfig(); err != nil {
			w.logger.Warnf("Failed to backup existing configuration: %v", err)
		}
	}

	// Save configuration using the loader
	loader := config.NewLoader()
	if err := loader.Save(cfg, w.config.ConfigPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	w.logger.Infof("Configuration saved to %s", w.config.ConfigPath)

	// Show next steps
	fmt.Printf("\n=== Setup Complete ===\n")
	fmt.Printf("Configuration saved to: %s\n", w.config.ConfigPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Review and customize the generated configuration")
	fmt.Println("2. Set up your environment variables (see README.md)")
	fmt.Println("3. Run 'git-pr-cli validate' to test your setup")
	fmt.Println("4. Run 'git-pr-cli check' to see available PRs")

	return nil
}

// backupConfig creates a timestamped backup of existing configuration
func (w *Wizard) backupConfig() error {
	if _, err := os.Stat(w.config.ConfigPath); os.IsNotExist(err) {
		return nil // No existing config to backup
	}

	// Create backup directory
	backupDir := filepath.Join(filepath.Dir(w.config.ConfigPath), ".backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	basename := filepath.Base(w.config.ConfigPath)
	timestamp := utils.FormatTimestamp()
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s.bak", basename, timestamp))

	// Copy file
	data, err := os.ReadFile(w.config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = os.WriteFile(backupPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	w.logger.Infof("Configuration backed up to %s", backupPath)
	return nil
}
