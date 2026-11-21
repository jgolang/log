package log

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"github.com/jgolang/errors"
	"github.com/jgolang/errors/codes"
	"github.com/jgolang/log/logger"
)

// recordingHandler stands in for a handler wrapper such as the otel one: it
// records the attributes it receives, flattened to dotted keys, before the
// handler it wraps encodes them.
type recordingHandler struct {
	slog.Handler
	mu    *sync.Mutex
	attrs map[string]any
}

func (h recordingHandler) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	record.Attrs(func(attr slog.Attr) bool {
		flattenAttr(h.attrs, "", attr)
		return true
	})
	return h.Handler.Handle(ctx, record)
}

func flattenAttr(into map[string]any, prefix string, attr slog.Attr) {
	value := attr.Value.Resolve()
	if value.Kind() == slog.KindGroup {
		for _, member := range value.Group() {
			flattenAttr(into, prefix+attr.Key+".", member)
		}
		return
	}
	into[prefix+attr.Key] = value.Any()
}

func newRecordingWrapper() (func(slog.Handler) slog.Handler, map[string]any) {
	attrs := map[string]any{}
	mu := new(sync.Mutex)
	return func(next slog.Handler) slog.Handler {
		return recordingHandler{Handler: next, mu: mu, attrs: attrs}
	}, attrs
}

func newCodedError() error {
	return errors.NewC(codes.New("PLN-111", "task not found"), "task 42 is missing")
}

func TestHandlerWrapperSeesExpandedCodedError(t *testing.T) {
	coded := newCodedError()
	tests := []struct {
		name string
		key  string
		log  func(*Logger)
	}{
		{name: "package first argument", key: "error", log: func(*Logger) { Error(coded) }},
		{name: "package later attribute", key: "err", log: func(*Logger) { Info("loading", "err", coded) }},
		{name: "instance attribute", key: "error", log: func(l *Logger) { l.Error("failed", "error", coded) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usePackageJSONHandler(t)
			wrap, attrs := newRecordingWrapper()
			SetHandlerWrapper(wrap)
			t.Cleanup(func() { SetHandlerWrapper(nil) })
			var buf bytes.Buffer
			instance := New(WithJSONHandler(&buf), WithHandlerWrapper(wrap))

			tt.log(instance)

			if attrs[tt.key+".code"] != "PLN-111" || attrs[tt.key+".msg"] != "task not found" {
				t.Fatalf("expected the wrapper to see %s.code and %s.msg, got %v", tt.key, tt.key, attrs)
			}
			if attrs[tt.key+".debug"] != "task 42 is missing" {
				t.Fatalf("expected the wrapper to see %s.debug, got %v", tt.key, attrs)
			}
		})
	}
}

func TestExpandedAttributeKeepsErrorText(t *testing.T) {
	err := fmt.Errorf("outer ctx: %w", newCodedError())
	var buf bytes.Buffer
	instance := New(WithJSONHandler(&buf))

	instance.Error("failed", "error", err)

	group, _ := decodeRecord(t, &buf)["error"].(map[string]any)
	if group["text"] != err.Error() {
		t.Fatalf("expected text %q, got %v", err.Error(), group)
	}
}

func TestFirstArgumentGroupHasNoText(t *testing.T) {
	record := logErrorRecord(t, fmt.Errorf("outer ctx: %w", newCodedError()))

	group, _ := record["error"].(map[string]any)
	if _, has := group["text"]; has {
		t.Fatalf("expected no text in the first-argument group, the message carries it; got %v", group)
	}
}

// redactedError hides the error it wraps from logs through slog.LogValuer.
type redactedError struct{ err error }

func (e redactedError) Error() string        { return e.err.Error() }
func (e redactedError) Unwrap() error        { return e.err }
func (e redactedError) LogValue() slog.Value { return slog.StringValue("redacted") }

func TestLogValuerErrorIsNotExpanded(t *testing.T) {
	err := redactedError{err: newCodedError()}
	tests := []struct {
		name string
		path []string
		log  func(*Logger)
	}{
		{name: "package first argument", path: []string{"error"}, log: func(*Logger) { Error(err) }},
		{name: "instance attribute", path: []string{"error"}, log: func(l *Logger) { l.Error("failed", "error", err) }},
		{name: "attribute inside a group", path: []string{"request", "error"}, log: func(l *Logger) {
			l.Error("failed", slog.Group("request", "error", err))
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := usePackageJSONHandler(t)
			instance := New(WithJSONHandler(buf))

			tt.log(instance)

			if got := errorAttrOf(decodeRecord(t, buf), tt.path...); got != "redacted" {
				t.Fatalf("expected LogValue to win over the coded error at %v, got %v", tt.path, got)
			}
		})
	}
}

func TestCodedErrorWithEmptyKeyIsNotExpanded(t *testing.T) {
	var buf bytes.Buffer
	instance := New(WithJSONHandler(&buf))

	instance.Error("failed", slog.Any("", newCodedError()))

	if msg := decodeRecord(t, &buf)["msg"]; msg != "failed" {
		t.Fatalf("expected the record message to stay %q, got %v in %q", "failed", msg, buf.String())
	}
}

func TestReplaceAttrLeavesErrorsUnchanged(t *testing.T) {
	coded := newCodedError()

	attr := logger.ReplaceAttr(nil, slog.Any("error", coded))

	if attr.Value.Kind() != slog.KindAny || attr.Value.Any() != coded {
		t.Fatalf("expected ReplaceAttr to keep the error value, got %v", attr)
	}
}
