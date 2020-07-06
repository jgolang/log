package log_test

import "github.com/jgolang/log"

func ExampleStackTrace_stacktrace() {
	// Use this function to see the trace of execution of the sentence.
	// This function is useful for tracking where errors are generated.
	func() {
		log.StackTrace("My message...")
	}()
	// Output:
	// 2020/07/05 21:27:22     ERROR   /Users/me/myWorkdir/example/example_test.go:43 (func1)       My message...
	// --- TRACE:
	// 		/Users/me/myWorkdir/example/fileA.go:43      func1()
	// 		/Users/me/myWorkdir/example/fileA.go:43      ExampleStackTrace()
	// 		/Users/me/myWorkdir/example/fileB.go:20      funcLevel1()
	// 		/Users/me/myWorkdir/example/fileC.go:17      funcLevel2()
	// 		...
	// ---
}
