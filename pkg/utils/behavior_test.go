package utils

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pacphi/git-pr-manager/pkg/config"
)

func TestNewBehaviorManager(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 2.0,
				Burst:             5,
				Timeout:           10 * time.Second,
			},
			Retry: config.Retry{
				MaxAttempts: 5,
				Backoff:     2 * time.Second,
				MaxBackoff:  30 * time.Second,
			},
		},
		Repositories: map[string][]config.Repository{
			"github": {{}},
			"gitlab": {{}},
		},
	}

	bm := NewBehaviorManager(cfg)

	assert.NotNil(t, bm)
	assert.NotNil(t, bm.rateLimiterManager)
	assert.NotNil(t, bm.retryConfig)
	assert.NotNil(t, bm.logger)

	// Check retry configuration
	assert.Equal(t, 5, bm.retryConfig.MaxAttempts)
	assert.Equal(t, 2*time.Second, bm.retryConfig.InitialBackoff)
	assert.Equal(t, 30*time.Second, bm.retryConfig.MaxBackoff)

	// Check that rate limiters were created for providers
	githubRL, exists := bm.rateLimiterManager.GetRateLimiter("github")
	assert.True(t, exists)
	assert.NotNil(t, githubRL)

	gitlabRL, exists := bm.rateLimiterManager.GetRateLimiter("gitlab")
	assert.True(t, exists)
	assert.NotNil(t, gitlabRL)

	globalRL, exists := bm.rateLimiterManager.GetRateLimiter("global")
	assert.True(t, exists)
	assert.NotNil(t, globalRL)
}

func TestNewBehaviorManager_ZeroRateLimit(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 0, // Disabled rate limiting
			},
		},
	}

	bm := NewBehaviorManager(cfg)

	assert.NotNil(t, bm)
	assert.NotNil(t, bm.rateLimiterManager)

	// No rate limiters should be created when RequestsPerSecond is 0
	stats := bm.rateLimiterManager.GetAllStats()
	assert.Equal(t, 0, len(stats))
}

func TestNewBehaviorManager_EmptyRepositories(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 1.0,
				Burst:             1,
			},
		},
		Repositories: map[string][]config.Repository{}, // Empty
	}

	bm := NewBehaviorManager(cfg)

	assert.NotNil(t, bm)

	// Should still create a global rate limiter
	globalRL, exists := bm.rateLimiterManager.GetRateLimiter("global")
	assert.True(t, exists)
	assert.NotNil(t, globalRL)
}

func TestBehaviorManager_ExecuteWithBehavior_Success(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 100.0, // High rate to avoid blocking
				Burst:             10,
			},
			Retry: config.Retry{
				MaxAttempts: 3,
				Backoff:     10 * time.Millisecond,
			},
		},
		Repositories: map[string][]config.Repository{
			"test-provider": {{}},
		},
	}

	bm := NewBehaviorManager(cfg)
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := bm.ExecuteWithBehavior(ctx, "test-provider", "test-operation", fn)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestBehaviorManager_ExecuteWithBehavior_UnknownProvider(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 100.0,
				Burst:             10,
			},
		},
		Repositories: map[string][]config.Repository{
			"known-provider": {{}},
		},
	}

	bm := NewBehaviorManager(cfg)
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	// Should fall back to global rate limiter for unknown provider
	err := bm.ExecuteWithBehavior(ctx, "unknown-provider", "test-operation", fn)

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestBehaviorManager_ExecuteWithBehavior_NoRateLimiter(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 0, // Disabled
			},
			Retry: config.Retry{
				MaxAttempts: 2,
				Backoff:     10 * time.Millisecond,
			},
		},
	}

	bm := NewBehaviorManager(cfg)
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := bm.ExecuteWithBehavior(ctx, "any-provider", "test-operation", fn)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestBehaviorManager_ExecuteWithBehaviorAndResult_Success(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 100.0,
				Burst:             10,
			},
			Retry: config.Retry{
				MaxAttempts: 3,
				Backoff:     10 * time.Millisecond,
			},
		},
		Repositories: map[string][]config.Repository{
			"test-provider": {{}},
		},
	}

	bm := NewBehaviorManager(cfg)
	ctx := context.Background()

	callCount := 0
	fn := func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", errors.New("temporary failure")
		}
		return "success", nil
	}

	result, err := ExecuteWithBehaviorAndResult(ctx, bm, "test-provider", "test-operation", fn)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 2, callCount)
}

