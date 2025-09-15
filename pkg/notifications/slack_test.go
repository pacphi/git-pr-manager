package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/merge"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/providers/common"
)

func TestNewSlackNotifier(t *testing.T) {
	config := SlackConfig{
		WebhookURL: "https://hooks.slack.com/test",
		Channel:    "#deployments",
		Username:   "TestBot",
	}

	notifier := NewSlackNotifier(config)

	assert.Equal(t, config.WebhookURL, notifier.webhookURL)
	assert.Equal(t, config.Channel, notifier.channel)
	assert.Equal(t, config.Username, notifier.username)
	assert.NotNil(t, notifier.logger)
}

func TestSlackNotifier_SendTestMessage(t *testing.T) {
	tests := []struct {
		name            string
		webhookURL      string
		serverResponse  func(w http.ResponseWriter, r *http.Request)
		expectError     bool
		validateRequest func(t *testing.T, r *http.Request)
	}{
		{
			name:        "empty webhook URL",
			webhookURL:  "",
			expectError: true,
		},
		{
			name:       "successful test message",
			webhookURL: "placeholder", // Will be replaced with server URL
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
			expectError: false,
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var message SlackMessage
				err = json.Unmarshal(body, &message)
				require.NoError(t, err)

				assert.Contains(t, message.Text, "Test message")
				assert.Equal(t, len(message.Attachments), 1)
				assert.Equal(t, message.Attachments[0].Color, "good")
				assert.Contains(t, message.Attachments[0].Title, "Configuration Test")
			},
		},
		{
			name:       "server error",
			webhookURL: "placeholder",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectError: true,
		},
		{
			name:       "slack returns error status",
			webhookURL: "placeholder",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalid_payload"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tt.validateRequest != nil {
						tt.validateRequest(t, r)
					}
					tt.serverResponse(w, r)
				}))
				defer server.Close()
				tt.webhookURL = server.URL
			}

			config := SlackConfig{
				WebhookURL: tt.webhookURL,
				Channel:    "#test",
				Username:   "TestBot",
			}
			notifier := NewSlackNotifier(config)

			err := notifier.SendTestMessage(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSlackNotifier_SendMergeResults(t *testing.T) {
	tests := []struct {
		name            string
		results         []merge.MergeResult
		serverResponse  func(w http.ResponseWriter, r *http.Request)
		expectError     bool
		validateRequest func(t *testing.T, r *http.Request, results []merge.MergeResult)
	}{
		{
			name: "successful merge results",
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo1",
					PullRequest: 123,
					Title:       "Add new feature",
					Author:      "developer1",
					Success:     true,
					MergeMethod: "squash",
					MergedAt:    time.Now(),
				},
				{
					Provider:    "github",
					Repository:  "owner/repo2",
					PullRequest: 456,
					Title:       "Bug fix",
					Author:      "developer2",
					Success:     true,
					MergeMethod: "merge",
					MergedAt:    time.Now(),
				},
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
			expectError: false,
			validateRequest: func(t *testing.T, r *http.Request, results []merge.MergeResult) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var message SlackMessage
				err = json.Unmarshal(body, &message)
				require.NoError(t, err)

				assert.Contains(t, message.Text, "Pull Request Merge Summary")
				assert.Contains(t, message.Text, "2 merged")
				assert.Len(t, message.Attachments, 2) // Summary + success details

				// Check summary attachment
				summaryAttachment := message.Attachments[0]
				assert.Equal(t, "good", summaryAttachment.Color)
				assert.Equal(t, "Merge Results Summary", summaryAttachment.Title)
				assert.Len(t, summaryAttachment.Fields, 4)
			},
		},
		{
			name: "mixed results with failures",
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo1",
					PullRequest: 123,
					Title:       "Success PR",
					Author:      "developer1",
					Success:     true,
					MergeMethod: "squash",
				},
				{
					Provider:    "github",
					Repository:  "owner/repo2",
					PullRequest: 456,
					Title:       "Failed PR",
					Author:      "developer2",
					Success:     false,
					Error:       fmt.Errorf("merge conflict"),
				},
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
			expectError: false,
			validateRequest: func(t *testing.T, r *http.Request, results []merge.MergeResult) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var message SlackMessage
				err = json.Unmarshal(body, &message)
				require.NoError(t, err)

				assert.Contains(t, message.Text, "1 merged")
				assert.Contains(t, message.Text, "1 failed")
				assert.Len(t, message.Attachments, 3) // Summary + success + failed

				// Check that the color reflects failures
				summaryAttachment := message.Attachments[0]
				assert.Equal(t, "danger", summaryAttachment.Color)

				// Check failed attachment exists
				failedAttachment := message.Attachments[1]
				assert.Equal(t, "danger", failedAttachment.Color)
				assert.Contains(t, failedAttachment.Title, "Failed Merges")
				assert.Contains(t, failedAttachment.Text, "merge conflict")
			},
		},
		{
			name: "skipped results",
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo1",
					PullRequest: 123,
					Title:       "Skipped PR",
					Author:      "developer1",
					Success:     true,
					Skipped:     true,
					Reason:      "dry run mode",
				},
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
			expectError: false,
			validateRequest: func(t *testing.T, r *http.Request, results []merge.MergeResult) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var message SlackMessage
				err = json.Unmarshal(body, &message)
				require.NoError(t, err)

				assert.Contains(t, message.Text, "1 skipped")

				// Color should be warning for skipped
				summaryAttachment := message.Attachments[0]
				assert.Equal(t, "warning", summaryAttachment.Color)
			},
		},
		{
			name: "many successful results (should limit display)",
			results: func() []merge.MergeResult {
				results := make([]merge.MergeResult, 15)
				for i := 0; i < 15; i++ {
					results[i] = merge.MergeResult{
						Provider:    "github",
						Repository:  fmt.Sprintf("owner/repo%d", i),
						PullRequest: i + 1,
						Title:       fmt.Sprintf("PR %d", i+1),
						Author:      "developer",
						Success:     true,
						MergeMethod: "squash",
					}
				}
				return results
			}(),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
			expectError: false,
			validateRequest: func(t *testing.T, r *http.Request, results []merge.MergeResult) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var message SlackMessage
				err = json.Unmarshal(body, &message)
				require.NoError(t, err)

				// Should limit successful results display
				successAttachment := message.Attachments[1]
				assert.Contains(t, successAttachment.Text, "... and 5 more")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateRequest != nil {
					tt.validateRequest(t, r, tt.results)
				}
				tt.serverResponse(w, r)
			}))
			defer server.Close()

			config := SlackConfig{
				WebhookURL: server.URL,
				Channel:    "#test",
				Username:   "TestBot",
			}
			notifier := NewSlackNotifier(config)

			err := notifier.SendMergeResults(context.Background(), tt.results)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSlackNotifier_SendMergeResults_EmptyWebhookURL(t *testing.T) {
	notifier := NewSlackNotifier(SlackConfig{})

	results := []merge.MergeResult{
		{Provider: "github", Repository: "owner/repo", PullRequest: 1},
	}

	err := notifier.SendMergeResults(context.Background(), results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook URL not configured")
}

func TestSlackNotifier_SendPRSummary(t *testing.T) {
	repositories := []common.Repository{
		{FullName: "owner/repo1"},
		{FullName: "owner/repo2"},
		{FullName: "owner/repo3"},
	}
	totalPRs := 15
	readyPRs := 5

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var message SlackMessage
		err = json.Unmarshal(body, &message)
		require.NoError(t, err)

		assert.Contains(t, message.Text, "Pull Request Status Summary")
		assert.Len(t, message.Attachments, 1)

		attachment := message.Attachments[0]
		assert.Equal(t, "good", attachment.Color)
		assert.Equal(t, "PR Status Check Results", attachment.Title)
		assert.Len(t, attachment.Fields, 4)

		// Verify field values
		fields := attachment.Fields
		assert.Equal(t, "Repositories Scanned", fields[0].Title)
		assert.Equal(t, "3", fields[0].Value)
		assert.Equal(t, "Total PRs Found", fields[1].Title)
		assert.Equal(t, "15", fields[1].Value)
		assert.Equal(t, "PRs Ready to Merge", fields[2].Title)
		assert.Equal(t, "5", fields[2].Value)
		assert.Equal(t, "Ready Percentage", fields[3].Title)
		assert.Equal(t, "33.3%", fields[3].Value)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#test",
		Username:   "TestBot",
	}
	notifier := NewSlackNotifier(config)

	err := notifier.SendPRSummary(context.Background(), repositories, totalPRs, readyPRs)
	assert.NoError(t, err)
}

