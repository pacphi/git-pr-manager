package commands

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestCommand(t *testing.T) {
	tests := []struct {
		name             string
		expectedUse      string
		expectedShort    string
		minExpectedFlags int
	}{
		{
			name:             "creates test command with correct properties",
			expectedUse:      "test",
			expectedShort:    "Test system functionality and integrations",
			minExpectedFlags: 8, // Minimum expected number of flags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewTestCommand()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.expectedUse, cmd.Use)
			assert.Equal(t, tt.expectedShort, cmd.Short)
			assert.NotEmpty(t, cmd.Long)
			assert.NotNil(t, cmd.RunE)

			// Check that command has the expected flags defined
			// Note: NFlag() returns number of flags set, not defined, so we count all defined flags
			flagCount := 0
			cmd.Flags().VisitAll(func(*pflag.Flag) { flagCount++ })
			assert.GreaterOrEqual(t, flagCount, tt.minExpectedFlags)

			// Check specific important flags
			flag := cmd.Flags().Lookup("notifications")
			assert.NotNil(t, flag, "notifications flag should exist")

			flag = cmd.Flags().Lookup("slack")
			assert.NotNil(t, flag, "slack flag should exist")

			flag = cmd.Flags().Lookup("email")
			assert.NotNil(t, flag, "email flag should exist")

			flag = cmd.Flags().Lookup("auth")
			assert.NotNil(t, flag, "auth flag should exist")

			flag = cmd.Flags().Lookup("verbose")
			assert.NotNil(t, flag, "verbose flag should exist")

			flag = cmd.Flags().Lookup("timeout")
			assert.NotNil(t, flag, "timeout flag should exist")
			assert.Equal(t, "30s", flag.DefValue)
		})
	}
}

func TestTestFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected TestFlags
		wantErr  bool
	}{
		{
			name: "default flags set notifications to true",
			args: []string{},
			expected: TestFlags{
				TestNotifications: true,
				Timeout:           "30s",
			},
			wantErr: false,
		},
		{
			name: "explicit slack flag",
			args: []string{"--slack"},
			expected: TestFlags{
				TestSlack: true,
				Timeout:   "30s",
			},
			wantErr: false,
		},
		{
			name: "explicit email flag",
			args: []string{"--email"},
			expected: TestFlags{
				TestEmail: true,
				Timeout:   "30s",
			},
			wantErr: false,
		},
		{
			name: "auth flag",
			args: []string{"--auth"},
			expected: TestFlags{
				Auth:    true,
				Timeout: "30s",
			},
			wantErr: false,
		},
		{
			name: "verbose flag",
			args: []string{"--verbose", "--notifications"},
			expected: TestFlags{
				TestNotifications: true,
				Verbose:           true,
				Timeout:           "30s",
			},
			wantErr: false,
		},
		{
			name: "custom timeout",
			args: []string{"--timeout", "60s", "--notifications"},
			expected: TestFlags{
				TestNotifications: true,
				Timeout:           "60s",
			},
			wantErr: false,
		},
		{
			name: "provider filter",
			args: []string{"--provider", "github", "--notifications"},
			expected: TestFlags{
				TestNotifications: true,
				Provider:          "github",
				Timeout:           "30s",
			},
			wantErr: false,
		},
		{
			name: "integration test flag",
			args: []string{"--integration"},
			expected: TestFlags{
				Integration: true,
				Timeout:     "30s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewTestCommand()
			cmd.SetArgs(tt.args)

			// Parse flags without executing
			err := cmd.ParseFlags(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Execute PreRunE to apply defaults
			if cmd.PreRunE != nil {
				err = cmd.PreRunE(cmd, tt.args)
				require.NoError(t, err)
			}

			// Get the flags
			notifications, _ := cmd.Flags().GetBool("notifications")
			slack, _ := cmd.Flags().GetBool("slack")
			email, _ := cmd.Flags().GetBool("email")
			auth, _ := cmd.Flags().GetBool("auth")
			verbose, _ := cmd.Flags().GetBool("verbose")
			timeout, _ := cmd.Flags().GetString("timeout")
			provider, _ := cmd.Flags().GetString("provider")
			integration, _ := cmd.Flags().GetBool("integration")

			assert.Equal(t, tt.expected.TestNotifications, notifications)
			assert.Equal(t, tt.expected.TestSlack, slack)
			assert.Equal(t, tt.expected.TestEmail, email)
			assert.Equal(t, tt.expected.Auth, auth)
			assert.Equal(t, tt.expected.Verbose, verbose)
			assert.Equal(t, tt.expected.Timeout, timeout)
			assert.Equal(t, tt.expected.Provider, provider)
			assert.Equal(t, tt.expected.Integration, integration)
		})
	}
}

