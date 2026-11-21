package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jgolang/log/logger"
)

// callSite is the source metadata a record is expected to carry.
type callSite struct {
	function string
	file     string
	line     int
}

// nextLine returns the call site of the line after its caller, which is where
// each test places the logging call it checks.
func nextLine() callSite {
	pc, file, line, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()
	return callSite{
		function: name[strings.LastIndex(name, ".")+1:],
		file:     filepath.Base(file),
		line:     line + 1,
	}
}

func decodeRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("decode %q: %v", buf.String(), err)
	}
	return record
}

func assertSource(t *testing.T, record map[string]any, want callSite) {
	t.Helper()
	source, ok := record["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected source object, got %v", record["source"])
	}
	got := callSite{
		function: fmt.Sprint(source["func"]),
		file:     fmt.Sprint(source["file"]),
	}
	if line, ok := source["line"].(float64); ok {
		got.line = int(line)
	}
	if got != want {
		t.Fatalf("source = %+v, want %+v", got, want)
	}
}

// expectPanic is deferred by the Panic cases: they must still panic after
// logging, and recovering lets the test inspect the record.
func expectPanic(t *testing.T) {
	if recover() == nil {
		t.Error("expected a panic")
	}
}

func usePackageJSONHandler(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	std.SetJSONHandler(&buf)
	old := SetLevel(logger.LevelTrace)
	t.Cleanup(func() {
		SetLevel(old)
		NewJSONHandler()
	})
	return &buf
}

func TestPackageLevelSourceIsCaller(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		log  func(t *testing.T) callSite
	}{
		{name: "Debug", log: func(t *testing.T) callSite {
			site := nextLine()
			Debug("hello")
			return site
		}},
		{name: "DebugC", log: func(t *testing.T) callSite {
			site := nextLine()
			DebugC(ctx, "hello")
			return site
		}},
		{name: "Print", log: func(t *testing.T) callSite {
			site := nextLine()
			Print("hello")
			return site
		}},
		{name: "PrintC", log: func(t *testing.T) callSite {
			site := nextLine()
			PrintC(ctx, "hello")
			return site
		}},
		{name: "Info", log: func(t *testing.T) callSite {
			site := nextLine()
			Info("hello")
			return site
		}},
		{name: "InfoC", log: func(t *testing.T) callSite {
			site := nextLine()
			InfoC(ctx, "hello")
			return site
		}},
		{name: "Warn", log: func(t *testing.T) callSite {
			site := nextLine()
			Warn("hello")
			return site
		}},
		{name: "WarnC", log: func(t *testing.T) callSite {
			site := nextLine()
			WarnC(ctx, "hello")
			return site
		}},
		{name: "Error", log: func(t *testing.T) callSite {
			site := nextLine()
			Error("hello")
			return site
		}},
		{name: "ErrorC", log: func(t *testing.T) callSite {
			site := nextLine()
			ErrorC(ctx, "hello")
			return site
		}},
		{name: "Panic", log: func(t *testing.T) (site callSite) {
			defer expectPanic(t)
			site = nextLine()
			Panic("hello")
			return site
		}},
		{name: "PanicC", log: func(t *testing.T) (site callSite) {
			defer expectPanic(t)
			site = nextLine()
			PanicC(ctx, "hello")
			return site
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := usePackageJSONHandler(t)

			want := tt.log(t)

			assertSource(t, decodeRecord(t, buf), want)
		})
	}
}

