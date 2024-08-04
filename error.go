package log

import (
	"context"
)

// Error logs an error-level message using the global logger.
// msg: The message to log.
// args: Additional arguments to format the message.
func Error(msg string, args ...any) {
	std.Error(msg, args...)
}

// ErrorC logs an error-level message with context using the global logger.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func ErrorC(ctx context.Context, msg string, args ...interface{}) {
	std.ErrorContext(ctx, msg, args...)
}

// Panic logs a panic-level message using the global logger and then panics.
// msg: The message to log.
// args: Additional arguments to format the message.
func Panic(msg string, args ...any) {
	std.Panic(msg, args...)
}

// PanicC logs a panic-level message with context using the global logger and then panics.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func PanicC(ctx context.Context, msg string, args ...any) {
	std.PanicContext(ctx, msg, args...)
}

// Fatal logs a fatal-level message using the global logger and then calls os.Exit(1).
// msg: The message to log.
// args: Additional arguments to format the message.
func Fatal(msg string, args ...any) {
	std.Fatal(msg, args...)
}

// FatalC logs a fatal-level message with context using the global logger and then calls os.Exit(1).
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func FatalC(ctx context.Context, msg string, args ...any) {
	std.FatalContext(ctx, msg, args...)
}
