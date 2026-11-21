package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

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

func newSpanLogger(t *testing.T) (*slog.Logger, context.Context, *recordingSpan) {
	t.Helper()
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)
	logger := slog.New(New(slog.NewJSONHandler(&bytes.Buffer{}, nil)))
	return logger, ctx, span
}

func eventAttr(t *testing.T, span *recordingSpan, key string) attribute.Value {
	t.Helper()
	if len(span.events) != 1 {
		t.Fatalf("events len = %d, want 1", len(span.events))
	}
	for _, attr := range span.events[0].attrs {
		if string(attr.Key) == key {
			return attr.Value
		}
	}
	t.Fatalf("event attribute %q not found in %v", key, span.events[0].attrs)
	return attribute.Value{}
}

func TestSpanEventAttributeKinds(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC)
	tests := []struct {
		name string
		attr slog.Attr
		want attribute.Value
	}{
		{"string", slog.String("k", "v"), attribute.StringValue("v")},
		{"bool", slog.Bool("k", true), attribute.BoolValue(true)},
		{"int64", slog.Int64("k", -7), attribute.Int64Value(-7)},
		{"float64", slog.Float64("k", 1.5), attribute.Float64Value(1.5)},
		{"time", slog.Time("k", when), attribute.StringValue(when.Format(time.RFC3339Nano))},
		{"duration as int64 nanoseconds", slog.Duration("k", 1500*time.Millisecond), attribute.Int64Value(1_500_000_000)},
		{"uint64 that fits int64", slog.Uint64("k", 42), attribute.Int64Value(42)},
		{"uint64 max int64", slog.Uint64("k", math.MaxInt64), attribute.Int64Value(math.MaxInt64)},
		{"uint64 above int64 as string", slog.Uint64("k", math.MaxUint64), attribute.StringValue("18446744073709551615")},
		{"string slice", slog.Any("k", []string{"a", "b"}), attribute.StringSliceValue([]string{"a", "b"})},
		{"error", slog.Any("k", errors.New("boom")), attribute.StringValue("ERROR: boom")},
		{"any as JSON", slog.Any("k", struct{ A int }{1}), attribute.StringValue(`{"A":1}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx, span := newSpanLogger(t)

			logger.LogAttrs(ctx, slog.LevelInfo, "msg", tt.attr)

			if got := eventAttr(t, span, "k"); got != tt.want {
				t.Fatalf("value = %v (%v), want %v (%v)", got.Emit(), got.Type(), tt.want.Emit(), tt.want.Type())
			}
		})
	}
}

func TestSpanEventNestedGroupKeys(t *testing.T) {
	tests := []struct {
		name    string
		logger  func(*slog.Logger) *slog.Logger
		attr    slog.Attr
		wantKey string
	}{
		{
			name:    "two level attr group",
			logger:  func(l *slog.Logger) *slog.Logger { return l },
			attr:    slog.Group("error", slog.Group("origin", slog.String("frame_0", "v"))),
			wantKey: "error.origin.frame_0",
		},
		{
			name:    "three level attr group keeps order",
			logger:  func(l *slog.Logger) *slog.Logger { return l },
			attr:    slog.Group("a", slog.Group("b", slog.Group("c", slog.String("d", "v")))),
			wantKey: "a.b.c.d",
		},
		{
			name:    "inline group with empty key adds no prefix",
			logger:  func(l *slog.Logger) *slog.Logger { return l },
			attr:    slog.Group("", slog.String("k", "v")),
			wantKey: "k",
		},
		{
			name:    "handler groups prefix record attrs",
			logger:  func(l *slog.Logger) *slog.Logger { return l.WithGroup("req").WithGroup("http") },
			attr:    slog.String("k", "v"),
			wantKey: "req.http.k",
		},
		{
			name:    "handler groups and attr groups combine in order",
			logger:  func(l *slog.Logger) *slog.Logger { return l.WithGroup("req").With("x", 1).WithGroup("http") },
			attr:    slog.Group("h", slog.Group("i", slog.String("k", "v"))),
			wantKey: "req.http.h.i.k",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx, span := newSpanLogger(t)

			tt.logger(logger).LogAttrs(ctx, slog.LevelInfo, "msg", tt.attr)

			if got := eventAttr(t, span, tt.wantKey); got.AsString() != "v" {
				t.Fatalf("%s = %q, want v", tt.wantKey, got.AsString())
			}
		})
	}
}

func TestSpanEventValuesAreCapped(t *testing.T) {
	long := strings.Repeat("x", 100*1024)
	wantCapped := strings.Repeat("x", maxAttributeValueLen) + truncationMarker
	tests := []struct {
		name string
		attr slog.Attr
		want string
	}{
		{"short string unchanged", slog.String("k", "short"), "short"},
		{"exactly at cap unchanged", slog.String("k", long[:maxAttributeValueLen]), long[:maxAttributeValueLen]},
		{"long string", slog.String("k", long), wantCapped},
		{"long any", slog.Any("k", struct{ S string }{long}), `{"S":"` + long[:maxAttributeValueLen-6] + truncationMarker},
		{"long error", slog.Any("k", errors.New(long)), "ERROR: " + long[:maxAttributeValueLen-7] + truncationMarker},
		{"cut on a rune boundary", slog.String("k", "a"+strings.Repeat("é", maxAttributeValueLen)), "a" + strings.Repeat("é", (maxAttributeValueLen-1)/2) + truncationMarker},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx, span := newSpanLogger(t)

			logger.LogAttrs(ctx, slog.LevelInfo, "msg", tt.attr)

			if got := eventAttr(t, span, "k").AsString(); got != tt.want {
				t.Fatalf("value len = %d, want len %d", len(got), len(tt.want))
			}
		})
	}
}

