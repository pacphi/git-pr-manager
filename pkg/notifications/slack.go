package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/merge"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// SlackNotifier sends notifications to Slack
type SlackNotifier struct {
	webhookURL string
	channel    string
	username   string
	logger     *utils.Logger
}

// SlackConfig contains Slack notification configuration
type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
}

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	TitleLink string       `json:"title_link,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Footer    string       `json:"footer,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(config SlackConfig) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: config.WebhookURL,
		channel:    config.Channel,
		username:   config.Username,
		logger:     utils.GetGlobalLogger().WithComponent("slack"),
	}
}

// SendMergeResults sends merge results to Slack
func (s *SlackNotifier) SendMergeResults(ctx context.Context, results []merge.MergeResult) error {
	if s.webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	message := s.buildMergeResultsMessage(results)
	return s.sendMessage(ctx, message)
}

// SendTestMessage sends a test message to verify Slack configuration
func (s *SlackNotifier) SendTestMessage(ctx context.Context) error {
	if s.webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	message := SlackMessage{
		Text:     "🧪 Test message from Git PR Automation",
		Username: s.getUsername(),
		Channel:  s.channel,
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Configuration Test",
				Text:  "If you can see this message, your Slack integration is working correctly!",
				Fields: []SlackField{
					{
						Title: "Status",
						Value: "✅ Success",
						Short: true,
					},
					{
						Title: "Timestamp",
						Value: time.Now().Format("2006-01-02 15:04:05 MST"),
						Short: true,
					},
				},
				Footer:    "Git PR Automation",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return s.sendMessage(ctx, message)
}

// buildMergeResultsMessage builds a Slack message for merge results
func (s *SlackNotifier) buildMergeResultsMessage(results []merge.MergeResult) SlackMessage {
	successful := 0
	failed := 0
	skipped := 0

	var attachments []SlackAttachment

	// Count results
	for _, result := range results {
		if result.Error != nil {
			failed++
		} else if result.Skipped {
			skipped++
		} else {
			successful++
		}
	}

	// Summary text
	var summaryText strings.Builder
	summaryText.WriteString("🤖 *Pull Request Merge Summary*\n")
	summaryText.WriteString(fmt.Sprintf("📊 Processed %d PRs: ", len(results)))
	if successful > 0 {
		summaryText.WriteString(fmt.Sprintf("✅ %d merged", successful))
	}
	if failed > 0 {
		if successful > 0 {
			summaryText.WriteString(", ")
		}
		summaryText.WriteString(fmt.Sprintf("❌ %d failed", failed))
	}
	if skipped > 0 {
		if successful > 0 || failed > 0 {
			summaryText.WriteString(", ")
		}
		summaryText.WriteString(fmt.Sprintf("⏭️ %d skipped", skipped))
	}

	// Create summary attachment
	color := "good"
	if failed > 0 {
		color = "danger"
	} else if skipped > 0 {
		color = "warning"
	}

	summaryAttachment := SlackAttachment{
		Color: color,
		Title: "Merge Results Summary",
		Fields: []SlackField{
			{
				Title: "Successful",
				Value: fmt.Sprintf("%d", successful),
				Short: true,
			},
			{
				Title: "Failed",
				Value: fmt.Sprintf("%d", failed),
				Short: true,
			},
			{
				Title: "Skipped",
				Value: fmt.Sprintf("%d", skipped),
				Short: true,
			},
			{
				Title: "Total",
				Value: fmt.Sprintf("%d", len(results)),
				Short: true,
			},
		},
		Footer:    "Git PR Automation",
		Timestamp: time.Now().Unix(),
	}

	attachments = append(attachments, summaryAttachment)

	// Add details for failed merges
	if failed > 0 {
		var failedDetails strings.Builder
		for _, result := range results {
			if result.Error != nil {
				failedDetails.WriteString(fmt.Sprintf("• *%s* #%d: %s\n",
					result.Repository,
					result.PullRequest,
					result.Error.Error()))
			}
		}

		if failedDetails.Len() > 0 {
			failedAttachment := SlackAttachment{
				Color: "danger",
				Title: "❌ Failed Merges",
				Text:  failedDetails.String(),
			}
			attachments = append(attachments, failedAttachment)
		}
	}

	// Add details for successful merges (limit to 10 to avoid message size limits)
	if successful > 0 {
		var successDetails strings.Builder
		count := 0
		for _, result := range results {
			if result.Error == nil && !result.Skipped {
				if count >= 10 {
					successDetails.WriteString(fmt.Sprintf("... and %d more\n", successful-10))
					break
				}
				successDetails.WriteString(fmt.Sprintf("• *%s* #%d: %s\n",
					result.Repository,
					result.PullRequest,
					result.Title))
				count++
			}
		}

		if successDetails.Len() > 0 {
			successAttachment := SlackAttachment{
				Color: "good",
				Title: "✅ Successful Merges",
				Text:  successDetails.String(),
			}
			attachments = append(attachments, successAttachment)
		}
	}

	return SlackMessage{
		Text:        summaryText.String(),
		Username:    s.getUsername(),
		Channel:     s.channel,
		Attachments: attachments,
	}
}

// sendMessage sends a message to Slack
func (s *SlackNotifier) sendMessage(ctx context.Context, message SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message to Slack: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.WithError(closeErr).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	s.logger.Debug("Slack notification sent successfully")
	return nil
}

// getUsername returns the username to use for Slack messages
func (s *SlackNotifier) getUsername() string {
	if s.username != "" {
		return s.username
	}
	return "Git PR Automation"
}

// SendPRSummary sends a summary of PR check results to Slack
func (s *SlackNotifier) SendPRSummary(ctx context.Context, repositories []common.Repository, totalPRs, readyPRs int) error {
	if s.webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	message := SlackMessage{
		Text:     "📋 *Pull Request Status Summary*",
		Username: s.getUsername(),
		Channel:  s.channel,
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "PR Status Check Results",
				Fields: []SlackField{
					{
						Title: "Repositories Scanned",
						Value: fmt.Sprintf("%d", len(repositories)),
						Short: true,
					},
					{
						Title: "Total PRs Found",
						Value: fmt.Sprintf("%d", totalPRs),
						Short: true,
					},
					{
						Title: "PRs Ready to Merge",
						Value: fmt.Sprintf("%d", readyPRs),
						Short: true,
					},
					{
						Title: "Ready Percentage",
						Value: fmt.Sprintf("%.1f%%", float64(readyPRs)/float64(totalPRs)*100),
						Short: true,
					},
				},
				Footer:    "Git PR Automation",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return s.sendMessage(ctx, message)
}
