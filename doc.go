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
// # Import path
//
// Import github.com/jgolang/log, this package, for both the package-level
// helpers (Info, ErrorC, ...) and the instance API ([New] with options such
// as [WithLevel]). Package github.com/jgolang/log/logger is the
// implementation behind it: [Logger] and [Option] are aliases of its types,
// so the same types are reachable from both paths. Import logger only for
// what this package does not re-export, such as the custom level
// logger.LevelTrace.
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
// attribute holding the error. Any other value is printed with fmt.Sprint.
//
// Any attribute whose value is an error with a jgolang/errors error anywhere
// in its chain (also behind fmt.Errorf) is logged as a group under its own
// key, with code, msg, debug and origin, for the package helpers and the
// Logger methods alike, also inside groups. The group is built before the
// handler, so handler wrappers see it too. code and msg come from the first
// non-nil code in the chain and are left out when there is none; debug and
// origin come from the first jgolang/errors error. Except for the first
// argument, whose message already carries it, the group starts with text,
// the full err.Error(). Any other error, and an attribute with an empty key,
// is logged as is. The chain is searched depth-first and the search stops at
// an error that implements slog.LogValuer, so an error that is one, or wraps
// one above its coded error, is never expanded: a type that must hide its
// wrapped detail implements it.
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
