package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_DefaultLevel(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")

	logger := NewLogger()

	assert.NotNil(t, logger)
	assert.Equal(t, logrus.InfoLevel, logger.Level)
	assert.NotNil(t, logger.fields)
	assert.Equal(t, 0, len(logger.fields))
}

func TestNewLogger_LogLevels(t *testing.T) {
	tests := []struct {
		envValue      string
		expectedLevel logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"DEBUG", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"INFO", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"WARN", logrus.WarnLevel},
		{"warning", logrus.WarnLevel},
		{"WARNING", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
		{"ERROR", logrus.ErrorLevel},
		{"fatal", logrus.FatalLevel},
		{"FATAL", logrus.FatalLevel},
		{"panic", logrus.PanicLevel},
		{"PANIC", logrus.PanicLevel},
		{"invalid", logrus.InfoLevel},
		{"", logrus.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.envValue)
			defer os.Unsetenv("LOG_LEVEL")

			logger := NewLogger()
			assert.Equal(t, tt.expectedLevel, logger.Level)
		})
	}
}

func TestNewLogger_LogFormats(t *testing.T) {
	tests := []struct {
		envValue        string
		expectedType    string
		timestampFormat string
	}{
		{"json", "JSONFormatter", "2006-01-02T15:04:05.000Z07:00"},
		{"JSON", "JSONFormatter", "2006-01-02T15:04:05.000Z07:00"},
		{"text", "TextFormatter", "2006-01-02 15:04:05"},
		{"TEXT", "TextFormatter", "2006-01-02 15:04:05"},
		{"", "TextFormatter", "2006-01-02 15:04:05"},
		{"invalid", "TextFormatter", "2006-01-02 15:04:05"},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("LOG_FORMAT", tt.envValue)
			defer os.Unsetenv("LOG_FORMAT")

			logger := NewLogger()

			formatter := logger.Formatter
			switch tt.expectedType {
			case "JSONFormatter":
				jsonFormatter, ok := formatter.(*logrus.JSONFormatter)
				assert.True(t, ok, "Expected JSONFormatter")
				if ok {
					assert.Equal(t, tt.timestampFormat, jsonFormatter.TimestampFormat)
				}
			case "TextFormatter":
				textFormatter, ok := formatter.(*logrus.TextFormatter)
				assert.True(t, ok, "Expected TextFormatter")
				if ok {
					assert.Equal(t, tt.timestampFormat, textFormatter.TimestampFormat)
					assert.True(t, textFormatter.FullTimestamp)
				}
			}
		})
	}
}

func TestLogger_WithField(t *testing.T) {
	logger := NewLogger()

	newLogger := logger.WithField("key1", "value1")

	// Original logger should be unchanged
	assert.Equal(t, 0, len(logger.fields))

	// New logger should have the field
	assert.Equal(t, 1, len(newLogger.fields))
	assert.Equal(t, "value1", newLogger.fields["key1"])

	// Chain another field
	chainedLogger := newLogger.WithField("key2", "value2")

	assert.Equal(t, 1, len(newLogger.fields)) // Original should be unchanged
	assert.Equal(t, 2, len(chainedLogger.fields))
	assert.Equal(t, "value1", chainedLogger.fields["key1"])
	assert.Equal(t, "value2", chainedLogger.fields["key2"])
}

func TestLogger_WithFields(t *testing.T) {
	logger := NewLogger().WithField("existing", "value")

	newFields := logrus.Fields{
		"key1": "value1",
		"key2": "value2",
	}

	newLogger := logger.WithFields(newFields)

	// Should have original field plus new fields
	assert.Equal(t, 3, len(newLogger.fields))
	assert.Equal(t, "value", newLogger.fields["existing"])
	assert.Equal(t, "value1", newLogger.fields["key1"])
	assert.Equal(t, "value2", newLogger.fields["key2"])

	// Original logger should be unchanged
	assert.Equal(t, 1, len(logger.fields))
}

func TestLogger_WithFields_Overwrites(t *testing.T) {
	logger := NewLogger().WithField("key1", "original")

	newFields := logrus.Fields{
		"key1": "overwritten",
		"key2": "value2",
	}

	newLogger := logger.WithFields(newFields)

	assert.Equal(t, 2, len(newLogger.fields))
	assert.Equal(t, "overwritten", newLogger.fields["key1"])
	assert.Equal(t, "value2", newLogger.fields["key2"])
}

func TestLogger_WithError(t *testing.T) {
	logger := NewLogger()
	testError := errors.New("test error")

	errorLogger := logger.WithError(testError)

	assert.Equal(t, 1, len(errorLogger.fields))
	assert.Equal(t, testError, errorLogger.fields["error"])
}

func TestLogger_WithProvider(t *testing.T) {
	logger := NewLogger()

	providerLogger := logger.WithProvider("github")

	assert.Equal(t, 1, len(providerLogger.fields))
	assert.Equal(t, "github", providerLogger.fields["provider"])
}