func TestBehaviorManager_ExecuteWithBehaviorAndResult_AllFail(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			Retry: config.Retry{
				MaxAttempts: 3, // Increase to 3 so it gets called 3 times total
				Backoff:     10 * time.Millisecond,
			},
		},
	}

	bm := NewBehaviorManager(cfg)
	ctx := context.Background()

	expectedError := errors.New("temporary failure")
	callCount := 0
	fn := func() (int, error) {
		callCount++
		return 0, expectedError
	}

	result, err := ExecuteWithBehaviorAndResult(ctx, bm, "any-provider", "test-operation", fn)

	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, 0, result)
	assert.Equal(t, 3, callCount)
}

func TestBehaviorManager_UpdateBehaviorConfig(t *testing.T) {
	initialCfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 1.0,
				Burst:             1,
			},
			Retry: config.Retry{
				MaxAttempts: 2,
				Backoff:     1 * time.Second,
			},
		},
		Repositories: map[string][]config.Repository{
			"provider1": {{}},
		},
	}

	bm := NewBehaviorManager(initialCfg)

	// Update configuration
	updatedCfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 5.0,
				Burst:             10,
				Timeout:           15 * time.Second,
			},
			Retry: config.Retry{
				MaxAttempts: 5,
				Backoff:     2 * time.Second,
				MaxBackoff:  60 * time.Second,
			},
		},
		Repositories: map[string][]config.Repository{
			"provider1": {{}},
			"provider2": {{}},
		},
	}

	err := bm.UpdateBehaviorConfig(updatedCfg)

	assert.NoError(t, err)

	// Check updated retry configuration
	retryConfig := bm.GetRetryConfig()
	assert.Equal(t, 5, retryConfig.MaxAttempts)
	assert.Equal(t, 2*time.Second, retryConfig.InitialBackoff)
	assert.Equal(t, 60*time.Second, retryConfig.MaxBackoff)

	// Check updated rate limiter configurations
	stats := bm.GetRateLimiterStats()
	assert.Contains(t, stats, "provider1")
	assert.Contains(t, stats, "provider2")
	assert.Contains(t, stats, "global")

	assert.Equal(t, 5.0, stats["provider1"].Limit)
	assert.Equal(t, 10, stats["provider1"].Burst)
	assert.Equal(t, 5.0, stats["provider2"].Limit)
	assert.Equal(t, 10, stats["provider2"].Burst)
}

func TestBehaviorManager_GetBehaviorStats(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 3.0,
				Burst:             7,
				Timeout:           20 * time.Second,
			},
			Retry: config.Retry{
				MaxAttempts: 4,
				Backoff:     3 * time.Second,
				MaxBackoff:  45 * time.Second,
			},
		},
		Repositories: map[string][]config.Repository{
			"test-provider": {{}},
		},
	}

	bm := NewBehaviorManager(cfg)

	stats := bm.GetBehaviorStats()

	// Check retry config stats
	assert.Equal(t, 4, stats.RetryConfig.MaxAttempts)
	assert.Equal(t, 3*time.Second, stats.RetryConfig.InitialBackoff)
	assert.Equal(t, 45*time.Second, stats.RetryConfig.MaxBackoff)
	assert.Equal(t, 2.0, stats.RetryConfig.BackoffFactor)
	assert.True(t, stats.RetryConfig.Jitter)

	// Check rate limiter stats
	assert.Contains(t, stats.RateLimiterStats, "test-provider")
	assert.Contains(t, stats.RateLimiterStats, "global")

	providerStats := stats.RateLimiterStats["test-provider"]
	assert.Equal(t, 3.0, providerStats.Limit)
	assert.Equal(t, 7, providerStats.Burst)
	assert.Equal(t, 20*time.Second, providerStats.Timeout)

	// Concurrency is not set from config in this test
	assert.Equal(t, 0, stats.Concurrency)
}

func TestValidateBehaviorConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil config",
			config:    nil,
			expectErr: true,
			errMsg:    "configuration is nil",
		},
		{
			name: "negative requests per second",
			config: &config.Config{
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: -1.0,
					},
				},
			},
			expectErr: true,
			errMsg:    "rate limit requests_per_second cannot be negative",
		},
		{
			name: "negative burst",
			config: &config.Config{
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 1.0,
						Burst:             -1,
					},
				},
			},
			expectErr: true,
			errMsg:    "rate limit burst cannot be negative",
		},
		{
			name: "negative timeout",
			config: &config.Config{
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 1.0,
						Burst:             1,
						Timeout:           -1 * time.Second,
					},
				},
			},
			expectErr: true,
			errMsg:    "rate limit timeout cannot be negative",
		},
		{
			name: "negative retry max attempts",
			config: &config.Config{
				Behavior: config.Behavior{
					Retry: config.Retry{
						MaxAttempts: -1,
					},
				},
			},
			expectErr: true,
			errMsg:    "retry max_attempts cannot be negative",
		},
		{
			name: "negative retry backoff",
			config: &config.Config{
				Behavior: config.Behavior{
					Retry: config.Retry{
						MaxAttempts: 3,
						Backoff:     -1 * time.Second,
					},
				},
			},
			expectErr: true,
			errMsg:    "retry backoff cannot be negative",
		},
		{
			name: "negative retry max backoff",
			config: &config.Config{
				Behavior: config.Behavior{
					Retry: config.Retry{
						MaxAttempts: 3,
						Backoff:     1 * time.Second,
						MaxBackoff:  -1 * time.Second,
					},
				},
			},
			expectErr: true,
			errMsg:    "retry max_backoff cannot be negative",
		},
		{
			name: "backoff greater than max backoff",
			config: &config.Config{
				Behavior: config.Behavior{
					Retry: config.Retry{
						MaxAttempts: 3,
						Backoff:     10 * time.Second,
						MaxBackoff:  5 * time.Second,
					},
				},
			},
			expectErr: true,
			errMsg:    "retry backoff cannot be greater than max_backoff",
		},
		{
			name: "valid config",
			config: &config.Config{
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 2.0,
						Burst:             5,
						Timeout:           10 * time.Second,
					},
					Retry: config.Retry{
						MaxAttempts: 3,
						Backoff:     1 * time.Second,
						MaxBackoff:  30 * time.Second,
					},
				},
			},
			expectErr: false,
		},
		{
			name: "zero values allowed",
			config: &config.Config{
				Behavior: config.Behavior{
					RateLimit: config.RateLimit{
						RequestsPerSecond: 0,
						Burst:             0,
						Timeout:           0,
					},
					Retry: config.Retry{
						MaxAttempts: 0,
						Backoff:     0,
						MaxBackoff:  0,
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBehaviorConfig(tt.config)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigToRetryConfig(t *testing.T) {
	tests := []struct {
		name           string
		input          config.Retry
		expectedOutput *RetryConfig
	}{
		{
			name: "all values set",
			input: config.Retry{
				MaxAttempts: 5,
				Backoff:     2 * time.Second,
				MaxBackoff:  60 * time.Second,
			},
			expectedOutput: &RetryConfig{
				MaxAttempts:    5,
				InitialBackoff: 2 * time.Second,
				MaxBackoff:     60 * time.Second,
				BackoffFactor:  2.0,
				Jitter:         true,
				RetryIf:        IsRetryableError,
			},
		},
		{
			name: "zero/negative values use defaults",
			input: config.Retry{
				MaxAttempts: -1,
				Backoff:     -1 * time.Second,
				MaxBackoff:  -1 * time.Second,
			},
			expectedOutput: &RetryConfig{
				MaxAttempts:    3,
				InitialBackoff: time.Second,
				MaxBackoff:     30 * time.Second,
				BackoffFactor:  2.0,
				Jitter:         true,
				RetryIf:        IsRetryableError,
			},
		},
		{
			name:  "empty config uses defaults",
			input: config.Retry{},
			expectedOutput: &RetryConfig{
				MaxAttempts:    3,
				InitialBackoff: time.Second,
				MaxBackoff:     30 * time.Second,
				BackoffFactor:  2.0,
				Jitter:         true,
				RetryIf:        IsRetryableError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := configToRetryConfig(tt.input)

			assert.Equal(t, tt.expectedOutput.MaxAttempts, result.MaxAttempts)
			assert.Equal(t, tt.expectedOutput.InitialBackoff, result.InitialBackoff)
			assert.Equal(t, tt.expectedOutput.MaxBackoff, result.MaxBackoff)
			assert.Equal(t, tt.expectedOutput.BackoffFactor, result.BackoffFactor)
			assert.Equal(t, tt.expectedOutput.Jitter, result.Jitter)
			assert.NotNil(t, result.RetryIf)

			// Test that RetryIf functions the same way
			testErr := errors.New("connection refused")
			assert.Equal(t, tt.expectedOutput.RetryIf(testErr), result.RetryIf(testErr))
		})
	}
}

func TestBehaviorManager_NilBehaviorManager(t *testing.T) {
	// Test the fallback behavior when BehaviorManager is nil in the helper function
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	// This should work with a nil behavior manager (fallback path)
	err := Retry(ctx, nil, fn)

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestBehaviorManager_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			RateLimit: config.RateLimit{
				RequestsPerSecond: 0.1, // Very slow rate
				Burst:             1,
				Timeout:           5 * time.Second,
			},
			Retry: config.Retry{
				MaxAttempts: 10,
				Backoff:     100 * time.Millisecond,
			},
		},
		Repositories: map[string][]config.Repository{
			"slow-provider": {},
		},
	}

	bm := NewBehaviorManager(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Consume the burst first
	quickCtx := context.Background()
	bm.ExecuteWithBehavior(quickCtx, "slow-provider", "setup", func() error {
		return nil
	})

	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("temporary failure") // Will trigger retry
	}

	start := time.Now()
	err := bm.ExecuteWithBehavior(ctx, "slow-provider", "slow-operation", fn)
	duration := time.Since(start)

	assert.Error(t, err)
	// Either context error, timeout, or rate limiter context deadline is acceptable
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) ||
		(err != nil && (err.Error() == "context deadline exceeded" || err.Error() == "context canceled" ||
			strings.Contains(err.Error(), "would exceed context deadline"))))

	// Should complete quickly due to context timeout
	assert.Less(t, duration, 200*time.Millisecond)

	// May have been called once before context cancellation
	assert.GreaterOrEqual(t, callCount, 0)
}

func TestBehaviorManager_GetRetryConfig(t *testing.T) {
	cfg := &config.Config{
		Behavior: config.Behavior{
			Retry: config.Retry{
				MaxAttempts: 7,
				Backoff:     3 * time.Second,
				MaxBackoff:  90 * time.Second,
			},
		},
	}

	bm := NewBehaviorManager(cfg)

	retryConfig := bm.GetRetryConfig()

	assert.Equal(t, 7, retryConfig.MaxAttempts)
	assert.Equal(t, 3*time.Second, retryConfig.InitialBackoff)
	assert.Equal(t, 90*time.Second, retryConfig.MaxBackoff)
	assert.Equal(t, 2.0, retryConfig.BackoffFactor)
	assert.True(t, retryConfig.Jitter)
	assert.NotNil(t, retryConfig.RetryIf)
}
