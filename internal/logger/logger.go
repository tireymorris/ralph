package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
	level         = new(slog.LevelVar)
)

// Init initializes the global logger with the specified verbosity.
// If verbose is true, debug-level logs are shown.
func Init(verbose bool) {
	once.Do(func() {
		if verbose {
			level.Set(slog.LevelDebug)
		} else {
			level.Set(slog.LevelInfo)
		}

		opts := &slog.HandlerOptions{
			Level: level,
		}

		handler := slog.NewTextHandler(os.Stderr, opts)
		defaultLogger = slog.New(handler)
	})
}

// InitWithWriter initializes the logger with a custom writer (useful for testing).
func InitWithWriter(w io.Writer, verbose bool) {
	if verbose {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(w, opts)
	defaultLogger = slog.New(handler)
}

// SetVerbose changes the log level dynamically.
func SetVerbose(verbose bool) {
	if verbose {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo)
	}
}

// get returns the logger, initializing with defaults if needed.
func get() *slog.Logger {
	if defaultLogger == nil {
		Init(false)
	}
	return defaultLogger
}

// Debug logs a debug message (only shown with --verbose).
func Debug(msg string, args ...any) {
	get().Debug(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	get().Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	get().Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	get().Error(msg, args...)
}

// With returns a logger with additional context attributes.
func With(args ...any) *slog.Logger {
	return get().With(args...)
}
