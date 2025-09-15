package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
)

// BehaviorManager manages rate limiting and retry behavior based on configuration
type BehaviorManager struct {
	rateLimiterManager *RateLimiterManager
	retryConfig        *RetryConfig
	logger             *Logger
}

// NewBehaviorManager creates a new behavior manager from configuration
func NewBehaviorManager(cfg *config.Config) *BehaviorManager {
	manager := &BehaviorManager{
		rateLimiterManager: NewRateLimiterManager(),
		logger:             GetGlobalLogger(),
	}

	// Set up retry configuration
	manager.retryConfig = configToRetryConfig(cfg.Behavior.Retry)

	// Set up rate limiters for each provider
	if cfg.Behavior.RateLimit.RequestsPerSecond > 0 {
		rateLimitConfig := &RateLimiterConfig{
			RequestsPerSecond: cfg.Behavior.RateLimit.RequestsPerSecond,
			Burst:             cfg.Behavior.RateLimit.Burst,
			Timeout:           cfg.Behavior.RateLimit.Timeout,
		}

		// Create rate limiters for known providers
		if len(cfg.Repositories) > 0 {
			for provider := range cfg.Repositories {
				providerConfig := *rateLimitConfig
				providerConfig.Name = provider
				manager.rateLimiterManager.GetOrCreateRateLimiter(provider, &providerConfig)
			}
		}

		// Create a default global rate limiter
		globalConfig := *rateLimitConfig
		globalConfig.Name = "global"
		manager.rateLimiterManager.GetOrCreateRateLimiter("global", &globalConfig)
	}

	return manager
}

// ExecuteWithBehavior executes a function with rate limiting and retry logic
func (bm *BehaviorManager) ExecuteWithBehavior(ctx context.Context, provider string, operation string, fn func() error) error {
	// Get or create rate limiter for this provider
	rateLimiter, _ := bm.rateLimiterManager.GetRateLimiter(provider)
	if rateLimiter == nil {
		// Fall back to global rate limiter
		rateLimiter, _ = bm.rateLimiterManager.GetRateLimiter("global")
	}

	// Wrap the function with rate limiting and retry
	return Retry(ctx, bm.retryConfig, func() error {
		return WaitWithRateLimiter(ctx, rateLimiter, func() error {
			bm.logger.Debugf("Executing %s operation for provider %s", operation, provider)
			return fn()
		})
	})
}

// ExecuteWithBehaviorAndResult executes a function with rate limiting and retry logic, returning a result
func ExecuteWithBehaviorAndResult[T any](ctx context.Context, bm *BehaviorManager, provider string, operation string, fn func() (T, error)) (T, error) {
	// Get or create rate limiter for this provider
	rateLimiter, _ := bm.rateLimiterManager.GetRateLimiter(provider)
	if rateLimiter == nil {
		// Fall back to global rate limiter
		rateLimiter, _ = bm.rateLimiterManager.GetRateLimiter("global")
	}

	// Wrap the function with rate limiting and retry
	return RetryWithResult(ctx, bm.retryConfig, func() (T, error) {
		return WaitWithRateLimiterAndResult(ctx, rateLimiter, func() (T, error) {
			bm.logger.Debugf("Executing %s operation for provider %s", operation, provider)
			return fn()
		})
	})
}

// UpdateBehaviorConfig updates the behavior configuration
func (bm *BehaviorManager) UpdateBehaviorConfig(cfg *config.Config) error {
	// Update retry configuration
	bm.retryConfig = configToRetryConfig(cfg.Behavior.Retry)

	// Update rate limiting configuration
	if cfg.Behavior.RateLimit.RequestsPerSecond > 0 {
		rateLimitConfig := &RateLimiterConfig{
			RequestsPerSecond: cfg.Behavior.RateLimit.RequestsPerSecond,
			Burst:             cfg.Behavior.RateLimit.Burst,
			Timeout:           cfg.Behavior.RateLimit.Timeout,
		}

		// Update rate limiters for each provider
		for provider := range cfg.Repositories {
			providerConfig := *rateLimitConfig
			providerConfig.Name = provider
			bm.rateLimiterManager.GetOrCreateRateLimiter(provider, &providerConfig)
		}

		// Update global rate limiter
		globalConfig := *rateLimitConfig
		globalConfig.Name = "global"
		bm.rateLimiterManager.GetOrCreateRateLimiter("global", &globalConfig)
	}

	return nil
}

