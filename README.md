# Log v1

Structured logging built on top of `log/slog`.

## Overview

This package wraps `slog` with:

- package-level helpers for common log levels
- source metadata on each record
- stack traces on debug logs
- optional OpenTelemetry correlation via `otel`

## Installation

`go get github.com/jgolang/log`

## Quick Start

```go
package main

import "github.com/jgolang/log"

func main(){
    log.Info("My info....")
}
```

### Output:

```terminal
{"time":"2026-05-02T10:00:00Z","level":"INFO","msg":"My info....","source":{"func":"main","file":"main.go","line":6}}
```

The first argument is the message; the rest are key-value attributes. A string is used as is. An error becomes the message and also adds an `error` attribute holding the error. Any other value is printed with `fmt.Sprint`, so `log.Info(42, "request_id", id)` logs the message `42`.

Any attribute whose value is an error with a `jgolang/errors` error anywhere in its chain (also behind `fmt.Errorf`) is logged as a group under its own key, with `code`, `msg`, `debug` and `origin`. This applies to the first-argument `error` attribute, to later attributes (`log.Info("msg", "err", err)`), to the instance API (`logger.Error("failed", "error", err)`) and to attributes inside groups. The group is built before the handler, so handler wrappers such as the `otel` handler see it too. `code` and `msg` come from the first non-nil code in the chain, also below an error whose code is nil, and are left out when the chain has no code; `debug` and `origin` come from the first `jgolang/errors` error. Except for the first argument, whose message already carries it, the group starts with `text`, the full `err.Error()`, so the context added by outer wrappers is kept. Any other error, and an attribute with an empty key, is logged as is:

```terminal
{"level":"ERROR","msg":"failed","error":{"text":"loading board: [PLN-111](task not found): task 42 is missing","code":"PLN-111","msg":"task not found","debug":"task 42 is missing","origin":{"frame_0":"..."}},"source":{...}}
```

The whole chain is searched, depth-first as `errors.As` does, and the search stops at the first error that implements `slog.LogValuer`. An error type that must keep its wrapped detail out of logs should implement it: an error that is a `LogValuer`, or wraps one above its coded error (`fmt.Errorf("ctx: %w", redacting)`), is never expanded, in any position, and slog logs the value as usual.

## Configuration

```go
log.SetLevel(slog.LevelWarn)
log.NewTextHandler()
log.SetSource(true)
log.SetDebugStackTrace(false)
```

The package no longer depends on a `MODE` environment variable.
Debug stack traces are now opt-in.

`source` reports the code that calls the logging function. If you log through your own helper, raise the call depth by one per wrapper function so `source` points at the helper's caller:

```go
log.SetCalldepth(4) // default 3; Logger.SetCalldepth works the same way
```

## Fatal

`Fatal`, `FatalC` and the `Logger` methods `Fatal` and `FatalContext` log the message and then call `os.Exit(1)`. Like `os.Exit`, they do not run deferred functions: a `defer` that flushes OpenTelemetry providers, closes a database or writes buffered output is skipped. Flush and close what you need before calling them, or keep the deferred cleanup in a function that returns the error, so it runs before you exit:

```go
func main() {
	if err := run(); err != nil {
		log.Fatal(err) // run's defers have already run
	}
}

func run() error {
	defer shutdownTelemetry()
	// ...
	return nil
}
```

## Instance API

```go
logger := log.New(
    log.WithLevel(slog.LevelInfo),
    log.WithTextHandler(os.Stdout),
    log.WithSource(true),
    log.WithDebugStackTrace(false),
)

logger.Info("service started")
```

### Handler wrappers

`WithHandlerWrapper` (instance) and `SetHandlerWrapper` (package-level logger) wrap the JSON or text handler the logger builds with any `func(slog.Handler) slog.Handler`, such as the `otel` handler below. The built handler still encodes the record, so the level, the custom level names and `source` keep working. The wrapper is kept when the handler changes later (`SetJSONHandler`, `SetTextHandler`, `NewJSONHandler`, `NewTextHandler`), so the option order does not matter. A nil wrapper removes it.

## OpenTelemetry

`github.com/jgolang/log/otel` is its own module since `log` v1.5.0, so importing `jgolang/log` adds no OpenTelemetry requirement.

Upgrading from `log` v1.4.x or earlier, where `otel` was part of the root module, raise both together:

```sh
go get github.com/jgolang/log@v1.6.0 github.com/jgolang/log/otel@v1.6.0
```

Raising only `log` makes Go report that no required module provides `github.com/jgolang/log/otel`; `log/otel` v1.5.0 and later require `log` v1.5.0 or later, so adding it raises `log` too. New users run `go get github.com/jgolang/log@v1.6.0 github.com/jgolang/log/otel@v1.6.0`. `SetHandlerWrapper` and `WithHandlerWrapper`, used below, need `log` v1.6.0.

Wrap the logger's handler with `otel.NewOtelHandler` so records logged with a context carry its `trace_id` and `span_id` and are added to the span as events:

```go
log.SetHandlerWrapper(otel.NewOtelHandler())
log.InfoC(ctx, "order created")

// or, for an instance:
logger := log.New(log.WithHandlerWrapper(otel.NewOtelHandler()))
logger.InfoContext(ctx, "order created")
```

Span events carry the record's attributes and, from `log/otel` v1.7.0, the attributes added with the handler's `WithAttrs`, with the same keys and size caps, for example on a `*slog.Logger` built on the handler:

```go
logger := slog.New(otel.New(slog.NewJSONHandler(os.Stdout, nil))).With("request_id", id)
logger.InfoContext(ctx, "order created") // the span event carries request_id
```

On `log/otel` v1.6.x and earlier, `WithAttrs` attributes reach the log line only.

Span event string values are capped at 1 KiB and end with `…`. From `log/otel` v1.8.0, slice values (`[]string`, `[]int`, `[]int64`, `[]float64`, `[]bool`) also keep only their first 128 elements; the log line still gets the full record.

The `otel` handler disables baggage logging by default.
If you need baggage in logs, enable it explicitly and prefer an allow-list:

```go
log.SetHandlerWrapper(otel.NewOtelHandler(
    otel.WithNoBaggage(false),
    otel.WithBaggageAllowList("request_id", "tenant"),
))
```

Avoid enabling all baggage in production unless the upstream context is already
sanitized. Baggage can contain tenant identifiers, tokens, or other sensitive
values; use `WithBaggageAllowList`, `WithBaggageDenyList`, or
`WithBaggageFilter` to keep log output intentional.

## Verification

```bash
go test ./...
go test -race ./...
go vet ./...
go test -bench=. -run=^$ ./...
(cd otel && go test ./... && go test -race ./... && go vet ./...)
```

Use the benchmarks to compare source metadata, disabled levels, and debug stack
traces before changing logger internals.

<hr>

Released under the [MIT License](LICENSE).
