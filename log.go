package log

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jgolang/errors"
	"github.com/jgolang/log/logger"
)

var pc = make([]uintptr, 10)

var std = logger.New(3, pc)

var level = slog.LevelDebug

func init() {
	NewJSONHandler()
}

func NewJSONHandler() {
	opts := &slog.HandlerOptions{
		Level: level,
		// ReplaceAttr: logger.ReplaceAttr,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

func NewTextHandler() {
	opts := &slog.HandlerOptions{
		Level: level,
		// ReplaceAttr: logger.ReplaceAttr,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

// SetCalldepth configure the number of stack frames
// to ascend, with 0 identifying the caller of Caller for default loggin
func SetCalldepth(calldepth int) {
	std.SetCalldepth(calldepth)
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
func SetLevel(newLevel slog.Level) (oldLevel slog.Level) {
	level = newLevel
	return std.SetLevel(newLevel)
}

// Level return current log level
func Level() slog.Level {
	return level
}

func validateArgs(args ...any) (string, []any) {
	if len(args) == 0 {
		return "", nil
	}
	var msg string
	var rest []interface{}
	switch v := args[0].(type) {
	case string:
		msg = v
		rest = args[1:]
	case error:
		msg = v.Error()
		rest = args[1:]
		err, ok := v.(*errors.Error)
		if ok {
			errGroup := slog.Group("error",
				"code", err.Code.Str(),
				"msg", err.Code.Msg(),
				"debug", err.Wrapper.Error(),
				"origin", err.StackTrace(),
			)
			rest = append(rest, errGroup)
		} else {
			rest = append(rest, "error", v)
		}
	default:
		msg = fmt.Sprintf("Unknown type: %T", v)
		rest = args[1:]
	}
	return msg, rest
}
