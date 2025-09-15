package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStringUtils_Truncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		suffix   string
		expected string
	}{
		{
			name:     "string longer than length",
			input:    "this is a long string",
			length:   10,
			suffix:   "...",
			expected: "this is...",
		},
		{
			name:     "string equal to length",
			input:    "exact",
			length:   5,
			suffix:   "...",
			expected: "exact",
		},
		{
			name:     "string shorter than length",
			input:    "short",
			length:   10,
			suffix:   "...",
			expected: "short",
		},
		{
			name:     "empty string",
			input:    "",
			length:   10,
			suffix:   "...",
			expected: "",
		},
		{
			name:     "no suffix",
			input:    "this is a long string",
			length:   10,
			suffix:   "",
			expected: "this is a ",
		},
		{
			name:     "suffix longer than length",
			input:    "test",
			length:   2,
			suffix:   "...",
			expected: "te",
		},
	}

	su := NewStringUtils()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := su.Truncate(tt.input, tt.length, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringUtils_ParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		expectError bool
	}{
		{
			name:        "valid go duration",
			input:       "5s",
			expected:    5 * time.Second,
			expectError: false,
		},
		{
			name:        "valid minutes",
			input:       "2m",
			expected:    2 * time.Minute,
			expectError: false,
		},
		{
			name:        "valid hours",
			input:       "1h",
			expected:    1 * time.Hour,
			expectError: false,
		},
		{
			name:        "number only as seconds",
			input:       "30",
			expected:    30 * time.Second,
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative duration",
			input:       "-5s",
			expected:    -5 * time.Second,
			expectError: false,
		},
		{
			name:        "days",
			input:       "5d",
			expected:    5 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "weeks",
			input:       "2w",
			expected:    2 * 7 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "years",
			input:       "1y",
			expected:    365 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "fractional days",
			input:       "1.5d",
			expected:    time.Duration(1.5 * float64(24*time.Hour)),
			expectError: false,
		},
		{
			name:        "30 days",
			input:       "30d",
			expected:    30 * 24 * time.Hour,
			expectError: false,
		},
	}

	su := NewStringUtils()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := su.ParseDuration(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStringUtils_FormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			input:    500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "seconds",
			input:    5 * time.Second,
			expected: "5.0s",
		},
		{
			name:     "minutes",
			input:    2 * time.Minute,
			expected: "2.0m",
		},
		{
			name:     "hours",
			input:    3 * time.Hour,
			expected: "3.0h",
		},
		{
			name:     "days",
			input:    2 * 24 * time.Hour,
			expected: "2.0d",
		},
		{
			name:     "fractional seconds",
			input:    1500 * time.Millisecond,
			expected: "1.5s",
		},
	}

	su := NewStringUtils()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := su.FormatDuration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test global utility functions
func TestGlobalUtilityFunctions(t *testing.T) {
	t.Run("ParseDuration global function", func(t *testing.T) {
		result, err := ParseDuration("5s")
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, result)
	})

	t.Run("Truncate global function", func(t *testing.T) {
		result := Truncate("long string", 5, "...")
		assert.Equal(t, "lo...", result)
	})
}

// Benchmark tests
func BenchmarkStringUtils_ParseDuration(b *testing.B) {
	su := NewStringUtils()
	input := "5m30s"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		su.ParseDuration(input)
	}
}

func BenchmarkStringUtils_Truncate(b *testing.B) {
	su := NewStringUtils()
	input := "This is a very long string that needs to be truncated"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		su.Truncate(input, 20, "...")
	}
}
