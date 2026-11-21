package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jgolang/log/logger"
)

// taggingHandler is a minimal handler wrapper: it counts the records it sees
// and tags each one, so a test can tell the wrapper ran before the handler
// the logger built.
type taggingHandler struct {
	slog.Handler
	handled *int
}

func (h taggingHandler) Handle(ctx context.Context, record slog.Record) error {
	*h.handled++
	record.AddAttrs(slog.Bool("wrapped", true))
	return h.Handler.Handle(ctx, record)
}

// newTaggingWrapper returns a wrapper and the counter of records it handled.
func newTaggingWrapper() (func(slog.Handler) slog.Handler, *int) {
	handled := new(int)
	return func(next slog.Handler) slog.Handler {
		return taggingHandler{Handler: next, handled: handled}
	}, handled
}

func TestWithHandlerWrapperIsIndependentOfOptionOrder(t *testing.T) {
	tests := []struct {
		name  string
		order func(handler, wrapper Option) []Option
	}{
		{name: "wrapper after handler", order: func(handler, wrapper Option) []Option {
			return []Option{handler, wrapper}
		}},
		{name: "wrapper before handler", order: func(handler, wrapper Option) []Option {
			return []Option{wrapper, handler}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			wrap, _ := newTaggingWrapper()
			opts := tt.order(WithJSONHandler(&buf), WithHandlerWrapper(wrap))
			instance := New(append(opts, WithLevel(logger.LevelTrace))...)

			site := nextLine()
			instance.Print("hello")

			record := decodeRecord(t, &buf)
			if record["wrapped"] != true {
				t.Fatalf("expected the wrapper to tag the record, got %v", record)
			}
			if record["level"] != "TRACE" {
				t.Fatalf("expected ReplaceAttr to name the level TRACE, got %v", record["level"])
			}
			assertSource(t, record, site)
		})
	}
}

func TestWithHandlerWrapperKeepsLevel(t *testing.T) {
	var buf bytes.Buffer
	wrap, handled := newTaggingWrapper()
	instance := New(WithHandlerWrapper(wrap), WithJSONHandler(&buf), WithLevel(slog.LevelWarn))

	instance.Info("hidden")

	if buf.Len() != 0 || *handled != 0 {
		t.Fatalf("expected a disabled level to skip the wrapper and the handler, got %d records and %q", *handled, buf.String())
	}
}

func TestHandlerWrapperSurvivesHandlerChange(t *testing.T) {
	wrap, _ := newTaggingWrapper()
	instance := New(WithHandlerWrapper(wrap))
	var buf bytes.Buffer

	instance.SetTextHandler(&buf)
	instance.Info("hello")

	if !strings.Contains(buf.String(), "wrapped=true") {
		t.Fatalf("expected the wrapper on the new text handler, got %q", buf.String())
	}
}

func TestNilHandlerWrapperRemovesWrapper(t *testing.T) {
	var buf bytes.Buffer
	wrap, handled := newTaggingWrapper()
	instance := New(WithJSONHandler(&buf), WithHandlerWrapper(wrap))

	instance.SetHandlerWrapper(nil)
	instance.Info("hello")

	if *handled != 0 || strings.Contains(buf.String(), "wrapped") {
		t.Fatalf("expected no wrapper, got %d records and %q", *handled, buf.String())
	}
}

func TestHandlerWrapperReturningNilKeepsBuiltHandler(t *testing.T) {
	var buf bytes.Buffer
	instance := New(WithJSONHandler(&buf), WithHandlerWrapper(func(slog.Handler) slog.Handler {
		return nil
	}))

	instance.Info("hello")

	if !strings.Contains(buf.String(), `"msg":"hello"`) {
		t.Fatalf("expected the built handler to write the record, got %q", buf.String())
	}
}

func TestSetHandlerWrapperWrapsPackageLogger(t *testing.T) {
	buf := usePackageJSONHandler(t)
	wrap, _ := newTaggingWrapper()
	SetHandlerWrapper(wrap)
	t.Cleanup(func() { SetHandlerWrapper(nil) })

	site := nextLine()
	InfoC(context.Background(), "hello")

	record := decodeRecord(t, buf)
	if record["wrapped"] != true {
		t.Fatalf("expected the wrapper to tag the record, got %v", record)
	}
	assertSource(t, record, site)
}

