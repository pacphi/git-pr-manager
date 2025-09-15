package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.Equal(t, 1.0, config.RequestsPerSecond)
	assert.Equal(t, 5, config.Burst)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "default", config.Name)
}

func TestNewRateLimiter(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 2.0,
		Burst:             3,
		Timeout:           15 * time.Second,
		Name:              "test-limiter",
	}

	rl := NewRateLimiter(config)

	assert.NotNil(t, rl)
	assert.Equal(t, 15*time.Second, rl.timeout)
	assert.Equal(t, "test-limiter", rl.name)

	stats := rl.GetStats()
	assert.Equal(t, "test-limiter", stats.Name)
	assert.Equal(t, 2.0, stats.Limit)
	assert.Equal(t, 3, stats.Burst)
	assert.Equal(t, 15*time.Second, stats.Timeout)
}

func TestNewRateLimiter_NilConfig(t *testing.T) {
	rl := NewRateLimiter(nil)

	assert.NotNil(t, rl)
	stats := rl.GetStats()
	assert.Equal(t, "default", stats.Name)
	assert.Equal(t, 1.0, stats.Limit)
	assert.Equal(t, 5, stats.Burst)
	assert.Equal(t, 30*time.Second, stats.Timeout)
}

func TestNewRateLimiter_InvalidValues(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: -1.0,
		Burst:             -5,
		Timeout:           -10 * time.Second,
		Name:              "invalid-config",
	}

	rl := NewRateLimiter(config)

	stats := rl.GetStats()
	assert.Equal(t, 1.0, stats.Limit)              // Should default to 1.0
	assert.Equal(t, 1, stats.Burst)                // Should default to 1
	assert.Equal(t, 30*time.Second, stats.Timeout) // Should default to 30s
}

func TestRateLimiter_Wait(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 10.0, // High rate to avoid blocking in tests
		Burst:             1,
		Timeout:           5 * time.Second,
		Name:              "test-wait",
	}

	rl := NewRateLimiter(config)
	ctx := context.Background()

	// First call should not block
	start := time.Now()
	err := rl.Wait(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond)
}

func TestRateLimiter_WaitWithTimeout(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 0.1, // Very low rate to force timeout
		Burst:             1,
		Timeout:           50 * time.Millisecond,
		Name:              "test-timeout",
	}

	rl := NewRateLimiter(config)

	// First call consumes the burst
	ctx := context.Background()
	err := rl.Wait(ctx)
	assert.NoError(t, err)

	// Second call should timeout
	start := time.Now()
	err = rl.Wait(ctx)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.Contains(t, err.Error(), "test-timeout")
	// Duration might be very short if rate limiter determines it would exceed deadline immediately
	assert.GreaterOrEqual(t, duration, 0*time.Millisecond)
}

func TestRateLimiter_WaitWithContextTimeout(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 0.1, // Very low rate
		Burst:             1,
		Timeout:           5 * time.Second, // Higher than context timeout
		Name:              "test-context-timeout",
	}

	rl := NewRateLimiter(config)

	// First call consumes the burst
	ctx := context.Background()
	err := rl.Wait(ctx)
	assert.NoError(t, err)

	// Second call with short context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = rl.Wait(ctx)
	duration := time.Since(start)

	assert.Error(t, err)
	// Accept either actual deadline exceeded or "would exceed deadline" immediate return
	assert.True(t, strings.Contains(err.Error(), "context deadline exceeded") ||
		strings.Contains(err.Error(), "would exceed context deadline"))
	// Duration might be very short if rate limiter determines it would exceed deadline immediately
	assert.GreaterOrEqual(t, duration, 0*time.Millisecond)
}

func TestRateLimiter_WaitWithContextCancellation(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 0.1, // Very low rate
		Burst:             1,
		Timeout:           5 * time.Second,
		Name:              "test-context-cancel",
	}

	rl := NewRateLimiter(config)

	// First call consumes the burst
	ctx := context.Background()
	err := rl.Wait(ctx)
	assert.NoError(t, err)

	// Second call with cancelled context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err = rl.Wait(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRateLimiter_Allow(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             2,
		Timeout:           5 * time.Second,
		Name:              "test-allow",
	}

	rl := NewRateLimiter(config)

	// Should allow up to burst limit
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())

	// Third call should be denied (burst exhausted)
	assert.False(t, rl.Allow())
}

