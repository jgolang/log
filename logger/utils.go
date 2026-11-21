package logger

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
)

func itoa(buf *[]byte, i int64, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// maxStackDepth is the number of frames a stack trace keeps.
const maxStackDepth = 10

// callers returns up to depth program counters of the current goroutine's
// stack, skipping skip frames as runtime.Callers does: 0 is runtime.Callers,
// 1 is callers and 2 is the function that calls callers.
func callers(skip, depth int) []uintptr {
	pcs := make([]uintptr, depth)
	return pcs[:runtime.Callers(skip, pcs)]
}

// callerPC returns the first program counter callers(skip, 1) would return,
// without allocating, or 0 when the stack is shorter than skip.
func callerPC(skip int) uintptr {
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])
	return pcs[0]
}

// sourceAttr builds the "source" group from pcs, whose first entry is the
// call site to report. withStack adds a stack trace that starts at that call
// site.
func sourceAttr(pcs []uintptr, withStack bool) slog.Attr {
	f, _ := runtime.CallersFrames(pcs).Next()
	var as []any
	if f.Function != "" {
		as = append(as, slog.String("func", getFuncName(f.Function)))
	}
	if f.File != "" {
		as = append(as, slog.String("file", filepath.Base(f.File)))
	}
	if f.Line != 0 {
		as = append(as, slog.Int("line", f.Line))
	}
	if withStack {
		as = append(as, getStackTrace(pcs))
	}
	return slog.Group("source", as...)
}

func getFuncName(function string) string {
	p := strings.LastIndex(function, ".")
	return function[p+1:]
}

func getStackTrace(pcs []uintptr) slog.Attr {
	num := len(pcs)
	frames := runtime.CallersFrames(pcs)
	var as []any
	level := 0
	for i := 0; i < num; i++ {
		frame, found := frames.Next()
		if found {
			var newbuf []byte
			newbuf = newbuf[:0]
			newbuf = append(newbuf, filepath.Base(frame.File)...)
			newbuf = append(newbuf, ':')
			itoa(&newbuf, int64(frame.Line), 2)
			function := getFuncName(frame.Function)
			newbuf = append(newbuf, ' ')
			newbuf = append(newbuf, '(')
			newbuf = append(newbuf, function...)
			newbuf = append(newbuf, ')')
			as = append(as, slog.String(fmt.Sprintf("frame_%v", level), string(newbuf)))
			level++
		}
	}
	return slog.Group("stack_trace", as...)
}
