package utils

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "returns environment variable value when set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			setEnv:       true,
			expected:     "env_value",
		},
		{
			name:         "returns default value when env var not set",
			key:          "TEST_VAR_NOT_SET",
			defaultValue: "default",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "returns default value when env var is empty",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "handles empty default value",
			key:          "TEST_VAR_NO_DEFAULT",
			defaultValue: "",
			setEnv:       false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		expected     int
	}{
		{
			name:         "returns parsed int from env var",
			key:          "TEST_INT_VAR",
			defaultValue: 42,
			envValue:     "123",
			setEnv:       true,
			expected:     123,
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_INT_NOT_SET",
			defaultValue: 42,
			setEnv:       false,
			expected:     42,
		},
		{
			name:         "returns default when env var is not valid int",
			key:          "TEST_INT_INVALID",
			defaultValue: 42,
			envValue:     "not_a_number",
			setEnv:       true,
			expected:     42,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_INT_EMPTY",
			defaultValue: 42,
			envValue:     "",
			setEnv:       true,
			expected:     42,
		},
		{
			name:         "handles negative numbers",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 42,
			envValue:     "-123",
			setEnv:       true,
			expected:     -123,
		},
		{
			name:         "handles zero",
			key:          "TEST_INT_ZERO",
			defaultValue: 42,
			envValue:     "0",
			setEnv:       true,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
	}{
		{
			name:         "returns true from env var 'true'",
			key:          "TEST_BOOL_VAR",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false from env var 'false'",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns true from env var '1'",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false from env var '0'",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_BOOL_NOT_SET",
			defaultValue: true,
			setEnv:       false,
			expected:     true,
		},
		{
			name:         "returns default when env var is invalid",
			key:          "TEST_BOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_BOOL_EMPTY",
			defaultValue: true,
			envValue:     "",
			setEnv:       true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnvBool(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		setEnv       bool
		expected     time.Duration
	}{
		{
			name:         "returns parsed duration from env var",
			key:          "TEST_DURATION_VAR",
			defaultValue: 5 * time.Second,
			envValue:     "10s",
			setEnv:       true,
			expected:     10 * time.Second,
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_DURATION_NOT_SET",
			defaultValue: 5 * time.Second,
			setEnv:       false,
			expected:     5 * time.Second,
		},
		{
			name:         "returns default when env var is invalid",
			key:          "TEST_DURATION_INVALID",
			defaultValue: 5 * time.Second,
			envValue:     "invalid",
			setEnv:       true,
			expected:     5 * time.Second,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_DURATION_EMPTY",
			defaultValue: 5 * time.Second,
			envValue:     "",
			setEnv:       true,
			expected:     5 * time.Second,
		},
		{
			name:         "handles minutes",
			key:          "TEST_DURATION_MINUTES",
			defaultValue: 5 * time.Second,
			envValue:     "2m",
			setEnv:       true,
			expected:     2 * time.Minute,
		},
		{
			name:         "handles hours",
			key:          "TEST_DURATION_HOURS",
			defaultValue: 5 * time.Second,
			envValue:     "1h",
			setEnv:       true,
			expected:     1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnvDuration(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustGetEnv(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		key := "TEST_MUST_GET_SET"
		value := "test_value"
		os.Setenv(key, value)
		defer os.Unsetenv(key)

		result := MustGetEnv(key)
		assert.Equal(t, value, result)
	})

	t.Run("panics when env var is not set", func(t *testing.T) {
		key := "TEST_MUST_GET_NOT_SET"
		os.Unsetenv(key) // Ensure it's not set

		assert.Panics(t, func() {
			MustGetEnv(key)
		})
	})

	t.Run("panics when env var is empty", func(t *testing.T) {
		key := "TEST_MUST_GET_EMPTY"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		assert.Panics(t, func() {
			MustGetEnv(key)
		})
	})
}

// TestSetEnvIfEmpty has been removed as SetEnvIfEmpty function was removed
/*
func TestSetEnvIfEmpty(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		existingValue string
		setExisting   bool
		expectError   bool
		expectedFinal string
	}{
		{
			name:          "sets env var when not set",
			key:           "TEST_SET_IF_EMPTY_NOT_SET",
			value:         "new_value",
			setExisting:   false,
			expectError:   false,
			expectedFinal: "new_value",
		},
		{
			name:          "does not set env var when already set",
			key:           "TEST_SET_IF_EMPTY_SET",
			value:         "new_value",
			existingValue: "existing_value",
			setExisting:   true,
			expectError:   false,
			expectedFinal: "existing_value",
		},
		{
			name:          "sets env var when existing is empty",
			key:           "TEST_SET_IF_EMPTY_EMPTY",
			value:         "new_value",
			existingValue: "",
			setExisting:   true,
			expectError:   false,
			expectedFinal: "new_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before and after
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setExisting {
				os.Setenv(tt.key, tt.existingValue)
			}

			err := SetEnvIfEmpty(tt.key, tt.value)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFinal, os.Getenv(tt.key))
			}
		})
	}
}
*/

// Test environment isolation
func TestEnvironmentIsolation(t *testing.T) {
	key := "TEST_ISOLATION"
	originalValue := os.Getenv(key)
	defer func() {
		if originalValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, originalValue)
		}
	}()

	// Test that modifications don't affect subsequent tests
	os.Setenv(key, "test1")
	assert.Equal(t, "test1", GetEnv(key, "default"))

	os.Setenv(key, "test2")
	assert.Equal(t, "test2", GetEnv(key, "default"))

	os.Unsetenv(key)
	assert.Equal(t, "default", GetEnv(key, "default"))
}

// Benchmark tests
func BenchmarkGetEnv(b *testing.B) {
	key := "BENCHMARK_ENV_VAR"
	os.Setenv(key, "benchmark_value")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnv(key, "default")
	}
}

func BenchmarkGetEnvInt(b *testing.B) {
	key := "BENCHMARK_INT_VAR"
	os.Setenv(key, "42")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnvInt(key, 0)
	}
}

func BenchmarkGetEnvBool(b *testing.B) {
	key := "BENCHMARK_BOOL_VAR"
	os.Setenv(key, "true")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnvBool(key, false)
	}
}

func BenchmarkGetEnvDuration(b *testing.B) {
	key := "BENCHMARK_DURATION_VAR"
	os.Setenv(key, "5s")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnvDuration(key, time.Second)
	}
}

// Edge case tests
func TestGetEnvEdgeCases(t *testing.T) {
	t.Run("very long environment variable", func(t *testing.T) {
		key := "TEST_LONG_VAR"
		longValue := string(make([]byte, 10000))
		for i := range longValue {
			longValue = string(rune('a' + (i % 26)))
		}

		os.Setenv(key, longValue)
		defer os.Unsetenv(key)

		result := GetEnv(key, "default")
		assert.Equal(t, longValue, result)
	})

	t.Run("unicode environment variable", func(t *testing.T) {
		key := "TEST_UNICODE_VAR"
		unicodeValue := "测试值🚀"

		os.Setenv(key, unicodeValue)
		defer os.Unsetenv(key)

		result := GetEnv(key, "default")
		assert.Equal(t, unicodeValue, result)
	})

	t.Run("special characters in environment variable", func(t *testing.T) {
		key := "TEST_SPECIAL_VAR"
		specialValue := "!@#$%^&*()_+-=[]{}|;':\",./<>?"

		os.Setenv(key, specialValue)
		defer os.Unsetenv(key)

		result := GetEnv(key, "default")
		assert.Equal(t, specialValue, result)
	})
}
