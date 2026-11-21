package log

import (
	"context"
	"log/slog"

	"github.com/jgolang/log/logger"
)

func Print(args ...interface{}) {
	if !levelEnabled(logger.LevelTrace) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.Print(msg, attrs...)
}

func PrintC(ctx context.Context, args ...interface{}) {
	if !levelEnabled(logger.LevelTrace) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.PrintContext(ctx, msg, attrs...)
}

func Info(args ...interface{}) {
	if !levelEnabled(slog.LevelInfo) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.Info(msg, attrs...)
}

func InfoC(ctx context.Context, args ...interface{}) {
	if !levelEnabled(slog.LevelInfo) {
		return
	}
	msg, attrs := validateArgs(args...)
	std.InfoContext(ctx, msg, attrs...)
}
