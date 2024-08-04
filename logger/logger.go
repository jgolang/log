// Package logger provides a flexible logging system with various log levels and contextual logging capabilities.

package logger

import (
	"context"
	"os"

	"log/slog"
)

// Priority represents the level of importance for log messages. Higher values indicate greater importance.
type Priority int8

// Logger manages logging operations with various log levels and modes.
type Logger struct {
	calldepth int       // Number of stack frames to ascend when generating log entries.
	pc        []uintptr // Stack trace information used for logging.
	prod      bool      // Indicates whether the logger is in production mode.
}

// New creates and initializes a new Logger instance.
// calldepth: Number of stack frames to ascend for log entries.
// pc: A slice of uintptr used to store stack trace information.
func New(calldepth int, pc []uintptr) *Logger {
	return &Logger{
		calldepth: calldepth,
		pc:        pc,
	}
}

// Debug logs a debug-level message with optional arguments.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Debug(msg string, args ...any) {
	source := sourceWithStackTrace(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Debug(msg, args...)
}

// DebugContext logs a debug-level message with optional arguments and context.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	source := sourceWithStackTrace(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.DebugContext(ctx, msg, args...)
}

// Warn logs a warning-level message with optional arguments.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Warn(msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Warn(msg, args...)
}

// WarnContext logs a warning-level message with optional arguments and context.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.WarnContext(ctx, msg, args...)
}

// Error logs an error-level message with optional arguments.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Error(msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Error(msg, args...)
}

// ErrorContext logs an error-level message with optional arguments and context.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.ErrorContext(ctx, msg, args...)
}

// Panic logs a panic-level message, then panics with the message.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Panic(msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	ctx := context.Background()
	slog.Log(ctx, LevelPanic, msg, args...)
	panic(msg)
}

// PanicC logs a panic-level message with context, then panics with the context.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) PanicContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Log(ctx, LevelPanic, msg, args...)
	panic(ctx)
}

// Fatal logs a fatal-level message, then exits the application.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Fatal(msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	ctx := context.Background()
	slog.Log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

// FatalC logs a fatal-level message with context, then exits the application.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

// Print logs a trace-level message with optional arguments.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Print(msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	ctx := context.Background()
	slog.Log(ctx, LevelTrace, msg, args...)
}

// PrintC logs a trace-level message with context and optional arguments.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) PrintContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Log(ctx, LevelTrace, msg, args...)
}

// Info logs an info-level message with optional arguments.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) Info(msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.Info(msg, args...)
}

// InfoC logs an info-level message with context and optional arguments.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	source := source(l.calldepth+1, l.pc)
	args = append(args, source)
	slog.InfoContext(ctx, msg, args...)
}

// StackTrace provides a stack trace of up to 10 layers from where the error or incident was generated.
func (l *Logger) StackTrace() slog.Attr {
	return sourceWithStackTrace(l.calldepth+1, l.pc)
}

// SetCalldepth configures the number of stack frames to ascend for logging.
// calldepth: The number of stack frames to ascend.
func (l *Logger) SetCalldepth(calldepth int) {
	l.calldepth = calldepth
}

// SetLevel sets the logging level for the Logger instance.
// This method updates the log level to the specified level and returns the previous log level.
//
// Parameters:
//
//	level (slog.Level) - The new log level to be set. This determines the severity of the logs
//	                     that will be captured. Common log levels include DEBUG, INFO, WARN, and ERROR.
//
// Returns:
//
//	oldLevel (slog.Level) - The previous log level before it was updated. This can be used to restore
//	                        the previous log level if needed.
//
// Example usage:
//
//	logger := &Logger{}
//	oldLevel := logger.SetLevel(slog.INFO)
//	// The log level is now set to INFO
//	// You can restore the old level if needed
//	logger.SetLevel(oldLevel)
func (l *Logger) SetLevel(level slog.Level) (oldLevel slog.Level) {
	return slog.SetLogLoggerLevel(level)
}
