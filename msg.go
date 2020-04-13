package log

// Println doc...
func Println(args ...interface{}) {
	log(defaultSkip, infoPriority, "", args)
	return
}

// Printf doc...
func Printf(template string, args ...interface{}) {
	log(defaultSkip, infoPriority, template, args)
	return
}

// Info doc...
func Info(args ...interface{}) {
	log(defaultSkip, infoPriority, "", args)
	return
}

// Infof doc...
func Infof(template string, args ...interface{}) {
	log(defaultSkip, infoPriority, template, args)
	return
}
