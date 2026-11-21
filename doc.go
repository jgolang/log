// GNU GENERAL PUBLIC LICENSE
// Version 3, 29 June 2007
//
// Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
// Everyone is permitted to copy and distribute verbatim copies
// of this license document, but changing it is not allowed.

// Package log provides structured logging helpers built on top of log/slog.
//
// The package adds source metadata to every record, stack traces to debug
// records, and OpenTelemetry integration through the separate module
// github.com/jgolang/log/otel, so importing this package adds no
// OpenTelemetry requirement.
//
// # Installation
//
// Run command in terminal:
//
//	go get github.com/jgolang/log
//
// # Quick Start
//
// This is a simple example of how the package is implemented with a basic function.
//
//	package main
//
//	import "github.com/jgolang/log"
//
//	func main(){
//	    log.Info("My info....")
//	}
//
// Output:
//
//	{"time":"2026-05-02T10:00:00Z","level":"INFO","msg":"My info....","source":{"func":"main","file":"main.go","line":6}}
//
// The first argument is the message; the rest are key-value attributes. A
// string is used as is. An error becomes the message and also adds an "error"
// attribute: for a jgolang/errors error anywhere in its chain (also behind
// fmt.Errorf) it is a group with code, msg, debug and origin (code and msg come
// from the first non-nil code in the chain and are left out when there is
// none; debug and origin come from the first jgolang/errors error); for any other error it is the error
// itself. Any other value is printed with fmt.Sprint.
//
// # Fatal
//
// Fatal, FatalC and the Logger methods Fatal and FatalContext call os.Exit(1)
// after logging, so deferred functions do not run. Keep deferred cleanup in a
// function that returns the error, and call Fatal after it returns.
//
// # Configuration
//
// You can configure the current level and handler programmatically:
//
//	log.SetLevel(slog.LevelWarn)
//	log.NewTextHandler()
//	log.SetSource(true)
//	log.SetDebugStackTrace(false)
//
// The package does not depend on environment variables, and debug stack
// traces are opt-in.
//
// SetHandlerWrapper and the WithHandlerWrapper option wrap the handler the
// logger builds, for example with the otel module's handler, which adds trace
// correlation to records logged with a context:
//
//	log.SetHandlerWrapper(otel.NewOtelHandler())
//
// The otel module (github.com/jgolang/log/otel) disables baggage logging by default. If baggage is
// enabled, prefer allow-lists or explicit filters so sensitive context values
// are not written to logs accidentally.
//
// You can also build isolated logger instances:
//
//	logger := log.New(
//	    log.WithLevel(slog.LevelInfo),
//	    log.WithTextHandler(os.Stdout),
//	    log.WithSource(true),
//	    log.WithDebugStackTrace(false),
//	)
//
//	logger.Info("service started")
package log
