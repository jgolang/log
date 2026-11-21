// Package codederror builds the log group of a jgolang/errors error, shared by
// the package-level helpers and the Logger.
package codederror

import (
	"log/slog"

	"github.com/jgolang/errors"
)

// Of returns the first jgolang/errors error in err's chain, or nil. It walks
// the chain depth-first, as errors.As does, and returns nil when it meets an
// error that implements slog.LogValuer before a coded one: slog logs that
// value instead, so an error that hides what it wraps keeps it hidden, even
// behind wrappers such as fmt.Errorf. The walk does not allocate.
func Of(err error) *errors.Error {
	coded, _ := find(err)
	return coded
}

// find reports, as found, whether the walk ended at a coded error or at a
// slog.LogValuer; either one ends the walk across every branch.
func find(err error) (coded *errors.Error, found bool) {
	for err != nil {
		if _, isLogValuer := err.(slog.LogValuer); isLogValuer {
			return nil, true
		}
		if coded, isCoded := err.(*errors.Error); isCoded {
			return coded, true
		}
		switch wrapper := err.(type) {
		case interface{ Unwrap() error }:
			err = wrapper.Unwrap()
		case interface{ Unwrap() []error }:
			for _, branch := range wrapper.Unwrap() {
				if coded, found := find(branch); found {
					return coded, true
				}
			}
			return nil, false
		default:
			return nil, false
		}
	}
	return nil, false
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
	debug := coded.Detail()
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
