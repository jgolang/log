package log

import (
	"go.uber.org/zap"
)

// Println doc...
func Println(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Info(args...)
}

// Printf doc...
func Printf(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Infof(template, args...)
}

// Printw doc...
func Printw(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Infow(msg, keysAndValues...)
}

// Info doc...
func Info(args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Info(args...)
}

// Infof doc...
func Infof(template string, args ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Infof(template, args...)
}

// Infow doc...
func Infow(msg string, keysAndValues ...interface{}) {
	logger := zap.S()
	defer logger.Sync()
	logger.Infow(msg, keysAndValues...)
}
