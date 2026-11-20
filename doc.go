// GNU GENERAL PUBLIC LICENSE
// Version 3, 29 June 2007
//
// Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
// Everyone is permitted to copy and distribute verbatim copies
// of this license document, but changing it is not allowed.

// Package log provides structured logging helpers built on top of log/slog.
//
// The package adds source metadata to every record, stack traces to debug
// records, and OpenTelemetry integration through the otel subpackage.
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
// fmt.Errorf) it is a group with code, msg, debug and origin (code and msg are
// left out when the error has no code); for any other error it is the error
// itself. Any other value is printed with fmt.Sprint.
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
// The otel subpackage disables baggage logging by default. If baggage is
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
