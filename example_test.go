package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jgolang/log"
)

// func ExampleStackTrace_stacktrace() {
// 	// Use this function to see the trace of execution of the sentence.
// 	// This function is useful for tracking where errors are generated.
// 	func() {
// 		log.StackTrace("My message...")
// 	}()
// 	// Output:
// 	// 2020/07/05 21:27:22     INFO   example_test.go:43 (func1)       My message...
// 	// --- TRACE:
// 	// 		/Users/me/myWorkdir/example/fileA.go:43      func1()
// 	// 		/Users/me/myWorkdir/example/fileA.go:43      ExampleStackTrace()
// 	// 		/Users/me/myWorkdir/example/fileB.go:20      funcLevel1()
// 	// 		/Users/me/myWorkdir/example/fileC.go:17      funcLevel2()
// 	// 		...
// 	// ---
// }

func ExampleInfo_info() {
	// Use this function to see the trace of execution of the sentence.
	// This function is useful for tracking where errors are generated.
	log.Info("Hello world!")
}

// serviceHandler tags every record with the service name before the handler
// it wraps encodes it.
type serviceHandler struct {
	slog.Handler
}

func (h serviceHandler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(slog.String("service", "billing"))
	return h.Handler.Handle(ctx, record)
}

func ExampleWithHandlerWrapper() {
	var buf bytes.Buffer
	logger := log.New(
		// The wrapper applies to the JSON handler whatever the option order.
		log.WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return serviceHandler{Handler: next}
		}),
		log.WithJSONHandler(&buf),
	)

	logger.Info("payment captured")

	var record struct{ Level, Msg, Service string }
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(record.Level, record.Msg, record.Service)
	// Output: INFO payment captured billing
}
