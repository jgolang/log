package otel

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

// BenchmarkHandleSpanEvent measures a record written with a recording span,
// for a plain logger and for a request-scoped logger built with With.
func BenchmarkHandleSpanEvent(b *testing.B) {
	plain := slog.New(New(slog.NewJSONHandler(io.Discard, nil)))
	loggers := []struct {
		name   string
		logger *slog.Logger
	}{
		{"no handler attrs", plain},
		{"with handler attrs", plain.With("request_id", "r-123", "tenant", "acme").WithGroup("req").With("path", "/orders")},
	}
	for _, l := range loggers {
		b.Run(l.name, func(b *testing.B) {
			ctx := trace.ContextWithSpan(context.Background(), &discardSpan{})
			b.ReportAllocs()
			for b.Loop() {
				l.logger.InfoContext(ctx, "order created", "order_id", 42, "status", "ok")
			}
		})
	}
}

// discardSpan is a recording span that drops events, so the benchmark measures
// the handler and not the recorder.
type discardSpan struct {
	recordingSpan
}

func (s *discardSpan) AddEvent(string, ...trace.EventOption) {}
