# Release Checklist

Before tagging a release:

- Run `go test ./...`, `go test -race ./...`, and `go vet ./...`.
- Run `go test -bench=. ./...` when logging performance changed.
- Review exported API changes with `go doc ./...`; deprecate before removing public names.
- Confirm README examples still match current defaults.
- Confirm OTel baggage logging remains opt-in and documented with filtering guidance.
