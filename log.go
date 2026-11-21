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

// stdCalldepthOffset is the extra frame the package-level helpers add: they
// call the std Logger method, which reports its caller at calldepth 3.
const stdCalldepthOffset = 1

var std = func() *logger.Logger {
	level.Set(slog.LevelDebug)
	return logger.New(3+stdCalldepthOffset, nil, level)
}()

func NewJSONHandler() {
	std.SetJSONHandler(os.Stderr)
}

func NewTextHandler() {
	std.SetTextHandler(os.Stderr)
}

// SetCalldepth configures the number of stack frames to ascend for the
// source metadata of package-level logs. The default, 3, reports the code that
// calls log.Info and the other helpers; add one for each wrapper function
// between that code and the helper.
func SetCalldepth(calldepth int) {
	std.SetCalldepth(calldepth + stdCalldepthOffset)
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

// levelEnabled reports whether package-level logs at lvl are written. The
// helpers check it before validateArgs, whose error handling (message,
// errors.As, code lookup, stack) is wasted work for a disabled level. The std
// handlers read the same level variable, so the answers match.
func levelEnabled(lvl slog.Level) bool {
	return lvl >= level.Level()
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

// codedErrorGroup builds the "error" group for a jgolang/errors error. code
// and msg come from errors.CodeOf, which looks past errors whose code is nil;
// they are left out when the chain has no code, so logging never panics on
// the error path it is meant to report. debug and origin come from coded, the
// first jgolang/errors error, which is where the chain was last wrapped.
func codedErrorGroup(coded *errors.Error) slog.Attr {
	debug := ""
	if coded.Wrapper != nil {
		debug = coded.Wrapper.Error()
	}

	code := errors.CodeOf(coded)
	if code == nil {
		return slog.Group("error", "debug", debug, "origin", coded.StackTrace())
	}

	return slog.Group("error",
		"code", code.Str(),
		"msg", code.Msg(),
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