func TestLogger_WithComponent(t *testing.T) {
	logger := NewLogger()

	componentLogger := logger.WithComponent("executor")

	assert.Equal(t, 1, len(componentLogger.fields))
	assert.Equal(t, "executor", componentLogger.fields["component"])
}

func TestLogger_WithRepo(t *testing.T) {
	logger := NewLogger()

	repoLogger := logger.WithRepo("owner/repo")

	assert.Equal(t, 1, len(repoLogger.fields))
	assert.Equal(t, "owner/repo", repoLogger.fields["repository"])
}

func TestLogger_WithPR(t *testing.T) {
	logger := NewLogger()

	prLogger := logger.WithPR(123)

	assert.Equal(t, 1, len(prLogger.fields))
	assert.Equal(t, 123, prLogger.fields["pr_number"])
}

func TestLogger_LoggingMethods_WithFields(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer

	logger := NewLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)

	// Set to JSON format for easier parsing
	logger.SetFormatter(&logrus.JSONFormatter{})

	contextLogger := logger.
		WithComponent("test").
		WithProvider("github").
		WithRepo("owner/repo").
		WithPR(123).
		WithError(errors.New("test error"))

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Debug",
			logFunc: func() {
				contextLogger.Debug("debug message")
			},
			expected: "debug message",
		},
		{
			name: "Debugf",
			logFunc: func() {
				contextLogger.Debugf("debug %s %d", "message", 42)
			},
			expected: "debug message 42",
		},
		{
			name: "Info",
			logFunc: func() {
				contextLogger.Info("info message")
			},
			expected: "info message",
		},
		{
			name: "Infof",
			logFunc: func() {
				contextLogger.Infof("info %s", "message")
			},
			expected: "info message",
		},
		{
			name: "Warn",
			logFunc: func() {
				contextLogger.Warn("warn message")
			},
			expected: "warn message",
		},
		{
			name: "Warnf",
			logFunc: func() {
				contextLogger.Warnf("warn %s", "message")
			},
			expected: "warn message",
		},
		{
			name: "Error",
			logFunc: func() {
				contextLogger.Error("error message")
			},
			expected: "error message",
		},
		{
			name: "Errorf",
			logFunc: func() {
				contextLogger.Errorf("error %s", "message")
			},
			expected: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			assert.NotEmpty(t, output)

			// Parse JSON log entry
			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)

			// Check message
			assert.Equal(t, tt.expected, logEntry["msg"])

			// Check fields are present
			assert.Equal(t, "test", logEntry["component"])
			assert.Equal(t, "github", logEntry["provider"])
			assert.Equal(t, "owner/repo", logEntry["repository"])
			assert.Equal(t, float64(123), logEntry["pr_number"]) // JSON numbers are float64
			assert.Equal(t, "test error", logEntry["error"])
		})
	}
}

