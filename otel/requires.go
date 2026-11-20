package otel

// This blank import makes the module require github.com/jgolang/log v1.5.0 or
// later. Versions before v1.5.0 still contained this package, so without it a
// program could select both modules and get an ambiguous import for
// github.com/jgolang/log/otel. See docs/adr/0001-otel-nested-module.md.
import _ "github.com/jgolang/log"
