package logger

import (
	"log/slog"

	"github.com/jgolang/log/internal/codederror"
)

// expandCodedErrors returns args with every attribute whose value is an error
// with a jgolang/errors error in its chain replaced by its group, also inside
// slog.Group values. It runs before the handler, so wrappers such as the otel
// handler see the group too. args is returned as is when nothing changes, so
// records without coded errors do not allocate; otherwise a copy is returned
// and the caller's slice is left untouched.
func expandCodedErrors(args []any) []any {
	var expanded []any
	for i := 0; i < len(args); {
		attr, consumed, changed := expandArg(args, i)
		if changed && expanded == nil {
			expanded = make([]any, i, len(args))
			copy(expanded, args[:i])
		}
		switch {
		case changed:
			expanded = append(expanded, attr)
		case expanded != nil:
			expanded = append(expanded, args[i:i+consumed]...)
		}
		i += consumed
	}
	if expanded == nil {
		return args
	}
	return expanded
}

// expandArg reads the attribute at args[i] the way slog does: a string key
// and the value after it, or a slog.Attr. It reports how many arguments the
// attribute uses and whether it was expanded.
func expandArg(args []any, i int) (attr slog.Attr, consumed int, changed bool) {
	switch arg := args[i].(type) {
	case string:
		if i+1 == len(args) {
			return slog.Attr{}, 1, false
		}
		attr, changed = expandAttr(slog.Any(arg, args[i+1]))
		return attr, 2, changed
	case slog.Attr:
		attr, changed = expandAttr(arg)
		return attr, 1, changed
	}
	return slog.Attr{}, 1, false
}

// expandAttr expands a coded error value, or the coded errors inside a group.
// An empty key is skipped: slog inlines a group with no key, and the error
// fields would then collide with the record's own msg.
func expandAttr(attr slog.Attr) (slog.Attr, bool) {
	switch attr.Value.Kind() {
	case slog.KindAny:
		err, isError := attr.Value.Any().(error)
		if !isError || attr.Key == "" {
			return attr, false
		}
		coded := codederror.Of(err)
		if coded == nil {
			return attr, false
		}
		return codederror.GroupWithText(attr.Key, err, coded), true
	case slog.KindGroup:
		members, changed := expandGroup(attr.Value.Group())
		if !changed {
			return attr, false
		}
		return slog.Attr{Key: attr.Key, Value: slog.GroupValue(members...)}, true
	}
	return attr, false
}

func expandGroup(members []slog.Attr) ([]slog.Attr, bool) {
	var expanded []slog.Attr
	for i, member := range members {
		next, changed := expandAttr(member)
		if changed && expanded == nil {
			expanded = make([]slog.Attr, i, len(members))
			copy(expanded, members[:i])
		}
		if expanded != nil {
			expanded = append(expanded, next)
		}
	}
	if expanded == nil {
		return members, false
	}
	return expanded, true
}
