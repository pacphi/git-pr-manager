package utils

import (
	"fmt"
	"time"
)

// FormatTimestamp formats a timestamp in a consistent way for filenames and IDs
func FormatTimestamp() string {
	return time.Now().Format("20060102-150405")
}

// FormatTimestampHuman formats a timestamp in a human-readable way for display
func FormatTimestampHuman() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}

// TimeAgo returns a human-readable "time ago" string for PR/commit display
func TimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		seconds := int(duration.Seconds())
		if seconds <= 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)

	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)

	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)

	case duration < 30*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)

	default:
		// For very old items, just show days
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
}
