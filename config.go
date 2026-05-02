package log

import (
	"io"
	"log/slog"

	"github.com/jgolang/log/logger"
)

type Logger = logger.Logger
type Option = logger.Option

func WithLevel(level slog.Level) Option {
	return logger.WithLevel(level)
}

func WithSource(enabled bool) Option {
	return logger.WithSource(enabled)
}

func WithDebugStackTrace(enabled bool) Option {
	return logger.WithDebugStackTrace(enabled)
}

func WithJSONHandler(w io.Writer) Option {
	return logger.WithJSONHandler(w)
}

func WithTextHandler(w io.Writer) Option {
	return logger.WithTextHandler(w)
}

// New creates a configurable logger instance without touching package-level state.
func New(opts ...Option) *Logger {
	return logger.NewWithOptions(opts...)
}

// SetSource controls whether package-level logs include source metadata.
func SetSource(enabled bool) {
	std.SetSource(enabled)
}

// SetDebugStackTrace controls whether package-level debug logs include stack traces.
func SetDebugStackTrace(enabled bool) {
	std.SetDebugStackTrace(enabled)
}
