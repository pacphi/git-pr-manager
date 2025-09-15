package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, time.Second, config.InitialBackoff)
	assert.Equal(t, 30*time.Second, config.MaxBackoff)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.True(t, config.Jitter)
	assert.NotNil(t, config.RetryIf)

	// Test the default RetryIf function
	assert.True(t, config.RetryIf(errors.New("test error")))
	assert.False(t, config.RetryIf(nil))
}

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxAttempts = 3

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Retry(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxAttempts = 3
	config.InitialBackoff = 10 * time.Millisecond
	config.MaxBackoff = 50 * time.Millisecond

	callCount := 0
	expectedError := errors.New("persistent error")
	fn := func() error {
		callCount++
		return expectedError
	}

	err := Retry(ctx, config, fn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 3 attempts")
	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, 3, callCount)
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultRetryConfig()
	config.MaxAttempts = 5
	config.InitialBackoff = 100 * time.Millisecond

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 2 {
			cancel() // Cancel after second attempt
		}
		return errors.New("test error")
	}

	err := Retry(ctx, config, fn)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.GreaterOrEqual(t, callCount, 2)
}

func TestRetry_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	config := DefaultRetryConfig()
	config.MaxAttempts = 5
	config.InitialBackoff = 100 * time.Millisecond

	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("test error")
	}

	err := Retry(ctx, config, fn)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxAttempts = 3
	config.RetryIf = func(err error) bool {
		return err.Error() != "non-retryable"
	}

	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("non-retryable")
	}

	err := Retry(ctx, config, fn)

	assert.Error(t, err)
	assert.Equal(t, "non-retryable", err.Error())
	assert.Equal(t, 1, callCount)
}

func TestRetry_NilConfig(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 2 {
			return errors.New("test error")
		}
		return nil
	}

	err := Retry(ctx, nil, fn)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRetryWithResult_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxAttempts = 3

	callCount := 0
	fn := func() (string, error) {
		callCount++
		if callCount < 2 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	}

	result, err := RetryWithResult(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 2, callCount)
}

func TestRetryWithResult_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxAttempts = 3
	config.InitialBackoff = 10 * time.Millisecond

	callCount := 0
	expectedError := errors.New("persistent error")
	fn := func() (int, error) {
		callCount++
		return 0, expectedError
	}

	result, err := RetryWithResult(ctx, config, fn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 3 attempts")
	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, 0, result)
	assert.Equal(t, 3, callCount)
}

func TestRetryWithResult_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultRetryConfig()
	config.MaxAttempts = 5
	config.InitialBackoff = 100 * time.Millisecond

	callCount := 0
	fn := func() (string, error) {
		callCount++
		if callCount == 2 {
			cancel()
		}
		return "", errors.New("test error")
	}

	result, err := RetryWithResult(ctx, config, fn)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, "", result)
	assert.GreaterOrEqual(t, callCount, 2)
}

