package log

import (
	stderrors "errors"
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

// codedErrorOf returns the first jgolang/errors error in err's chain, or nil.
// The type assertion keeps the common unwrapped case free of the allocation
// errors.As costs; errors.As finds one behind wrappers such as fmt.Errorf.
func codedErrorOf(err error) *errors.Error {
	if coded, ok := err.(*errors.Error); ok {
		return coded
	}

	var coded *errors.Error
	if stderrors.As(err, &coded) {
		return coded
	}

	return nil
}

// codedErrorGroup builds the "error" group for a jgolang/errors error. The code
// attributes are left out when the code is nil, so logging never panics on the
// error path it is meant to report.
func codedErrorGroup(coded *errors.Error) slog.Attr {
	debug := ""
	if coded.Wrapper != nil {
		debug = coded.Wrapper.Error()
	}

	if coded.Code == nil {
		return slog.Group("error", "debug", debug, "origin", coded.StackTrace())
	}

	return slog.Group("error",
		"code", coded.Code.Str(),
		"msg", coded.Code.Msg(),
		"debug", debug,
		"origin", coded.StackTrace(),
	)
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
		if coded := codedErrorOf(v); coded != nil {
			rest = append(rest, codedErrorGroup(coded))
		} else {
			rest = append(rest, "error", v)
		}
	default:
		// Keep the caller's value instead of dropping it (LOG-03).
		msg = fmt.Sprint(v)
		rest = args[1:]
	}
	return msg, rest
}
