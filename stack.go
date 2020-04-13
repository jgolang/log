package log

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// StackTrace doc ...
func StackTrace(v interface{}) {

	trace := stackTrace()

	log(defaultSkip, getPriority(v), "%v\n%v", []interface{}{v, trace})

	return

}

func stackTrace() string {

	var info string
	buffer := bytes.NewBufferString(info)

	pc := make([]uintptr, 10)
	num := runtime.Callers(3, pc)

	frames := runtime.CallersFrames(pc[0 : num-1])

	buffer.WriteString(fmt.Sprintf("--- TRACE: "))

	for i := 0; i < num; i++ {

		frame, found := frames.Next()
		if found {
			frameInfo := fmt.Sprintf("%v:%v", frame.File, frame.Line)
			p := strings.LastIndex(frame.Function, ".")
			function := frame.Function[p+1:]
			buffer.WriteString(fmt.Sprintf("\n\t\t%s\t%s()", frameInfo, function))
		}
	}

	buffer.WriteString(fmt.Sprintf("\n---"))

	return fmt.Sprintf("%v", buffer.String())

}
