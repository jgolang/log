package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
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

func TestInstanceJSONOutputIncludesLevelMessageAndSource(t *testing.T) {
	var buf bytes.Buffer
	instance := New(
		WithJSONHandler(&buf),
		WithLevel(slog.LevelInfo),
	)

	instance.Info("hello", "request_id", "abc")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output = %q", err, buf.String())
	}
	if entry["level"] != "INFO" {
		t.Fatalf("level = %v, want INFO", entry["level"])
	}
	if entry["msg"] != "hello" {
		t.Fatalf("msg = %v, want hello", entry["msg"])
	}
	if entry["request_id"] != "abc" {
		t.Fatalf("request_id = %v, want abc", entry["request_id"])
	}
	if _, ok := entry["source"].(map[string]any); !ok {
		t.Fatalf("expected source object in output, got %v", entry["source"])
	}
}

func TestDisabledLevelDoesNotWrite(t *testing.T) {
	var buf bytes.Buffer
	instance := New(
		WithJSONHandler(&buf),
		WithLevel(slog.LevelWarn),
	)

	instance.Info("hidden")

	if got := buf.String(); got != "" {
		t.Fatalf("expected no output for disabled level, got %q", got)
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

func TestLoggerSupportsConcurrentConfigurationAndLogging(t *testing.T) {
	var buf bytes.Buffer
	instance := New(
		WithJSONHandler(&buf),
		WithLevel(slog.LevelDebug),
	)

	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			instance.SetSource(i%2 == 0)
			instance.SetDebugStackTrace(i%3 == 0)
			instance.SetCalldepth(3)
		}(i)
		go func() {
			defer wg.Done()
			instance.Debug("hello")
		}()
	}
	wg.Wait()
}
