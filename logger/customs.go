package logger

import (
	"log/slog"
)

// Levels this package adds to the slog levels. Trace is below slog.LevelDebug;
// Fatal and Panic are above slog.LevelError.
const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
	LevelPanic = slog.Level(13)
)

// LevelNames maps the custom levels to the names [ReplaceAttr] writes for them.
var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
	LevelPanic: "PANIC",
}

// ReplaceAttr normalizes custom levels before the handler encodes them.
func ReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key != slog.LevelKey {
		return attr
	}

	level, ok := attr.Value.Any().(slog.Level)
	if !ok {
		return attr
	}

	if name, found := LevelNames[level]; found {
		attr.Value = slog.StringValue(name)
	}

	return attr
}
