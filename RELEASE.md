# Release Checklist

Before tagging a release:

- Run `gofmt -l .` (no output), `go test ./...`, `go test -race ./...`, `go vet ./...`, and `govulncheck ./...` in the root and in `otel/` (its own module). CI runs the same checks; do not tag a red master.
- Run `go test -bench=. ./...` when logging performance changed.
- Review exported API changes with `go doc ./...`; deprecate before removing public names.
- Confirm README examples still match current defaults.
- Confirm OTel baggage logging remains opt-in and documented with filtering guidance.

Tags:

- The root module is tagged `vX.Y.Z`; the `otel` module is tagged `otel/vX.Y.Z`.
- Do not run `go mod tidy` in `otel/` expecting it to drop `github.com/jgolang/log`: `otel` imports it on purpose so `log/otel` raises `log` past v1.4.x (where `otel` was part of the root module) and avoids an ambiguous import.
