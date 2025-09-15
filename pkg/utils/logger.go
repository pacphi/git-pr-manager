package utils

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus with additional functionality
type Logger struct {
	*logrus.Logger
	fields logrus.Fields
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	logger := logrus.New()

	// Set log level from environment
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set output format
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	switch format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     shouldUseColors(),
		})
	}

	return &Logger{
		Logger: logger,
		fields: make(logrus.Fields),
	}
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newFields := make(logrus.Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &Logger{
		Logger: l.Logger,
		fields: newFields,
	}
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields logrus.Fields) *Logger {
	newFields := make(logrus.Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		Logger: l.Logger,
		fields: newFields,
	}
}

// WithError adds an error field to the logger context
func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err)
}

// WithProvider adds provider context to logs
func (l *Logger) WithProvider(provider string) *Logger {
	return l.WithField("provider", provider)
}

// WithComponent adds component context to logs
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField("component", component)
}

// WithRepo adds repository context to logs
func (l *Logger) WithRepo(repo string) *Logger {
	return l.WithField("repository", repo)
}

// WithPR adds pull request context to logs
func (l *Logger) WithPR(prNumber int) *Logger {
	return l.WithField("pr_number", prNumber)
}

// Debug logs a debug message with context fields
func (l *Logger) Debug(args ...interface{}) {
	l.Logger.WithFields(l.fields).Debug(args...)
}

// Debugf logs a formatted debug message with context fields
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Debugf(format, args...)
}

// Info logs an info message with context fields
func (l *Logger) Info(args ...interface{}) {
	l.Logger.WithFields(l.fields).Info(args...)
}

// Infof logs a formatted info message with context fields
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Infof(format, args...)
}

// Warn logs a warning message with context fields
func (l *Logger) Warn(args ...interface{}) {
	l.Logger.WithFields(l.fields).Warn(args...)
}

// Warnf logs a formatted warning message with context fields
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Warnf(format, args...)
}

// Error logs an error message with context fields
func (l *Logger) Error(args ...interface{}) {
	l.Logger.WithFields(l.fields).Error(args...)
}

// Errorf logs a formatted error message with context fields
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Errorf(format, args...)
}

// Fatal logs a fatal message with context fields and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.WithFields(l.fields).Fatal(args...)
}

// Fatalf logs a formatted fatal message with context fields and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Fatalf(format, args...)
}

// shouldUseColors determines if colored output should be used
func shouldUseColors() bool {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if FORCE_COLOR environment variable is set
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Default to no colors for safety
	return false
}

// Global logger instance
var defaultLogger = NewLogger()

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	defaultLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return defaultLogger
}

// Global logger functions for convenience

// Global logging functions removed as per remediation plan.
// The codebase now uses structured logging with component-based loggers.
// Use utils.GetGlobalLogger().WithComponent("component") for structured logging.
