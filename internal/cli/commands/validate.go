package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cphillipson/multi-gitter-pr-automation/internal/executor"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/validation"
)

// ValidateFlags contains flags for the validate command
type ValidateFlags struct {
	ConfigPath  string
	CheckAuth   bool
	CheckRepos  bool
	Verbose     bool
	CheckConfig bool
	CheckDeps   bool
	Provider    string
	Repos       string
	ShowDetails bool
	Timeout     string
}

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	var flags ValidateFlags

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration and connectivity",
		Long: `Validates the configuration file and tests connectivity to Git providers.

This command performs the following checks:
- Configuration file syntax and structure
- Required fields and valid values
- Authentication with configured providers
- Access to configured repositories (optional)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "config.yaml", "configuration file path")
	cmd.Flags().BoolVar(&flags.CheckAuth, "check-auth", true, "verify authentication with providers")
	cmd.Flags().BoolVar(&flags.CheckRepos, "check-repos", false, "verify access to configured repositories")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().BoolVar(&flags.CheckConfig, "check-config", true, "validate configuration file structure and syntax")
	cmd.Flags().BoolVar(&flags.CheckDeps, "check-deps", false, "check system dependencies and requirements")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", "validate specific provider only (github, gitlab, bitbucket)")
	cmd.Flags().StringVar(&flags.Repos, "repos", "", "validate specific repositories only (comma-separated or pattern)")
	cmd.Flags().BoolVar(&flags.ShowDetails, "show-details", false, "show detailed validation information")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "timeout duration for validation operations")

	return cmd
}

// runValidate performs configuration validation
func runValidate(ctx context.Context, flags ValidateFlags) error {
	logger := utils.GetGlobalLogger()

	// Parse timeout
	timeout, err := time.ParseDuration(flags.Timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout duration: %w", err)
	}

	// Create context with timeout
	validateCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	logger.Info("Starting configuration validation...")
	if flags.ShowDetails {
		logger.Infof("Validation timeout: %s", flags.Timeout)
		if flags.Provider != "" {
			logger.Infof("Provider filter: %s", flags.Provider)
		}
		if flags.Repos != "" {
			logger.Infof("Repository filter: %s", flags.Repos)
		}
	}

	var cfg *config.Config

	// Step 1: Load and validate configuration (if enabled)
	if flags.CheckConfig {
		logger.Info("Validating configuration file...")
		if flags.ConfigPath != "" && flags.ConfigPath != "config.yaml" {
			cfg, err = config.LoadConfigFromPath(flags.ConfigPath)
		} else {
			cfg, err = LoadConfig()
		}
		if err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		// Create validator and validate configuration structure
		validator := validation.New()
		if err := validator.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		// Validate behavior configuration
		logger.Info("Validating behavior configuration...")
		if err := utils.ValidateBehaviorConfig(cfg); err != nil {
			return fmt.Errorf("behavior configuration validation failed: %w", err)
		}

		// Validate provider configurations
		logger.Info("Validating provider configurations...")
		for providerName := range cfg.Repositories {
			provider := config.Provider(providerName)
			if !provider.IsValid() {
				return fmt.Errorf("invalid provider name: %s", providerName)
			}
		}
		logger.Info("✅ Configuration file is valid")

		// Check environment variables
		logger.Info("Checking environment variables...")
		missing := validator.CheckEnvironmentVariables(cfg)
		if len(missing) > 0 {
			logger.Warnf("Missing environment variables: %v", missing)
			if flags.ShowDetails {
				for _, varName := range missing {
					logger.Warnf("  - %s", varName)
				}
			}
		} else {
			logger.Info("✅ All required environment variables are set")
		}
	} else {
		// Still need to load config for other validations
		cfg, err = LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Step 2: Check system dependencies (if enabled)
	if flags.CheckDeps {
		logger.Info("Checking system dependencies...")
		if err := checkSystemDependencies(flags.ShowDetails); err != nil {
			logger.Warnf("System dependencies check failed: %v", err)
		} else {
			logger.Info("✅ System dependencies verified")
		}
	}

	// Step 3: Test provider authentication (filtered by provider if specified)
	if flags.CheckAuth {
		logger.Info("Testing provider authentication...")
		if err := validateProviderAuth(validateCtx, cfg, flags); err != nil {
			return fmt.Errorf("authentication validation failed: %w", err)
		}
		logger.Info("✅ Provider authentication successful")
	}

	// Step 4: Check repository access (filtered by repos if specified)
	if flags.CheckRepos {
		logger.Info("Checking repository access...")
		if err := validateRepositoryAccess(validateCtx, cfg, flags); err != nil {
			logger.Warnf("Repository access check failed: %v", err)
			logger.Info("Some repositories may not be accessible or may not exist")
		} else {
			logger.Info("✅ Repository access verified")
		}
	}

	// Step 5: Validation summary
	logger.Info("Configuration validation completed successfully")

	// Print summary
	printValidationSummary(cfg, flags.ShowDetails || flags.Verbose)

	return nil
}

// validateProviderAuth tests authentication with configured providers
func validateProviderAuth(ctx context.Context, cfg *config.Config, flags ValidateFlags) error {
	logger := utils.GetGlobalLogger()

	// Use the executor's TestAuthentication method for comprehensive testing
	exec, err := executor.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// If provider filter is specified, test only that provider
	if flags.Provider != "" {
		providers := exec.GetProviders()
		provider, exists := providers[flags.Provider]
		if !exists {
			return fmt.Errorf("provider %s not configured", flags.Provider)
		}

		if flags.Verbose || flags.ShowDetails {
			logger.Infof("Testing %s authentication...", flags.Provider)
		}

		if err := provider.Authenticate(ctx); err != nil {
			return fmt.Errorf("authentication failed for %s: %w", flags.Provider, err)
		}

		if flags.Verbose || flags.ShowDetails {
			logger.Infof("✅ %s authentication successful", flags.Provider)
		}
		return nil
	}

	// Test all providers using executor's method
	return exec.TestAuthentication(ctx)
}

// validateRepositoryAccess tests access to configured repositories
func validateRepositoryAccess(ctx context.Context, cfg *config.Config, flags ValidateFlags) error {
	logger := utils.GetGlobalLogger()

	factory := providers.NewFactory(cfg)
	providerMap, err := factory.CreateProviders()
	if err != nil {
		return fmt.Errorf("failed to create providers: %w", err)
	}

	totalRepos := 0
	accessibleRepos := 0
	var errors []error

	for providerName, repos := range cfg.Repositories {
		// Skip if provider filter is specified and doesn't match
		if flags.Provider != "" && flags.Provider != providerName {
			continue
		}

		provider, exists := providerMap[providerName]
		if !exists {
			logger.Warnf("Provider %s not available, skipping repository checks", providerName)
			continue
		}

		for _, repo := range repos {
			// Skip if repos filter is specified and doesn't match
			if flags.Repos != "" && !matchesRepoFilter(repo.Name, flags.Repos) {
				continue
			}

			totalRepos++
			if flags.Verbose || flags.ShowDetails {
				logger.Infof("Checking access to %s...", repo.Name)
			}

			owner, name, err := parseRepository(repo.Name)
			if err != nil {
				errors = append(errors, fmt.Errorf("invalid repository name %s: %w", repo.Name, err))
				continue
			}

			_, err = provider.GetRepository(ctx, owner, name)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to access %s: %w", repo.Name, err))
				if flags.Verbose || flags.ShowDetails {
					logger.Warnf("❌ Cannot access %s: %v", repo.Name, err)
				}
			} else {
				accessibleRepos++
				if flags.Verbose || flags.ShowDetails {
					logger.Infof("✅ %s is accessible", repo.Name)
				}
			}
		}
	}

	logger.Infof("Repository access check: %d/%d repositories accessible", accessibleRepos, totalRepos)

	if len(errors) > 0 && (flags.Verbose || flags.ShowDetails) {
		logger.Warn("Repository access errors:")
		for _, err := range errors {
			logger.Warnf("  %v", err)
		}
	}

	// Don't return error if some repos are inaccessible, just warn
	return nil
}

// printValidationSummary prints a summary of the validation results
func printValidationSummary(cfg *config.Config, verbose bool) {
	logger := utils.GetGlobalLogger()

	logger.Info("\n=== Validation Summary ===")

	// Count configured items
	totalRepos := 0
	for _, repos := range cfg.Repositories {
		totalRepos += len(repos)
	}

	logger.Infof("Providers configured: %d", len(cfg.Repositories))
	logger.Infof("Total repositories: %d", totalRepos)
	logger.Infof("Allowed actors: %d", len(cfg.PRFilters.AllowedActors))

	if verbose {
		logger.Info("\n=== Configuration Details ===")

		// Show providers
		for provider := range cfg.Repositories {
			logger.Infof("Provider: %s", provider)
		}

		// Show allowed actors
		if len(cfg.PRFilters.AllowedActors) > 0 {
			logger.Info("Allowed actors:")
			for _, actor := range cfg.PRFilters.AllowedActors {
				logger.Infof("  - %s", actor)
			}
		}

		// Show skip labels
		if len(cfg.PRFilters.SkipLabels) > 0 {
			logger.Info("Skip labels:")
			for _, label := range cfg.PRFilters.SkipLabels {
				logger.Infof("  - %s", label)
			}
		}

		// Show behavior settings
		logger.Infof("Concurrency: %d", cfg.Behavior.Concurrency)
		logger.Infof("Dry run: %t", cfg.Behavior.DryRun)
		if cfg.PRFilters.MaxAge != "" {
			logger.Infof("Max PR age: %s", cfg.PRFilters.MaxAge)
		}
	}
}

// checkSystemDependencies checks for required system dependencies
func checkSystemDependencies(showDetails bool) error {
	logger := utils.GetGlobalLogger()

	dependencies := []struct {
		name     string
		command  string
		args     []string
		required bool
	}{
		{"Git", "git", []string{"--version"}, true},
		{"curl", "curl", []string{"--version"}, false},
		{"jq", "jq", []string{"--version"}, false},
	}

	var errors []error
	passedCount := 0

	for _, dep := range dependencies {
		if showDetails {
			logger.Infof("Checking %s...", dep.name)
		}

		cmd := exec.Command(dep.command, dep.args...)
		if err := cmd.Run(); err != nil {
			if dep.required {
				errors = append(errors, fmt.Errorf("required dependency %s not found", dep.name))
				if showDetails {
					logger.Errorf("❌ %s not found (required)", dep.name)
				}
			} else {
				if showDetails {
					logger.Warnf("⚠️ %s not found (optional)", dep.name)
				}
			}
		} else {
			passedCount++
			if showDetails {
				logger.Infof("✅ %s found", dep.name)
			}
		}
	}

	if showDetails {
		logger.Infof("System dependencies: %d/%d available", passedCount, len(dependencies))
	}

	if len(errors) > 0 {
		return fmt.Errorf("missing required dependencies: %v", errors)
	}

	return nil
}

// matchesRepoFilter checks if a repository name matches the filter
func matchesRepoFilter(repoName, filter string) bool {
	if filter == "" {
		return true
	}

	// Support comma-separated list
	filters := strings.Split(filter, ",")
	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f == repoName || strings.Contains(repoName, f) {
			return true
		}
	}

	return false
}

// parseRepository parses a repository name into owner and name parts
func parseRepository(repoName string) (owner, name string, err error) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("repository name must be in format 'owner/name'")
	}
	return parts[0], parts[1], nil
}
