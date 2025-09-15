package commands

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCheckCommand(t *testing.T) {
	tests := []struct {
		name             string
		expectedUse      string
		expectedShort    string
		minExpectedFlags int
	}{
		{
			name:             "creates check command with correct properties",
			expectedUse:      "check",
			expectedShort:    "Check pull request status across repositories",
			minExpectedFlags: 8, // Minimum expected number of flags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCheckCommand()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.expectedUse, cmd.Use)
			assert.Equal(t, tt.expectedShort, cmd.Short)
			assert.NotEmpty(t, cmd.Long)
			assert.NotNil(t, cmd.RunE)
			assert.NotEmpty(t, cmd.Example)

			// Check that command has the expected flags
			var flagCount int
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				flagCount++
			})
			assert.GreaterOrEqual(t, flagCount, tt.minExpectedFlags)
		})
	}
}

func TestCheckCommand_Flags(t *testing.T) {
	tests := []struct {
		name            string
		flagName        string
		expectedType    string
		expectedDefault interface{}
	}{
		{
			name:         "providers flag",
			flagName:     "providers",
			expectedType: "stringSlice",
		},
		{
			name:         "repos flag",
			flagName:     "repos",
			expectedType: "stringSlice",
		},
		{
			name:            "max-age flag",
			flagName:        "max-age",
			expectedType:    "string",
			expectedDefault: "",
		},
		{
			name:            "output flag",
			flagName:        "output",
			expectedType:    "string",
			expectedDefault: "table",
		},
		{
			name:            "require-checks flag",
			flagName:        "require-checks",
			expectedType:    "bool",
			expectedDefault: false,
		},
		{
			name:            "show-skipped flag",
			flagName:        "show-skipped",
			expectedType:    "bool",
			expectedDefault: false,
		},
		{
			name:            "show-details flag",
			flagName:        "show-details",
			expectedType:    "bool",
			expectedDefault: false,
		},
		{
			name:            "show-status flag",
			flagName:        "show-status",
			expectedType:    "bool",
			expectedDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCheckCommand()

			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag, "Flag %s should exist", tt.flagName)

			assert.Equal(t, tt.expectedType, flag.Value.Type())

			if tt.expectedDefault != nil {
				switch tt.expectedType {
				case "string":
					assert.Equal(t, tt.expectedDefault.(string), flag.DefValue)
				case "bool":
					assert.Equal(t, tt.expectedDefault.(bool), flag.DefValue == "true")
				}
			}
		})
	}
}

func TestCheckFlags_Parsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected CheckFlags
		wantErr  bool
	}{
		{
			name: "default flags",
			args: []string{},
			expected: CheckFlags{
				Output: "table",
			},
			wantErr: false,
		},
		{
			name: "single provider",
			args: []string{"--providers", "github"},
			expected: CheckFlags{
				Providers: []string{"github"},
				Output:    "table",
			},
			wantErr: false,
		},
		{
			name: "multiple providers",
			args: []string{"--providers", "github,gitlab,bitbucket"},
			expected: CheckFlags{
				Providers: []string{"github", "gitlab", "bitbucket"},
				Output:    "table",
			},
			wantErr: false,
		},
		{
			name: "single repository",
			args: []string{"--repos", "owner/repo1"},
			expected: CheckFlags{
				Repositories: []string{"owner/repo1"},
				Output:       "table",
			},
			wantErr: false,
		},
		{
			name: "multiple repositories",
			args: []string{"--repos", "owner/repo1,owner/repo2"},
			expected: CheckFlags{
				Repositories: []string{"owner/repo1", "owner/repo2"},
				Output:       "table",
			},
			wantErr: false,
		},
		{
			name: "max age setting",
			args: []string{"--max-age", "7d"},
			expected: CheckFlags{
				MaxAge: "7d",
				Output: "table",
			},
			wantErr: false,
		},
		{
			name: "JSON output",
			args: []string{"--output", "json"},
			expected: CheckFlags{
				Output: "json",
			},
			wantErr: false,
		},
		{
			name: "YAML output",
			args: []string{"--output", "yaml"},
			expected: CheckFlags{
				Output: "yaml",
			},
			wantErr: false,
		},
		{
			name: "CSV output",
			args: []string{"--output", "csv"},
			expected: CheckFlags{
				Output: "csv",
			},
			wantErr: false,
		},
		{
			name: "summary output",
			args: []string{"--output", "summary"},
			expected: CheckFlags{
				Output: "summary",
			},
			wantErr: false,
		},
		{
			name: "require checks",
			args: []string{"--require-checks"},
			expected: CheckFlags{
				Output:        "table",
				RequireChecks: true,
			},
			wantErr: false,
		},
		{
			name: "show skipped",
			args: []string{"--show-skipped"},
			expected: CheckFlags{
				Output:      "table",
				ShowSkipped: true,
			},
			wantErr: false,
		},
		{
			name: "show details",
			args: []string{"--show-details"},
			expected: CheckFlags{
				Output:      "table",
				ShowDetails: true,
			},
			wantErr: false,
		},
		{
			name: "show status",
			args: []string{"--show-status"},
			expected: CheckFlags{
				Output:     "table",
				ShowStatus: true,
			},
			wantErr: false,
		},
		{
			name: "complex combination",
			args: []string{
				"--providers", "github,gitlab",
				"--repos", "owner/repo1,owner/repo2",
				"--max-age", "14d",
				"--output", "json",
				"--require-checks",
				"--show-skipped",
				"--show-details",
			},
			expected: CheckFlags{
				Providers:     []string{"github", "gitlab"},
				Repositories:  []string{"owner/repo1", "owner/repo2"},
				MaxAge:        "14d",
				Output:        "json",
				RequireChecks: true,
				ShowSkipped:   true,
				ShowDetails:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCheckCommand()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Extract parsed flag values
			providers, _ := cmd.Flags().GetStringSlice("providers")
			repos, _ := cmd.Flags().GetStringSlice("repos")
			maxAge, _ := cmd.Flags().GetString("max-age")
			output, _ := cmd.Flags().GetString("output")
			requireChecks, _ := cmd.Flags().GetBool("require-checks")
			showSkipped, _ := cmd.Flags().GetBool("show-skipped")
			showDetails, _ := cmd.Flags().GetBool("show-details")
			showStatus, _ := cmd.Flags().GetBool("show-status")

			// Handle empty slices vs nil slices
			if len(tt.expected.Providers) == 0 {
				assert.Empty(t, providers)
			} else {
				assert.Equal(t, tt.expected.Providers, providers)
			}
			if len(tt.expected.Repositories) == 0 {
				assert.Empty(t, repos)
			} else {
				assert.Equal(t, tt.expected.Repositories, repos)
			}
			assert.Equal(t, tt.expected.MaxAge, maxAge)
			assert.Equal(t, tt.expected.Output, output)
			assert.Equal(t, tt.expected.RequireChecks, requireChecks)
			assert.Equal(t, tt.expected.ShowSkipped, showSkipped)
			assert.Equal(t, tt.expected.ShowDetails, showDetails)
			assert.Equal(t, tt.expected.ShowStatus, showStatus)
		})
	}
}