func TestHandlerWrapperFactoryCanLogThroughSameLogger(t *testing.T) {
	tests := []struct {
		name    string
		rebuild func(instance *Logger, wrap func(slog.Handler) slog.Handler)
	}{
		{name: "SetHandlerWrapper", rebuild: func(instance *Logger, wrap func(slog.Handler) slog.Handler) {
			instance.SetHandlerWrapper(wrap)
		}},
		{name: "SetTextHandler after the wrapper", rebuild: func(instance *Logger, wrap func(slog.Handler) slog.Handler) {
			instance.SetHandlerWrapper(wrap)
			instance.SetTextHandler(io.Discard)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			instance := New(WithJSONHandler(&buf))
			wrap := func(next slog.Handler) slog.Handler {
				instance.Info("building wrapper")
				return next
			}
			done := make(chan struct{})

			go func() {
				defer close(done)
				tt.rebuild(instance, wrap)
			}()

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("rebuilding the handler deadlocked when the wrapper logged through the same logger")
			}
			if !strings.Contains(buf.String(), "building wrapper") {
				t.Fatalf("expected the wrapper's record, got %q", buf.String())
			}
		})
	}
}

func TestPanickingHandlerWrapperIsNotKept(t *testing.T) {
	var buf bytes.Buffer
	instance := New(WithJSONHandler(io.Discard))
	func() {
		defer func() { _ = recover() }()
		instance.SetHandlerWrapper(func(slog.Handler) slog.Handler { panic("broken wrapper") })
	}()

	instance.SetJSONHandler(&buf)
	instance.Info("hello")

	if !strings.Contains(buf.String(), `"msg":"hello"`) {
		t.Fatalf("expected the new handler to write unwrapped, got %q", buf.String())
	}
}

func TestSetHandlerWrapperOnZeroValueLoggerDoesNotPanic(t *testing.T) {
	wrap, handled := newTaggingWrapper()
	tests := []struct {
		name string
		wrap func(slog.Handler) slog.Handler
	}{
		{name: "nil wrapper", wrap: nil},
		{name: "wrapper", wrap: wrap},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var zero Logger

			zero.SetHandlerWrapper(tt.wrap)
			zero.Info("dropped")

			if *handled != 0 {
				t.Fatalf("expected no handler to run without a built handler, got %d records", *handled)
			}
		})
	}
}

// pcRecordingHandler records the call site that record.PC points at, which is
// what a wrapper with AddSource or an OpenTelemetry bridge reports.
type pcRecordingHandler struct {
	slog.Handler
	site *callSite
}

func (h pcRecordingHandler) Handle(ctx context.Context, record slog.Record) error {
	frame, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
	*h.site = callSite{
		function: frame.Function[strings.LastIndex(frame.Function, ".")+1:],
		file:     filepath.Base(frame.File),
		line:     frame.Line,
	}
	return h.Handler.Handle(ctx, record)
}

func newPCRecordingWrapper() (func(slog.Handler) slog.Handler, *callSite) {
	site := new(callSite)
	return func(next slog.Handler) slog.Handler {
		return pcRecordingHandler{Handler: next, site: site}
	}, site
}

func TestHandlerWrapperSeesCallerPC(t *testing.T) {
	t.Run("instance API", func(t *testing.T) {
		wrap, got := newPCRecordingWrapper()
		instance := New(WithJSONHandler(io.Discard), WithHandlerWrapper(wrap))

		want := nextLine()
		instance.Info("hello")

		if *got != want {
			t.Fatalf("record.PC = %+v, want %+v", *got, want)
		}
	})

	t.Run("instance API without source metadata", func(t *testing.T) {
		wrap, got := newPCRecordingWrapper()
		instance := New(WithJSONHandler(io.Discard), WithHandlerWrapper(wrap), WithSource(false))

		want := nextLine()
		instance.WarnContext(context.Background(), "hello")

		if *got != want {
			t.Fatalf("record.PC = %+v, want %+v", *got, want)
		}
	})

	t.Run("package-level helpers", func(t *testing.T) {
		usePackageJSONHandler(t)
		wrap, got := newPCRecordingWrapper()
		SetHandlerWrapper(wrap)
		t.Cleanup(func() { SetHandlerWrapper(nil) })

		want := nextLine()
		InfoC(context.Background(), "hello")

		if *got != want {
			t.Fatalf("record.PC = %+v, want %+v", *got, want)
		}
	})
}

// contextRecordingHandler records whether Handle received a nil context.
type contextRecordingHandler struct {
	slog.Handler
	gotNil *bool
}

func (h contextRecordingHandler) Handle(ctx context.Context, record slog.Record) error {
	*h.gotNil = ctx == nil
	return h.Handler.Handle(ctx, record)
}

func TestHandlerWrapperNeverSeesNilContext(t *testing.T) {
	gotNil := new(bool)
	instance := New(WithJSONHandler(io.Discard), WithHandlerWrapper(func(next slog.Handler) slog.Handler {
		return contextRecordingHandler{Handler: next, gotNil: gotNil}
	}))
	var nilContext context.Context

	instance.InfoContext(nilContext, "hello")

	if *gotNil {
		t.Fatal("expected a nil context to reach the handler as context.Background()")
	}
}
