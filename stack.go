package log

import "log/slog"

// StackTrace allows you to view the exact place where the error or incident originated within the code.
// Shows a trace of up to 10 layers from where the error or incident was generated.
func StackTrace() slog.Attr {
	return std.StackTrace()
}
