package log

import (
	"context"
)

// Error logs an error-level message using the global logger.
func Error(args ...any) {
	msg, attrs := validateArgs(args...)
	std.Error(msg, attrs...)
}

// ErrorC logs an error-level message with context using the global logger.
// ctx: The context for the log entry..
func ErrorC(ctx context.Context, args ...interface{}) {
	msg, attrs := validateArgs(args...)
	std.ErrorContext(ctx, msg, attrs...)
}

// Panic logs a panic-level message using the global logger and then panics.
func Panic(args ...any) {
	msg, attrs := validateArgs(args...)
	std.Panic(msg, attrs...)
}

// PanicC logs a panic-level message with context using the global logger and then panics.
// ctx: The context for the log entry.
func PanicC(ctx context.Context, args ...any) {
	msg, attrs := validateArgs(args...)
	std.PanicContext(ctx, msg, attrs...)
}

// Fatal logs a fatal-level message using the global logger and then calls os.Exit(1).
func Fatal(args ...any) {
	msg, attrs := validateArgs(args...)
	std.Fatal(msg, attrs...)
}

// FatalC logs a fatal-level message with context using the global logger and then calls os.Exit(1).
// ctx: The context for the log entry.
func FatalC(ctx context.Context, args ...any) {
	msg, attrs := validateArgs(args...)
	std.FatalContext(ctx, msg, attrs...)
}