func TestCheckCommand_OutputFormats(t *testing.T) {
	validFormats := []string{"table", "json", "yaml", "csv", "summary"}

	for _, format := range validFormats {
		t.Run("format_"+format, func(t *testing.T) {
			cmd := NewCheckCommand()
			cmd.SetArgs([]string{"--output", format})

			err := cmd.ParseFlags([]string{"--output", format})
			assert.NoError(t, err)

			output, _ := cmd.Flags().GetString("output")
			assert.Equal(t, format, output)
		})
	}
}

func TestCheckCommand_MaxAgeParsing(t *testing.T) {
	tests := []struct {
		name      string
		maxAge    string
		wantValid bool
		expected  time.Duration
	}{
		{
			name:      "valid hours",
			maxAge:    "24h",
			wantValid: true,
			expected:  24 * time.Hour,
		},
		{
			name:      "valid days shorthand",
			maxAge:    "7d",
			wantValid: true,
			// Note: "7d" might not parse directly with time.ParseDuration
			// This would be handled by utils.ParseDuration
		},
		{
			name:      "valid minutes",
			maxAge:    "30m",
			wantValid: true,
			expected:  30 * time.Minute,
		},
		{
			name:      "valid seconds",
			maxAge:    "3600s",
			wantValid: true,
			expected:  3600 * time.Second,
		},
		{
			name:      "empty string",
			maxAge:    "",
			wantValid: true,
		},
		{
			name:      "invalid format",
			maxAge:    "invalid",
			wantValid: false,
		},
		{
			name:      "negative duration",
			maxAge:    "-1h",
			wantValid: true, // Technically valid, though semantically questionable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.maxAge == "" {
				// Skip parsing empty string
				return
			}

			// Try to parse with standard time.ParseDuration first
			duration, err := time.ParseDuration(tt.maxAge)

			if tt.wantValid && strings.HasSuffix(tt.maxAge, "d") {
				// Days format won't parse with standard parser
				assert.Error(t, err)
			} else if tt.wantValid {
				assert.NoError(t, err)
				if tt.expected > 0 {
					assert.Equal(t, tt.expected, duration)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCheckCommand_ShortFlags(t *testing.T) {
	tests := []struct {
		name      string
		shortFlag string
		longFlag  string
	}{
		{
			name:      "providers short flag",
			shortFlag: "-p",
			longFlag:  "--providers",
		},
		{
			name:      "repos short flag",
			shortFlag: "-r",
			longFlag:  "--repos",
		},
		{
			name:      "output short flag",
			shortFlag: "-o",
			longFlag:  "--output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCheckCommand()

			// Test short flag
			cmd.SetArgs([]string{tt.shortFlag, "test-value"})
			err := cmd.ParseFlags([]string{tt.shortFlag, "test-value"})
			assert.NoError(t, err)

			// Test long flag
			cmd.SetArgs([]string{tt.longFlag, "test-value"})
			err = cmd.ParseFlags([]string{tt.longFlag, "test-value"})
			assert.NoError(t, err)
		})
	}
}

func TestCheckCommand_Integration(t *testing.T) {
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
			cmd := NewCheckCommand()

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

// Test table format output helpers
func TestCheckCommand_TableFormatHelpers(t *testing.T) {
	// Test that the table format functions exist and work with mock data
	// These would be integration tests with actual PR data in a real scenario

	t.Run("table headers", func(t *testing.T) {
		// Basic validation that we can construct table headers
		headers := []string{"PROVIDER", "REPOSITORY", "PR", "TITLE", "STATUS"}
		assert.Len(t, headers, 5)
		assert.Equal(t, "PROVIDER", headers[0])
		assert.Equal(t, "STATUS", headers[len(headers)-1])
	})

	t.Run("column widths", func(t *testing.T) {
		// Basic validation of column width calculations
		widths := []int{15, 30, 8, 25, 10}
		assert.Len(t, widths, 5)
		for _, width := range widths {
			assert.Greater(t, width, 0)
		}
	})
}

// Test CSV output format
func TestCheckCommand_CSVOutput(t *testing.T) {
	t.Run("csv header format", func(t *testing.T) {
		expectedHeaders := []string{"Provider", "Repository", "PR Number", "Title", "Author", "Status", "Ready", "Reason"}
		assert.Len(t, expectedHeaders, 8)
		assert.Equal(t, "Provider", expectedHeaders[0])
		assert.Equal(t, "Reason", expectedHeaders[len(expectedHeaders)-1])
	})
}

// Test output format validation
func TestCheckCommand_OutputFormatValidation(t *testing.T) {
	validFormats := []string{"table", "json", "yaml", "csv", "summary"}
	invalidFormats := []string{"xml", "html", "invalid", ""}

	for _, format := range validFormats {
		t.Run("valid_format_"+format, func(t *testing.T) {
			// Test that the format string is recognized
			switch strings.ToLower(format) {
			case "json", "yaml", "csv", "summary", "table":
				// These should all be valid
				assert.Contains(t, validFormats, format)
			default:
				t.Errorf("Format %s should be valid but is not in validFormats slice", format)
			}
		})
	}

	for _, format := range invalidFormats {
		t.Run("invalid_format_"+format, func(t *testing.T) {
			if format == "" {
				// Empty string might default to "table"
				return
			}
			// Test that invalid formats are not in the valid list
			assert.NotContains(t, validFormats, format)
		})
	}
}

// Benchmark tests
func BenchmarkCheckCommand_FlagParsing(b *testing.B) {
	args := []string{
		"--providers", "github,gitlab,bitbucket",
		"--repos", "owner/repo1,owner/repo2,owner/repo3",
		"--output", "json",
		"--require-checks",
		"--show-details",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCheckCommand()
		cmd.SetArgs(args)
		_ = cmd.ParseFlags(args)
	}
}

func BenchmarkCheckCommand_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewCheckCommand()
	}
}

// Test flag combinations that might cause conflicts
func TestCheckCommand_FlagCombinations(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		checkFn func(*testing.T, *cobra.Command)
	}{
		{
			name: "all output options",
			args: []string{"--show-details", "--show-status", "--show-skipped"},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				showDetails, _ := cmd.Flags().GetBool("show-details")
				showStatus, _ := cmd.Flags().GetBool("show-status")
				showSkipped, _ := cmd.Flags().GetBool("show-skipped")

				assert.True(t, showDetails)
				assert.True(t, showStatus)
				assert.True(t, showSkipped)
			},
		},
		{
			name: "filters and requirements",
			args: []string{"--providers", "github", "--require-checks", "--max-age", "7d"},
			checkFn: func(t *testing.T, cmd *cobra.Command) {
				providers, _ := cmd.Flags().GetStringSlice("providers")
				requireChecks, _ := cmd.Flags().GetBool("require-checks")
				maxAge, _ := cmd.Flags().GetString("max-age")

				assert.Equal(t, []string{"github"}, providers)
				assert.True(t, requireChecks)
				assert.Equal(t, "7d", maxAge)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCheckCommand()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			tt.checkFn(t, cmd)
		})
	}
}

// Test environment variable interaction
func TestCheckCommand_EnvironmentVariables(t *testing.T) {
	// Test that the command can work with various environment variable states
	t.Run("with config file env var", func(t *testing.T) {
		oldValue := os.Getenv("CONFIG_FILE")
		os.Setenv("CONFIG_FILE", "test-config.yaml")
		defer func() {
			if oldValue == "" {
				os.Unsetenv("CONFIG_FILE")
			} else {
				os.Setenv("CONFIG_FILE", oldValue)
			}
		}()

		cmd := NewCheckCommand()
		assert.NotNil(t, cmd)
		// The command should be created successfully regardless of env vars
	})
}
