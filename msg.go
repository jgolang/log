package log

import "context"

func Print(args ...interface{}) {
	msg, attrs := validateArgs(args...)
	std.Print(msg, attrs...)
}

func Printf(ctx context.Context, args ...interface{}) {
	msg, attrs := validateArgs(args...)
	std.PrintContext(ctx, msg, attrs...)
}

func Info(args ...interface{}) {
	msg, attrs := validateArgs(args...)
	std.Info(msg, attrs...)
}

func InfoC(ctx context.Context, args ...interface{}) {
	msg, attrs := validateArgs(args...)
	std.InfoContext(ctx, msg, attrs...)
}
