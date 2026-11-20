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

The first argument is the message; the rest are key-value attributes. A string is used as is. An error becomes the message and also adds an `error` attribute: for a `jgolang/errors` error anywhere in its chain (also behind `fmt.Errorf`) it is a group with `code`, `msg`, `debug` and `origin` (`code` and `msg` come from the first non-nil code in the chain, also below an error whose code is nil, and are left out when the chain has no code; `debug` and `origin` come from the first `jgolang/errors` error); for any other error it is the error itself. Any other value is printed with `fmt.Sprint`, so `log.Info(42, "request_id", id)` logs the message `42`.

## Configuration

```go
log.SetLevel(slog.LevelWarn)
log.NewTextHandler()
log.SetSource(true)
log.SetDebugStackTrace(false)
```

The package no longer depends on a `MODE` environment variable.
Debug stack traces are now opt-in.

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

## OpenTelemetry

`github.com/jgolang/log/otel` is its own module since `log` v1.5.0, so importing `jgolang/log` adds no OpenTelemetry requirement.

Upgrading from `log` v1.4.x or earlier, where `otel` was part of the root module, raise both together:

```sh
go get github.com/jgolang/log@v1.5.0 github.com/jgolang/log/otel@v1.5.0
```

Raising only `log` makes Go report that no required module provides `github.com/jgolang/log/otel`; `log/otel` v1.5.0 requires `log` v1.5.0, so adding it raises `log` too. New users run `go get github.com/jgolang/log/otel`.

The `otel` handler now disables baggage logging by default.
If you need baggage in logs, enable it explicitly and prefer an allow-list:

```go
handler := otel.New(
    slog.NewJSONHandler(os.Stdout, nil),
    otel.WithNoBaggage(false),
    otel.WithBaggageAllowList("request_id", "tenant"),
)
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

Released under the [GPL-3.0](LICENSE).
