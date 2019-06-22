package log

import "go.uber.org/zap"

// Error doc...
func Error(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Error(args...)
}

// Errorf doc...
func Errorf(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Errorf(template, args...)
}

// Errorw doc...
func Errorw(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Errorw(msg, keysAndValues...)
}

// Panic doc...
func Panic(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Panic(args...)
}

// Panicf doc...
func Panicf(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Panicf(template, args...)
}

// Panicw doc...
func Panicw(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Panicw(msg, keysAndValues...)
}

// Fatal doc...
func Fatal(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Fatal(args...)
}

// Fatalf doc...
func Fatalf(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Fatalf(template, args...)
}

// Fatalw doc...
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Fatalw(msg, keysAndValues...)
}
