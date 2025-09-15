package utils

import (
	"fmt"
	"strconv"
	"time"
)

// StringUtils provides utility functions for string manipulation
type StringUtils struct{}

// NewStringUtils creates a new StringUtils instance
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// Truncate truncates a string to the specified length
func (su *StringUtils) Truncate(s string, length int, suffix string) string {
	if len(s) <= length {
		return s
	}
	if length <= len(suffix) {
		return s[:length]
	}
	return s[:length-len(suffix)] + suffix
}

// ParseDuration parses a duration string with support for common formats
func (su *StringUtils) ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	// Try standard Go duration format first
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Try parsing as seconds if it's just a number
	if seconds, err := strconv.Atoi(s); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}

// FormatDuration formats a duration in a human-readable way
func (su *StringUtils) FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// Global string utilities instance
var stringUtils = NewStringUtils()

// Global utility functions for convenience

// Truncate truncates a string to the specified length
func Truncate(s string, length int, suffix string) string {
	return stringUtils.Truncate(s, length, suffix)
}

// ParseDuration parses a duration string with support for common formats
func ParseDuration(s string) (time.Duration, error) {
	return stringUtils.ParseDuration(s)
}

