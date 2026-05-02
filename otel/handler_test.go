package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type recordedEvent struct {
	name       string
	attrs      []attribute.KeyValue
	stackTrace bool
}

type recordingSpan struct {
	noop.Span

	spanContext       trace.SpanContext
	events            []recordedEvent
	statusCode        codes.Code
	statusDescription string
}

func (s *recordingSpan) IsRecording() bool {
	return true
}

func (s *recordingSpan) SpanContext() trace.SpanContext {
	return s.spanContext
}

func (s *recordingSpan) AddEvent(name string, options ...trace.EventOption) {
	cfg := trace.NewEventConfig(options...)
	s.events = append(s.events, recordedEvent{
		name:       name,
		attrs:      cfg.Attributes(),
		stackTrace: cfg.StackTrace(),
	})
}

func (s *recordingSpan) SetStatus(code codes.Code, description string) {
	s.statusCode = code
	s.statusDescription = description
}

func TestBaggageIsDisabledByDefault(t *testing.T) {
	var buf bytes.Buffer
	handler := New(slog.NewJSONHandler(&buf, nil))
	logger := slog.New(handler)

	member, err := baggage.NewMember("tenant", "acme")
	if err != nil {
		t.Fatalf("NewMember() error = %v", err)
	}

	bag, err := baggage.New(member)
	if err != nil {
		t.Fatalf("baggage.New() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(context.Background(), bag), "hello")

	if strings.Contains(buf.String(), "\"tenant\"") {
		t.Fatalf("expected baggage to be omitted by default, got %q", buf.String())
	}
}

func TestBaggageAllowList(t *testing.T) {
	var buf bytes.Buffer
	handler := New(
		slog.NewJSONHandler(&buf, nil),
		WithNoBaggage(false),
		WithBaggageAllowList("tenant"),
	)
	logger := slog.New(handler)

	tenant, err := baggage.NewMember("tenant", "acme")
	if err != nil {
		t.Fatalf("NewMember(tenant) error = %v", err)
	}
	token, err := baggage.NewMember("token", "secret")
	if err != nil {
		t.Fatalf("NewMember(token) error = %v", err)
	}

	bag, err := baggage.New(tenant, token)
	if err != nil {
		t.Fatalf("baggage.New() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(context.Background(), bag), "hello")

	output := buf.String()
	if !strings.Contains(output, "\"tenant\":\"acme\"") {
		t.Fatalf("expected allow-listed baggage in output, got %q", output)
	}
	if strings.Contains(output, "\"token\":\"secret\"") {
		t.Fatalf("expected non-allow-listed baggage to be omitted, got %q", output)
	}
}

func TestBaggageDenyList(t *testing.T) {
	var buf bytes.Buffer
	handler := New(
		slog.NewJSONHandler(&buf, nil),
		WithNoBaggage(false),
		WithBaggageDenyList("token"),
	)
	logger := slog.New(handler)

	tenant, err := baggage.NewMember("tenant", "acme")
	if err != nil {
		t.Fatalf("NewMember(tenant) error = %v", err)
	}
	token, err := baggage.NewMember("token", "secret")
	if err != nil {
		t.Fatalf("NewMember(token) error = %v", err)
	}
	bag, err := baggage.New(tenant, token)
	if err != nil {
		t.Fatalf("baggage.New() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(context.Background(), bag), "hello")

	output := buf.String()
	if !strings.Contains(output, "\"tenant\":\"acme\"") {
		t.Fatalf("expected non-denied baggage in output, got %q", output)
	}
	if strings.Contains(output, "\"token\":\"secret\"") {
		t.Fatalf("expected denied baggage to be omitted, got %q", output)
	}
}

func TestBaggageFilterCanRedactValues(t *testing.T) {
	var buf bytes.Buffer
	handler := New(
		slog.NewJSONHandler(&buf, nil),
		WithNoBaggage(false),
		WithBaggageFilter(func(key, value string) (string, bool) {
			if key == "token" {
				return "redacted", true
			}
			return value, true
		}),
	)
	logger := slog.New(handler)

	token, err := baggage.NewMember("token", "secret")
	if err != nil {
		t.Fatalf("NewMember(token) error = %v", err)
	}
	bag, err := baggage.New(token)
	if err != nil {
		t.Fatalf("baggage.New() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(context.Background(), bag), "hello")

	output := buf.String()
	if !strings.Contains(output, "\"token\":\"redacted\"") {
		t.Fatalf("expected redacted baggage in output, got %q", output)
	}
	if strings.Contains(output, "secret") {
		t.Fatalf("expected original secret to be omitted, got %q", output)
	}
}

func TestHandleAddsTraceCorrelationSpanEventAndErrorStatus(t *testing.T) {
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		t.Fatalf("TraceIDFromHex() error = %v", err)
	}
	spanID, err := trace.SpanIDFromHex("0102030405060708")
	if err != nil {
		t.Fatalf("SpanIDFromHex() error = %v", err)
	}

	var buf bytes.Buffer
	handler := New(slog.NewJSONHandler(&buf, nil))
	logger := slog.New(handler)
	span := &recordingSpan{
		spanContext: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		}),
	}
	ctx := trace.ContextWithSpan(context.Background(), span)

	logger.ErrorContext(ctx, "failed", "request_id", "abc")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output = %q", err, buf.String())
	}
	if entry[TraceIDKey] != traceID.String() {
		t.Fatalf("trace_id = %v, want %s", entry[TraceIDKey], traceID.String())
	}
	if entry[SpanIDKey] != spanID.String() {
		t.Fatalf("span_id = %v, want %s", entry[SpanIDKey], spanID.String())
	}
	if len(span.events) != 1 {
		t.Fatalf("events len = %d, want 1", len(span.events))
	}
	if span.events[0].name != "log.error" {
		t.Fatalf("event name = %q, want log.error", span.events[0].name)
	}
	if got := attrValue(span.events[0].attrs, "request_id"); got != "abc" {
		t.Fatalf("event request_id = %q, want abc", got)
	}
	if span.statusCode != codes.Error {
		t.Fatalf("status code = %v, want %v", span.statusCode, codes.Error)
	}
	if span.statusDescription != "failed" {
		t.Fatalf("status description = %q, want failed", span.statusDescription)
	}
}

func attrValue(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}