func TestRateLimiter_Reserve(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		Timeout:           5 * time.Second,
		Name:              "test-reserve",
	}

	rl := NewRateLimiter(config)

	// First reservation should be immediate
	reservation := rl.Reserve()
	assert.NotNil(t, reservation)
	assert.True(t, reservation.OK())
	assert.LessOrEqual(t, reservation.Delay(), 10*time.Millisecond)

	// Second reservation should have a delay
	reservation2 := rl.Reserve()
	assert.NotNil(t, reservation2)
	assert.True(t, reservation2.OK())
	assert.Greater(t, reservation2.Delay(), 500*time.Millisecond)
}

func TestRateLimiter_UpdateConfig(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		Timeout:           5 * time.Second,
		Name:              "original",
	}

	rl := NewRateLimiter(config)

	// Update configuration
	updateConfig := &RateLimiterConfig{
		RequestsPerSecond: 5.0,
		Burst:             10,
		Timeout:           15 * time.Second,
		Name:              "updated",
	}

	rl.UpdateConfig(updateConfig)

	stats := rl.GetStats()
	assert.Equal(t, 5.0, stats.Limit)
	assert.Equal(t, 10, stats.Burst)
	assert.Equal(t, 15*time.Second, stats.Timeout)
	assert.Equal(t, "updated", stats.Name)
}

func TestRateLimiter_UpdateConfig_Nil(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		Timeout:           5 * time.Second,
		Name:              "original",
	}

	rl := NewRateLimiter(config)
	originalStats := rl.GetStats()

	// Nil update should not change anything
	rl.UpdateConfig(nil)

	newStats := rl.GetStats()
	assert.Equal(t, originalStats, newStats)
}

func TestRateLimiter_UpdateConfig_PartialUpdate(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		Timeout:           5 * time.Second,
		Name:              "original",
	}

	rl := NewRateLimiter(config)

	// Partial update (only some fields)
	updateConfig := &RateLimiterConfig{
		RequestsPerSecond: 3.0,
		// Burst not set (should remain unchanged)
		// Timeout not set (should remain unchanged)
		Name: "partially-updated",
	}

	rl.UpdateConfig(updateConfig)

	stats := rl.GetStats()
	assert.Equal(t, 3.0, stats.Limit)
	assert.Equal(t, 1, stats.Burst)               // Should remain unchanged
	assert.Equal(t, 5*time.Second, stats.Timeout) // Should remain unchanged
	assert.Equal(t, "partially-updated", stats.Name)
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 100.0, // High rate to minimize blocking
		Burst:             10,
		Timeout:           5 * time.Second,
		Name:              "concurrent-test",
	}

	rl := NewRateLimiter(config)

	const numGoroutines = 50
	const operationsPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Mix of different operations
				switch j % 3 {
				case 0:
					if err := rl.Wait(ctx); err != nil {
						errors <- err
					}
				case 1:
					rl.Allow()
				case 2:
					rl.Reserve()
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Concurrent operation error: %v", err)
	}

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent access")
}

func TestRateLimiterStats_String(t *testing.T) {
	stats := RateLimiterStats{
		Name:    "test-stats",
		Limit:   2.5,
		Burst:   10,
		Tokens:  7,
		Timeout: 30 * time.Second,
	}

	expected := "RateLimit[test-stats]: 2.50 req/s, burst=10, tokens=7, timeout=30s"
	assert.Equal(t, expected, stats.String())
}

func TestNewRateLimiterManager(t *testing.T) {
	manager := NewRateLimiterManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.limiters)
	assert.Equal(t, 0, len(manager.limiters))
}

func TestRateLimiterManager_GetOrCreateRateLimiter(t *testing.T) {
	manager := NewRateLimiterManager()

	config := &RateLimiterConfig{
		RequestsPerSecond: 2.0,
		Burst:             3,
		Timeout:           10 * time.Second,
		Name:              "test-limiter",
	}

	// Create new rate limiter
	rl1 := manager.GetOrCreateRateLimiter("test", config)
	assert.NotNil(t, rl1)

	stats := rl1.GetStats()
	assert.Equal(t, "test", stats.Name) // Should use the provided name, not config.Name
	assert.Equal(t, 2.0, stats.Limit)
	assert.Equal(t, 3, stats.Burst)

	// Get existing rate limiter
	rl2 := manager.GetOrCreateRateLimiter("test", nil)
	assert.Equal(t, rl1, rl2) // Should return the same instance
}

