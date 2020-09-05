package log

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	std.Debug(args...)
	return
}

// Debugf uses fmt.Sprintf to log a templated message
func Debugf(template string, args ...interface{}) {
	std.Debugf(template, args...)
	return
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	std.Warn(args...)
	return
}

// Warnf uses fmt.Sprintf to log a templated message
func Warnf(template string, args ...interface{}) {
	std.Warnf(template, args...)
	return
}
