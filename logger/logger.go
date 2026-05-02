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
type Priority int8

// Logger manages logging operations with various log levels and modes.
type Logger struct {
	calldepth int // Number of stack frames to ascend when generating log entries.
	prod      bool
	level     *slog.LevelVar
	mu        sync.RWMutex
	logger    *slog.Logger
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
	}
	l.SetJSONHandler(os.Stderr)
	return l
}

func (l *Logger) backend() *slog.Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logger
}

func (l *Logger) setBackend(next *slog.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger = next
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
	source := sourceWithStackTrace(l.calldepth + 1)
	args = append(args, source)
	l.backend().Debug(msg, args...)
}

// DebugContext logs a debug-level message with optional arguments and context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	source := sourceWithStackTrace(l.calldepth + 1)
	args = append(args, source)
	l.backend().DebugContext(ctx, msg, args...)
}

// Warn logs a warning-level message with optional arguments.
func (l *Logger) Warn(msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Warn(msg, args...)
}

// WarnContext logs a warning-level message with optional arguments and context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().WarnContext(ctx, msg, args...)
}

// Error logs an error-level message with optional arguments.
func (l *Logger) Error(msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Error(msg, args...)
}

// ErrorContext logs an error-level message with optional arguments and context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().ErrorContext(ctx, msg, args...)
}

// Panic logs a panic-level message, then panics with the message.
func (l *Logger) Panic(msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Log(context.Background(), LevelPanic, msg, args...)
	panic(msg)
}

// PanicContext logs a panic-level message with context, then panics with the message.
func (l *Logger) PanicContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Log(ctx, LevelPanic, msg, args...)
	panic(msg)
}

// Fatal logs a fatal-level message, then exits the application.
func (l *Logger) Fatal(msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

// FatalContext logs a fatal-level message with context, then exits the application.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

// Print logs a trace-level message with optional arguments.
func (l *Logger) Print(msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Log(context.Background(), LevelTrace, msg, args...)
}

// PrintContext logs a trace-level message with context and optional arguments.
func (l *Logger) PrintContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Log(ctx, LevelTrace, msg, args...)
}

// Info logs an info-level message with optional arguments.
func (l *Logger) Info(msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().Info(msg, args...)
}

// InfoContext logs an info-level message with context and optional arguments.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth + 1)
	args = append(args, source)
	l.backend().InfoContext(ctx, msg, args...)
}

// StackTrace provides a stack trace of up to 10 layers from where the error or incident was generated.
func (l *Logger) StackTrace() slog.Attr {
	return sourceWithStackTrace(l.calldepth + 1)
}

// SetCalldepth configures the number of stack frames to ascend for logging.
func (l *Logger) SetCalldepth(calldepth int) {
	l.calldepth = calldepth
}

// SetLevel sets the logging level for the Logger instance and returns the previous level.
func (l *Logger) SetLevel(level slog.Level) (oldLevel slog.Level) {
	oldLevel = l.level.Level()
	l.level.Set(level)
	return oldLevel
}
