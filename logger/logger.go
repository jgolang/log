// Package logger provides a flexible logging system with various log levels and contextual logging capabilities.

package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Priority represents the level of importance for log messages. Higher values indicate greater importance.
//
// Deprecated: use slog.Level instead.
type Priority int8

// Option configures a Logger instance.
type Option func(*Logger)

// Logger manages logging operations with various log levels and modes.
type Logger struct {
	calldepth  int // Number of stack frames to ascend when generating log entries.
	level      *slog.LevelVar
	addSource  bool
	debugStack bool
	mu         sync.RWMutex
	logger     *slog.Logger
}

type loggerConfig struct {
	calldepth  int
	addSource  bool
	debugStack bool
	logger     *slog.Logger
}

// WithLevel sets the initial logging level.
func WithLevel(level slog.Level) Option {
	return func(l *Logger) {
		l.level.Set(level)
	}
}

// WithSource controls whether source metadata is attached to log records.
func WithSource(enabled bool) Option {
	return func(l *Logger) {
		l.addSource = enabled
	}
}

// WithDebugStackTrace controls whether debug logs include a stack trace.
func WithDebugStackTrace(enabled bool) Option {
	return func(l *Logger) {
		l.debugStack = enabled
	}
}

// WithJSONHandler configures a JSON handler that writes to w.
func WithJSONHandler(w io.Writer) Option {
	return func(l *Logger) {
		l.SetJSONHandler(w)
	}
}

// WithTextHandler configures a text handler that writes to w.
func WithTextHandler(w io.Writer) Option {
	return func(l *Logger) {
		l.SetTextHandler(w)
	}
}

// New creates and initializes a new Logger instance.
// calldepth: Number of stack frames to ascend for log entries.
// pc: Deprecated and ignored. Kept for backward compatibility.
func New(calldepth int, _ []uintptr, levels ...*slog.LevelVar) *Logger {
	level := new(slog.LevelVar)
	level.Set(slog.LevelDebug)
	if len(levels) > 0 && levels[0] != nil {
		level = levels[0]
	}

	l := &Logger{
		calldepth: calldepth,
		level:     level,
		addSource: true,
	}
	l.SetJSONHandler(os.Stderr)
	return l
}

// NewWithOptions creates a new Logger instance using functional options.
func NewWithOptions(opts ...Option) *Logger {
	l := New(3, nil)
	for _, opt := range opts {
		if opt != nil {
			opt(l)
		}
	}
	return l
}

func (l *Logger) snapshot() loggerConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return loggerConfig{
		calldepth:  l.calldepth,
		addSource:  l.addSource,
		debugStack: l.debugStack,
		logger:     l.logger,
	}
}

func (l *Logger) setBackend(next *slog.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger = next
}

func (c loggerConfig) sourceAttr(withStack bool) slog.Attr {
	if !c.addSource {
		return slog.Attr{}
	}
	if withStack {
		return sourceWithStackTrace(c.calldepth + 1)
	}
	return source(c.calldepth + 1)
}

func (c loggerConfig) appendSource(args []any, withStack bool) []any {
	attr := c.sourceAttr(withStack)
	if attr.Key == "" {
		return args
	}
	return append(args, attr)
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args []any, withStack bool) {
	cfg := l.snapshot()
	logWithConfig(ctx, level, msg, args, cfg, withStack)
}

func logWithConfig(ctx context.Context, level slog.Level, msg string, args []any, cfg loggerConfig, withStack bool) {
	if cfg.logger == nil || !cfg.logger.Enabled(ctx, level) {
		return
	}
	args = cfg.appendSource(args, withStack)
	cfg.logger.Log(ctx, level, msg, args...)
}

// SetJSONHandler configures the logger to emit JSON logs to the provided writer.
func (l *Logger) SetJSONHandler(w io.Writer) {
	opts := &slog.HandlerOptions{
		Level:       l.level,
		ReplaceAttr: ReplaceAttr,
	}
	l.setBackend(slog.New(slog.NewJSONHandler(w, opts)))
}

// SetTextHandler configures the logger to emit text logs to the provided writer.
func (l *Logger) SetTextHandler(w io.Writer) {
	opts := &slog.HandlerOptions{
		Level:       l.level,
		ReplaceAttr: ReplaceAttr,
	}
	l.setBackend(slog.New(slog.NewTextHandler(w, opts)))
}

// Debug logs a debug-level message with optional arguments.
func (l *Logger) Debug(msg string, args ...any) {
	cfg := l.snapshot()
	logWithConfig(context.Background(), slog.LevelDebug, msg, args, cfg, cfg.debugStack)
}

// DebugContext logs a debug-level message with optional arguments and context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	cfg := l.snapshot()
	logWithConfig(ctx, slog.LevelDebug, msg, args, cfg, cfg.debugStack)
}

// Warn logs a warning-level message with optional arguments.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, args, false)
}

// WarnContext logs a warning-level message with optional arguments and context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args, false)
}

// Error logs an error-level message with optional arguments.
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), slog.LevelError, msg, args, false)
}

// ErrorContext logs an error-level message with optional arguments and context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args, false)
}

// Panic logs a panic-level message, then panics with the message.
func (l *Logger) Panic(msg string, args ...any) {
	l.log(context.Background(), LevelPanic, msg, args, false)
	panic(msg)
}

// PanicContext logs a panic-level message with context, then panics with the message.
func (l *Logger) PanicContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelPanic, msg, args, false)
	panic(msg)
}

// Fatal logs a fatal-level message, then exits the application.
func (l *Logger) Fatal(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args, false)
	os.Exit(1)
}

// FatalContext logs a fatal-level message with context, then exits the application.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args, false)
	os.Exit(1)
}

// Print logs a trace-level message with optional arguments.
func (l *Logger) Print(msg string, args ...any) {
	l.log(context.Background(), LevelTrace, msg, args, false)
}

// PrintContext logs a trace-level message with context and optional arguments.
func (l *Logger) PrintContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelTrace, msg, args, false)
}

// Info logs an info-level message with optional arguments.
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, args, false)
}

// InfoContext logs an info-level message with context and optional arguments.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args, false)
}

// StackTrace provides a stack trace of up to 10 layers from where the error or incident was generated.
func (l *Logger) StackTrace() slog.Attr {
	return sourceWithStackTrace(l.snapshot().calldepth + 1)
}

// SetCalldepth configures the number of stack frames to ascend for logging.
func (l *Logger) SetCalldepth(calldepth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.calldepth = calldepth
}

// SetSource controls whether source metadata is attached to log records.
func (l *Logger) SetSource(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.addSource = enabled
}

// SetDebugStackTrace controls whether debug logs include a stack trace.
func (l *Logger) SetDebugStackTrace(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugStack = enabled
}

// SetLevel sets the logging level for the Logger instance and returns the previous level.
func (l *Logger) SetLevel(level slog.Level) (oldLevel slog.Level) {
	oldLevel = l.level.Level()
	l.level.Set(level)
	return oldLevel
}
