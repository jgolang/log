package log

import (
	"io"
)

type out interface {
	Write(p []byte) error
}

// Out doc
type Out struct {
	// wr is standar output
	wr io.Writer
}

// Write doc ..
func (o Out) Write(p []byte) error {
	_, err := o.wr.Write(p)
	return err
}
