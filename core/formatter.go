package core

import "time"

// Formatter define custom format output to implement in a new logger
type Formatter interface {
	Development(buf *[]byte, t time.Time, file string, line int, function string, p Priority, s string, stack [][]byte)
	Production(buf *[]byte, t time.Time, file string, line int, function string, p Priority, s string, stack [][]byte)
}