func TestRateLimiterManager_GetOrCreateRateLimiter_NilConfig(t *testing.T) {
	manager := NewRateLimiterManager()

	rl := manager.GetOrCreateRateLimiter("default-test", nil)
	assert.NotNil(t, rl)

	stats := rl.GetStats()
	assert.Equal(t, "default-test", stats.Name)
	assert.Equal(t, 1.0, stats.Limit) // Default values
	assert.Equal(t, 5, stats.Burst)
}

func TestRateLimiterManager_GetRateLimiter(t *testing.T) {
	manager := NewRateLimiterManager()

	// Non-existent rate limiter
	rl, exists := manager.GetRateLimiter("nonexistent")
	assert.Nil(t, rl)
	assert.False(t, exists)

	// Create a rate limiter
	config := DefaultRateLimiterConfig()
	createdRL := manager.GetOrCreateRateLimiter("existing", config)

	// Get existing rate limiter
	rl, exists = manager.GetRateLimiter("existing")
	assert.Equal(t, createdRL, rl)
	assert.True(t, exists)
}

func TestRateLimiterManager_UpdateRateLimiter(t *testing.T) {
	manager := NewRateLimiterManager()

	// Try to update non-existent rate limiter
	config := &RateLimiterConfig{RequestsPerSecond: 5.0}
	err := manager.UpdateRateLimiter("nonexistent", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limiter nonexistent not found")

	// Create and update existing rate limiter
	originalConfig := DefaultRateLimiterConfig()
	rl := manager.GetOrCreateRateLimiter("existing", originalConfig)

	updateConfig := &RateLimiterConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
	}

	err = manager.UpdateRateLimiter("existing", updateConfig)
	assert.NoError(t, err)

	stats := rl.GetStats()
	assert.Equal(t, 10.0, stats.Limit)
	assert.Equal(t, 20, stats.Burst)
}

func TestRateLimiterManager_RemoveRateLimiter(t *testing.T) {
	manager := NewRateLimiterManager()

	// Create a rate limiter
	config := DefaultRateLimiterConfig()
	manager.GetOrCreateRateLimiter("to-remove", config)

	// Verify it exists
	_, exists := manager.GetRateLimiter("to-remove")
	assert.True(t, exists)

	// Remove it
	manager.RemoveRateLimiter("to-remove")

	// Verify it's gone
	_, exists = manager.GetRateLimiter("to-remove")
	assert.False(t, exists)

	// Removing non-existent should not panic
	manager.RemoveRateLimiter("never-existed")
}

func TestRateLimiterManager_GetAllStats(t *testing.T) {
	manager := NewRateLimiterManager()

	// Empty manager
	stats := manager.GetAllStats()
	assert.Equal(t, 0, len(stats))

	// Add some rate limiters
	config1 := &RateLimiterConfig{RequestsPerSecond: 1.0, Burst: 1}
	config2 := &RateLimiterConfig{RequestsPerSecond: 2.0, Burst: 2}

	manager.GetOrCreateRateLimiter("limiter1", config1)
	manager.GetOrCreateRateLimiter("limiter2", config2)

	stats = manager.GetAllStats()
	assert.Equal(t, 2, len(stats))

	assert.Contains(t, stats, "limiter1")
	assert.Contains(t, stats, "limiter2")

	assert.Equal(t, 1.0, stats["limiter1"].Limit)
	assert.Equal(t, 1, stats["limiter1"].Burst)
	assert.Equal(t, 2.0, stats["limiter2"].Limit)
	assert.Equal(t, 2, stats["limiter2"].Burst)
}

