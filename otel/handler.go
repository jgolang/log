// Package otel provides a log/slog handler that adds OpenTelemetry trace
// context, and optionally filtered baggage, to records written through
// github.com/jgolang/log. It is its own module, so programs that do not use it
// do not depend on OpenTelemetry; it has been since log v1.5.0. Upgrading from
// log v1.4.x or earlier, raise both modules together:
//
//	go get github.com/jgolang/log@latest github.com/jgolang/log/otel@latest
//
// Attach the handler to a jgolang logger with log.SetHandlerWrapper or the
// log.WithHandlerWrapper option (log v1.6.0 or later):
//
//	log.SetHandlerWrapper(otel.NewOtelHandler())
package otel

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDKey is the key used by the Otel handler
	// to inject the trace ID in the log record.
	TraceIDKey = "trace_id"
	// SpanIDKey is the key used by the Otel handler
	// to inject the span ID in the log record.
	SpanIDKey = "span_id"
	// SpanEventKey is the prefix key used by the Otel handler
	// to inject the log record in the recording span, as a span event.
	SpanEventKey = "log"
)

const (
	// maxAttributeValueLen caps, in bytes, every string the handler writes to
	// a span: event attribute values (including each []string element), the
	// event message and the error status description. Longer strings are cut
	// on a UTF-8 boundary and end with truncationMarker, so a capped value is
	// at most maxAttributeValueLen+len(truncationMarker) bytes. The next
	// handler still receives the full record.
	maxAttributeValueLen = 1024
	truncationMarker     = "…"
	// maxAttributeSliceLen caps the element count of every slice attribute
	// value ([]string, []int, []int64, []float64, []bool) the handler writes
	// to a span event; later elements are dropped. 128 is the same value as
	// the OpenTelemetry SDK's default per-event attribute count limit, which
	// counts attributes, not slice elements. []string elements are also
	// capped by maxAttributeValueLen.
	maxAttributeSliceLen = 128
)

// OtelHandler is an implementation of slog's Handler interface.
// Its role is to ensure correlation between logs and OTel spans
// by:
//
// 1. Adding otel span and trace IDs to the log record.
// 2. Adding otel context baggage members to the log record.
// 3. Setting slog record as otel span event.
// 4. Adding slog record attributes to the otel span event.
// 5. Setting span status based on slog record level (only if >= slog.LevelError).
//
// Span event attribute keys are the slog group path joined with ".", handler
// groups (WithGroup) first, e.g. "req.error.origin". slog.Duration values are
// recorded as int64 nanoseconds; slog.Uint64 values as int64, or as a decimal
// string when they exceed math.MaxInt64. String values longer than 1 KiB are
// truncated and end with "…"; slice values ([]string, []int, []int64,
// []float64, []bool) keep their first 128 elements. Attributes added with
// WithAttrs (logger.With) are recorded on every span event the derived handler
// adds, before the record's own attributes, with the same keys, kinds and
// caps; they are converted once, when WithAttrs is called.
type OtelHandler struct {
	// Next represents the next handler in the chain.
	Next slog.Handler
	// NoBaggage determines whether to add context baggage members to the log record.
	NoBaggage bool
	// NoTraceEvents determines whether to record an event for every log on the active trace.
	NoTraceEvents bool
	// BaggageFilter can allow, drop, or redact baggage members before they are
	// logged. The baggage filter options add to it; setting the field directly
	// replaces it.
	BaggageFilter BaggageFilter

	// groupPrefix is the dot-joined path of the groups opened with WithGroup.
	groupPrefix string
	// attrs are the span event attributes added with WithAttrs, converted
	// with the group prefix in effect when they were added.
	attrs []attribute.KeyValue
}

// OtelHandlerOpt configures an [OtelHandler] built with [New] or
// [NewOtelHandler].
type OtelHandlerOpt func(handler *OtelHandler)

// BaggageFilter returns the value to log and whether the member should be included.
type BaggageFilter func(key, value string) (string, bool)

// HandlerFn defines the handler used by slog.Handler as return value.
type HandlerFn func(slog.Handler) slog.Handler

// WithNoBaggage returns an OtelHandlerOpt, which sets the NoBaggage flag
func WithNoBaggage(noBaggage bool) OtelHandlerOpt {
	return func(handler *OtelHandler) {
		handler.NoBaggage = noBaggage
	}
}

// WithNoTraceEvents returns an OtelHandlerOpt, which sets the NoTraceEvents flag
func WithNoTraceEvents(noTraceEvents bool) OtelHandlerOpt {
	return func(handler *OtelHandler) {
		handler.NoTraceEvents = noTraceEvents
	}
}

// WithBaggageFilter adds a baggage filter for selective logging/redaction.
// Baggage filter options (WithBaggageFilter, WithBaggageAllowList,
// WithBaggageDenyList) compose in the order given: a member is logged only
// when every filter keeps it, and each filter receives the value returned by
// the previous one. A nil filter is ignored.
func WithBaggageFilter(filter BaggageFilter) OtelHandlerOpt {
	return func(handler *OtelHandler) {
		handler.BaggageFilter = composeBaggageFilters(handler.BaggageFilter, filter)
	}
}

