package log

import "go.uber.org/zap"

// Debug doc...
func Debug(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Debug(args...)
}

// Debugf doc...
func Debugf(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Debugf(template, args...)
}

// Debugw doc...
func Debugw(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Debugw(msg, keysAndValues...)
}

// Warn doc...
func Warn(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Warn(args...)
}

// Warnf doc...
func Warnf(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Warnf(template, args...)
}

// Warnw doc...
func Warnw(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Warnw(msg, keysAndValues...)
}
