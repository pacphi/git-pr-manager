package validation

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// Validator provides configuration validation functionality
type Validator struct {
	logger *utils.Logger
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		logger: utils.GetGlobalLogger().WithComponent("validator"),
	}
}

// ValidateConfig validates the entire configuration structure and business rules
func (v *Validator) ValidateConfig(cfg *config.Config) error {
	var errors []string

	// Validate PR filters
	if err := v.validatePRFilters(cfg.PRFilters); err != nil {
		errors = append(errors, fmt.Sprintf("PR filters: %v", err))
	}

	// Validate repositories
	if err := v.validateRepositories(cfg.Repositories); err != nil {
		errors = append(errors, fmt.Sprintf("repositories: %v", err))
	}

	// Validate authentication
	if err := v.validateAuth(cfg.Auth, cfg.Repositories); err != nil {
		errors = append(errors, fmt.Sprintf("authentication: %v", err))
	}

	// Validate behavior settings
	if err := v.validateBehavior(cfg.Behavior); err != nil {
		errors = append(errors, fmt.Sprintf("behavior: %v", err))
	}

	// Validate notifications (optional)
	if err := v.validateNotifications(cfg.Notifications); err != nil {
		errors = append(errors, fmt.Sprintf("notifications: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validatePRFilters validates PR filter configuration
func (v *Validator) validatePRFilters(filters config.PRFilters) error {
	var errors []string

	// Check allowed actors
	if len(filters.AllowedActors) == 0 {
		errors = append(errors, "at least one allowed actor must be specified")
	}

	for i, actor := range filters.AllowedActors {
		if strings.TrimSpace(actor) == "" {
			errors = append(errors, fmt.Sprintf("allowed actor at index %d is empty", i))
		}
	}

	// Validate max age if specified
	if filters.MaxAge != "" {
		if _, err := utils.ParseDuration(filters.MaxAge); err != nil {
			errors = append(errors, fmt.Sprintf("invalid max_age format '%s': %v", filters.MaxAge, err))
		}
	}

	// Validate skip labels
	for i, label := range filters.SkipLabels {
		if strings.TrimSpace(label) == "" {
			errors = append(errors, fmt.Sprintf("skip label at index %d is empty", i))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// validateRepositories validates repository configuration
func (v *Validator) validateRepositories(repositories map[string][]config.Repository) error {
	var errors []string

	if len(repositories) == 0 {
		errors = append(errors, "at least one provider must be configured")
	}

	repoNameRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

	for provider, repos := range repositories {
		if len(repos) == 0 {
			continue
		}

		for i, repo := range repos {
			// Validate repository name format
			if !repoNameRegex.MatchString(repo.Name) {
				errors = append(errors, fmt.Sprintf("provider '%s', repository %d: invalid name format '%s' (expected 'owner/name')", provider, i, repo.Name))
			}

			// Validate merge strategy
			if string(repo.MergeStrategy) != "" {
				validStrategies := []string{"merge", "squash", "rebase"}
				valid := false
				for _, strategy := range validStrategies {
					if string(repo.MergeStrategy) == strategy {
						valid = true
						break
					}
				}
				if !valid {
					errors = append(errors, fmt.Sprintf("provider '%s', repository '%s': invalid merge strategy '%s' (must be one of: %s)",
						provider, repo.Name, string(repo.MergeStrategy), strings.Join(validStrategies, ", ")))
				}
			}

			// Validate skip labels
			for j, label := range repo.SkipLabels {
				if strings.TrimSpace(label) == "" {
					errors = append(errors, fmt.Sprintf("provider '%s', repository '%s': skip label at index %d is empty", provider, repo.Name, j))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// validateAuth validates authentication configuration
func (v *Validator) validateAuth(auth config.Auth, repositories map[string][]config.Repository) error {
	var errors []string
	hasProvider := false

	// Helper function to check if a provider has repositories
	hasRepos := func(provider string) bool {
		repos, exists := repositories[provider]
		return exists && len(repos) > 0
	}

	// Check GitHub auth - only validate if GitHub repositories are configured
	if hasRepos("github") {
		if auth.GitHub.Token != "" {
			hasProvider = true
			if !v.isValidEnvVarReference(auth.GitHub.Token) && !v.isValidToken(auth.GitHub.Token, "GitHub") {
				errors = append(errors, "GitHub token appears invalid")
			}
		} else {
			errors = append(errors, "GitHub token is required when GitHub repositories are configured")
		}
	}

	// Check GitLab auth - only validate if GitLab repositories are configured
	if hasRepos("gitlab") {
		if auth.GitLab.Token != "" {
			hasProvider = true
			if !v.isValidEnvVarReference(auth.GitLab.Token) && !v.isValidToken(auth.GitLab.Token, "GitLab") {
				errors = append(errors, "GitLab token appears invalid")
			}

			// Validate URL if provided
			if auth.GitLab.URL != "" && !v.isValidEnvVarReference(auth.GitLab.URL) {
				if !strings.HasPrefix(auth.GitLab.URL, "http://") && !strings.HasPrefix(auth.GitLab.URL, "https://") {
					errors = append(errors, "GitLab URL must start with http:// or https://")
				}
			}
		} else {
			errors = append(errors, "GitLab token is required when GitLab repositories are configured")
		}
	}

	// Check Bitbucket auth - only validate if Bitbucket repositories are configured
	if hasRepos("bitbucket") {
		if auth.Bitbucket.Username != "" || auth.Bitbucket.AppPassword != "" {
			hasProvider = true
			if auth.Bitbucket.Username == "" {
				errors = append(errors, "Bitbucket username is required when app password is provided")
			}
			if auth.Bitbucket.AppPassword == "" {
				errors = append(errors, "Bitbucket app password is required when username is provided")
			}
		} else {
			errors = append(errors, "Bitbucket username and app password are required when Bitbucket repositories are configured")
		}
	}

	// Check if we have any repositories configured at all
	hasAnyRepos := false
	for _, repos := range repositories {
		if len(repos) > 0 {
			hasAnyRepos = true
			break
		}
	}

	// Only require provider authentication if repositories are configured
	if hasAnyRepos && !hasProvider {
		errors = append(errors, "at least one provider authentication must be configured when repositories are present")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// validateBehavior validates behavior configuration
func (v *Validator) validateBehavior(behavior config.Behavior) error {
	var errors []string

	if behavior.Concurrency <= 0 {
		errors = append(errors, "concurrency must be greater than 0")
	}

	if behavior.Concurrency > 50 {
		errors = append(errors, "concurrency should not exceed 50 for rate limiting")
	}

	if behavior.RateLimit.RequestsPerSecond < 0 {
		errors = append(errors, "rate limit requests per second cannot be negative")
	}

	if behavior.RateLimit.Burst < 0 {
		errors = append(errors, "rate limit burst cannot be negative")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// validateNotifications validates notification configuration
func (v *Validator) validateNotifications(notifications config.Notifications) error {
	var errors []string

	// Validate Slack configuration
	if notifications.Slack.WebhookURL != "" {
		if !v.isValidEnvVarReference(notifications.Slack.WebhookURL) {
			if !strings.HasPrefix(notifications.Slack.WebhookURL, "https://hooks.slack.com/") {
				errors = append(errors, "Slack webhook URL should start with https://hooks.slack.com/")
			}
		}
	}

	// Validate email configuration
	if notifications.Email.SMTPHost != "" {
		if notifications.Email.SMTPPort <= 0 || notifications.Email.SMTPPort > 65535 {
			errors = append(errors, "SMTP port must be between 1 and 65535")
		}

		if len(notifications.Email.To) == 0 {
			errors = append(errors, "at least one email recipient must be specified when SMTP host is configured")
		}

		// Basic email format validation
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

		if notifications.Email.From != "" && !emailRegex.MatchString(notifications.Email.From) {
			errors = append(errors, "invalid 'from' email address format")
		}

		for i, email := range notifications.Email.To {
			if !emailRegex.MatchString(email) {
				errors = append(errors, fmt.Sprintf("invalid 'to' email address at index %d: %s", i, email))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// CheckEnvironmentVariables checks if required environment variables are set
func (v *Validator) CheckEnvironmentVariables(cfg *config.Config) []string {
	var missing []string

	// Check authentication environment variables
	if v.isValidEnvVarReference(cfg.Auth.GitHub.Token) {
		envVar := v.extractEnvVarName(cfg.Auth.GitHub.Token)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Auth.GitLab.Token) {
		envVar := v.extractEnvVarName(cfg.Auth.GitLab.Token)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Auth.GitLab.URL) {
		envVar := v.extractEnvVarName(cfg.Auth.GitLab.URL)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Auth.Bitbucket.Username) {
		envVar := v.extractEnvVarName(cfg.Auth.Bitbucket.Username)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Auth.Bitbucket.AppPassword) {
		envVar := v.extractEnvVarName(cfg.Auth.Bitbucket.AppPassword)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Auth.Bitbucket.Workspace) {
		envVar := v.extractEnvVarName(cfg.Auth.Bitbucket.Workspace)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	// Check notification environment variables
	if v.isValidEnvVarReference(cfg.Notifications.Slack.WebhookURL) {
		envVar := v.extractEnvVarName(cfg.Notifications.Slack.WebhookURL)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Notifications.Email.SMTPUsername) {
		envVar := v.extractEnvVarName(cfg.Notifications.Email.SMTPUsername)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if v.isValidEnvVarReference(cfg.Notifications.Email.SMTPPassword) {
		envVar := v.extractEnvVarName(cfg.Notifications.Email.SMTPPassword)
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	return missing
}

// Helper functions

// isValidEnvVarReference checks if a value is an environment variable reference
func (v *Validator) isValidEnvVarReference(value string) bool {
	return len(value) > 1 && value[0] == '$'
}

// extractEnvVarName extracts the environment variable name from a reference
func (v *Validator) extractEnvVarName(value string) string {
	if v.isValidEnvVarReference(value) {
		return value[1:]
	}
	return value
}

// isValidToken performs basic token format validation
func (v *Validator) isValidToken(token, provider string) bool {
	if len(token) < 10 {
		return false
	}

	switch provider {
	case "GitHub":
		return strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "github_pat_") || len(token) >= 40
	case "GitLab":
		return strings.HasPrefix(token, "glpat-") || len(token) >= 20
	default:
		return len(token) >= 10
	}
}
