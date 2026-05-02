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
