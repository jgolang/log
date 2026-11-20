package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/jgolang/errors"
	"github.com/jgolang/errors/codes"
)

// logErrorRecord logs err with the package-level Error and returns the
// decoded JSON record.
func logErrorRecord(t *testing.T, err error) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	std.SetJSONHandler(&buf)
	t.Cleanup(NewJSONHandler)

	Error(err)

	var record map[string]any
	if decodeErr := json.Unmarshal(buf.Bytes(), &record); decodeErr != nil {
		t.Fatalf("decode %q: %v", buf.String(), decodeErr)
	}
	return record
}

func TestErrorLogsCodeOfWrappedCodedError(t *testing.T) {
	notFound := codes.New("PLN-111", "task not found")
	coded := errors.NewC(notFound, "task 42 is missing")

	tests := []struct {
		name string
		err  error
	}{
		{name: "unwrapped", err: coded},
		{name: "behind fmt.Errorf", err: fmt.Errorf("loading board: %w", coded)},
		{name: "behind two wrappers", err: fmt.Errorf("handler: %w", fmt.Errorf("loading board: %w", coded))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := logErrorRecord(t, tt.err)

			group, ok := record["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected an error group, got %v", record["error"])
			}
			if group["code"] != "PLN-111" || group["msg"] != "task not found" {
				t.Fatalf("expected code PLN-111 and its message, got %v", group)
			}
			if origin, ok := group["origin"].(map[string]any); !ok || len(origin) == 0 {
				t.Fatalf("expected a non-empty origin, got %v", group["origin"])
			}
			if record["msg"] != tt.err.Error() {
				t.Fatalf("expected message %q, got %v", tt.err.Error(), record["msg"])
			}
		})
	}
}

func TestErrorWithNilCodeDoesNotPanic(t *testing.T) {
	nilCode := errors.WrapC(fmt.Errorf("plain failure"), nil)

	tests := []struct {
		name string
		err  error
	}{
		{name: "unwrapped", err: nilCode},
		{name: "behind fmt.Errorf", err: fmt.Errorf("loading board: %w", nilCode)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := logErrorRecord(t, tt.err)

			group, ok := record["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected an error group, got %v", record["error"])
			}
			for _, key := range []string{"code", "msg"} {
				if _, has := group[key]; has {
					t.Fatalf("expected no %q attribute for a nil code, got %v", key, group)
				}
			}
			if _, has := group["debug"]; !has {
				t.Fatalf("expected a debug attribute, got %v", group)
			}
			if origin, ok := group["origin"].(map[string]any); !ok || len(origin) == 0 {
				t.Fatalf("expected a non-empty origin, got %v", group["origin"])
			}
		})
	}
}

func TestErrorWithTypedNilCodedErrorDoesNotPanic(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "unwrapped", err: (*errors.Error)(nil)},
		{name: "behind fmt.Errorf", err: fmt.Errorf("x: %w", (*errors.Error)(nil))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := logErrorRecord(t, tt.err)

			if _, isGroup := record["error"].(map[string]any); isGroup {
				t.Fatalf("expected the plain error attribute, got group %v", record["error"])
			}
		})
	}
}

var errLogTitleRequired = errors.Define(codes.New("PLN-001", "title is required").WithKind(codes.KindValidation), "title is required")

// failValidation returns the sentinel through Wrap, so its stack starts here.
//
//go:noinline
func failValidation() error {
	return errors.Wrap(errLogTitleRequired)
}

func TestErrorLogsDefinedSentinelBehindFmtErrorf(t *testing.T) {
	record := logErrorRecord(t, fmt.Errorf("creating task: %w", failValidation()))

	group, ok := record["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected an error group, got %v", record["error"])
	}
	if group["code"] != "PLN-001" {
		t.Fatalf("expected code PLN-001, got %v", group["code"])
	}
	origin, _ := group["origin"].(map[string]any)
	if top, _ := origin["frame_0"].(string); !strings.Contains(top, "failValidation") {
		t.Fatalf("expected origin to start at failValidation, got %v", origin["frame_0"])
	}
}

func TestErrorLogsDeeperCodeBehindNilCode(t *testing.T) {
	coded := errors.NewC(codes.New("PLN-111", "task not found"), "task 42 is missing")
	nilCode := errors.WrapC(coded, nil)

	tests := []struct {
		name string
		err  error
	}{
		{name: "unwrapped", err: nilCode},
		{name: "behind fmt.Errorf", err: fmt.Errorf("loading: %w", nilCode)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := logErrorRecord(t, tt.err)

			group, ok := record["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected an error group, got %v", record["error"])
			}
			if group["code"] != "PLN-111" || group["msg"] != "task not found" {
				t.Fatalf("expected the deeper code PLN-111, got %v", group)
			}
			if _, has := group["debug"]; !has {
				t.Fatalf("expected a debug attribute, got %v", group)
			}
			if origin, ok := group["origin"].(map[string]any); !ok || len(origin) == 0 {
				t.Fatalf("expected a non-empty origin, got %v", group["origin"])
			}
		})
	}
}
