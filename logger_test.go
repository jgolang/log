package log_test

import (
	"testing"

	"github.com/jgolang/log"
)

func init() {
}

func TestLogger_Output(t *testing.T) {
	log.Info("Test message")
}
