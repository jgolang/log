package log

import (
	"fmt"
	golog "log"
	"os"
	"runtime"
	"strings"
)

// A priority is a logging priority. Higher levels are more important.
type priority int8

const (
	debugPriority priority = iota - 1
	infoPriority
	warnPriority
	errorPriority
	dPanicPriority
	panicPriority
	fatalPriority

	_minLevel = debugPriority
	_maxLevel = fatalPriority
)

const defaultSkip = 3

func log(skip int, p priority, template string, args []interface{}) {

	if prod && p < 2 {
		return
	}

	// timeStamp := time.Now().Local().Format("2006-01-02T15:04:05.999-0700")
	typeMsg := getTypeMsg(p)

	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}

	pc := make([]uintptr, 12)
	num := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[0 : num-1])

	var frameInfo, function string
	frame, found := frames.Next()
	if found {
		frameInfo = fmt.Sprintf("%v:%v", frame.File, frame.Line)
		p := strings.LastIndex(frame.Function, ".")
		function = frame.Function[p+1:]
	}

	// s := fmt.Sprintf("%v\t%v\t%v (%v)\t%v\n", timeStamp, typeMsg, frameInfo, function, msg)
	s := fmt.Sprintf("\t%v\t%v (%v)\t%v\n", typeMsg, frameInfo, function, msg)

	var std = golog.New(os.Stderr, "", golog.LstdFlags)

	switch p {
	case fatalPriority:
		std.Output(2, s)
		os.Exit(1)
		return
	case panicPriority:
		fallthrough
	case dPanicPriority:
		std.Output(2, s)
		panic(msg)
	default:
		std.Output(2, s)
	}

}

func getTypeMsg(p priority) string {
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

func getPriority(v interface{}) priority {

	switch v.(type) {
	case error:
		return errorPriority
	default:
		return infoPriority
	}

}
