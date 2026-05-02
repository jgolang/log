package log

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jgolang/errors"
	"github.com/jgolang/log/logger"
)

var level = new(slog.LevelVar)

var std = func() *logger.Logger {
	level.Set(slog.LevelDebug)
	return logger.New(3, nil, level)
}()

func NewJSONHandler() {
	std.SetJSONHandler(os.Stderr)
}

func NewTextHandler() {
	std.SetTextHandler(os.Stderr)
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
	return std.SetLevel(newLevel)
}

// Level return current log level
func Level() slog.Level {
	return level.Level()
}

func validateArgs(args ...any) (string, []any) {
	if len(args) == 0 {
		return "", nil
	}
	var msg string
	var rest []any
	switch v := args[0].(type) {
	case string:
		msg = v
		rest = args[1:]
	case error:
		msg = v.Error()
		rest = args[1:]
		err, ok := v.(*errors.Error)
		if ok {
			debug := ""
			if err.Wrapper != nil {
				debug = err.Wrapper.Error()
			}
			errGroup := slog.Group("error",
				"code", err.Code.Str(),
				"msg", err.Code.Msg(),
				"debug", debug,
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
