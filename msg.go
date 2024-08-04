package log

import "context"

func Print(msg string, args ...interface{}) {
	std.Print(msg, args...)
}

func Printf(ctx context.Context, msg string, args ...interface{}) {
	std.PrintContext(ctx, msg, args...)
}

func Info(msg string, args ...interface{}) {
	std.Info(msg, args...)
}

func InfoC(ctx context.Context, msg string, args ...interface{}) {
	std.InfoContext(ctx, msg, args...)
}
