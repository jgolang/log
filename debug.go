package log

import (
	"fmt"
)

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {

	trace := stackTrace()
	args = append(args, fmt.Sprintf("\n%v", trace))

	log(defaultSkip, debugPriority, "", args)

	return
}

// Debugf uses fmt.Sprintf to log a templated message
func Debugf(template string, args ...interface{}) {

	trace := stackTrace()
	args = append(args, fmt.Sprintf("\n%v", trace))

	log(defaultSkip, debugPriority, template+"%v", args)

	return

}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	log(defaultSkip, warnPriority, "", args)
	return
}

// Warnf uses fmt.Sprintf to log a templated message
func Warnf(template string, args ...interface{}) {
	log(defaultSkip, warnPriority, template, args)
	return
}
