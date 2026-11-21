# Release Checklist

Before tagging a release:

- Run `gofmt -l .` (no output), `go test ./...`, `go test -race ./...`, `go vet ./...`, and `govulncheck ./...` in the root and in `otel/` (its own module). CI runs the same checks; do not tag a red master.
- Run `go test -bench=. ./...` when logging performance changed.
- Review exported API changes with `go doc ./...`; deprecate before removing public names.
- Confirm README examples still match current defaults.
- Confirm OTel baggage logging remains opt-in and documented with filtering guidance.

Tags:

- The root module is tagged `vX.Y.Z`; the `otel` module is tagged `otel/vX.Y.Z`.
- Release in dependency order: `errors`, then `log`, then `otel`. Before every `otel/vX.Y.Z` tag, raise `github.com/jgolang/log` in `otel/go.mod` to the latest `log` tag (its `errors` requirement follows), even when `otel` does not need the new code: otherwise `otel` users keep an older `log` and `errors`, and dependencies those releases dropped, in their module graph.
- Do not run `go mod tidy` in `otel/` expecting it to drop `github.com/jgolang/log`: `otel` imports it on purpose so `log/otel` raises `log` past v1.4.x (where `otel` was part of the root module) and avoids an ambiguous import.
