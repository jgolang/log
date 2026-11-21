package log

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/jgolang/errors"
	"github.com/jgolang/errors/codes"
)

func BenchmarkLoggerInfo(b *testing.B) {
	coded := errors.NewC(codes.AppValidation, "invalid value")
	plain := stderrors.New("invalid value")
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
			name:      "plain_error_attr",
			addSource: true,
			level:     slog.LevelInfo,
			log: func(logger *Logger) {
				logger.Error("failed", "error", plain)
			},
		},
		{
			name:      "coded_error_attr",
			addSource: true,
			level:     slog.LevelInfo,
			log: func(logger *Logger) {
				logger.Error("failed", "error", coded)
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

func BenchmarkValidateArgs(b *testing.B) {
	coded := errors.NewC(codes.AppValidation, "invalid value")
	tests := []struct {
		name string
		args []any
	}{
		{name: "string", args: []any{"hello", "request_id", "abc"}},
		{name: "other_value", args: []any{42, "request_id", "abc"}},
		{name: "coded_error", args: []any{coded, "request_id", "abc"}},
		{name: "wrapped_coded_error", args: []any{fmt.Errorf("loading: %w", coded), "request_id", "abc"}},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				validateArgs(tt.args...)
			}
		})
	}
}

func BenchmarkPackageLevel(b *testing.B) {
	coded := errors.NewC(codes.AppValidation, "invalid value")
	tests := []struct {
		name string
		log  func()
	}{
		{name: "enabled_string", log: func() { Info("hello", "request_id", "abc") }},
		{name: "enabled_coded_error", log: func() { Error(coded, "request_id", "abc") }},
		{name: "disabled_string", log: func() { Debug("hello", "request_id", "abc") }},
		{name: "disabled_coded_error", log: func() { Debug(coded, "request_id", "abc") }},
		{name: "disabled_coded_error_context", log: func() { DebugC(context.Background(), coded, "request_id", "abc") }},
	}

	std.SetJSONHandler(io.Discard)
	old := SetLevel(slog.LevelInfo)
	b.Cleanup(func() {
		SetLevel(old)
		NewJSONHandler()
	})

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				tt.log()
			}
		})
	}
}
