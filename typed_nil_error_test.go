package log

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"log/slog"
	"testing"
)

// nilUnsafeError dereferences its receiver in Error and Unwrap, as many
// application error types do.
type nilUnsafeError struct{ inner error }

func (e *nilUnsafeError) Error() string { return "wrapped: " + e.inner.Error() }
func (e *nilUnsafeError) Unwrap() error { return e.inner }

func TestTypedNilWrapperErrorDoesNotPanic(t *testing.T) {
	var typedNil *nilUnsafeError
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "attribute", err: typedNil, want: "<nil>"},
		{name: "behind fmt.Errorf", err: fmt.Errorf("x: %w", typedNil), want: "x: <nil>"},
		// Join's Error calls the nil wrapper's Error on a non-nil receiver;
		// slog writes this panic as text, and so does the logger.
		{name: "behind errors.Join", err: stderrors.Join(typedNil), want: "!PANIC: runtime error: invalid memory address or nil pointer dereference"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			instance := New(WithJSONHandler(&buf))

			instance.Error("failed", "error", tt.err, slog.Group("req", slog.Any("error", tt.err)))

			record := decodeRecord(t, &buf)
			if got := record["error"]; got != tt.want {
				t.Fatalf("error = %v, want %q (line %s)", got, tt.want, buf.String())
			}
			if got := errorAttrOf(record, "req", "error"); got != tt.want {
				t.Fatalf("req.error = %v, want %q (line %s)", got, tt.want, buf.String())
			}
		})
	}
}

func TestPackageLevelTypedNilErrorAsMessageDoesNotPanic(t *testing.T) {
	var typedNil *nilUnsafeError

	record := logErrorRecord(t, typedNil)

	if record["msg"] != "<nil>" || record["error"] != "<nil>" {
		t.Fatalf("msg = %v, error = %v, want <nil>", record["msg"], record["error"])
	}
}

func TestTypedNilJoinedWithCodedErrorDoesNotPanic(t *testing.T) {
	var typedNil *nilUnsafeError
	joined := stderrors.Join(typedNil, newCodedError())
	const panicText = "!PANIC: runtime error: invalid memory address or nil pointer dereference"

	t.Run("attribute", func(t *testing.T) {
		var buf bytes.Buffer
		instance := New(WithJSONHandler(&buf))

		instance.Error("failed", "error", joined)

		if got := errorAttrOf(decodeRecord(t, &buf), "error", "text"); got != panicText {
			t.Fatalf("error.text = %v, want %q (line %s)", got, panicText, buf.String())
		}
	})
	t.Run("package-level message", func(t *testing.T) {
		record := logErrorRecord(t, joined)

		if record["msg"] != panicText {
			t.Fatalf("msg = %v, want %q", record["msg"], panicText)
		}
	})
}