// composeBaggageFilters returns a filter that applies first and then second.
func composeBaggageFilters(first, second BaggageFilter) BaggageFilter {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return func(key, value string) (string, bool) {
		value, ok := first(key, value)
		if !ok {
			return "", false
		}
		return second(key, value)
	}
}

// WithBaggageAllowList only logs baggage members whose keys are explicitly
// listed. It composes with the other baggage filter options.
func WithBaggageAllowList(keys ...string) OtelHandlerOpt {
	allowed := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		allowed[key] = struct{}{}
	}

	return WithBaggageFilter(func(key, value string) (string, bool) {
		_, ok := allowed[key]
		return value, ok
	})
}

// WithBaggageDenyList drops baggage members whose keys are explicitly listed.
// It composes with the other baggage filter options.
func WithBaggageDenyList(keys ...string) OtelHandlerOpt {
	denied := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		denied[key] = struct{}{}
	}

	return WithBaggageFilter(func(key, value string) (string, bool) {
		_, deniedKey := denied[key]
		return value, !deniedKey
	})
}

// New creates a new OtelHandler to use with log/slog
func New(next slog.Handler, opts ...OtelHandlerOpt) *OtelHandler {
	ret := &OtelHandler{
		Next:      next,
		NoBaggage: true,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

// NewOtelHandler creates and returns a new HandlerFn, which wraps a handler with OtelHandler to use with log/slog.
func NewOtelHandler(opts ...OtelHandlerOpt) HandlerFn {
	return func(next slog.Handler) slog.Handler {
		return New(next, opts...)
	}
}

// Handle handles the provided log record and adds correlation between a slog record and an Open-Telemetry span.
func (h OtelHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		return h.Next.Handle(ctx, record)
	}
	if !h.NoBaggage {
		// Adding context baggage members to log record.
		b := baggage.FromContext(ctx)
		for _, m := range b.Members() {
			key := m.Key()
			value := m.Value()
			if h.BaggageFilter != nil {
				filteredValue, ok := h.BaggageFilter(key, value)
				if !ok {
					continue
				}
				value = filteredValue
			}
			record.AddAttrs(slog.String(key, value))
		}
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return h.Next.Handle(ctx, record)
	}

	if !h.NoTraceEvents {
		// Adding log info to span event.
		eventAttrs := make([]attribute.KeyValue, 0, 3+len(h.attrs)+record.NumAttrs())
		eventAttrs = append(eventAttrs, attribute.String(slog.MessageKey, capString(record.Message)))
		eventAttrs = append(eventAttrs, attribute.String(slog.LevelKey, record.Level.String()))
		eventAttrs = append(eventAttrs, attribute.String(slog.TimeKey, record.Time.Format(time.RFC3339Nano)))
		eventAttrs = append(eventAttrs, h.attrs...)
		record.Attrs(func(attr slog.Attr) bool {
			eventAttrs = h.appendEventAttrs(eventAttrs, attr)
			return true
		})

		spanKey := fmt.Sprintf("%s.%s", SpanEventKey, strings.ToLower(record.Level.String()))
		span.AddEvent(spanKey, trace.WithAttributes(eventAttrs...))
	}

	// Adding span info to log record.
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		traceID := spanContext.TraceID().String()
		record.AddAttrs(slog.String(TraceIDKey, traceID))
	}

	if spanContext.HasSpanID() {
		spanID := spanContext.SpanID().String()
		record.AddAttrs(slog.String(SpanIDKey, spanID))
	}

	// Setting span status if the log is an error.
	// Purposely leaving as codes.Unset (default) otherwise.
	if record.Level >= slog.LevelError {
		span.SetStatus(codes.Error, capString(record.Message))
	}

	return h.Next.Handle(ctx, record)
}

// WithAttrs returns a new Otel whose attributes consists of handler's attributes followed by attrs.
func (h OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return OtelHandler{
		Next:          h.Next.WithAttrs(attrs),
		NoBaggage:     h.NoBaggage,
		NoTraceEvents: h.NoTraceEvents,
		BaggageFilter: h.BaggageFilter,
		groupPrefix:   h.groupPrefix,
		attrs:         h.appendAttrs(attrs),
	}
}

// appendAttrs returns a new slice with the handler's span event attributes
// followed by attrs converted under the current group prefix. It never
// appends to h.attrs, so handlers derived from the same parent do not share
// their attributes.
func (h OtelHandler) appendAttrs(attrs []slog.Attr) []attribute.KeyValue {
	converted := make([]attribute.KeyValue, len(h.attrs), len(h.attrs)+len(attrs))
	copy(converted, h.attrs)
	for _, attr := range attrs {
		converted = h.appendEventAttrs(converted, attr)
	}
	return converted
}

// appendEventAttrs converts attr under the current group prefix and appends
// the valid results to eventAttrs. Record and WithAttrs attributes both go
// through it, so they get the same keys, kinds and caps.
func (h OtelHandler) appendEventAttrs(eventAttrs []attribute.KeyValue, attr slog.Attr) []attribute.KeyValue {
	for _, otelAttr := range h.slogAttrToOtelAttr(attr, h.groupPrefix) {
		if otelAttr.Valid() {
			eventAttrs = append(eventAttrs, otelAttr)
		}
	}
	return eventAttrs
}

