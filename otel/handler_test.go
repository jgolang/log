package otel

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/baggage"
)

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
