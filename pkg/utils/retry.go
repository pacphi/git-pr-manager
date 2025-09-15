package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig defines configuration for retry operations
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
	Jitter         bool
	RetryIf        func(error) bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
		Jitter:         true,
		RetryIf:        func(err error) bool { return err != nil },
	}
}

// RetryFunc represents a function that can be retried
type RetryFunc func() error

// RetryFuncWithResult represents a function that returns a value and can be retried
type RetryFuncWithResult[T any] func() (T, error)

// Retry executes a function with retry logic
func Retry(ctx context.Context, config *RetryConfig, fn RetryFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastError error
	logger := GetGlobalLogger()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			if attempt > 1 {
				logger.Infof("Operation succeeded on attempt %d", attempt)
			}
			return nil
		}

		lastError = err

		// Check if we should retry this error
		if !config.RetryIf(err) {
			logger.Debugf("Error not retryable: %v", err)
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts {
			logger.Warnf("All %d attempts failed, giving up", config.MaxAttempts)
			break
		}

		// Calculate backoff duration
		backoff := config.calculateBackoff(attempt)
		logger.Warnf("Attempt %d failed, retrying in %v: %v", attempt, backoff, err)

		// Sleep with context cancellation support
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastError)
}

// RetryWithResult executes a function with retry logic and returns a result
func RetryWithResult[T any](ctx context.Context, config *RetryConfig, fn RetryFuncWithResult[T]) (T, error) {
	var zero T
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastError error
	logger := GetGlobalLogger()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, err := fn()
		if err == nil {
			if attempt > 1 {
				logger.Infof("Operation succeeded on attempt %d", attempt)
			}
			return result, nil
		}

		lastError = err

		// Check if we should retry this error
		if !config.RetryIf(err) {
			logger.Debugf("Error not retryable: %v", err)
			return zero, err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts {
			logger.Warnf("All %d attempts failed, giving up", config.MaxAttempts)
			break
		}

		// Calculate backoff duration
		backoff := config.calculateBackoff(attempt)
		logger.Warnf("Attempt %d failed, retrying in %v: %v", attempt, backoff, err)

		// Sleep with context cancellation support
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}

	return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastError)
}

// calculateBackoff calculates the backoff duration for a given attempt
func (c *RetryConfig) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff
	backoff := time.Duration(float64(c.InitialBackoff) * math.Pow(c.BackoffFactor, float64(attempt-1)))

	// Apply maximum backoff limit
	if backoff > c.MaxBackoff {
		backoff = c.MaxBackoff
	}

	// Add jitter to prevent thundering herd
	if c.Jitter {
		jitterRange := time.Duration(float64(backoff) * 0.1) // 10% jitter
		if jitterRange > 0 {
			jitter := time.Duration(rand.Int63n(int64(jitterRange)))
			backoff = backoff + jitter - jitterRange/2 // Add random jitter (+/- 5%)
		}
	}

	// Ensure minimum backoff
	if backoff < 0 {
		backoff = c.InitialBackoff
	}

	return backoff
}

// IsRetryableError is a helper function to determine if common errors are retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Network timeouts and temporary errors are retryable
	if isTemporaryError(err) {
		return true
	}

	// Check for specific error patterns that are retryable
	errorStr := err.Error()
	retryablePatterns := []string{
		"connection reset by peer",
		"connection refused",
		"no such host",
		"temporary failure",
		"timeout",
		"rate limited",
		"server error",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
	}

	for _, pattern := range retryablePatterns {
		if contains(errorStr, pattern) {
			return true
		}
	}

	return false
}

// isTemporaryError checks if an error implements the Temporary interface
func isTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}

	if te, ok := err.(temporary); ok {
		return te.Temporary()
	}

	return false
}

// contains is a case-insensitive string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexContains(s, substr) >= 0)))
}

// indexContains is a helper for case-insensitive substring search
func indexContains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// RetryableFunc creates a retry configuration function for common scenarios
type RetryableFunc func(error) bool

// Common retry functions
var (
	// RetryOnAnyError retries on any error
	RetryOnAnyError RetryableFunc = func(err error) bool {
		return err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
	}

	// RetryOnNetworkError retries on network-related errors
	RetryOnNetworkError RetryableFunc = IsRetryableError

	// RetryOnTemporaryError retries on temporary errors
	RetryOnTemporaryError RetryableFunc = func(err error) bool {
		return err != nil && isTemporaryError(err)
	}
)

// WithMaxAttempts sets the maximum number of retry attempts
func (c *RetryConfig) WithMaxAttempts(attempts int) *RetryConfig {
	c.MaxAttempts = attempts
	return c
}

// WithBackoff sets the initial backoff and factor
func (c *RetryConfig) WithBackoff(initial time.Duration, factor float64) *RetryConfig {
	c.InitialBackoff = initial
	c.BackoffFactor = factor
	return c
}

// WithMaxBackoff sets the maximum backoff duration
func (c *RetryConfig) WithMaxBackoff(max time.Duration) *RetryConfig {
	c.MaxBackoff = max
	return c
}

// WithJitter enables or disables jitter
func (c *RetryConfig) WithJitter(enable bool) *RetryConfig {
	c.Jitter = enable
	return c
}

// WithRetryIf sets the condition for retrying
func (c *RetryConfig) WithRetryIf(fn func(error) bool) *RetryConfig {
	c.RetryIf = fn
	return c
}