// GetRateLimiterStats returns statistics for all rate limiters
func (bm *BehaviorManager) GetRateLimiterStats() map[string]RateLimiterStats {
	return bm.rateLimiterManager.GetAllStats()
}

// GetRetryConfig returns the current retry configuration
func (bm *BehaviorManager) GetRetryConfig() *RetryConfig {
	return bm.retryConfig
}

// ValidateBehaviorConfig validates behavior configuration
func ValidateBehaviorConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate rate limit configuration
	if cfg.Behavior.RateLimit.RequestsPerSecond < 0 {
		return fmt.Errorf("rate limit requests_per_second cannot be negative")
	}

	if cfg.Behavior.RateLimit.Burst < 0 {
		return fmt.Errorf("rate limit burst cannot be negative")
	}

	if cfg.Behavior.RateLimit.Timeout < 0 {
		return fmt.Errorf("rate limit timeout cannot be negative")
	}

	// Validate retry configuration
	if cfg.Behavior.Retry.MaxAttempts < 0 {
		return fmt.Errorf("retry max_attempts cannot be negative")
	}

	if cfg.Behavior.Retry.Backoff < 0 {
		return fmt.Errorf("retry backoff cannot be negative")
	}

	if cfg.Behavior.Retry.MaxBackoff < 0 {
		return fmt.Errorf("retry max_backoff cannot be negative")
	}

	if cfg.Behavior.Retry.MaxBackoff > 0 && cfg.Behavior.Retry.Backoff > cfg.Behavior.Retry.MaxBackoff {
		return fmt.Errorf("retry backoff cannot be greater than max_backoff")
	}

	return nil
}

// configToRetryConfig converts config.Retry to utils.RetryConfig
func configToRetryConfig(retryConfig config.Retry) *RetryConfig {
	// Set defaults if not specified
	maxAttempts := retryConfig.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	initialBackoff := retryConfig.Backoff
	if initialBackoff <= 0 {
		initialBackoff = time.Second
	}

	maxBackoff := retryConfig.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}

	return &RetryConfig{
		MaxAttempts:    maxAttempts,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
		BackoffFactor:  2.0,
		Jitter:         true,
		RetryIf:        IsRetryableError,
	}
}


// BehaviorStats contains statistics about behavior management
type BehaviorStats struct {
	RetryConfig      RetryConfigStats            `json:"retry_config"`
	RateLimiterStats map[string]RateLimiterStats `json:"rate_limiter_stats"`
	Concurrency      int                         `json:"concurrency"`
}

// RetryConfigStats contains statistics about retry configuration
type RetryConfigStats struct {
	MaxAttempts    int           `json:"max_attempts"`
	InitialBackoff time.Duration `json:"initial_backoff"`
	MaxBackoff     time.Duration `json:"max_backoff"`
	BackoffFactor  float64       `json:"backoff_factor"`
	Jitter         bool          `json:"jitter"`
}

// GetBehaviorStats returns statistics about the current behavior configuration
func (bm *BehaviorManager) GetBehaviorStats() BehaviorStats {
	retryStats := RetryConfigStats{
		MaxAttempts:    bm.retryConfig.MaxAttempts,
		InitialBackoff: bm.retryConfig.InitialBackoff,
		MaxBackoff:     bm.retryConfig.MaxBackoff,
		BackoffFactor:  bm.retryConfig.BackoffFactor,
		Jitter:         bm.retryConfig.Jitter,
	}

	return BehaviorStats{
		RetryConfig:      retryStats,
		RateLimiterStats: bm.GetRateLimiterStats(),
		Concurrency:      0, // This would be set from the main config
	}
}
