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

func SetVerbose(verbose bool) {
	if verbose {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo)
	}
}

func get() *slog.Logger {
	if defaultLogger == nil {
		Init(false)
	}
	return defaultLogger
}

func Debug(msg string, args ...any) {
	get().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	get().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	get().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	get().Error(msg, args...)
}

func With(args ...any) *slog.Logger {
	return get().With(args...)
}
