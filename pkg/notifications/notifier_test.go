package notifications

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/merge"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
)

// MockNotifier is a mock implementation of the Notifier interface
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendMergeResults(ctx context.Context, results []merge.MergeResult) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *MockNotifier) SendTestMessage(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockNotifier) SendPRSummary(ctx context.Context, repositories []common.Repository, totalPRs, readyPRs int) error {
	args := m.Called(ctx, repositories, totalPRs, readyPRs)
	return args.Error(0)
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedCount int
		checkSlack    bool
		checkEmail    bool
	}{
		{
			name: "empty config - no notifiers",
			config: &config.Config{
				Notifications: config.Notifications{},
			},
			expectedCount: 0,
			checkSlack:    false,
			checkEmail:    false,
		},
		{
			name: "slack only",
			config: &config.Config{
				Notifications: config.Notifications{
					Slack: config.SlackConfig{
						WebhookURL: "https://hooks.slack.com/test",
						Channel:    "#deployments",
					},
				},
			},
			expectedCount: 1,
			checkSlack:    true,
			checkEmail:    false,
		},
		{
			name: "email only",
			config: &config.Config{
				Notifications: config.Notifications{
					Email: config.EmailConfig{
						SMTPHost: "smtp.example.com",
						SMTPPort: 587,
						To:       []string{"admin@example.com"},
						From:     "bot@example.com",
					},
				},
			},
			expectedCount: 1,
			checkSlack:    false,
			checkEmail:    true,
		},
		{
			name: "both slack and email",
			config: &config.Config{
				Notifications: config.Notifications{
					Slack: config.SlackConfig{
						WebhookURL: "https://hooks.slack.com/test",
						Channel:    "#deployments",
					},
					Email: config.EmailConfig{
						SMTPHost: "smtp.example.com",
						SMTPPort: 587,
						To:       []string{"admin@example.com"},
						From:     "bot@example.com",
					},
				},
			},
			expectedCount: 2,
			checkSlack:    true,
			checkEmail:    true,
		},
		{
			name: "incomplete email config - only slack",
			config: &config.Config{
				Notifications: config.Notifications{
					Slack: config.SlackConfig{
						WebhookURL: "https://hooks.slack.com/test",
						Channel:    "#deployments",
					},
					Email: config.EmailConfig{
						SMTPHost: "smtp.example.com",
						// Missing To field
					},
				},
			},
			expectedCount: 1,
			checkSlack:    true,
			checkEmail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, manager.GetNotifierCount())
			assert.Equal(t, tt.expectedCount > 0, manager.HasNotifiers())
		})
	}
}

func TestManager_SendMergeResults(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func() []Notifier
		results        []merge.MergeResult
		expectError    bool
		errorSubstring string
	}{
		{
			name: "no notifiers configured",
			setupMocks: func() []Notifier {
				return []Notifier{}
			},
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo",
					PullRequest: 1,
					Success:     true,
				},
			},
			expectError: false,
		},
		{
			name: "single notifier success",
			setupMocks: func() []Notifier {
				mock1 := &MockNotifier{}
				mock1.On("SendMergeResults", mock.Anything, mock.Anything).Return(nil)
				return []Notifier{mock1}
			},
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo",
					PullRequest: 1,
					Success:     true,
				},
			},
			expectError: false,
		},
		{
			name: "multiple notifiers success",
			setupMocks: func() []Notifier {
				mock1 := &MockNotifier{}
				mock1.On("SendMergeResults", mock.Anything, mock.Anything).Return(nil)
				mock2 := &MockNotifier{}
				mock2.On("SendMergeResults", mock.Anything, mock.Anything).Return(nil)
				return []Notifier{mock1, mock2}
			},
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo",
					PullRequest: 1,
					Success:     true,
				},
			},
			expectError: false,
		},
		{
			name: "one notifier fails",
			setupMocks: func() []Notifier {
				mock1 := &MockNotifier{}
				mock1.On("SendMergeResults", mock.Anything, mock.Anything).Return(errors.New("slack webhook failed"))
				mock2 := &MockNotifier{}
				mock2.On("SendMergeResults", mock.Anything, mock.Anything).Return(nil)
				return []Notifier{mock1, mock2}
			},
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo",
					PullRequest: 1,
					Success:     true,
				},
			},
			expectError:    true,
			errorSubstring: "failed to send notifications",
		},
		{
			name: "all notifiers fail",
			setupMocks: func() []Notifier {
				mock1 := &MockNotifier{}
				mock1.On("SendMergeResults", mock.Anything, mock.Anything).Return(errors.New("slack failed"))
				mock2 := &MockNotifier{}
				mock2.On("SendMergeResults", mock.Anything, mock.Anything).Return(errors.New("email failed"))
				return []Notifier{mock1, mock2}
			},
			results: []merge.MergeResult{
				{
					Provider:    "github",
					Repository:  "owner/repo",
					PullRequest: 1,
					Success:     true,
				},
			},
			expectError:    true,
			errorSubstring: "failed to send notifications",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifiers := tt.setupMocks()
			manager := createTestManager(notifiers)

			err := manager.SendMergeResults(context.Background(), tt.results)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all mocks were called as expected
			for _, notifier := range notifiers {
				if mockNotifier, ok := notifier.(*MockNotifier); ok {
					mockNotifier.AssertExpectations(t)
				}
			}
		})
	}
}

