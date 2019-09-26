package log

import (
	"testing"
)

func TestLog(*testing.T) {
	ChangeCallerSkip(0)

	hola()
	hola2()
	ChangeCallerSkip(1)
	hola()
	hola2()
	ChangeCallerSkip(-2)
	hola()
	hola2()
}