func TestRunTest_TimeoutParsing(t *testing.T) {
	tests := []struct {
		name           string
		timeout        string
		wantError      bool
		expectDuration time.Duration
	}{
		{
			name:           "valid timeout",
			timeout:        "30s",
			wantError:      false,
			expectDuration: 30 * time.Second,
		},
		{
			name:           "valid timeout with minutes",
			timeout:        "2m",
			wantError:      false,
			expectDuration: 2 * time.Minute,
		},
		{
			name:      "invalid timeout",
			timeout:   "invalid",
			wantError: true,
		},
		{
			name:      "negative timeout",
			timeout:   "-10s",
			wantError: false, // Go's time.ParseDuration accepts negative durations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := TestFlags{
				Timeout: tt.timeout,
			}

			duration, err := time.ParseDuration(flags.Timeout)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectDuration > 0 {
					assert.Equal(t, tt.expectDuration, duration)
				}
			}
		})
	}
}

func TestMatchesTestRepoFilter(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		filter   string
		expected bool
	}{
		{
			name:     "empty filter matches all",
			repoName: "owner/repo",
			filter:   "",
			expected: true,
		},
		{
			name:     "exact match",
			repoName: "owner/repo",
			filter:   "owner/repo",
			expected: true,
		},
		{
			name:     "partial match",
			repoName: "owner/my-repo",
			filter:   "my-repo",
			expected: true,
		},
		{
			name:     "comma-separated list first match",
			repoName: "owner/repo1",
			filter:   "repo1,repo2,repo3",
			expected: true,
		},
		{
			name:     "comma-separated list middle match",
			repoName: "owner/repo2",
			filter:   "repo1,repo2,repo3",
			expected: true,
		},
		{
			name:     "comma-separated list last match",
			repoName: "owner/repo3",
			filter:   "repo1,repo2,repo3",
			expected: true,
		},
		{
			name:     "no match",
			repoName: "owner/different",
			filter:   "repo1,repo2,repo3",
			expected: false,
		},
		{
			name:     "partial match in comma list",
			repoName: "owner/my-special-repo",
			filter:   "other,special,third",
			expected: true,
		},
		{
			name:     "whitespace handling",
			repoName: "owner/repo",
			filter:   " repo1 , repo , repo3 ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesTestRepoFilter(tt.repoName, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMockPRs(t *testing.T) {
	mockPRs := generateMockPRs()

	assert.NotEmpty(t, mockPRs, "Should generate at least one mock PR")

	for i, pr := range mockPRs {
		t.Run(t.Name()+"_pr_"+string(rune('0'+i)), func(t *testing.T) {
			assert.NotEmpty(t, pr.Title, "PR should have a title")
			assert.NotEmpty(t, pr.Author, "PR should have an author")
			assert.NotEmpty(t, pr.Status, "PR should have a status")
			assert.Greater(t, pr.ID, 0, "PR should have a positive ID")
		})
	}

	// Check that we have different types of PRs for comprehensive testing
	var mergeablePRs, nonMergeablePRs int
	for _, pr := range mockPRs {
		if pr.Mergeable {
			mergeablePRs++
		} else {
			nonMergeablePRs++
		}
	}

	assert.Greater(t, mergeablePRs, 0, "Should have at least one mergeable PR for testing")
	assert.Greater(t, nonMergeablePRs, 0, "Should have at least one non-mergeable PR for testing")
}

func TestProcessMockPR(t *testing.T) {
	tests := []struct {
		name    string
		pr      MockPR
		wantErr bool
	}{
		{
			name: "valid PR",
			pr: MockPR{
				ID:        1,
				Title:     "feat: add new feature",
				Author:    "developer",
				Status:    "open",
				Mergeable: true,
			},
			wantErr: false,
		},
		{
			name: "PR with empty title",
			pr: MockPR{
				ID:     2,
				Title:  "",
				Author: "developer",
				Status: "open",
			},
			wantErr: true,
		},
		{
			name: "PR with empty author",
			pr: MockPR{
				ID:     3,
				Title:  "fix: bug fix",
				Author: "",
				Status: "open",
			},
			wantErr: true,
		},
		{
			name: "valid non-mergeable PR",
			pr: MockPR{
				ID:        4,
				Title:     "docs: update documentation",
				Author:    "writer",
				Status:    "open",
				Mergeable: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processMockPR(tt.pr)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResolveEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "regular value",
			value:    "regular-value",
			expected: "regular-value",
		},
		{
			name:     "empty value",
			value:    "",
			expected: "",
		},
		{
			name:     "environment variable with value",
			value:    "$TEST_VAR",
			envVar:   "TEST_VAR",
			envValue: "test-value",
			expected: "test-value",
		},
		{
			name:     "environment variable not set",
			value:    "$UNSET_VAR",
			envVar:   "OTHER_VAR",
			envValue: "other-value",
			expected: "",
		},
		{
			name:     "dollar sign without variable name",
			value:    "$",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable if needed
			if tt.envVar != "" && tt.envValue != "" {
				oldValue := os.Getenv(tt.envVar)
				os.Setenv(tt.envVar, tt.envValue)
				defer func() {
					if oldValue == "" {
						os.Unsetenv(tt.envVar)
					} else {
						os.Setenv(tt.envVar, oldValue)
					}
				}()
			}

			result := resolveEnvVar(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration-style tests that exercise command execution paths
func TestTestCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		reason  string
	}{
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
			reason:  "help should always work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewTestCommand()

			// Set up a minimal execution context
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			cmd.SetArgs(tt.args)

			err := cmd.ExecuteContext(ctx)

			if tt.wantErr {
				assert.Error(t, err, tt.reason)
			} else {
				// Help command returns ErrHelp which is not actually an error
				if err != nil && err.Error() != "help requested" {
					assert.NoError(t, err, tt.reason)
				}
			}
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkMatchesTestRepoFilter(b *testing.B) {
	testCases := []struct {
		repoName string
		filter   string
	}{
		{"owner/repo1", "repo1,repo2,repo3,repo4,repo5"},
		{"owner/my-long-repository-name", "short,medium-name,my-long-repository-name,another-repo"},
		{"org/complex-repo-name-with-many-parts", "simple,complex-repo-name-with-many-parts"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc := testCases[i%len(testCases)]
		matchesTestRepoFilter(tc.repoName, tc.filter)
	}
}

func BenchmarkProcessMockPR(b *testing.B) {
	pr := MockPR{
		ID:        1,
		Title:     "feat: add new feature with a reasonably long title",
		Author:    "developer-username",
		Status:    "open",
		Mergeable: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processMockPR(pr)
	}
}

func BenchmarkGenerateMockPRs(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateMockPRs()
	}
}

// Test helper functions
func TestMockPR_Validation(t *testing.T) {
	validPR := MockPR{
		ID:        1,
		Title:     "Valid PR",
		Author:    "author",
		Status:    "open",
		Mergeable: true,
	}

	// Test JSON serialization
	data, err := json.Marshal(validPR)
	require.NoError(t, err)

	var unmarshaled MockPR
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, validPR.ID, unmarshaled.ID)
	assert.Equal(t, validPR.Title, unmarshaled.Title)
	assert.Equal(t, validPR.Author, unmarshaled.Author)
	assert.Equal(t, validPR.Status, unmarshaled.Status)
	assert.Equal(t, validPR.Mergeable, unmarshaled.Mergeable)
}

// Test command flag combinations
func TestTestCommand_FlagCombinations(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		checkFn func(*testing.T, *cobra.Command)
	}{
		{
			name: "all notification flags",
			args: []string{"--notifications", "--slack", "--email"},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				notifications, _ := cmd.Flags().GetBool("notifications")
				slack, _ := cmd.Flags().GetBool("slack")
				email, _ := cmd.Flags().GetBool("email")

				assert.True(t, notifications)
				assert.True(t, slack)
				assert.True(t, email)
			},
		},
		{
			name: "all test types",
			args: []string{"--auth", "--integration", "--load-test", "--performance"},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				auth, _ := cmd.Flags().GetBool("auth")
				integration, _ := cmd.Flags().GetBool("integration")
				loadTest, _ := cmd.Flags().GetBool("load-test")
				performance, _ := cmd.Flags().GetBool("performance")

				assert.True(t, auth)
				assert.True(t, integration)
				assert.True(t, loadTest)
				assert.True(t, performance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewTestCommand()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			tt.checkFn(t, cmd)
		})
	}
}
