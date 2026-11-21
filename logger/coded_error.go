package logger

import (
	stderrors "errors"
	"log/slog"

	"github.com/jgolang/errors"
)

// codedErrorOf returns the first jgolang/errors error in err's chain, or nil.
// errors.AsType finds one behind wrappers such as fmt.Errorf without the
// allocation errors.As costs, which ReplaceAttr would otherwise add to every
// plain error attribute.
func codedErrorOf(err error) *errors.Error {
	coded, _ := stderrors.AsType[*errors.Error](err)
	return coded
}

// codedErrorGroup builds the error group, under key, for a jgolang/errors
// error. code and msg come from errors.CodeOf, which looks past errors whose
// code is nil; they are left out when the chain has no code, so logging never
// panics on the error path it is meant to report. debug and origin come from
// coded, the first jgolang/errors error, which is where the chain was last
// wrapped.
func codedErrorGroup(key string, coded *errors.Error) slog.Attr {
	debug := ""
	if coded.Wrapper != nil {
		debug = coded.Wrapper.Error()
	}

	code := errors.CodeOf(coded)
	if code == nil {
		return slog.Group(key, "debug", debug, "origin", coded.StackTrace())
	}

	return slog.Group(key,
		"code", code.Str(),
		"msg", code.Msg(),
		"debug", debug,
		"origin", coded.StackTrace(),
	)
}
