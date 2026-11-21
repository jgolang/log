package otel

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/baggage"
)

func TestBaggageFilterOptionsCompose(t *testing.T) {
	redactTenant := WithBaggageFilter(func(key, value string) (string, bool) {
		if key == "tenant" {
			return "[REDACTED]", true
		}
		return value, true
	})
	tests := []struct {
		name    string
		opts    []OtelHandlerOpt
		want    []string
		notWant []string
	}{
		{
			name:    "allow list then deny list",
			opts:    []OtelHandlerOpt{WithBaggageAllowList("request_id", "password"), WithBaggageDenyList("password")},
			want:    []string{`"request_id":"r1"`},
			notWant: []string{"session_token", "password", "tenant"},
		},
		{
			name:    "deny list then allow list",
			opts:    []OtelHandlerOpt{WithBaggageDenyList("password"), WithBaggageAllowList("request_id", "password")},
			want:    []string{`"request_id":"r1"`},
			notWant: []string{"session_token", "password", "tenant"},
		},
		{
			name:    "allow list then custom filter",
			opts:    []OtelHandlerOpt{WithBaggageAllowList("request_id", "tenant"), redactTenant},
			want:    []string{`"request_id":"r1"`, `"tenant":"[REDACTED]"`},
			notWant: []string{"session_token", "acme"},
		},
		{
			name:    "nil filter keeps the allow list",
			opts:    []OtelHandlerOpt{WithBaggageAllowList("request_id"), WithBaggageFilter(nil)},
			want:    []string{`"request_id":"r1"`},
			notWant: []string{"session_token", "password", "tenant"},
		},
		{
			name:    "custom filter then allow list",
			opts:    []OtelHandlerOpt{redactTenant, WithBaggageAllowList("request_id", "tenant")},
			want:    []string{`"request_id":"r1"`, `"tenant":"[REDACTED]"`},
			notWant: []string{"session_token", "acme"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			opts := append([]OtelHandlerOpt{WithNoBaggage(false)}, tt.opts...)
			logger := slog.New(New(slog.NewJSONHandler(&out, nil), opts...))
			bag, err := baggage.Parse("request_id=r1,session_token=s3cr3t,password=p,tenant=acme")
			if err != nil {
				t.Fatalf("baggage.Parse() error = %v", err)
			}

			logger.InfoContext(baggage.ContextWithBaggage(context.Background(), bag), "x")

			for _, want := range tt.want {
				if !strings.Contains(out.String(), want) {
					t.Errorf("output %s does not contain %s", out.String(), want)
				}
			}
			for _, notWant := range tt.notWant {
				if strings.Contains(out.String(), notWant) {
					t.Errorf("output %s contains %s", out.String(), notWant)
				}
			}
		})
	}
}
