package log

// StackTrace allows you to view the exact place where the error or incident originated within the code.
// Shows a trace of up to 10 layers from where the error or incident was generated.
func StackTrace(v interface{}) {
	std.StackTrace(v)
	return
}
