package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

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
