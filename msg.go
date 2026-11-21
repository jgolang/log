package log

import (
	"context"
	"log/slog"

	"github.com/jgolang/log/logger"
)

// Print logs a trace-level message using the package-level logger. The trace
// level ([logger.LevelTrace]) is below the default debug level, so nothing is
// written until [SetLevel] lowers the level to it. Arguments are handled as
// described in the package documentation.
func Print(args ...interface{}) {
	if !levelEnabled(logger.LevelTrace) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.Print(msg, attrs...)
}

// PrintC is like [Print] and passes ctx to the handler, for example for trace
// correlation.
func PrintC(ctx context.Context, args ...interface{}) {
	if !levelEnabled(logger.LevelTrace) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.PrintContext(ctx, msg, attrs...)
}

// Info logs an info-level message using the package-level logger. Arguments
// are handled as described in the package documentation.
func Info(args ...interface{}) {
	if !levelEnabled(slog.LevelInfo) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.Info(msg, attrs...)
}

// InfoC is like [Info] and passes ctx to the handler, for example for trace
// correlation.
func InfoC(ctx context.Context, args ...interface{}) {
	if !levelEnabled(slog.LevelInfo) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.InfoContext(ctx, msg, attrs...)
}
