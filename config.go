package log

import (
	"io"
	"log/slog"

	"github.com/jgolang/log/logger"
)

// Logger is a configurable logger instance, created with [New]. It is an
// alias of the implementation type in package logger, so values of either
// name are interchangeable.
type Logger = logger.Logger

// Option configures a [Logger] created with [New].
type Option = logger.Option

// WithLevel sets the initial logging level of a [Logger].
func WithLevel(level slog.Level) Option {
	return logger.WithLevel(level)
}

// WithSource controls whether a [Logger] attaches source metadata to records.
func WithSource(enabled bool) Option {
	return logger.WithSource(enabled)
}

// WithDebugStackTrace controls whether a [Logger] adds a stack trace to debug
// records.
func WithDebugStackTrace(enabled bool) Option {
	return logger.WithDebugStackTrace(enabled)
}

// WithJSONHandler makes a [Logger] write JSON records to w.
func WithJSONHandler(w io.Writer) Option {
	return logger.WithJSONHandler(w)
}

// WithTextHandler makes a [Logger] write text records to w.
func WithTextHandler(w io.Writer) Option {
	return logger.WithTextHandler(w)
}

// WithHandlerWrapper wraps the JSON or text handler the logger builds with
// wrap, in any option order and across later handler changes. Level and
// attribute normalization keep working. A nil wrap means no wrapper.
func WithHandlerWrapper(wrap func(slog.Handler) slog.Handler) Option {
	return logger.WithHandlerWrapper(wrap)
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

// SetHandlerWrapper wraps the handler of the package-level logger with wrap,
// for example otel.NewOtelHandler(), so log.InfoC and the other helpers
// correlate records with the span in their context. The wrapper is kept when
// NewJSONHandler or NewTextHandler replaces the handler. A nil wrap removes
// it.
func SetHandlerWrapper(wrap func(slog.Handler) slog.Handler) {
	std.SetHandlerWrapper(wrap)
}
