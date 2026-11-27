# AGENTS.md

## Project

**log**: Structured logging for Go over log/slog with source metadata, debug stack traces and jgolang/errors groups; optional OpenTelemetry module log/otel

- **Stack:** Go (`go.mod`).
- **Layout** (from the card `context/repos/log.md` of the workbench):
  - `log.go`, `msg.go`, `error.go`, `debug.go`, `stack.go`: the package-level logger and helpers (`log.go:19`).
  - `config.go`: the instance API, `New` and the `With*` options, plus `SetSource`, `SetHandlerWrapper` (`config.go:19`).
  - `logger/`: the `Logger` implementation that `log.Logger` aliases (`config.go:13`, `logger/logger.go:30`).
  - `internal/codederror/`: builds the log group of a `jgolang/errors` error (`internal/codederror/codederror.go:1`).
  - `otel/`: nested module with the `OtelHandler` wrapper (`otel/handler.go:195`, `otel/handler.go:207`).
  - `docs/adr/`, `RELEASE.md`: decisions and the release checklist.
- **Workbench:** part of the jgolang project; its workbench is `jgolang/workbench-jgolang`.

## Commands

```sh
go vet ./...                                             # from CI: .github/workflows/ci.yml
go test -race -cover ./...                               # from CI: .github/workflows/ci.yml
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...   # from CI: .github/workflows/ci.yml
go mod edit -replace github.com/jgolang/log=../          # from CI: .github/workflows/ci.yml
go mod tidy                                              # from CI: .github/workflows/ci.yml
go test -race ./...                                      # from CI: .github/workflows/ci.yml
```

## Inherited rules

Read [docs/agents/foundations.md](docs/agents/foundations.md) (global rules, managed by the template) and [docs/agents/project.md](docs/agents/project.md) (this project's rules, synced from the workbench by agents-sync) before you work. The rules here and the invariants below apply even if you cannot read them.

## Invariants

No level can change these, not even with a deviation:

- **I1.** Never commit secrets, credentials or personal data to git.
- **I2.** Never force-push or rewrite the history of the main branch or of a branch with an open PR.
- **I3.** Nothing enters the main branch without a PR and without its checks passing.
- **I4.** Never disable or skip tests or checks to make something pass.
- **I5.** Report vulnerabilities privately, never in public issues.
- **I6.** A change is not done until the documentation it affects is updated.
- **I7.** Ask for the owner's decisions, do not assume them; confirm what is hard to revert before doing it.
- **I8.** Respect licenses and ownership: do not copy code or content from third parties or other organizations without permission.

## Deviations

None. When this repo changes something a higher level (global, organization or project) defines, add one line here: what changes, why, and a link to the issue or decision. Delete the line when it no longer applies. Invariants cannot be deviated.
