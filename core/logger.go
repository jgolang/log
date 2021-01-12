package core

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Priority is a logging priority. Higher levels are more important.
type Priority int8

const (
	debugPriority Priority = iota - 1
	infoPriority
	warnPriority
	errorPriority
	dPanicPriority
	panicPriority
	fatalPriority

	_minLevel = debugPriority
	_maxLevel = fatalPriority
)

// Logger struct doc
type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	buf       []byte     // for accumulating text to write
	out       Output     // destination for output
	formater  Formater   // out formatter
	calldepth int
	prod      bool
}

// New creates a new Logger. The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.
func New(f Formater, out Output, calldepth int) Logger {
	return Logger{out: out, formater: f, calldepth: calldepth}
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.z
func (l *Logger) Output(calldepth int, p Priority, template string, args []interface{}, stack [][]byte) error {
	s := template
	if s == "" && len(args) > 0 {
		s = fmt.Sprint(args...)
	} else if s != "" && len(args) > 0 {
		s = fmt.Sprintf(template, args...)
	}
	now := time.Now() // get this early.
	pc, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	f := runtime.FuncForPC(pc)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = l.buf[:0]
	if l.prod {
		if p < 2 {
			return nil
		}
		l.formater.Production(&l.buf, now, file, line, f.Name(), p, s, stack)
	} else {
		l.formater.Development(&l.buf, now, file, line, f.Name(), p, s, stack)
	}
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	err := l.out.Write(l.buf)
	return err
}

// GetTypeMsg get message type
func GetTypeMsg(prod bool, p Priority) string {
	if prod {
		switch p {
		case debugPriority:
			return "debug"
		case infoPriority:
			return "info"
		case warnPriority:
			return "warn"
		case errorPriority:
			return "error"
		case dPanicPriority:
			return "dpanic"
		case panicPriority:
			return "panic"
		case fatalPriority:
			return "fatal"
		default:
			return fmt.Sprintf("LEVEL(%d)", p)
		}
	}
	switch p {
	case debugPriority:
		return "DEBUG"
	case infoPriority:
		return "INFO"
	case warnPriority:
		return "WARN"
	case errorPriority:
		return "ERROR"
	case dPanicPriority:
		return "DPANIC"
	case panicPriority:
		return "PANIC"
	case fatalPriority:
		return "FATAL"
	default:
		return fmt.Sprintf("LEVEL(%d)", p)
	}
}

// GetPriority message
func GetPriority(v interface{}) Priority {
	switch v.(type) {
	case error:
		return errorPriority
	default:
		return infoPriority
	}
}

// GetStackTrace doc ...
func GetStackTrace(calldepth int) [][]byte {
	var buf [][]byte
	buf = buf[:0]
	pc := make([]uintptr, 10)
	num := runtime.Callers(calldepth, pc)
	frames := runtime.CallersFrames(pc[0 : num-1])
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
			buf = append(buf, newbuf)
		}
	}
	return buf
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Logger) Debug(args ...interface{}) {
	stackTrace := GetStackTrace(l.calldepth + 1)
	l.Output(l.calldepth, debugPriority, "", args, stackTrace)
	return
}

// Debugf uses fmt.Sprintf to log a templated message
func (l *Logger) Debugf(template string, args ...interface{}) {
	stackTrace := GetStackTrace(l.calldepth + 1)
	l.Output(l.calldepth, debugPriority, template+"%v", args, stackTrace)
	return
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *Logger) Warn(args ...interface{}) {
	l.Output(l.calldepth, warnPriority, "", args, nil)
	return
}

// Warnf uses fmt.Sprintf to log a templated message
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.Output(l.calldepth, warnPriority, template, args, nil)
	return
}

// Error uses fmt.Sprint to construct and log a message.
// Error logs a message at ErrorLevel.
func (l *Logger) Error(args ...interface{}) {
	l.Output(l.calldepth, errorPriority, "", args, nil)
	return
}

// Errorf uses fmt.Sprintf to log a templated message
// Errorf logs a message at ErrorLevel with format.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.Output(l.calldepth, errorPriority, template, args, nil)
	return
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
// Panic logs a message at PanicLevel. The logger then panics,
// even if logging at PanicLevel is disabled.
func (l *Logger) Panic(args ...interface{}) {
	l.Output(l.calldepth, panicPriority, "", args, nil)
	panic(fmt.Sprint(args...))
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
// Panicf logs a message at PanicLevel whit format. The logger then panics,
// even if logging at PanicLevel is disabled.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.Output(l.calldepth, panicPriority, template, args, nil)
	s := template
	if s == "" && len(args) > 0 {
		s = fmt.Sprint(args...)
	} else if s != "" && len(args) > 0 {
		s = fmt.Sprintf(template, args...)
	}
	panic(s)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
// Fatal logs a message at FatalLevel.
func (l *Logger) Fatal(args ...interface{}) {
	l.Output(l.calldepth, fatalPriority, "", args, nil)
	os.Exit(1)
	return
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
// Fatalf logs a message at FatalLevel with format.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.Output(l.calldepth, fatalPriority, template, args, nil)
	os.Exit(1)
	return
}

// Println uses fmt.Sprint to construct and log a message.
// Println logs a message at InfoLevel.
func (l *Logger) Println(args ...interface{}) {
	l.Output(l.calldepth, infoPriority, "", args, nil)
	return
}

// Printf uses fmt.Sprintf to log a templated message.
// Printf logs a message at InfoLevel whit format.
func (l *Logger) Printf(template string, args ...interface{}) {
	l.Output(l.calldepth, infoPriority, template, args, nil)
	return
}

// Info uses fmt.Sprint to construct and log a message.
// Info logs a message at InfoLevel.
func (l *Logger) Info(args ...interface{}) {
	l.Output(l.calldepth, infoPriority, "", args, nil)
	return
}

// Infof uses fmt.Sprintf to log a templated message.
// Infof logs a message at InfoLevel whit format.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.Output(l.calldepth, infoPriority, template, args, nil)
	return
}

// StackTrace allows you to view the exact place where the error or incident originated within the code.
// Shows a trace of up to 10 layers from where the error or incident was generated.
func (l *Logger) StackTrace(v interface{}) {
	stackTrace := GetStackTrace(l.calldepth + 1)
	args := []interface{}{v}
	l.Output(l.calldepth, GetPriority(v), "", args, stackTrace)
	return
}

// SetNewFormat configure your custom outputs development and production format
func (l *Logger) SetNewFormat(f Formater) {
	l.formater = f
}

// SetNewOutput Set custom log output destination
func (l *Logger) SetNewOutput(o Output) {
	l.out = o
}

// ProductionMode set production mode logger
func (l *Logger) ProductionMode() {
	l.prod = true
}

// DevelopmentMode set development mode logger
func (l *Logger) DevelopmentMode() {
	l.prod = false
}

// SetCalldepth configure the number of stack frames
// to ascend, with 0 identifying the caller of Caller.
func (l *Logger) SetCalldepth(calldepth int) {
	l.calldepth = calldepth
}

// GetMode doc ...
func (l *Logger) GetMode() string {
	if l.prod {
		return "PROD"
	}
	return "DEV"
}