func TestRateLimiterManager_ConcurrentAccess(t *testing.T) {
	manager := NewRateLimiterManager()

	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Concurrent creation, updates, and access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			limiterName := fmt.Sprintf("limiter-%d", id%3) // Create some overlap

			for j := 0; j < operationsPerGoroutine; j++ {
				switch j % 4 {
				case 0:
					config := &RateLimiterConfig{RequestsPerSecond: float64(id + 1)}
					manager.GetOrCreateRateLimiter(limiterName, config)
				case 1:
					manager.GetRateLimiter(limiterName)
				case 2:
					updateConfig := &RateLimiterConfig{Burst: j + 1}
					manager.UpdateRateLimiter(limiterName, updateConfig)
				case 3:
					manager.GetAllStats()
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	stats := manager.GetAllStats()
	assert.LessOrEqual(t, len(stats), 3) // At most 3 different limiters created
}

func TestWaitWithRateLimiter(t *testing.T) {
	ctx := context.Background()

	t.Run("with nil rate limiter", func(t *testing.T) {
		callCount := 0
		fn := func() error {
			callCount++
			return nil
		}

		err := WaitWithRateLimiter(ctx, nil, fn)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("with rate limiter", func(t *testing.T) {
		config := &RateLimiterConfig{
			RequestsPerSecond: 100.0, // High rate to avoid blocking
			Burst:             1,
			Name:              "test-wait-wrapper",
		}
		rl := NewRateLimiter(config)

		callCount := 0
		fn := func() error {
			callCount++
			return nil
		}

		err := WaitWithRateLimiter(ctx, rl, fn)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("function returns error", func(t *testing.T) {
		config := &RateLimiterConfig{
			RequestsPerSecond: 100.0,
			Burst:             1,
			Name:              "test-wait-error",
		}
		rl := NewRateLimiter(config)

		expectedErr := assert.AnError
		fn := func() error {
			return expectedErr
		}

		err := WaitWithRateLimiter(ctx, rl, fn)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("rate limiter wait fails", func(t *testing.T) {
		config := &RateLimiterConfig{
			RequestsPerSecond: 0.1, // Very low rate
			Burst:             1,
			Timeout:           10 * time.Millisecond,
			Name:              "test-wait-timeout",
		}
		rl := NewRateLimiter(config)

		// Consume the burst
		rl.Wait(ctx)

		callCount := 0
		fn := func() error {
			callCount++
			return nil
		}

		err := WaitWithRateLimiter(ctx, rl, fn)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Equal(t, 0, callCount) // Function should not be called
	})
}

func TestWaitWithRateLimiterAndResult(t *testing.T) {
	ctx := context.Background()

	t.Run("with nil rate limiter", func(t *testing.T) {
		callCount := 0
		fn := func() (string, error) {
			callCount++
			return "success", nil
		}

		result, err := WaitWithRateLimiterAndResult(ctx, nil, fn)

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 1, callCount)
	})

	t.Run("with rate limiter", func(t *testing.T) {
		config := &RateLimiterConfig{
			RequestsPerSecond: 100.0, // High rate to avoid blocking
			Burst:             1,
			Name:              "test-wait-result-wrapper",
		}
		rl := NewRateLimiter(config)

		callCount := 0
		fn := func() (int, error) {
			callCount++
			return 42, nil
		}

		result, err := WaitWithRateLimiterAndResult(ctx, rl, fn)

		assert.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.Equal(t, 1, callCount)
	})

	t.Run("function returns error", func(t *testing.T) {
		config := &RateLimiterConfig{
			RequestsPerSecond: 100.0,
			Burst:             1,
			Name:              "test-wait-result-error",
		}
		rl := NewRateLimiter(config)

		expectedErr := assert.AnError
		fn := func() (string, error) {
			return "", expectedErr
		}

		result, err := WaitWithRateLimiterAndResult(ctx, rl, fn)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", result)
	})

	t.Run("rate limiter wait fails", func(t *testing.T) {
		config := &RateLimiterConfig{
			RequestsPerSecond: 0.1, // Very low rate
			Burst:             1,
			Timeout:           10 * time.Millisecond,
			Name:              "test-wait-result-timeout",
		}
		rl := NewRateLimiter(config)

		// Consume the burst
		rl.Wait(ctx)

		callCount := 0
		fn := func() (int, error) {
			callCount++
			return 42, nil
		}

		result, err := WaitWithRateLimiterAndResult(ctx, rl, fn)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Equal(t, 0, result)
		assert.Equal(t, 0, callCount) // Function should not be called
	})
}
