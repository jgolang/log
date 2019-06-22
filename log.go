package log

import (
	"log"

	"go.uber.org/zap"
)

var newLogger *zap.Logger

func init() {
	var err error
	newLogger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	newLogger = newLogger.WithOptions(zap.AddCallerSkip(1)).WithOptions(zap.AddStacktrace(zap.FatalLevel))
	zap.ReplaceGlobals(newLogger)
}

// ChangeCallerSkip doc...
func ChangeCallerSkip(n int) {
	newLogger = newLogger.WithOptions(zap.AddCallerSkip(n)).WithOptions(zap.AddStacktrace(zap.FatalLevel))
	zap.ReplaceGlobals(newLogger)
}
