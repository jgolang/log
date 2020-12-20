package log_test

import (
	"testing"

	"github.com/jgolang/log"
)

func TestLogger_Output(t *testing.T) {
	log.ProductionMode()
	log.Error("Test message")
}
