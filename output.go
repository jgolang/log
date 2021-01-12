package log

import (
	"io"
)

// Output doc
type Output struct {
	// wr is standar output
	wr io.Writer
}

// Write doc ..
func (o Output) Write(p []byte) error {
	_, err := o.wr.Write(p)
	return err
}
