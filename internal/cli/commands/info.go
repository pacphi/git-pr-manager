package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pacphi/git-pr-manager/internal/executor"
	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// InfoFlags contains flags for the info command
type InfoFlags struct {
	ConfigPath    string
	ShowProviders bool
	ShowConfig    bool
	ShowBehavior  bool
	Verbose       bool
}

// NewInfoCommand creates the info command
func NewInfoCommand() *cobra.Command {
	var flags InfoFlags

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show configuration and provider information",
		Long: `Show detailed information about the current configuration and providers.

This command displays:
- Configured providers and their status
- Current configuration settings
- Behavior management statistics (rate limiting, retry)
- Repository configuration summary`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "config.yaml", "configuration file path")
	cmd.Flags().BoolVar(&flags.ShowProviders, "providers", true, "show provider information")
	cmd.Flags().BoolVar(&flags.ShowConfig, "config-details", false, "show detailed configuration")
	cmd.Flags().BoolVar(&flags.ShowBehavior, "behavior", false, "show behavior management statistics")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "verbose output")

	return cmd
}

// runInfo displays configuration and provider information
func runInfo(ctx context.Context, flags InfoFlags) error {
	logger := utils.GetGlobalLogger()

	// Load configuration
	var cfg *config.Config
	var err error
	if flags.ConfigPath != "" && flags.ConfigPath != "config.yaml" {
		cfg, err = config.LoadConfigFromPath(flags.ConfigPath)
	} else {
		cfg, err = LoadConfig()
	}
	if err != nil {
		return HandleConfigError(err, "info")
	}

	// Create executor to get provider information
	exec, err := executor.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	fmt.Println("Git PR Automation - Configuration Information")
	fmt.Println("=" + strings.Repeat("=", 44))
	fmt.Println()

	// Show provider information
	if flags.ShowProviders {
		fmt.Println("📡 Configured Providers:")
		providers := exec.GetProviders()
		if len(providers) == 0 {
			fmt.Println("  No providers configured")
		} else {
			for name, provider := range providers {
				fmt.Printf("  • %s (%s)\n", name, provider.GetProviderName())
				if flags.Verbose {
					// Test authentication status
					if err := provider.Authenticate(ctx); err != nil {
						fmt.Printf("    Status: ❌ Authentication failed (%v)\n", err)
					} else {
						fmt.Printf("    Status: ✅ Authenticated\n")
					}
				}
			}
		}
		fmt.Println()
	}

	// Show repository summary
	fmt.Println("📁 Repository Configuration:")
	totalRepos := 0
	for providerName, repos := range cfg.Repositories {
		count := len(repos)
		totalRepos += count
		fmt.Printf("  • %s: %d repositories\n", providerName, count)
		if flags.Verbose {
			for _, repo := range repos {
				fmt.Printf("    - %s\n", repo.Name)
			}
		}
	}
	fmt.Printf("  Total: %d repositories\n", totalRepos)
	fmt.Println()

	// Show behavior configuration
	if flags.ShowBehavior {
		fmt.Println("⚙️  Behavior Management:")
		fmt.Printf("  Rate Limiting:\n")
		fmt.Printf("    Requests per second: %.1f\n", cfg.Behavior.RateLimit.RequestsPerSecond)
		fmt.Printf("    Burst: %d\n", cfg.Behavior.RateLimit.Burst)
		fmt.Printf("    Timeout: %v\n", cfg.Behavior.RateLimit.Timeout)
		fmt.Printf("  Retry Configuration:\n")
		fmt.Printf("    Max attempts: %d\n", cfg.Behavior.Retry.MaxAttempts)
		fmt.Printf("    Initial backoff: %v\n", cfg.Behavior.Retry.Backoff)
		fmt.Printf("    Max backoff: %v\n", cfg.Behavior.Retry.MaxBackoff)
		fmt.Printf("  Concurrency: %d\n", cfg.Behavior.Concurrency)
		fmt.Printf("  Dry run: %t\n", cfg.Behavior.DryRun)
		fmt.Println()
	}

	// Show PR filters
	fmt.Println("🔍 PR Filters:")
	fmt.Printf("  Allowed actors: %s\n", strings.Join(cfg.PRFilters.AllowedActors, ", "))
	if len(cfg.PRFilters.SkipLabels) > 0 {
		fmt.Printf("  Skip labels: %s\n", strings.Join(cfg.PRFilters.SkipLabels, ", "))
	}
	if cfg.PRFilters.MaxAge != "" {
		fmt.Printf("  Max age: %s\n", cfg.PRFilters.MaxAge)
	}
	fmt.Println()

	// Show notification configuration
	hasNotifications := cfg.Notifications.Slack.Enabled || cfg.Notifications.Email.Enabled
	fmt.Println("📢 Notifications:")
	if !hasNotifications {
		fmt.Println("  No notifications configured")
	} else {
		if cfg.Notifications.Slack.Enabled {
			fmt.Printf("  • Slack: ✅ Enabled (channel: %s)\n", cfg.Notifications.Slack.Channel)
		}
		if cfg.Notifications.Email.Enabled {
			fmt.Printf("  • Email: ✅ Enabled (%d recipients)\n", len(cfg.Notifications.Email.To))
		}
	}
	fmt.Println()

	// Show detailed configuration if requested
	if flags.ShowConfig {
		fmt.Println("📄 Detailed Configuration:")
		cfg := exec.GetConfig()

		// Show authentication status
		fmt.Println("  Authentication:")
		if cfg.Auth.GitHub.Token != "" {
			fmt.Println("    • GitHub: Configured")
		}
		if cfg.Auth.GitLab.Token != "" {
			fmt.Println("    • GitLab: Configured")
		}
		if cfg.Auth.Bitbucket.Username != "" {
			fmt.Println("    • Bitbucket: Configured")
		}
		fmt.Println()
	}

	logger.Info("Configuration information displayed successfully")
	return nil
}
