package notifications

import (
	"context"
	"fmt"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/merge"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// Notifier interface for sending notifications
type Notifier interface {
	SendMergeResults(ctx context.Context, results []merge.MergeResult) error
	SendTestMessage(ctx context.Context) error
	SendPRSummary(ctx context.Context, repositories []common.Repository, totalPRs, readyPRs int) error
}

// Manager manages multiple notification channels
type Manager struct {
	notifiers []Notifier
	logger    *utils.Logger
}

// NewManager creates a new notification manager
func NewManager(cfg *config.Config) (*Manager, error) {
	var notifiers []Notifier

	// Add Slack notifier if configured
	if cfg.Notifications.Slack.WebhookURL != "" {
		slackConfig := SlackConfig{
			WebhookURL: resolveEnvVar(cfg.Notifications.Slack.WebhookURL),
			Channel:    cfg.Notifications.Slack.Channel,
			Username:   "Git PR Bot", // Default username
		}
		notifiers = append(notifiers, NewSlackNotifier(slackConfig))
	}

	// Add email notifier if configured
	if cfg.Notifications.Email.SMTPHost != "" && len(cfg.Notifications.Email.To) > 0 {
		emailConfig := EmailConfig{
			SMTPHost: cfg.Notifications.Email.SMTPHost,
			SMTPPort: cfg.Notifications.Email.SMTPPort,
			Username: resolveEnvVar(cfg.Notifications.Email.SMTPUsername),
			Password: resolveEnvVar(cfg.Notifications.Email.SMTPPassword),
			From:     cfg.Notifications.Email.From,
			To:       cfg.Notifications.Email.To,
			UseTLS:   true, // Default to TLS
		}
		notifiers = append(notifiers, NewEmailNotifier(emailConfig))
	}

	return &Manager{
		notifiers: notifiers,
		logger:    utils.GetGlobalLogger().WithComponent("notifications"),
	}, nil
}

// SendMergeResults sends merge results to all configured notifiers
func (m *Manager) SendMergeResults(ctx context.Context, results []merge.MergeResult) error {
	if len(m.notifiers) == 0 {
		m.logger.Debug("No notifiers configured, skipping notifications")
		return nil
	}

	var errors []error

	for i, notifier := range m.notifiers {
		if err := notifier.SendMergeResults(ctx, results); err != nil {
			m.logger.WithError(err).Errorf("Failed to send merge results via notifier %d", i)
			errors = append(errors, fmt.Errorf("notifier %d: %w", i, err))
		} else {
			m.logger.Debugf("Successfully sent merge results via notifier %d", i)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send notifications: %v", errors)
	}

	return nil
}

// SendPRSummary sends PR summary to all configured notifiers
func (m *Manager) SendPRSummary(ctx context.Context, repositories []common.Repository, totalPRs, readyPRs int) error {
	if len(m.notifiers) == 0 {
		m.logger.Debug("No notifiers configured, skipping PR summary")
		return nil
	}

	var errors []error

	for i, notifier := range m.notifiers {
		if err := notifier.SendPRSummary(ctx, repositories, totalPRs, readyPRs); err != nil {
			m.logger.WithError(err).Errorf("Failed to send PR summary via notifier %d", i)
			errors = append(errors, fmt.Errorf("notifier %d: %w", i, err))
		} else {
			m.logger.Debugf("Successfully sent PR summary via notifier %d", i)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send PR summary: %v", errors)
	}

	return nil
}

// GetNotifierCount returns the number of configured notifiers
func (m *Manager) GetNotifierCount() int {
	return len(m.notifiers)
}

// HasNotifiers returns true if any notifiers are configured
func (m *Manager) HasNotifiers() bool {
	return len(m.notifiers) > 0
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

	return value
}
