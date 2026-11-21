package logger

import (
	"log/slog"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
	LevelPanic = slog.Level(13)
)

var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
	LevelPanic: "PANIC",
}

// ReplaceAttr normalizes custom levels and expands errors before the handler
// encodes them. Any attribute whose value is an error with a jgolang/errors
// error in its chain, also inside a group, becomes a group under the same key
// with code, msg, debug and origin; other errors are left as they are.
func ReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Value.Kind() != slog.KindAny {
		return attr
	}

	switch value := attr.Value.Any().(type) {
	case slog.Level:
		if name, found := LevelNames[value]; found && attr.Key == slog.LevelKey {
			attr.Value = slog.StringValue(name)
		}
	case error:
		if coded := codedErrorOf(value); coded != nil {
			return codedErrorGroup(attr.Key, coded)
		}
	}

	return attr
}