func TestLogger_LogLevels_Filtering(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.WarnLevel) // Only warn and above

	tests := []struct {
		name      string
		logFunc   func()
		shouldLog bool
	}{
		{
			name:      "Debug (should not log)",
			logFunc:   func() { logger.Debug("debug message") },
			shouldLog: false,
		},
		{
			name:      "Info (should not log)",
			logFunc:   func() { logger.Info("info message") },
			shouldLog: false,
		},
		{
			name:      "Warn (should log)",
			logFunc:   func() { logger.Warn("warn message") },
			shouldLog: true,
		},
		{
			name:      "Error (should log)",
			logFunc:   func() { logger.Error("error message") },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			if tt.shouldLog {
				assert.NotEmpty(t, output)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestShouldUseColors(t *testing.T) {
	tests := []struct {
		name           string
		noColor        string
		forceColor     string
		expectedColors bool
	}{
		{
			name:           "Default (no env vars)",
			noColor:        "",
			forceColor:     "",
			expectedColors: false,
		},
		{
			name:           "NO_COLOR set",
			noColor:        "1",
			forceColor:     "",
			expectedColors: false,
		},
		{
			name:           "FORCE_COLOR set",
			noColor:        "",
			forceColor:     "1",
			expectedColors: true,
		},
		{
			name:           "Both set (NO_COLOR takes precedence)",
			noColor:        "1",
			forceColor:     "1",
			expectedColors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.noColor != "" {
				os.Setenv("NO_COLOR", tt.noColor)
				defer os.Unsetenv("NO_COLOR")
			} else {
				os.Unsetenv("NO_COLOR")
			}

			if tt.forceColor != "" {
				os.Setenv("FORCE_COLOR", tt.forceColor)
				defer os.Unsetenv("FORCE_COLOR")
			} else {
				os.Unsetenv("FORCE_COLOR")
			}

			result := shouldUseColors()
			assert.Equal(t, tt.expectedColors, result)
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Save original global logger
	originalLogger := GetGlobalLogger()

	// Create a new logger and set it as global
	newLogger := NewLogger().WithComponent("test-global")
	SetGlobalLogger(newLogger)

	// Verify it was set
	retrievedLogger := GetGlobalLogger()
	assert.Equal(t, newLogger, retrievedLogger)
	assert.Equal(t, "test-global", retrievedLogger.fields["component"])

	// Restore original logger
	SetGlobalLogger(originalLogger)
}

func TestLogger_ImmutableFields(t *testing.T) {
	originalLogger := NewLogger().WithField("original", "value")

	// Create multiple derived loggers
	logger1 := originalLogger.WithField("logger1", "value1")
	logger2 := originalLogger.WithField("logger2", "value2")

	// Verify they don't affect each other
	assert.Equal(t, 1, len(originalLogger.fields))
	assert.Equal(t, "value", originalLogger.fields["original"])

	assert.Equal(t, 2, len(logger1.fields))
	assert.Equal(t, "value", logger1.fields["original"])
	assert.Equal(t, "value1", logger1.fields["logger1"])
	assert.NotContains(t, logger1.fields, "logger2")

	assert.Equal(t, 2, len(logger2.fields))
	assert.Equal(t, "value", logger2.fields["original"])
	assert.Equal(t, "value2", logger2.fields["logger2"])
	assert.NotContains(t, logger2.fields, "logger1")
}

func TestLogger_ChainedContextMethods(t *testing.T) {
	logger := NewLogger().
		WithProvider("github").
		WithComponent("executor").
		WithRepo("owner/repo").
		WithPR(456).
		WithError(errors.New("chained error")).
		WithField("custom", "value")

	expectedFields := map[string]interface{}{
		"provider":   "github",
		"component":  "executor",
		"repository": "owner/repo",
		"pr_number":  456,
		"error":      errors.New("chained error"),
		"custom":     "value",
	}

	assert.Equal(t, len(expectedFields), len(logger.fields))
	for key, expectedValue := range expectedFields {
		actualValue := logger.fields[key]
		if err, ok := expectedValue.(error); ok {
			assert.Equal(t, err.Error(), actualValue.(error).Error())
		} else {
			assert.Equal(t, expectedValue, actualValue)
		}
	}
}

// Test that logging methods don't panic and produce output
func TestLogger_LoggingMethodsProduceOutput(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)

	contextLogger := logger.WithField("test", "value")

	// Test all logging methods
	assert.NotPanics(t, func() {
		contextLogger.Debug("debug")
		contextLogger.Debugf("debugf %s", "test")
		contextLogger.Info("info")
		contextLogger.Infof("infof %s", "test")
		contextLogger.Warn("warn")
		contextLogger.Warnf("warnf %s", "test")
		contextLogger.Error("error")
		contextLogger.Errorf("errorf %s", "test")
	})

	output := buf.String()
	assert.NotEmpty(t, output)

	// Should contain all log messages
	assert.Contains(t, output, "debug")
	assert.Contains(t, output, "debugf test")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "infof test")
	assert.Contains(t, output, "warn")
	assert.Contains(t, output, "warnf test")
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "errorf test")
}

func TestLogger_TextFormatterConfiguration(t *testing.T) {
	os.Unsetenv("LOG_FORMAT") // Use default text format
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("FORCE_COLOR")

	logger := NewLogger()

	formatter, ok := logger.Formatter.(*logrus.TextFormatter)
	require.True(t, ok, "Expected TextFormatter")

	assert.True(t, formatter.FullTimestamp)
	assert.Equal(t, "2006-01-02 15:04:05", formatter.TimestampFormat)
	assert.False(t, formatter.ForceColors) // Default should be false
}

func TestLogger_JSONFormatterConfiguration(t *testing.T) {
	os.Setenv("LOG_FORMAT", "json")
	defer os.Unsetenv("LOG_FORMAT")

	logger := NewLogger()

	formatter, ok := logger.Formatter.(*logrus.JSONFormatter)
	require.True(t, ok, "Expected JSONFormatter")

	assert.Equal(t, "2006-01-02T15:04:05.000Z07:00", formatter.TimestampFormat)
}

// Benchmark logger performance
func BenchmarkLogger_WithField(b *testing.B) {
	logger := NewLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithField("key", "value")
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	logger := NewLogger()
	fields := logrus.Fields{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields)
	}
}

func BenchmarkLogger_Info(b *testing.B) {
	logger := NewLogger()
	logger.SetOutput(bytes.NewBuffer(nil)) // Discard output

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

func BenchmarkLogger_InfoWithFields(b *testing.B) {
	logger := NewLogger().
		WithComponent("benchmark").
		WithProvider("test").
		WithField("iteration", 0)

	logger.SetOutput(bytes.NewBuffer(nil)) // Discard output

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message with fields")
	}
}
