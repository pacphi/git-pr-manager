package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatTimestamp(t *testing.T) {
	// Since this function uses time.Now(), we can only test the format
	result := FormatTimestamp()
	assert.Len(t, result, 15) // Format: 20060102-150405 (15 characters)
	assert.Regexp(t, `^\d{8}-\d{6}$`, result)
}

func TestFormatTimestampHuman(t *testing.T) {
	// Since this function uses time.Now(), we can only test the format
	result := FormatTimestampHuman()
	assert.Contains(t, result, "-") // Should contain date separators
	assert.Contains(t, result, ":") // Should contain time separators
	assert.Contains(t, result, " ") // Should contain space between date and time
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "1 second ago",
			input:    now.Add(-1 * time.Second),
			expected: "1 second ago",
		},
		{
			name:     "30 seconds ago",
			input:    now.Add(-30 * time.Second),
			expected: "30 seconds ago",
		},
		{
			name:     "1 minute ago",
			input:    now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "30 minutes ago",
			input:    now.Add(-30 * time.Minute),
			expected: "30 minutes ago",
		},
		{
			name:     "1 hour ago",
			input:    now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "5 hours ago",
			input:    now.Add(-5 * time.Hour),
			expected: "5 hours ago",
		},
		{
			name:     "1 day ago",
			input:    now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "7 days ago",
			input:    now.Add(-7 * 24 * time.Hour),
			expected: "7 days ago",
		},
		{
			name:     "30 days ago",
			input:    now.Add(-30 * 24 * time.Hour),
			expected: "30 days ago",
		},
		{
			name:     "90 days ago",
			input:    now.Add(-90 * 24 * time.Hour),
			expected: "90 days ago",
		},
		{
			name:     "365 days ago",
			input:    now.Add(-365 * 24 * time.Hour),
			expected: "365 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeAgo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// BeginningOfDay, EndOfDay, BeginningOfWeek, EndOfWeek, BeginningOfMonth, EndOfMonth
// IsToday, IsYesterday, IsThisWeek, IsThisMonth functions have been removed
// as they were over-engineered for this use case.

// Benchmark tests
func BenchmarkTimeAgo(b *testing.B) {
	testTime := time.Now().Add(-2 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeAgo(testTime)
	}
}

func BenchmarkFormatTimestamp(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatTimestamp()
	}
}

func BenchmarkFormatTimestampHuman(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatTimestampHuman()
	}
}
