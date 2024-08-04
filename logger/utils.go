package logger

import (
	"fmt"
	"log/slog"
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

// Source returns a Source for the log event.
// If the Record was created without the necessary information,
// or if the location is unavailable, it returns a non-nil *Source
// with zero fields.
func source(calldepth int, pc []uintptr) slog.Attr {
	num := runtime.Callers(calldepth, pc)
	fs := runtime.CallersFrames(pc[0:num])
	f, _ := fs.Next()
	var as []any
	if f.Function != "" {
		as = append(as, slog.String("func", f.Function))
	}
	if f.File != "" {
		as = append(as, slog.String("file", f.File))
	}
	if f.Line != 0 {
		as = append(as, slog.Int("line", f.Line))
	}
	return slog.Group("source", as...)
}

func sourceWithStackTrace(calldepth int, pc []uintptr) slog.Attr {
	num := runtime.Callers(calldepth, pc)
	fs := runtime.CallersFrames(pc[0:num])
	f, _ := fs.Next()
	var as []any
	if f.Function != "" {
		as = append(as, slog.String("func", f.Function))
	}
	if f.File != "" {
		as = append(as, slog.String("file", f.File))
	}
	if f.Line != 0 {
		as = append(as, slog.Int("line", f.Line))
	}
	stack := getStackTrace(calldepth+1, pc)
	as = append(as, stack)
	return slog.Group("source", as...)
}

func getFuncName(function string) string {
	p := strings.LastIndex(function, ".")
	return function[p+1:]
}

func getStackTrace(calldepth int, pc []uintptr) slog.Attr {
	num := runtime.Callers(calldepth, pc)
	frames := runtime.CallersFrames(pc[0:num])
	var as []any
	level := 0
	for i := 0; i < num; i++ {
		frame, found := frames.Next()
		if found {
			var newbuf []byte
			newbuf = newbuf[:0]
			newbuf = append(newbuf, frame.File...)
			newbuf = append(newbuf, ':')
			itoa(&newbuf, int64(frame.Line), 2)
			p := strings.LastIndex(frame.Function, ".")
			function := frame.Function[p+1:]
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
