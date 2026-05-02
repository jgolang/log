package log

import (
	"io"
	"log/slog"
	"testing"
)

func BenchmarkLoggerInfo(b *testing.B) {
	tests := []struct {
		name       string
		addSource  bool
		debugStack bool
		level      slog.Level
		log        func(*Logger)
	}{
		{
			name:      "source_on",
			addSource: true,
			level:     slog.LevelInfo,
			log: func(logger *Logger) {
				logger.Info("hello", "request_id", "abc")
			},
		},
		{
			name:      "source_off",
			addSource: false,
			level:     slog.LevelInfo,
			log: func(logger *Logger) {
				logger.Info("hello", "request_id", "abc")
			},
		},
		{
			name:       "debug_stack_on",
			addSource:  true,
			debugStack: true,
			level:      slog.LevelDebug,
			log: func(logger *Logger) {
				logger.Debug("hello", "request_id", "abc")
			},
		},
		{
			name:      "disabled_level",
			addSource: true,
			level:     slog.LevelWarn,
			log: func(logger *Logger) {
				logger.Info("hello", "request_id", "abc")
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			logger := New(
				WithJSONHandler(io.Discard),
				WithLevel(tt.level),
				WithSource(tt.addSource),
				WithDebugStackTrace(tt.debugStack),
			)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tt.log(logger)
			}
		})
	}
}