func TestSpanEventStringSliceElementsAreCapped(t *testing.T) {
	logger, ctx, span := newSpanLogger(t)
	long := strings.Repeat("x", 2*maxAttributeValueLen)

	logger.InfoContext(ctx, "msg", slog.Any("k", []string{"a", long}))

	got := eventAttr(t, span, "k").AsStringSlice()
	want := []string{"a", long[:maxAttributeValueLen] + truncationMarker}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("slice lens = %d, want capped elements", len(got))
	}
}

func TestSpanEventSliceLengthIsCapped(t *testing.T) {
	const huge = 100_000
	ints := sequence(huge, func(i int) int { return i })
	int64s := sequence(huge, func(i int) int64 { return int64(i) })
	float64s := sequence(huge, func(i int) float64 { return float64(i) })
	bools := sequence(huge, func(i int) bool { return i%3 == 0 })
	strs := sequence(huge, strconv.Itoa)
	tests := []struct {
		name  string
		value func() any // a fresh copy, to check the logged slice is not modified
		want  attribute.Value
	}{
		{"[]int", func() any { return slices.Clone(ints) }, attribute.IntSliceValue(ints[:maxAttributeSliceLen])},
		{"[]int64", func() any { return slices.Clone(int64s) }, attribute.Int64SliceValue(int64s[:maxAttributeSliceLen])},
		{"[]float64", func() any { return slices.Clone(float64s) }, attribute.Float64SliceValue(float64s[:maxAttributeSliceLen])},
		{"[]bool", func() any { return slices.Clone(bools) }, attribute.BoolSliceValue(bools[:maxAttributeSliceLen])},
		{"[]string", func() any { return slices.Clone(strs) }, attribute.StringSliceValue(strs[:maxAttributeSliceLen])},
		{"exactly at cap unchanged", func() any { return slices.Clone(ints[:maxAttributeSliceLen]) }, attribute.IntSliceValue(ints[:maxAttributeSliceLen])},
		{"short slice unchanged", func() any { return []int{1, 2, 3} }, attribute.IntSliceValue([]int{1, 2, 3})},
	}
	sources := []struct {
		name  string
		write func(l *slog.Logger, ctx context.Context, value any)
	}{
		{"record attr", func(l *slog.Logger, ctx context.Context, value any) {
			l.InfoContext(ctx, "msg", slog.Any("k", value))
		}},
		{"With attr", func(l *slog.Logger, ctx context.Context, value any) {
			l.With(slog.Any("k", value)).InfoContext(ctx, "msg")
		}},
	}
	for _, tt := range tests {
		for _, source := range sources {
			t.Run(tt.name+" "+source.name, func(t *testing.T) {
				logger, ctx, span := newSpanLogger(t)
				value := tt.value()

				source.write(logger, ctx, value)

				if got := eventAttr(t, span, "k"); got != tt.want {
					t.Fatalf("value = %.80s, want its first elements %.80s", got.Emit(), tt.want.Emit())
				}
				if !reflect.DeepEqual(value, tt.value()) {
					t.Fatal("the caller's slice was modified")
				}
			})
		}
	}
}

// sequence returns n elements built from their index.
func sequence[T any](n int, element func(int) T) []T {
	values := make([]T, n)
	for i := range values {
		values[i] = element(i)
	}
	return values
}

func TestSpanEventStringSliceIsCappedInLengthAndElements(t *testing.T) {
	logger, ctx, span := newSpanLogger(t)
	long := strings.Repeat("x", 2*maxAttributeValueLen)
	values := make([]string, 2*maxAttributeSliceLen)
	for i := range values {
		values[i] = long
	}

	logger.InfoContext(ctx, "msg", slog.Any("k", values))

	got := eventAttr(t, span, "k").AsStringSlice()
	if len(got) != maxAttributeSliceLen {
		t.Fatalf("slice len = %d, want %d", len(got), maxAttributeSliceLen)
	}
	want := long[:maxAttributeValueLen] + truncationMarker
	for i, value := range got {
		if value != want {
			t.Fatalf("element %d len = %d, want %d", i, len(value), len(want))
		}
	}
	if len(values) != 2*maxAttributeSliceLen || values[0] != long {
		t.Fatal("the caller's slice was modified")
	}
}

