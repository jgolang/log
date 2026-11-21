package log

import (
	"context"
	"log/slog"
)

// Debug logs a debug-level message using the global logger.
// msg: The message to log.
// args: Additional arguments to format the message.
func Debug(args ...any) {
	if !levelEnabled(slog.LevelDebug) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.Debug(msg, attrs...)
}

// DebugC logs a debug-level message with context using the global logger.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func DebugC(ctx context.Context, args ...any) {
	if !levelEnabled(slog.LevelDebug) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.DebugContext(ctx, msg, attrs...)
}

// Warn logs a warning-level message using the global logger.
// msg: The message to log.
// args: Additional arguments to format the message.
func Warn(args ...any) {
	if !levelEnabled(slog.LevelWarn) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.Warn(msg, attrs...)
}

// WarnC logs a warning-level message with context using the global logger.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func WarnC(ctx context.Context, args ...interface{}) {
	if !levelEnabled(slog.LevelWarn) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.WarnContext(ctx, msg, attrs...)
}
