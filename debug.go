package log

import (
	"fmt"
)

// Debug doc...
func Debug(args ...interface{}) {

	trace := stackTrace()
	args = append(args, fmt.Sprintf("\n%v", trace))

	log(defaultSkip, debugPriority, "", args)

	return
}

// Debugf doc...
func Debugf(template string, args ...interface{}) {

	trace := stackTrace()
	args = append(args, fmt.Sprintf("\n%v", trace))

	log(defaultSkip, debugPriority, template+"%v", args)

	return

}

// Warn doc...
func Warn(args ...interface{}) {
	log(defaultSkip, warnPriority, "", args)
	return
}

// Warnf doc...
func Warnf(template string, args ...interface{}) {
	log(defaultSkip, warnPriority, template, args)
	return
}
