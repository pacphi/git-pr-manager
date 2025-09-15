package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/notifications"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// TestFlags contains flags for the test command
type TestFlags struct {
	TestNotifications bool
	TestSlack         bool
	TestEmail         bool
	Auth              bool
	Config            string
	Integration       bool
	LoadTest          bool
	MockPRs           bool
	Performance       bool
	Provider          string
	Repos             string
	Timeout           string
	Verbose           bool
}

// NewTestCommand creates the test command
func NewTestCommand() *cobra.Command {
	var flags TestFlags

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test system functionality and integrations",
		Long: `Test various aspects of the Git PR automation system.

Available tests:
- Notification systems (Slack, email)
- Provider authentication (automatically tested)
- Configuration validation (use validate command)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest(cmd.Context(), flags)
		},
	}

	cmd.Flags().BoolVar(&flags.TestNotifications, "notifications", false, "test notification systems")
	cmd.Flags().BoolVar(&flags.TestSlack, "slack", false, "test Slack notifications only")
	cmd.Flags().BoolVar(&flags.TestEmail, "email", false, "test email notifications only")
	cmd.Flags().BoolVar(&flags.Auth, "auth", false, "test provider authentication")
	cmd.Flags().StringVar(&flags.Config, "config", "", "path to configuration file")
	cmd.Flags().BoolVar(&flags.Integration, "integration", false, "run integration tests")
	cmd.Flags().BoolVar(&flags.LoadTest, "load-test", false, "run load tests")
	cmd.Flags().BoolVar(&flags.MockPRs, "mock-prs", false, "test with mock pull requests")
	cmd.Flags().BoolVar(&flags.Performance, "performance", false, "run performance tests")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", "test specific provider (github, gitlab, bitbucket)")
	cmd.Flags().StringVar(&flags.Repos, "repos", "", "test specific repositories (comma-separated or pattern)")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "test timeout duration")
	cmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "enable verbose test output")

	// If no specific flags are provided, test notifications by default
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !flags.TestNotifications && !flags.TestSlack && !flags.TestEmail && !flags.Auth &&
			!flags.Integration && !flags.LoadTest && !flags.MockPRs && !flags.Performance {
			flags.TestNotifications = true
		}
		return nil
	}

	return cmd
}

// runTest performs the requested tests
func runTest(ctx context.Context, flags TestFlags) error {
	logger := utils.GetGlobalLogger()

	// Parse timeout
	timeout, err := time.ParseDuration(flags.Timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout duration: %w", err)
	}

	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	logger.Info("Starting system tests...")
	if flags.Verbose {
		logger.Infof("Test configuration: timeout=%s, provider=%s, repos=%s",
			flags.Timeout, flags.Provider, flags.Repos)
	}

	// Load configuration
	var cfg *config.Config
	if flags.Config != "" {
		cfg, err = config.LoadConfigFromPath(flags.Config)
	} else {
		cfg, err = LoadConfig()
	}
	if err != nil {
		return HandleConfigError(err, "test")
	}

	var testsPassed, testsFailed int

	// Test authentication if requested
	if flags.Auth {
		logger.Info("Testing provider authentication...")
		passed, failed := testAuthentication(testCtx, cfg, flags)
		testsPassed += passed
		testsFailed += failed
	}

	// Test notifications if requested
	if flags.TestNotifications || flags.TestSlack || flags.TestEmail {
		logger.Info("Testing notification systems...")
		passed, failed := testNotifications(testCtx, cfg, flags)
		testsPassed += passed
		testsFailed += failed
	}

	// Run integration tests if requested
	if flags.Integration {
		logger.Info("Running integration tests...")
		passed, failed := testIntegration(testCtx, cfg, flags)
		testsPassed += passed
		testsFailed += failed
	}

	// Run load tests if requested
	if flags.LoadTest {
		logger.Info("Running load tests...")
		passed, failed := testLoad(testCtx, cfg, flags)
		testsPassed += passed
		testsFailed += failed
	}

	// Test with mock PRs if requested
	if flags.MockPRs {
		logger.Info("Testing with mock pull requests...")
		passed, failed := testMockPRs(testCtx, cfg, flags)
		testsPassed += passed
		testsFailed += failed
	}

	// Run performance tests if requested
	if flags.Performance {
		logger.Info("Running performance tests...")
		passed, failed := testPerformance(testCtx, cfg, flags)
		testsPassed += passed
		testsFailed += failed
	}

	// Print test summary
	logger.Info("=== Test Summary ===")
	logger.Infof("Tests passed: %d", testsPassed)
	logger.Infof("Tests failed: %d", testsFailed)

	if testsFailed > 0 {
		logger.Error("Some tests failed. Please check your configuration and setup.")
		return fmt.Errorf("%d tests failed", testsFailed)
	}

	logger.Info("All tests passed successfully!")
	return nil
}

// testNotifications tests the notification systems
func testNotifications(ctx context.Context, cfg *config.Config, flags TestFlags) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	// Create notification manager
	notificationManager, err := notifications.NewManager(cfg)
	if err != nil {
		logger.WithError(err).Error("Failed to create notification manager")
		return 0, 1
	}

	if !notificationManager.HasNotifiers() {
		logger.Warn("No notification systems configured - skipping notification tests")
		return 0, 1
	}

	// Test individual notification systems if specific flags are set
	if flags.TestSlack && !flags.TestEmail {
		return testSlackOnly(ctx, cfg)
	}

	if flags.TestEmail && !flags.TestSlack {
		return testEmailOnly(ctx, cfg)
	}

	// Test all configured notification systems
	logger.Info("Testing all configured notification systems...")

	if err := notificationManager.SendTestMessage(ctx); err != nil {
		logger.WithError(err).Error("Notification test failed")
		return 0, 1
	}

	logger.Info("✅ Notification test passed")
	return 1, 0
}

// testSlackOnly tests only Slack notifications
func testSlackOnly(ctx context.Context, cfg *config.Config) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	logger.Info("Testing Slack notifications...")

	if cfg.Notifications.Slack.WebhookURL == "" {
		logger.Error("Slack webhook URL not configured")
		return 0, 1
	}

	// Create Slack notifier directly
	slackConfig := notifications.SlackConfig{
		WebhookURL: resolveEnvVar(cfg.Notifications.Slack.WebhookURL),
		Channel:    cfg.Notifications.Slack.Channel,
		Username:   "Git PR Bot",
	}

	slackNotifier := notifications.NewSlackNotifier(slackConfig)

	if err := slackNotifier.SendTestMessage(ctx); err != nil {
		logger.WithError(err).Error("Slack test failed")
		return 0, 1
	}

	logger.Info("✅ Slack test passed")
	return 1, 0
}

// testEmailOnly tests only email notifications
func testEmailOnly(ctx context.Context, cfg *config.Config) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	logger.Info("Testing email notifications...")

	if cfg.Notifications.Email.SMTPHost == "" {
		logger.Error("Email SMTP host not configured")
		return 0, 1
	}

	if len(cfg.Notifications.Email.To) == 0 {
		logger.Error("Email recipients not configured")
		return 0, 1
	}

	// Create email notifier directly
	emailConfig := notifications.EmailConfig{
		SMTPHost: cfg.Notifications.Email.SMTPHost,
		SMTPPort: cfg.Notifications.Email.SMTPPort,
		Username: resolveEnvVar(cfg.Notifications.Email.SMTPUsername),
		Password: resolveEnvVar(cfg.Notifications.Email.SMTPPassword),
		From:     cfg.Notifications.Email.From,
		To:       cfg.Notifications.Email.To,
		UseTLS:   true,
	}

	emailNotifier := notifications.NewEmailNotifier(emailConfig)

	if err := emailNotifier.SendTestMessage(ctx); err != nil {
		logger.WithError(err).Error("Email test failed")
		return 0, 1
	}

	logger.Info("✅ Email test passed")
	return 1, 0
}

// testAuthentication tests provider authentication
func testAuthentication(ctx context.Context, cfg *config.Config, flags TestFlags) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	providerConfigs := getFilteredProviderConfigs(cfg, flags.Provider, flags.Repos)

	for providerType, repos := range providerConfigs {
		if flags.Verbose {
			logger.Infof("Testing authentication for %s provider...", providerType)
		}

		provider, err := providers.CreateProvider(providerType, repos)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create %s provider", providerType)
			failed++
			continue
		}

		// Test authentication by making a simple API call
		if err := testProviderAuthentication(ctx, provider, providerType); err != nil {
			logger.WithError(err).Errorf("Authentication test failed for %s", providerType)
			failed++
		} else {
			logger.Infof("✅ Authentication test passed for %s", providerType)
			passed++
		}
	}

	return passed, failed
}

// testIntegration runs integration tests
func testIntegration(ctx context.Context, cfg *config.Config, flags TestFlags) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	// Integration test: full workflow test
	if flags.Verbose {
		logger.Info("Running full workflow integration test...")
	}

	// Test configuration loading
	if cfg == nil {
		logger.Error("Configuration not loaded")
		return 0, 1
	}

	// Test provider creation and authentication
	authPassed, authFailed := testAuthentication(ctx, cfg, flags)
	passed += authPassed
	failed += authFailed

	// Test notification system if configured
	if cfg.Notifications.Slack.WebhookURL != "" || cfg.Notifications.Email.SMTPHost != "" {
		notifPassed, notifFailed := testNotifications(ctx, cfg, flags)
		passed += notifPassed
		failed += notifFailed
	}

	logger.Info("✅ Integration tests completed")
	return passed, failed
}

// testLoad runs load tests
func testLoad(ctx context.Context, cfg *config.Config, flags TestFlags) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	if flags.Verbose {
		logger.Info("Running load tests...")
	}

	providerConfigs := getFilteredProviderConfigs(cfg, flags.Provider, flags.Repos)

	for providerType, repos := range providerConfigs {
		logger.Infof("Load testing %s provider with %d repositories...", providerType, len(repos))

		provider, err := providers.CreateProvider(providerType, repos)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create %s provider", providerType)
			failed++
			continue
		}

		// Simulate concurrent requests
		start := time.Now()
		concurrency := 5
		results := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(repoIndex int) {
				if repoIndex < len(repos) {
					_, err := provider.ListPullRequests(ctx, repos[repoIndex], common.ListPROptions{})
					results <- err
				} else {
					results <- nil
				}
			}(i)
		}

		loadTestPassed := 0
		loadTestFailed := 0
		for i := 0; i < concurrency; i++ {
			if err := <-results; err != nil {
				loadTestFailed++
			} else {
				loadTestPassed++
			}
		}

		duration := time.Since(start)
		if flags.Verbose {
			logger.Infof("Load test completed in %v: %d passed, %d failed",
				duration, loadTestPassed, loadTestFailed)
		}

		if loadTestFailed == 0 {
			logger.Infof("✅ Load test passed for %s", providerType)
			passed++
		} else {
			logger.Errorf("Load test failed for %s: %d/%d requests failed",
				providerType, loadTestFailed, concurrency)
			failed++
		}
	}

	return passed, failed
}

// testMockPRs tests with mock pull request data
func testMockPRs(_ context.Context, _ *config.Config, flags TestFlags) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	if flags.Verbose {
		logger.Info("Testing with mock pull requests...")
	}

	// Generate mock PR data
	mockPRs := generateMockPRs()

	for i, mockPR := range mockPRs {
		if flags.Verbose {
			logger.Infof("Processing mock PR %d: %s", i+1, mockPR.Title)
		}

		// Test PR processing logic without actual API calls
		if err := processMockPR(mockPR); err != nil {
			logger.WithError(err).Errorf("Failed to process mock PR %d", i+1)
			failed++
		} else {
			passed++
		}
	}

	logger.Infof("✅ Mock PR tests completed: %d passed, %d failed", passed, failed)
	return passed, failed
}

// testPerformance runs performance tests
func testPerformance(ctx context.Context, cfg *config.Config, flags TestFlags) (passed, failed int) {
	logger := utils.GetGlobalLogger()

	if flags.Verbose {
		logger.Info("Running performance tests...")
	}

	providerConfigs := getFilteredProviderConfigs(cfg, flags.Provider, flags.Repos)

	for providerType, repos := range providerConfigs {
		logger.Infof("Performance testing %s provider...", providerType)

		provider, err := providers.CreateProvider(providerType, repos)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create %s provider", providerType)
			failed++
			continue
		}

		// Performance test: measure response times
		var totalDuration time.Duration
		iterations := 3

		for i := 0; i < iterations && len(repos) > 0; i++ {
			start := time.Now()
			_, err := provider.ListPullRequests(ctx, repos[0], common.ListPROptions{})
			duration := time.Since(start)
			totalDuration += duration

			if err != nil {
				logger.WithError(err).Errorf("Performance test iteration %d failed", i+1)
				failed++
				break
			}
		}

		avgDuration := totalDuration / time.Duration(iterations)
		if flags.Verbose {
			logger.Infof("Average response time for %s: %v", providerType, avgDuration)
		}

		// Consider test passed if average response time is under 5 seconds
		if avgDuration < 5*time.Second {
			logger.Infof("✅ Performance test passed for %s (avg: %v)", providerType, avgDuration)
			passed++
		} else {
			logger.Errorf("Performance test failed for %s: avg response time %v exceeds threshold",
				providerType, avgDuration)
			failed++
		}
	}

	return passed, failed
}

// testProviderAuthentication tests authentication for a specific provider
func testProviderAuthentication(ctx context.Context, provider common.Provider, providerType string) error {
	// Try to make a simple API call to test authentication
	repos, err := provider.ListRepositories(ctx)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found for %s provider", providerType)
	}

	return nil
}

// getFilteredProviderConfigs returns provider configs filtered by flags
func getFilteredProviderConfigs(cfg *config.Config, providerFilter, repoFilter string) map[string][]common.Repository {
	result := make(map[string][]common.Repository)

	// Convert config to provider configs
	if providerFilter == "" || providerFilter == "github" {
		if githubRepos, exists := cfg.Repositories["github"]; exists && len(githubRepos) > 0 {
			repos := make([]common.Repository, 0, len(githubRepos))
			for _, repo := range githubRepos {
				if repoFilter == "" || matchesTestRepoFilter(repo.Name, repoFilter) {
					repos = append(repos, common.Repository{
						Name:     repo.Name,
						FullName: repo.Name, // Assuming name contains owner/repo
						Provider: "github",
						WebURL:   fmt.Sprintf("https://github.com/%s", repo.Name),
					})
				}
			}
			if len(repos) > 0 {
				result["github"] = repos
			}
		}
	}

	if providerFilter == "" || providerFilter == "gitlab" {
		if gitlabRepos, exists := cfg.Repositories["gitlab"]; exists && len(gitlabRepos) > 0 {
			repos := make([]common.Repository, 0, len(gitlabRepos))
			for _, repo := range gitlabRepos {
				if repoFilter == "" || matchesTestRepoFilter(repo.Name, repoFilter) {
					repos = append(repos, common.Repository{
						Name:     repo.Name,
						FullName: repo.Name, // Assuming name contains owner/repo
						Provider: "gitlab",
						WebURL:   fmt.Sprintf("https://gitlab.com/%s", repo.Name),
					})
				}
			}
			if len(repos) > 0 {
				result["gitlab"] = repos
			}
		}
	}

	if providerFilter == "" || providerFilter == "bitbucket" {
		if bitbucketRepos, exists := cfg.Repositories["bitbucket"]; exists && len(bitbucketRepos) > 0 {
			repos := make([]common.Repository, 0, len(bitbucketRepos))
			for _, repo := range bitbucketRepos {
				if repoFilter == "" || matchesTestRepoFilter(repo.Name, repoFilter) {
					repos = append(repos, common.Repository{
						Name:     repo.Name,
						FullName: repo.Name, // Assuming name contains owner/repo
						Provider: "bitbucket",
						WebURL:   fmt.Sprintf("https://bitbucket.org/%s", repo.Name),
					})
				}
			}
			if len(repos) > 0 {
				result["bitbucket"] = repos
			}
		}
	}

	return result
}

// matchesTestRepoFilter checks if a repository name matches the filter
func matchesTestRepoFilter(repoName, filter string) bool {
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

// generateMockPRs generates mock pull request data for testing
func generateMockPRs() []MockPR {
	return []MockPR{
		{
			ID:        1,
			Title:     "feat: add new authentication system",
			Author:    "developer1",
			Status:    "open",
			Mergeable: true,
		},
		{
			ID:        2,
			Title:     "fix: resolve memory leak in worker process",
			Author:    "developer2",
			Status:    "open",
			Mergeable: true,
		},
		{
			ID:        3,
			Title:     "docs: update API documentation",
			Author:    "tech-writer",
			Status:    "open",
			Mergeable: false, // Simulate conflicts
		},
	}
}

// processMockPR processes a mock pull request for testing
func processMockPR(pr MockPR) error {
	// Simulate processing logic
	if pr.Title == "" {
		return fmt.Errorf("PR title cannot be empty")
	}

	if pr.Author == "" {
		return fmt.Errorf("PR author cannot be empty")
	}

	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	return nil
}

// MockPR represents a mock pull request for testing
type MockPR struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Status    string `json:"status"`
	Mergeable bool   `json:"mergeable"`
}

// resolveEnvVar resolves environment variable references in config values
func resolveEnvVar(value string) string {
	if value == "" {
		return ""
	}

	// If value starts with $, treat it as an environment variable
	if len(value) > 1 && value[0] == '$' {
		envVar := value[1:]
		return utils.GetEnv(envVar, "")
	}

	// If value is just a lone $, return empty string
	if value == "$" {
		return ""
	}

	return value
}
