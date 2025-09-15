package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for API calls
type RateLimiter struct {
	limiter *rate.Limiter
	timeout time.Duration
	name    string
	mu      sync.RWMutex
}

// RateLimiterConfig contains configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond float64       `json:"requests_per_second"`
	Burst             int           `json:"burst"`
	Timeout           time.Duration `json:"timeout"`
	Name              string        `json:"name,omitempty"`
}

// DefaultRateLimiterConfig returns a default rate limiter configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerSecond: 1.0, // 1 request per second
		Burst:             5,   // Allow bursts of up to 5 requests
		Timeout:           30 * time.Second,
		Name:              "default",
	}
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	// Ensure minimum values
	if config.RequestsPerSecond <= 0 {
		config.RequestsPerSecond = 1.0
	}
	if config.Burst <= 0 {
		config.Burst = 1
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)

	return &RateLimiter{
		limiter: limiter,
		timeout: config.Timeout,
		name:    config.Name,
	}
}

// Wait blocks until the rate limiter allows a request or the context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	logger := GetGlobalLogger()

	// Create a context with timeout if no deadline is set
	waitCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && rl.timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, rl.timeout)
		defer cancel()
	}

	// Wait for permission to proceed
	logger.Debugf("Rate limiter %s: waiting for permission", rl.name)

	start := time.Now()
	if err := rl.limiter.Wait(waitCtx); err != nil {
		if err == context.DeadlineExceeded {
			return fmt.Errorf("rate limiter %s: timeout after %v", rl.name, time.Since(start))
		}
		return fmt.Errorf("rate limiter %s: %w", rl.name, err)
	}

	waitDuration := time.Since(start)
	if waitDuration > time.Millisecond {
		logger.Debugf("Rate limiter %s: waited %v for permission", rl.name, waitDuration)
	}

	return nil
}

// Allow checks if a request is allowed without blocking
func (rl *RateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.limiter.Allow()
}

// Reserve returns a reservation for a request
func (rl *RateLimiter) Reserve() *rate.Reservation {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.limiter.Reserve()
}

// UpdateConfig updates the rate limiter configuration
func (rl *RateLimiter) UpdateConfig(config *RateLimiterConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if config == nil {
		return
	}

	// Update rate limit if specified
	if config.RequestsPerSecond > 0 {
		rl.limiter.SetLimit(rate.Limit(config.RequestsPerSecond))
	}

	// Update burst if specified
	if config.Burst > 0 {
		rl.limiter.SetBurst(config.Burst)
	}

	// Update timeout if specified
	if config.Timeout > 0 {
		rl.timeout = config.Timeout
	}

	// Update name if specified
	if config.Name != "" {
		rl.name = config.Name
	}
}

// GetStats returns current statistics about the rate limiter
func (rl *RateLimiter) GetStats() RateLimiterStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return RateLimiterStats{
		Name:    rl.name,
		Limit:   float64(rl.limiter.Limit()),
		Burst:   rl.limiter.Burst(),
		Tokens:  int(rl.limiter.Tokens()),
		Timeout: rl.timeout,
	}
}

// RateLimiterStats contains statistics about the rate limiter
type RateLimiterStats struct {
	Name    string        `json:"name"`
	Limit   float64       `json:"limit"`
	Burst   int           `json:"burst"`
	Tokens  int           `json:"tokens"`
	Timeout time.Duration `json:"timeout"`
}

// String returns a string representation of the rate limiter stats
func (s RateLimiterStats) String() string {
	return fmt.Sprintf("RateLimit[%s]: %.2f req/s, burst=%d, tokens=%d, timeout=%v",
		s.Name, s.Limit, s.Burst, s.Tokens, s.Timeout)
}

// RateLimiterManager manages multiple rate limiters for different providers or endpoints
type RateLimiterManager struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewRateLimiterManager creates a new rate limiter manager
func NewRateLimiterManager() *RateLimiterManager {
	return &RateLimiterManager{
		limiters: make(map[string]*RateLimiter),
	}
}

// GetOrCreateRateLimiter gets an existing rate limiter or creates a new one
func (rlm *RateLimiterManager) GetOrCreateRateLimiter(name string, config *RateLimiterConfig) *RateLimiter {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if rl, exists := rlm.limiters[name]; exists {
		// Update existing limiter if config is provided
		if config != nil {
			rl.UpdateConfig(config)
		}
		return rl
	}

	// Create new rate limiter
	if config == nil {
		config = DefaultRateLimiterConfig()
	}
	config.Name = name

	rl := NewRateLimiter(config)
	rlm.limiters[name] = rl
	return rl
}

// GetRateLimiter gets an existing rate limiter by name
func (rlm *RateLimiterManager) GetRateLimiter(name string) (*RateLimiter, bool) {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	rl, exists := rlm.limiters[name]
	return rl, exists
}

// UpdateRateLimiter updates the configuration of an existing rate limiter
func (rlm *RateLimiterManager) UpdateRateLimiter(name string, config *RateLimiterConfig) error {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	rl, exists := rlm.limiters[name]
	if !exists {
		return fmt.Errorf("rate limiter %s not found", name)
	}

	rl.UpdateConfig(config)
	return nil
}

// RemoveRateLimiter removes a rate limiter
func (rlm *RateLimiterManager) RemoveRateLimiter(name string) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	delete(rlm.limiters, name)
}

// GetAllStats returns statistics for all managed rate limiters
func (rlm *RateLimiterManager) GetAllStats() map[string]RateLimiterStats {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	stats := make(map[string]RateLimiterStats)
	for name, rl := range rlm.limiters {
		stats[name] = rl.GetStats()
	}

	return stats
}

// WaitWithRateLimiter wraps a function call with rate limiting
func WaitWithRateLimiter(ctx context.Context, rl *RateLimiter, fn func() error) error {
	if rl == nil {
		return fn()
	}

	if err := rl.Wait(ctx); err != nil {
		return err
	}

	return fn()
}

// WaitWithRateLimiterAndResult wraps a function call with rate limiting and returns a result
func WaitWithRateLimiterAndResult[T any](ctx context.Context, rl *RateLimiter, fn func() (T, error)) (T, error) {
	var zero T

	if rl == nil {
		return fn()
	}

	if err := rl.Wait(ctx); err != nil {
		return zero, err
	}

	return fn()
}
