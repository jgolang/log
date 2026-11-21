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
