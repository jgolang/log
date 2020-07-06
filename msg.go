package log

// Println uses fmt.Sprint to construct and log a message.
// Println logs a message at InfoLevel.
func Println(args ...interface{}) {
	log(defaultSkip, infoPriority, "", args)
	return
}

// Printf uses fmt.Sprintf to log a templated message.
// Printf logs a message at InfoLevel whit format.
func Printf(template string, args ...interface{}) {
	log(defaultSkip, infoPriority, template, args)
	return
}

// Info uses fmt.Sprint to construct and log a message.
// Info logs a message at InfoLevel.
func Info(args ...interface{}) {
	log(defaultSkip, infoPriority, "", args)
	return
}

// Infof uses fmt.Sprintf to log a templated message.
// Infof logs a message at InfoLevel whit format.
func Infof(template string, args ...interface{}) {
	log(defaultSkip, infoPriority, template, args)
	return
}
