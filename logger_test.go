package log

import (
	"testing"
)

func TestLogger_Output(t *testing.T) {
	func() {
		std.SetNewFormat(NewFormaterConfig(LstdDevFlags, LstdProdFlags, "test"))
		Info(std.GetMode())
	}()
}
