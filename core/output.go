package core

// Output define custom destination out to implement in a new logger
type Output interface {
	Write(p []byte) error
}