func TestManager_SendPRSummary(t *testing.T) {
	repositories := []common.Repository{
		{FullName: "owner/repo1"},
		{FullName: "owner/repo2"},
	}
	totalPRs := 10
	readyPRs := 3

	tests := []struct {
		name        string
		setupMocks  func() []Notifier
		expectError bool
	}{
		{
			name: "no notifiers",
			setupMocks: func() []Notifier {
				return []Notifier{}
			},
			expectError: false,
		},
		{
			name: "single notifier success",
			setupMocks: func() []Notifier {
				mock1 := &MockNotifier{}
				mock1.On("SendPRSummary", mock.Anything, repositories, totalPRs, readyPRs).Return(nil)
				return []Notifier{mock1}
			},
			expectError: false,
		},
		{
			name: "notifier failure",
			setupMocks: func() []Notifier {
				mock1 := &MockNotifier{}
				mock1.On("SendPRSummary", mock.Anything, repositories, totalPRs, readyPRs).
					Return(errors.New("notification failed"))
				return []Notifier{mock1}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifiers := tt.setupMocks()
			manager := createTestManager(notifiers)

			err := manager.SendPRSummary(context.Background(), repositories, totalPRs, readyPRs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all mocks were called as expected
			for _, notifier := range notifiers {
				if mockNotifier, ok := notifier.(*MockNotifier); ok {
					mockNotifier.AssertExpectations(t)
				}
			}
		})
	}
}

func TestResolveEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envValue string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			envValue: "",
			expected: "",
		},
		{
			name:     "regular string",
			input:    "regular-value",
			envValue: "",
			expected: "regular-value",
		},
		{
			name:     "environment variable",
			input:    "$TEST_VAR",
			envValue: "env-value",
			expected: "env-value",
		},
		{
			name:     "environment variable not set",
			input:    "$NONEXISTENT_VAR",
			envValue: "",
			expected: "",
		},
		{
			name:     "string starting with dollar but not env var pattern",
			input:    "$",
			envValue: "",
			expected: "$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable if needed
			if tt.input != "" && tt.input[0] == '$' && len(tt.input) > 1 {
				envVar := tt.input[1:]
				if tt.envValue != "" {
					t.Setenv(envVar, tt.envValue)
				}
			}

			result := resolveEnvVar(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_GetNotifierCount(t *testing.T) {
	manager := &Manager{
		notifiers: []Notifier{&MockNotifier{}, &MockNotifier{}},
	}
	assert.Equal(t, 2, manager.GetNotifierCount())

	emptyManager := &Manager{
		notifiers: []Notifier{},
	}
	assert.Equal(t, 0, emptyManager.GetNotifierCount())
}

func TestManager_HasNotifiers(t *testing.T) {
	manager := &Manager{
		notifiers: []Notifier{&MockNotifier{}},
	}
	assert.True(t, manager.HasNotifiers())

	emptyManager := &Manager{
		notifiers: []Notifier{},
	}
	assert.False(t, emptyManager.HasNotifiers())
}

// Helper to create test manager without using NewManager constructor
func createTestManager(notifiers []Notifier) *Manager {
	// Use the same approach as NewManager but avoid config dependency
	testConfig := &config.Config{
		Notifications: config.Notifications{},
	}
	manager, _ := NewManager(testConfig)
	// Replace notifiers with our test mocks
	manager.notifiers = notifiers
	return manager
}

// Integration test with real config
func TestNewManager_WithRealConfig(t *testing.T) {
	// Test with environment variable resolution
	t.Setenv("TEST_WEBHOOK_URL", "https://hooks.slack.com/test")
	t.Setenv("TEST_EMAIL_PASSWORD", "secret123")

	cfg := &config.Config{
		Notifications: config.Notifications{
			Slack: config.SlackConfig{
				WebhookURL: "$TEST_WEBHOOK_URL",
				Channel:    "#test",
			},
			Email: config.EmailConfig{
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     587,
				SMTPUsername: "test@example.com",
				SMTPPassword: "$TEST_EMAIL_PASSWORD",
				From:         "test@example.com",
				To:           []string{"admin@example.com"},
			},
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, manager.GetNotifierCount())
	assert.True(t, manager.HasNotifiers())
}

func TestManager_SendMergeResults_WithTimeout(t *testing.T) {
	// Test that the manager can handle context cancellation
	// This test just verifies that a timeout error from a notifier is handled properly
	mock1 := &MockNotifier{}
	mock1.On("SendMergeResults", mock.Anything, mock.Anything).
		Return(context.DeadlineExceeded)

	manager := createTestManager([]Notifier{mock1})

	results := []merge.MergeResult{
		{Provider: "github", Repository: "owner/repo", PullRequest: 1, Success: true},
	}

	err := manager.SendMergeResults(context.Background(), results)
	assert.Error(t, err)
	// The manager should wrap the deadline exceeded error in its own error message
	assert.Contains(t, err.Error(), "failed to send notifications")

	mock1.AssertExpectations(t)
}
