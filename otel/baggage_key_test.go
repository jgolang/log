package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

func TestBaggageCannotOverrideRecordKeys(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(New(slog.NewJSONHandler(&out, nil), WithNoBaggage(false)))
	bag, err := baggage.Parse("level=INFO,msg=user%20logged%20out,time=0,source=x,trace_id=00000000000000000000000000000001,span_id=0000000000000001,order=forged,request_id=r1")
	if err != nil {
		t.Fatalf("baggage.Parse() error = %v", err)
	}

	logger.ErrorContext(baggage.ContextWithBaggage(context.Background(), bag), "payment fraud detected", "order", "o-1")

	var line map[string]any
	if err := json.Unmarshal(out.Bytes(), &line); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", out.String(), err)
	}
	want := map[string]any{"level": "ERROR", "msg": "payment fraud detected", "order": "o-1"}
	for key, value := range want {
		if line[key] != value {
			t.Errorf("%s = %v, want %v (line %s)", key, line[key], value, out.String())
		}
	}
	for _, key := range []string{"trace_id", "span_id", "source"} {
		if _, ok := line[key]; ok {
			t.Errorf("baggage set top-level %q: %s", key, out.String())
		}
	}
	if line["time"] == "0" {
		t.Errorf("baggage replaced time: %s", out.String())
	}
	bagGroup, ok := line[BaggageKey].(map[string]any)
	if !ok {
		t.Fatalf("baggage group = %v, want members under %q (line %s)", line[BaggageKey], BaggageKey, out.String())
	}
	for _, key := range []string{"level", "msg", "time", "source", "trace_id", "span_id", "order", "request_id"} {
		if _, ok := bagGroup[key]; !ok {
			t.Errorf("baggage member %q missing from the group (line %s)", key, out.String())
		}
	}
}

func TestBaggageDoesNotReplaceSpanIDs(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(New(slog.NewJSONHandler(&out, nil), WithNoBaggage(false)))
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0xaa},
		SpanID:  trace.SpanID{0xbb},
	})
	ctx := trace.ContextWithSpan(context.Background(), &recordingSpan{spanContext: spanContext})
	bag, err := baggage.Parse("trace_id=00000000000000000000000000000001,span_id=0000000000000001")
	if err != nil {
		t.Fatalf("baggage.Parse() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(ctx, bag), "x")

	// Before the fix the baggage IDs were top-level keys too; the decoder keeps
	// the last duplicate, so check the raw line.
	for _, key := range []string{TraceIDKey, SpanIDKey} {
		if got := strings.Count(out.String(), `"`+key+`"`); got != 2 {
			t.Errorf("%s keys = %d, want 2 (record and baggage group): %s", key, got, out.String())
		}
	}
	if !strings.Contains(out.String(), `"`+BaggageKey+`":{`) {
		t.Errorf("no %q group: %s", BaggageKey, out.String())
	}
	var line map[string]any
	if err := json.Unmarshal(out.Bytes(), &line); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", out.String(), err)
	}
	if line[TraceIDKey] != spanContext.TraceID().String() || line[SpanIDKey] != spanContext.SpanID().String() {
		t.Fatalf("trace/span IDs = %v/%v, want the span's (line %s)", line[TraceIDKey], line[SpanIDKey], out.String())
	}
}

func TestBaggageGroupOmittedWhenFilterDropsAllMembers(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(New(slog.NewJSONHandler(&out, nil), WithNoBaggage(false), WithBaggageAllowList("tenant")))
	bag, err := baggage.Parse("token=secret")
	if err != nil {
		t.Fatalf("baggage.Parse() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(context.Background(), bag), "x")

	if strings.Contains(out.String(), `"`+BaggageKey+`"`) {
		t.Fatalf("expected no baggage group, got %s", out.String())
	}
}

func TestSpanEventCarriesBaggageUnderGroupKey(t *testing.T) {
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)
	logger := slog.New(New(slog.NewJSONHandler(&bytes.Buffer{}, nil), WithNoBaggage(false)))
	bag, err := baggage.Parse("msg=forged,request_id=r1")
	if err != nil {
		t.Fatalf("baggage.Parse() error = %v", err)
	}

	logger.InfoContext(baggage.ContextWithBaggage(ctx, bag), "real")

	if got := eventAttr(t, span, BaggageKey+".request_id").AsString(); got != "r1" {
		t.Errorf("baggage.request_id = %q, want r1", got)
	}
	if got := eventAttr(t, span, "msg").AsString(); got != "real" {
		t.Errorf("msg = %q, want real", got)
	}
}
