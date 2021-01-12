package core

import "time"

// Formater interface for implement a logger...
type Formater interface {
	Development(buf *[]byte, t time.Time, file string, line int, function string, p Priority, s string, stack [][]byte)
	Production(buf *[]byte, t time.Time, file string, line int, function string, p Priority, s string, stack [][]byte)
}
