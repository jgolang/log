package log

// Error doc...
func Error(args ...interface{}) {
	log(defaultSkip, errorPriority, "", args)
	return
}

// Errorf doc...
func Errorf(template string, args ...interface{}) {
	log(defaultSkip, errorPriority, template, args)
	return
}

// Panic doc ...
func Panic(args ...interface{}) {
	log(defaultSkip, panicPriority, "", args)
	return
}

// Panicf doc...
func Panicf(template string, args ...interface{}) {
	log(defaultSkip, panicPriority, template, args)
	return
}

// Fatal doc...
func Fatal(args ...interface{}) {
	log(defaultSkip, fatalPriority, "", args)
	return
}

// Fatalf doc...
func Fatalf(template string, args ...interface{}) {
	log(defaultSkip, fatalPriority, template, args)
	return
}
