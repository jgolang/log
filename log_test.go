package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestSetLevelUpdatesCurrentLevel(t *testing.T) {
	old := SetLevel(slog.LevelWarn)
	t.Cleanup(func() {
		SetLevel(old)
	})

	if got := Level(); got != slog.LevelWarn {
		t.Fatalf("Level() = %v, want %v", got, slog.LevelWarn)
	}
}

func TestPanicContextPanicsWithMessage(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered != "boom" {
			t.Fatalf("panic = %v, want %q", recovered, "boom")
		}
	}()

	std.PanicContext(context.Background(), "boom")
}

func TestInfoUsesConfiguredHandler(t *testing.T) {
	var buf bytes.Buffer
	std.SetJSONHandler(&buf)
	t.Cleanup(NewJSONHandler)

	Info("hello")

	if !strings.Contains(buf.String(), "\"msg\":\"hello\"") {
		t.Fatalf("expected message in output, got %q", buf.String())
	}
}

func TestNewSupportsDisablingSource(t *testing.T) {
	var buf bytes.Buffer
	instance := New(
		WithJSONHandler(&buf),
		WithSource(false),
	)

	instance.Info("hello")

	if strings.Contains(buf.String(), "\"source\"") {
		t.Fatalf("expected source metadata to be disabled, got %q", buf.String())
	}
}

func TestDebugStackTraceIsOptIn(t *testing.T) {
	var withoutStack bytes.Buffer
	instance := New(
		WithJSONHandler(&withoutStack),
		WithLevel(slog.LevelDebug),
	)
	instance.Debug("hello")

	if strings.Contains(withoutStack.String(), "stack_trace") {
		t.Fatalf("expected debug stack trace to be disabled by default, got %q", withoutStack.String())
	}

	var withStack bytes.Buffer
	instanceWithStack := New(
		WithJSONHandler(&withStack),
		WithLevel(slog.LevelDebug),
		WithDebugStackTrace(true),
	)
	instanceWithStack.Debug("hello")

	if !strings.Contains(withStack.String(), "stack_trace") {
		t.Fatalf("expected debug stack trace when enabled, got %q", withStack.String())
	}
}
