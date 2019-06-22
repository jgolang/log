package log

import (
	"testing"
)

func TestLog(*testing.T) {
	hola()
	ChangeCallerSkip(2)
	hola()
	hola()
}

func hola() {
	Println("hola")
}
