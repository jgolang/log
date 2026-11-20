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

The first argument is the message; the rest are key-value attributes. A string is used as is. An error becomes the message and also adds an `error` attribute: for a `jgolang/errors` error anywhere in its chain (also behind `fmt.Errorf`) it is a group with `code`, `msg`, `debug` and `origin` (`code` and `msg` are left out when the error has no code); for any other error it is the error itself. Any other value is printed with `fmt.Sprint`, so `log.Info(42, "request_id", id)` logs the message `42`.

## Configuration

```go
log.SetLevel(slog.LevelWarn)
log.NewTextHandler()
log.SetSource(true)
log.SetDebugStackTrace(false)
```

The package no longer depends on a `MODE` environment variable.
Debug stack traces are now opt-in.

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
```

Use the benchmarks to compare source metadata, disabled levels, and debug stack
traces before changing logger internals.

<hr>

Released under the [GPL-3.0](LICENSE).