func TestInstanceSourceIsCaller(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		log  func(*testing.T, *Logger) callSite
	}{
		{name: "Debug", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.Debug("hello")
			return site
		}},
		{name: "DebugContext", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.DebugContext(ctx, "hello")
			return site
		}},
		{name: "Print", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.Print("hello")
			return site
		}},
		{name: "PrintContext", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.PrintContext(ctx, "hello")
			return site
		}},
		{name: "Info", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.Info("hello")
			return site
		}},
		{name: "InfoContext", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.InfoContext(ctx, "hello")
			return site
		}},
		{name: "Warn", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.Warn("hello")
			return site
		}},
		{name: "WarnContext", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.WarnContext(ctx, "hello")
			return site
		}},
		{name: "Error", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.Error("hello")
			return site
		}},
		{name: "ErrorContext", log: func(t *testing.T, l *Logger) callSite {
			site := nextLine()
			l.ErrorContext(ctx, "hello")
			return site
		}},
		{name: "Panic", log: func(t *testing.T, l *Logger) (site callSite) {
			defer expectPanic(t)
			site = nextLine()
			l.Panic("hello")
			return site
		}},
		{name: "PanicContext", log: func(t *testing.T, l *Logger) (site callSite) {
			defer expectPanic(t)
			site = nextLine()
			l.PanicContext(ctx, "hello")
			return site
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			instance := New(WithJSONHandler(&buf), WithLevel(logger.LevelTrace))

			want := tt.log(t, instance)

			assertSource(t, decodeRecord(t, &buf), want)
		})
	}
}

func assertStackStartsAt(t *testing.T, record map[string]any, want callSite) {
	t.Helper()
	source, _ := record["source"].(map[string]any)
	stack, ok := source["stack_trace"].(map[string]any)
	if !ok {
		t.Fatalf("expected source.stack_trace object, got %v", record["source"])
	}
	wantFrame := fmt.Sprintf("%s:%d (%s)", want.file, want.line, want.function)
	if stack["frame_0"] != wantFrame {
		t.Fatalf("stack_trace.frame_0 = %v, want %q", stack["frame_0"], wantFrame)
	}
}

func TestPackageLevelDebugStackStartsAtCaller(t *testing.T) {
	buf := usePackageJSONHandler(t)
	SetDebugStackTrace(true)
	t.Cleanup(func() { SetDebugStackTrace(false) })

	want := nextLine()
	Debug("hello")

	record := decodeRecord(t, buf)
	assertSource(t, record, want)
	assertStackStartsAt(t, record, want)
}

func TestInstanceDebugStackStartsAtCaller(t *testing.T) {
	var buf bytes.Buffer
	instance := New(WithJSONHandler(&buf), WithDebugStackTrace(true))

	want := nextLine()
	instance.Debug("hello")

	record := decodeRecord(t, &buf)
	assertSource(t, record, want)
	assertStackStartsAt(t, record, want)
}

func TestStackTraceStartsAtCaller(t *testing.T) {
	want := nextLine()
	attr := StackTrace()

	record := map[string]any{"source": attrToMap(attr)}
	assertSource(t, record, want)
	assertStackStartsAt(t, record, want)
}

func TestInstanceStackTraceStartsAtCaller(t *testing.T) {
	instance := New(WithJSONHandler(&bytes.Buffer{}))

	want := nextLine()
	attr := instance.StackTrace()

	record := map[string]any{"source": attrToMap(attr)}
	assertSource(t, record, want)
	assertStackStartsAt(t, record, want)
}

// attrToMap encodes attr the way the JSON handler does, so StackTrace can be
// checked with the same assertions as a record.
func attrToMap(attr slog.Attr) map[string]any {
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("", attr)
	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		return nil
	}
	source, _ := record[attr.Key].(map[string]any)
	return source
}

// logThroughWrapper stands for a consumer's own logging helper.
func logThroughWrapper(msg string) {
	Info(msg)
}

func (w instanceWrapper) info(msg string) {
	w.logger.Info(msg)
}

type instanceWrapper struct {
	logger *Logger
}

func TestSetCalldepthSkipsWrapperFrames(t *testing.T) {
	buf := usePackageJSONHandler(t)
	SetCalldepth(4)
	t.Cleanup(func() { SetCalldepth(3) })

	want := nextLine()
	logThroughWrapper("hello")

	assertSource(t, decodeRecord(t, buf), want)
}

func TestInstanceSetCalldepthSkipsWrapperFrames(t *testing.T) {
	var buf bytes.Buffer
	wrapper := instanceWrapper{logger: New(WithJSONHandler(&buf))}
	wrapper.logger.SetCalldepth(4)

	want := nextLine()
	wrapper.info("hello")

	assertSource(t, decodeRecord(t, &buf), want)
}