func TestSpanEventMessageAndStatusAreCapped(t *testing.T) {
	logger, ctx, span := newSpanLogger(t)
	long := strings.Repeat("x", 2*maxAttributeValueLen)
	want := long[:maxAttributeValueLen] + truncationMarker

	logger.ErrorContext(ctx, long)

	if got := eventAttr(t, span, slog.MessageKey).AsString(); got != want {
		t.Fatalf("event message len = %d, want %d", len(got), len(want))
	}
	if span.statusDescription != want {
		t.Fatalf("status description len = %d, want %d", len(span.statusDescription), len(want))
	}
}

func TestCapStringKeepsInvalidUTF8UpToTheCap(t *testing.T) {
	value := strings.Repeat("\x80", 2000)

	capped := capString(value)

	want := maxAttributeValueLen + len(truncationMarker)
	if len(capped) != want {
		t.Fatalf("expected %d bytes, got %d", want, len(capped))
	}
}

func TestSpanEventRecordsHandlerAttrs(t *testing.T) {
	tests := []struct {
		name    string
		logger  func(*slog.Logger) *slog.Logger
		wantKey string
		want    attribute.Value
	}{
		{
			name:    "attr added with With",
			logger:  func(l *slog.Logger) *slog.Logger { return l.With("request_id", "r1") },
			wantKey: "request_id",
			want:    attribute.StringValue("r1"),
		},
		{
			name:    "kind converted as record attrs",
			logger:  func(l *slog.Logger) *slog.Logger { return l.With(slog.Duration("elapsed", time.Second)) },
			wantKey: "elapsed",
			want:    attribute.Int64Value(int64(time.Second)),
		},
		{
			name:    "attr after WithGroup gets the group prefix",
			logger:  func(l *slog.Logger) *slog.Logger { return l.WithGroup("req").With("id", 1) },
			wantKey: "req.id",
			want:    attribute.Int64Value(1),
		},
		{
			name:    "attr before WithGroup keeps the outer prefix",
			logger:  func(l *slog.Logger) *slog.Logger { return l.With("id", 1).WithGroup("req") },
			wantKey: "id",
			want:    attribute.Int64Value(1),
		},
		{
			name:    "group attr nests under handler groups",
			logger:  func(l *slog.Logger) *slog.Logger { return l.WithGroup("req").With(slog.Group("user", "id", "u1")) },
			wantKey: "req.user.id",
			want:    attribute.StringValue("u1"),
		},
		{
			name:    "long value is capped",
			logger:  func(l *slog.Logger) *slog.Logger { return l.With("k", strings.Repeat("x", 2*maxAttributeValueLen)) },
			wantKey: "k",
			want:    attribute.StringValue(strings.Repeat("x", maxAttributeValueLen) + truncationMarker),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx, span := newSpanLogger(t)

			tt.logger(logger).InfoContext(ctx, "msg")

			if got := eventAttr(t, span, tt.wantKey); got != tt.want {
				t.Fatalf("%s = %v, want %v", tt.wantKey, got.Emit(), tt.want.Emit())
			}
		})
	}
}

func TestSpanEventHandlerAttrsComeBeforeRecordAttrs(t *testing.T) {
	logger, ctx, span := newSpanLogger(t)

	logger.With("a", 1).With("b", 2).InfoContext(ctx, "msg", "c", 3)

	if len(span.events) != 1 {
		t.Fatalf("events len = %d, want 1", len(span.events))
	}
	var keys []string
	for _, attr := range span.events[0].attrs[3:] { // after message, level and time
		keys = append(keys, string(attr.Key))
	}
	if got := strings.Join(keys, ","); got != "a,b,c" {
		t.Fatalf("event attribute keys = %s, want a,b,c", got)
	}
}

func TestSpanEventHandlerAttrsAreNotSharedBetweenSiblings(t *testing.T) {
	logger, ctx, span := newSpanLogger(t)
	parent := logger.With("p", 1, "q", 2, "r", 3) // three attrs leave spare capacity on a plain append
	first := parent.With("first", 1)
	second := parent.With("second", 2)

	first.InfoContext(ctx, "msg")
	second.InfoContext(ctx, "msg")

	if len(span.events) != 2 {
		t.Fatalf("events len = %d, want 2", len(span.events))
	}
	tests := []struct {
		name    string
		attrs   []attribute.KeyValue
		want    []string
		notWant string
	}{
		{"first", span.events[0].attrs, []string{"p", "q", "r", "first"}, "second"},
		{"second", span.events[1].attrs, []string{"p", "q", "r", "second"}, "first"},
	}
	for _, tt := range tests {
		for _, key := range tt.want {
			if !hasAttr(tt.attrs, key) {
				t.Fatalf("%s logger event lacks %q: %v", tt.name, key, tt.attrs)
			}
		}
		if hasAttr(tt.attrs, tt.notWant) {
			t.Fatalf("%s logger event has sibling attr %q: %v", tt.name, tt.notWant, tt.attrs)
		}
	}
}

