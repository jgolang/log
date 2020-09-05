package log

// Error uses fmt.Sprint to construct and log a message.
// Error logs a message at ErrorLevel.
func Error(args ...interface{}) {
	std.Error(args...)
	return
}

// Errorf uses fmt.Sprintf to log a templated message
// Errorf logs a message at ErrorLevel with format.
func Errorf(template string, args ...interface{}) {
	std.Errorf(template, args...)
	return
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
// Panic logs a message at PanicLevel. The logger then panics,
// even if logging at PanicLevel is disabled.
func Panic(args ...interface{}) {
	std.Panic(args...)
	return
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
// Panicf logs a message at PanicLevel whit format. The logger then panics,
// even if logging at PanicLevel is disabled.
func Panicf(template string, args ...interface{}) {
	std.Panicf(template, args...)
	return
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
// Fatal logs a message at FatalLevel.
func Fatal(args ...interface{}) {
	std.Fatal(args...)
	return
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
// Fatalf logs a message at FatalLevel with format.
func Fatalf(template string, args ...interface{}) {
	std.Fatalf(template, args...)
	return
}
