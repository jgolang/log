// Package codederror builds the log group of a jgolang/errors error, shared by
// the package-level helpers and the Logger.
package codederror

import (
	stderrors "errors"
	"log/slog"

	"github.com/jgolang/errors"
)

// Of returns the first jgolang/errors error in err's chain, or nil. It also
// returns nil when err implements slog.LogValuer: slog resolves that value
// first, so an error that hides its wrapped detail keeps it hidden.
// errors.AsType finds a coded error behind wrappers such as fmt.Errorf without
// the allocation errors.As costs, which plain errors would otherwise pay.
func Of(err error) *errors.Error {
	if _, isLogValuer := err.(slog.LogValuer); isLogValuer {
		return nil
	}
	coded, _ := stderrors.AsType[*errors.Error](err)
	return coded
}

// Group builds the group, under key, for coded: code, msg, debug and origin.
// code and msg come from errors.CodeOf, which looks past errors whose code is
// nil; they are left out when the chain has no code, so logging never panics
// on the error path it is meant to report. debug and origin come from coded,
// the first jgolang/errors error, which is where the chain was last wrapped.
func Group(key string, coded *errors.Error) slog.Attr {
	return slog.Attr{Key: key, Value: slog.GroupValue(fields(coded)...)}
}

// GroupWithText is Group with a leading text attribute holding err.Error(),
// so the context added by wrappers outside coded is not lost when the error
// is not also the record's message.
func GroupWithText(key string, err error, coded *errors.Error) slog.Attr {
	attrs := append([]slog.Attr{slog.String("text", err.Error())}, fields(coded)...)
	return slog.Attr{Key: key, Value: slog.GroupValue(attrs...)}
}

func fields(coded *errors.Error) []slog.Attr {
	debug := ""
	if coded.Wrapper != nil {
		debug = coded.Wrapper.Error()
	}

	origin := slog.Attr{Key: "origin", Value: coded.StackTrace()}
	code := errors.CodeOf(coded)
	if code == nil {
		return []slog.Attr{slog.String("debug", debug), origin}
	}

	return []slog.Attr{
		slog.String("code", code.Str()),
		slog.String("msg", code.Msg()),
		slog.String("debug", debug),
		origin,
	}
}
