package log

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/jgolang/log/internal/codederror"
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

// NewJSONHandler makes the package-level logger write JSON records to
// os.Stderr, its default. A wrapper set with [SetHandlerWrapper] is kept.
func NewJSONHandler() {
	std.SetJSONHandler(os.Stderr)
}

// NewTextHandler makes the package-level logger write text records to
// os.Stderr. A wrapper set with [SetHandlerWrapper] is kept.
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
		msg = codederror.Text(v)
		// Clip so append never writes into spare capacity of the caller's slice.
		rest = slices.Clip(args[1:])
		// The group has no text attribute: msg already carries v.Error().
		if coded := codederror.Of(v); coded != nil {
			rest = append(rest, codederror.Group("error", coded))
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
