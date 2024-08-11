package otel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

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

// OtelHandler is an implementation of slog's Handler interface.
// Its role is to ensure correlation between logs and OTel spans
// by:
//
// 1. Adding otel span and trace IDs to the log record.
// 2. Adding otel context baggage members to the log record.
// 3. Setting slog record as otel span event.
// 4. Adding slog record attributes to the otel span event.
// 5. Setting span status based on slog record level (only if >= slog.LevelError).
type OtelHandler struct {
	// Next represents the next handler in the chain.
	Next slog.Handler
	// NoBaggage determines whether to add context baggage members to the log record.
	NoBaggage bool
	// NoTraceEvents determines whether to record an event for every log on the active trace.
	NoTraceEvents bool
}

type OtelHandlerOpt func(handler *OtelHandler)

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

// New creates a new OtelHandler to use with log/slog
func New(next slog.Handler, opts ...OtelHandlerOpt) *OtelHandler {
	ret := &OtelHandler{
		Next: next,
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
			record.AddAttrs(slog.String(m.Key(), m.Value()))
		}
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return h.Next.Handle(ctx, record)
	}

	if !h.NoTraceEvents {
		// Adding log info to span event.
		eventAttrs := make([]attribute.KeyValue, 0, record.NumAttrs())
		eventAttrs = append(eventAttrs, attribute.String(slog.MessageKey, record.Message))
		eventAttrs = append(eventAttrs, attribute.String(slog.LevelKey, record.Level.String()))
		eventAttrs = append(eventAttrs, attribute.String(slog.TimeKey, record.Time.Format(time.RFC3339Nano)))
		record.Attrs(func(attr slog.Attr) bool {
			otelAttrs := h.slogAttrToOtelAttr(attr)
			for _, otelAttr := range otelAttrs {
				if otelAttr.Valid() {
					eventAttrs = append(eventAttrs, otelAttr)
				}
			}
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
		span.SetStatus(codes.Error, record.Message)
	}

	return h.Next.Handle(ctx, record)
}

// WithAttrs returns a new Otel whose attributes consists of handler's attributes followed by attrs.
func (h OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return OtelHandler{
		Next:          h.Next.WithAttrs(attrs),
		NoBaggage:     h.NoBaggage,
		NoTraceEvents: h.NoTraceEvents,
	}
}

// WithGroup returns a new Otel with a group, provided the group's name.
func (h OtelHandler) WithGroup(name string) slog.Handler {
	return OtelHandler{
		Next:          h.Next.WithGroup(name),
		NoBaggage:     h.NoBaggage,
		NoTraceEvents: h.NoTraceEvents,
	}
}

// Enabled reports whether the logger emits log records at the given context and level.
// Note: We handover the decision down to the next handler.
func (h OtelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Next.Enabled(ctx, level)
}

// slogAttrToOtelAttr converts a slog attribute to an OTel one.
// Note: returns an empty attribute if the provided slog attribute is empty.
func (h OtelHandler) slogAttrToOtelAttr(attr slog.Attr, groupKeys ...string) []attribute.KeyValue {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return []attribute.KeyValue{}
	}

	key := func(k string, prefixes ...string) string {
		for _, prefix := range prefixes {
			k = fmt.Sprintf("%s.%s", prefix, k)
		}

		return k
	}(attr.Key, groupKeys...)

	value := attr.Value.Resolve()

	switch attr.Value.Kind() {
	case slog.KindBool:
		return []attribute.KeyValue{attribute.Bool(key, value.Bool())}
	case slog.KindFloat64:
		return []attribute.KeyValue{attribute.Float64(key, value.Float64())}
	case slog.KindInt64:
		return []attribute.KeyValue{attribute.Int64(key, value.Int64())}
	case slog.KindString:
		return []attribute.KeyValue{attribute.String(key, value.String())}
	case slog.KindTime:
		return []attribute.KeyValue{attribute.String(key, value.Time().Format(time.RFC3339Nano))}
	case slog.KindGroup:
		groupAttrs := value.Group()
		if len(groupAttrs) == 0 {
			return nil
		}
		var attrs []attribute.KeyValue
		for _, groupAttr := range groupAttrs {
			childAttr := h.slogAttrToOtelAttr(groupAttr, append(groupKeys, key)...)
			attrs = append(attrs, childAttr...)
		}
		return attrs
	case slog.KindAny:
		switch v := attr.Value.Any().(type) {
		case []string:
			return []attribute.KeyValue{attribute.StringSlice(key, v)}
		case []int:
			return []attribute.KeyValue{attribute.IntSlice(key, v)}
		case []int64:
			return []attribute.KeyValue{attribute.Int64Slice(key, v)}
		case []float64:
			return []attribute.KeyValue{attribute.Float64Slice(key, v)}
		case []bool:
			return []attribute.KeyValue{attribute.BoolSlice(key, v)}
		case error:
			return []attribute.KeyValue{attribute.String(key, fmt.Sprintf("ERROR: %+v", v))}
		default:
			return []attribute.KeyValue{attribute.String(key, fmt.Sprintf("%+v", v))}
		}
	default:
		return []attribute.KeyValue{}
	}
}