func TestSpanEventHandlerAttrsReachNextOnce(t *testing.T) {
	var buf bytes.Buffer
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)
	logger := slog.New(New(slog.NewJSONHandler(&buf, nil))).With("request_id", "r1")

	logger.InfoContext(ctx, "msg")

	if got := strings.Count(buf.String(), `"request_id"`); got != 1 {
		t.Fatalf("request_id appears %d times in %s, want 1", got, buf.String())
	}
}

func TestNoTraceEventsSkipsHandlerAttrs(t *testing.T) {
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)
	logger := slog.New(New(slog.NewJSONHandler(&bytes.Buffer{}, nil), WithNoTraceEvents(true))).With("request_id", "r1")

	logger.InfoContext(ctx, "msg")

	if len(span.events) != 0 {
		t.Fatalf("events len = %d, want 0", len(span.events))
	}
}

func hasAttr(attrs []attribute.KeyValue, key string) bool {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return true
		}
	}
	return false
}

type loginRequest struct {
	User     string
	Password string `json:"-"`
}

type redactingJSON string

func (redactingJSON) MarshalJSON() ([]byte, error) { return []byte(`"[REDACTED]"`), nil }

type redactingText string

func (redactingText) MarshalText() ([]byte, error) { return []byte("[REDACTED]"), nil }

type leakyStringer struct {
	Updates  chan int
	password string
}

func (s leakyStringer) String() string { return s.password }

type valueMarshaler struct{}

func (valueMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"[REDACTED]"`), nil }

type nilUnsafeError struct{ message string }

func (e *nilUnsafeError) Error() string { return e.message }

type panickingText struct{}

func (panickingText) MarshalText() ([]byte, error) { panic("boom") }

func TestSpanEventValuesFollowLogRedaction(t *testing.T) {
	var typedNil *nilUnsafeError
	tests := []struct {
		name string
		attr slog.Attr
		want string
	}{
		{"json dash field", slog.Any("k", loginRequest{User: "bob", Password: "hunter2"}), `{"User":"bob"}`},
		{"pointer to json dash field", slog.Any("k", &loginRequest{User: "bob", Password: "hunter2"}), `{"User":"bob"}`},
		{"MarshalJSON string is unquoted", slog.Any("k", redactingJSON("sk_live_123")), "[REDACTED]"},
		{"MarshalText", slog.Any("k", redactingText("sk_live_123")), "[REDACTED]"},
		{"unexported field stays out", slog.Any("k", struct{ token string }{"s3cr3t"}), "{}"},
		{"Stringer is not used when JSON fails", slog.Any("k", leakyStringer{password: "s3cr3t"}), "!ERROR: json: unsupported type: chan int"},
		{"typed nil marshaler", slog.Any("k", (*valueMarshaler)(nil)), "null"},
		{"no HTML escaping", slog.Any("k", struct{ URL string }{"a?b=1&c=<2>"}), `{"URL":"a?b=1&c=<2>"}`},
		{"map as JSON", slog.Any("k", map[string]int{"a": 1}), `{"a":1}`},
		{"nil", slog.Any("k", nil), "null"},
		{"typed nil error", slog.Any("k", error(typedNil)), "ERROR: <nil>"},
		{"panicking method", slog.Any("k", panickingText{}), "!PANIC: boom"},
		{"unsupported value", slog.Any("k", make(chan int)), "!ERROR: json: unsupported type: chan int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx, span := newSpanLogger(t)

			logger.LogAttrs(ctx, slog.LevelInfo, "msg", tt.attr)

			if got := eventAttr(t, span, "k").AsString(); got != tt.want {
				t.Fatalf("value = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSpanEventAndLogLineHideJSONDashField(t *testing.T) {
	var out bytes.Buffer
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)
	logger := slog.New(New(slog.NewJSONHandler(&out, nil)))

	logger.InfoContext(ctx, "login", "req", loginRequest{User: "bob", Password: "hunter2"})

	if strings.Contains(out.String(), "hunter2") {
		t.Fatalf("log line leaks the password: %s", out.String())
	}
	if got := eventAttr(t, span, "req").AsString(); strings.Contains(got, "hunter2") {
		t.Fatalf("span event leaks the password: %q", got)
	}
}