func TestSlackNotifier_SendPRSummary_EmptyWebhookURL(t *testing.T) {
	notifier := NewSlackNotifier(SlackConfig{})

	repositories := []common.Repository{
		{FullName: "owner/repo"},
	}

	err := notifier.SendPRSummary(context.Background(), repositories, 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook URL not configured")
}

func TestSlackNotifier_GetUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected string
	}{
		{
			name:     "custom username set",
			username: "CustomBot",
			expected: "CustomBot",
		},
		{
			name:     "empty username uses default",
			username: "",
			expected: "Git PR Automation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SlackConfig{
				WebhookURL: "https://hooks.slack.com/test",
				Username:   tt.username,
			}
			notifier := NewSlackNotifier(config)

			result := notifier.getUsername()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSlackNotifier_SendMessage_WithTimeout(t *testing.T) {
	// Create a server that takes longer than the timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#test",
		Username:   "TestBot",
	}
	notifier := NewSlackNotifier(config)

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	message := SlackMessage{
		Text:     "Test message",
		Username: "TestBot",
		Channel:  "#test",
	}

	err := notifier.sendMessage(ctx, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestSlackNotifier_BuildMergeResultsMessage(t *testing.T) {
	notifier := NewSlackNotifier(SlackConfig{
		Username: "TestBot",
		Channel:  "#test",
	})

	results := []merge.MergeResult{
		{
			Provider:    "github",
			Repository:  "owner/repo1",
			PullRequest: 1,
			Success:     true,
			Title:       "Feature PR",
			Author:      "dev1",
		},
		{
			Provider:    "github",
			Repository:  "owner/repo2",
			PullRequest: 2,
			Success:     false,
			Title:       "Bug fix PR",
			Author:      "dev2",
			Error:       fmt.Errorf("merge conflict"),
		},
		{
			Provider:    "github",
			Repository:  "owner/repo3",
			PullRequest: 3,
			Success:     true,
			Skipped:     true,
			Title:       "Docs PR",
			Author:      "dev3",
			Reason:      "dry run",
		},
	}

	message := notifier.buildMergeResultsMessage(results)

	assert.Contains(t, message.Text, "Pull Request Merge Summary")
	assert.Contains(t, message.Text, "1 merged")
	assert.Contains(t, message.Text, "1 failed")
	assert.Contains(t, message.Text, "1 skipped")
	assert.Equal(t, "TestBot", message.Username)
	assert.Equal(t, "#test", message.Channel)

	// Should have 3 attachments: summary, failed details, success details
	assert.Len(t, message.Attachments, 3)

	// Summary attachment should be danger color due to failures
	assert.Equal(t, "danger", message.Attachments[0].Color)

	// Check field values in summary
	fields := message.Attachments[0].Fields
	assert.Equal(t, "1", fields[0].Value) // Successful
	assert.Equal(t, "1", fields[1].Value) // Failed
	assert.Equal(t, "1", fields[2].Value) // Skipped
	assert.Equal(t, "3", fields[3].Value) // Total
}

func TestSlackNotifier_HTTPClientTimeout(t *testing.T) {
	// Test that the HTTP client has proper timeout configuration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response that should timeout (longer than 10 second client timeout)
		time.Sleep(15 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#test",
	}
	notifier := NewSlackNotifier(config)

	message := SlackMessage{Text: "Test"}

	start := time.Now()
	err := notifier.sendMessage(context.Background(), message)
	duration := time.Since(start)

	// Should timeout due to HTTP client timeout (10 seconds in implementation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message to Slack")
	// Should timeout around 10 seconds, not wait for the full 15 seconds
	assert.Less(t, duration, 12*time.Second, "Request should timeout before server response")
}

func TestSlackMessage_JSONMarshaling(t *testing.T) {
	message := SlackMessage{
		Text:     "Test message",
		Username: "TestBot",
		Channel:  "#test",
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Test Attachment",
				Fields: []SlackField{
					{Title: "Field1", Value: "Value1", Short: true},
					{Title: "Field2", Value: "Value2", Short: false},
				},
			},
		},
	}

	jsonData, err := json.Marshal(message)
	require.NoError(t, err)

	var unmarshaled SlackMessage
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, message.Text, unmarshaled.Text)
	assert.Equal(t, message.Username, unmarshaled.Username)
	assert.Equal(t, message.Channel, unmarshaled.Channel)
	assert.Len(t, unmarshaled.Attachments, 1)
	assert.Equal(t, message.Attachments[0].Title, unmarshaled.Attachments[0].Title)
	assert.Len(t, unmarshaled.Attachments[0].Fields, 2)
}

func TestSlackNotifier_EdgeCases(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var message SlackMessage
			err = json.Unmarshal(body, &message)
			require.NoError(t, err)

			assert.Contains(t, message.Text, "0 PRs")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := NewSlackNotifier(SlackConfig{WebhookURL: server.URL})
		err := notifier.SendMergeResults(context.Background(), []merge.MergeResult{})
		assert.NoError(t, err)
	})

	t.Run("nil results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := NewSlackNotifier(SlackConfig{WebhookURL: server.URL})
		err := notifier.SendMergeResults(context.Background(), nil)
		assert.NoError(t, err)
	})

	t.Run("very long repository names", func(t *testing.T) {
		longName := strings.Repeat("a", 1000)
		results := []merge.MergeResult{
			{
				Repository:  longName,
				PullRequest: 1,
				Success:     true,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), longName)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := NewSlackNotifier(SlackConfig{WebhookURL: server.URL})
		err := notifier.SendMergeResults(context.Background(), results)
		assert.NoError(t, err)
	})
}
