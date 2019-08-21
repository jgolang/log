package log

import (
	"testing"
)

func TestLog(*testing.T) {
	ChangeCallerSkip(0)

	hola()
	hola2()
	ChangeCallerSkip(2)
	hola()
	hola2()
	ChangeCallerSkip(-2)
	hola()
	hola2()
}

func hola() {
	Println("hola")
}

func hola2() {
	hola()
}