func TestRetryConfig_CalculateBackoff(t *testing.T) {
	tests := []struct {
		name     string
		config   *RetryConfig
		attempt  int
		expected time.Duration
	}{
		{
			name: "first attempt",
			config: &RetryConfig{
				InitialBackoff: time.Second,
				BackoffFactor:  2.0,
				MaxBackoff:     30 * time.Second,
				Jitter:         false,
			},
			attempt:  1,
			expected: time.Second,
		},
		{
			name: "second attempt",
			config: &RetryConfig{
				InitialBackoff: time.Second,
				BackoffFactor:  2.0,
				MaxBackoff:     30 * time.Second,
				Jitter:         false,
			},
			attempt:  2,
			expected: 2 * time.Second,
		},
		{
			name: "max backoff reached",
			config: &RetryConfig{
				InitialBackoff: time.Second,
				BackoffFactor:  2.0,
				MaxBackoff:     5 * time.Second,
				Jitter:         false,
			},
			attempt:  10,
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.config.calculateBackoff(tt.attempt)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRetryConfig_CalculateBackoffWithJitter(t *testing.T) {
	config := &RetryConfig{
		InitialBackoff: time.Second,
		BackoffFactor:  2.0,
		MaxBackoff:     30 * time.Second,
		Jitter:         true,
	}

	// Run multiple times to ensure jitter is working
	backoffs := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		backoffs[i] = config.calculateBackoff(2)
	}

	// Check that we have some variation (not all the same)
	allSame := true
	first := backoffs[0]
	for _, backoff := range backoffs[1:] {
		if backoff != first {
			allSame = false
			break
		}
	}
	assert.False(t, allSame, "jitter should create variation in backoff times")

	// Check that all values are within reasonable bounds (base ± 10%)
	baseBackoff := 2 * time.Second
	minExpected := time.Duration(float64(baseBackoff) * 0.85)
	maxExpected := time.Duration(float64(baseBackoff) * 1.15)

	for _, backoff := range backoffs {
		assert.GreaterOrEqual(t, backoff, minExpected)
		assert.LessOrEqual(t, backoff, maxExpected)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "connection reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New("timeout occurred"),
			expected: true,
		},
		{
			name:     "rate limited",
			err:      errors.New("rate limited"),
			expected: true,
		},
		{
			name:     "server error",
			err:      errors.New("server error"),
			expected: true,
		},
		{
			name:     "service unavailable",
			err:      errors.New("service unavailable"),
			expected: true,
		},
		{
			name:     "bad gateway",
			err:      errors.New("bad gateway"),
			expected: true,
		},
		{
			name:     "gateway timeout",
			err:      errors.New("gateway timeout"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsRetryableError(tt.err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

type temporaryError struct {
	msg       string
	temporary bool
}

func (e *temporaryError) Error() string {
	return e.msg
}

func (e *temporaryError) Temporary() bool {
	return e.temporary
}

func TestIsRetryableError_TemporaryInterface(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "temporary error - true",
			err:      &temporaryError{msg: "temp error", temporary: true},
			expected: true,
		},
		{
			name:     "temporary error - false",
			err:      &temporaryError{msg: "non-temp error", temporary: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsRetryableError(tt.err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRetryableFunc_Presets(t *testing.T) {
	t.Run("RetryOnAnyError", func(t *testing.T) {
		assert.True(t, RetryOnAnyError(errors.New("any error")))
		assert.False(t, RetryOnAnyError(nil))
		assert.False(t, RetryOnAnyError(context.Canceled))
		assert.False(t, RetryOnAnyError(context.DeadlineExceeded))
	})

	t.Run("RetryOnNetworkError", func(t *testing.T) {
		assert.True(t, RetryOnNetworkError(errors.New("connection refused")))
		assert.False(t, RetryOnNetworkError(errors.New("validation error")))
		assert.False(t, RetryOnNetworkError(nil))
	})

	t.Run("RetryOnTemporaryError", func(t *testing.T) {
		tempErr := &temporaryError{msg: "temp", temporary: true}
		nonTempErr := &temporaryError{msg: "non-temp", temporary: false}
		regularErr := errors.New("regular")

		assert.True(t, RetryOnTemporaryError(tempErr))
		assert.False(t, RetryOnTemporaryError(nonTempErr))
		assert.False(t, RetryOnTemporaryError(regularErr))
		assert.False(t, RetryOnTemporaryError(nil))
	})
}

func TestRetryConfig_ChainableMethods(t *testing.T) {
	config := DefaultRetryConfig().
		WithMaxAttempts(5).
		WithBackoff(2*time.Second, 3.0).
		WithMaxBackoff(60 * time.Second).
		WithJitter(false).
		WithRetryIf(func(err error) bool { return true })

	assert.Equal(t, 5, config.MaxAttempts)
	assert.Equal(t, 2*time.Second, config.InitialBackoff)
	assert.Equal(t, 3.0, config.BackoffFactor)
	assert.Equal(t, 60*time.Second, config.MaxBackoff)
	assert.False(t, config.Jitter)
	assert.NotNil(t, config.RetryIf)
	assert.True(t, config.RetryIf(errors.New("test")))
}

func TestRetry_FirstAttemptSuccess(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := Retry(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetryWithResult_FirstAttemptSuccess(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "immediate success", nil
	}

	result, err := RetryWithResult(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, "immediate success", result)
	assert.Equal(t, 1, callCount)
}

func TestContainsHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "test",
			substr:   "test",
			expected: true,
		},
		{
			name:     "substring at start",
			s:        "testing",
			substr:   "test",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "unittest",
			substr:   "test",
			expected: true,
		},
		{
			name:     "substring in middle",
			s:        "atestb",
			substr:   "test",
			expected: true,
		},
		{
			name:     "not found",
			s:        "hello world",
			substr:   "test",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "substring longer than string",
			s:        "hi",
			substr:   "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIndexContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{
			name:     "found at beginning",
			s:        "testing",
			substr:   "test",
			expected: 0,
		},
		{
			name:     "found at end",
			s:        "unittest",
			substr:   "test",
			expected: 4,
		},
		{
			name:     "found in middle",
			s:        "atestb",
			substr:   "test",
			expected: 1,
		},
		{
			name:     "not found",
			s:        "hello world",
			substr:   "test",
			expected: -1,
		},
		{
			name:     "multiple occurrences",
			s:        "testtest",
			substr:   "test",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := indexContains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRetryWithResult_NilConfig(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	fn := func() (int, error) {
		callCount++
		if callCount < 2 {
			return 0, errors.New("test error")
		}
		return 42, nil
	}

	result, err := RetryWithResult(ctx, nil, fn)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 2, callCount)
}
