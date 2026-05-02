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

## Configuration

```go
log.SetLevel(slog.LevelWarn)
log.NewTextHandler()
```

The package no longer depends on a `MODE` environment variable.

<hr>

Released under the [GPL-3.0](LICENSE).