// WithGroup returns a new Otel with a group, provided the group's name.
func (h OtelHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return OtelHandler{
		Next:          h.Next.WithGroup(name),
		NoBaggage:     h.NoBaggage,
		NoTraceEvents: h.NoTraceEvents,
		BaggageFilter: h.BaggageFilter,
		groupPrefix:   joinKey(h.groupPrefix, name),
		attrs:         h.attrs,
	}
}

// Enabled reports whether the logger emits log records at the given context and level.
// Note: We handover the decision down to the next handler.
func (h OtelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Next.Enabled(ctx, level)
}

// slogAttrToOtelAttr converts a slog attribute to OTel ones, prefixing its key
// with the dot-joined group path in prefix.
// Note: returns no attributes if the provided slog attribute is empty.
func (h OtelHandler) slogAttrToOtelAttr(attr slog.Attr, prefix string) []attribute.KeyValue {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return []attribute.KeyValue{}
	}

	value := attr.Value
	if value.Kind() == slog.KindGroup {
		// A group with an empty key is inlined, as slog handlers do.
		groupPrefix := prefix
		if attr.Key != "" {
			groupPrefix = joinKey(prefix, attr.Key)
		}
		var attrs []attribute.KeyValue
		for _, groupAttr := range value.Group() {
			attrs = append(attrs, h.slogAttrToOtelAttr(groupAttr, groupPrefix)...)
		}
		return attrs
	}

	key := joinKey(prefix, attr.Key)
	switch value.Kind() {
	case slog.KindBool:
		return []attribute.KeyValue{attribute.Bool(key, value.Bool())}
	case slog.KindFloat64:
		return []attribute.KeyValue{attribute.Float64(key, value.Float64())}
	case slog.KindInt64:
		return []attribute.KeyValue{attribute.Int64(key, value.Int64())}
	case slog.KindUint64:
		u := value.Uint64()
		if u <= math.MaxInt64 {
			return []attribute.KeyValue{attribute.Int64(key, int64(u))}
		}
		return []attribute.KeyValue{attribute.String(key, strconv.FormatUint(u, 10))}
	case slog.KindDuration:
		// Nanoseconds, as slog's JSON handler writes durations.
		return []attribute.KeyValue{attribute.Int64(key, int64(value.Duration()))}
	case slog.KindString:
		return []attribute.KeyValue{attribute.String(key, capString(value.String()))}
	case slog.KindTime:
		return []attribute.KeyValue{attribute.String(key, value.Time().Format(time.RFC3339Nano))}
	case slog.KindAny:
		switch v := value.Any().(type) {
		case []string:
			return []attribute.KeyValue{attribute.StringSlice(key, capStrings(capSlice(v)))}
		case []int:
			return []attribute.KeyValue{attribute.IntSlice(key, capSlice(v))}
		case []int64:
			return []attribute.KeyValue{attribute.Int64Slice(key, capSlice(v))}
		case []float64:
			return []attribute.KeyValue{attribute.Float64Slice(key, capSlice(v))}
		case []bool:
			return []attribute.KeyValue{attribute.BoolSlice(key, capSlice(v))}
		case error:
			return []attribute.KeyValue{attribute.String(key, capString(fmt.Sprintf("ERROR: %+v", v)))}
		default:
			return []attribute.KeyValue{attribute.String(key, capString(fmt.Sprintf("%+v", v)))}
		}
	default:
		return []attribute.KeyValue{}
	}
}

// joinKey appends key to the dot-joined group path in prefix.
func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// capString truncates s to maxAttributeValueLen bytes on a UTF-8 boundary and
// appends truncationMarker when it was longer.
func capString(s string) string {
	if len(s) <= maxAttributeValueLen {
		return s
	}
	// Step back at most to the start of a rune; invalid UTF-8 (no rune start
	// within utf8.UTFMax bytes) is cut at the cap so the value is not lost.
	cut := maxAttributeValueLen
	for cut > maxAttributeValueLen-utf8.UTFMax && !utf8.RuneStart(s[cut]) {
		cut--
	}
	if !utf8.RuneStart(s[cut]) {
		cut = maxAttributeValueLen
	}
	return s[:cut] + truncationMarker
}

// capSlice returns the first maxAttributeSliceLen elements of values. It
// reslices without copying; the attribute constructors copy the elements.
func capSlice[T any](values []T) []T {
	if len(values) <= maxAttributeSliceLen {
		return values
	}
	return values[:maxAttributeSliceLen]
}

// capStrings applies capString to every element, copying values only when one
// needs truncation so the caller's slice is never modified.
func capStrings(values []string) []string {
	for i, value := range values {
		if len(value) > maxAttributeValueLen {
			capped := make([]string, len(values))
			copy(capped, values)
			for j := i; j < len(capped); j++ {
				capped[j] = capString(capped[j])
			}
			return capped
		}
	}
	return values
}
