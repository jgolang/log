package log

import "context"

// Debug logs a debug-level message using the global logger.
// msg: The message to log.
// args: Additional arguments to format the message.
func Debug(msg string, args ...any) {
	std.Debug(msg, args...)
}

// DebugC logs a debug-level message with context using the global logger.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func DebugC(ctx context.Context, msg string, args ...any) {
	std.DebugContext(ctx, msg, args...)
}

// Warn logs a warning-level message using the global logger.
// msg: The message to log.
// args: Additional arguments to format the message.
func Warn(msg string, args ...any) {
	std.Warn(msg, args...)
}

// WarnC logs a warning-level message with context using the global logger.
// ctx: The context for the log entry.
// msg: The message to log.
// args: Additional arguments to format the message.
func WarnC(ctx context.Context, msg string, args ...interface{}) {
	std.WarnContext(ctx, msg, args...)
}
