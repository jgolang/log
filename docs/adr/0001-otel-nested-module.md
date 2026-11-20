# 0001 — `log/otel` is its own module

- Status: accepted (2026-11-21, owner decision Q5)
- Issue: LOG-04 (#18)

## Context

The root `go.mod` required OpenTelemetry although only the `otel` package used it, so every importer of `jgolang/log` got an OTel version floor (constitution Principle VI).

## Decision

- `github.com/jgolang/log/otel` is a nested module with its own `go.mod`, keeping OTel at v1.28.0 so its users get no new floor.
- Root `log` v1.5.0 no longer contains `otel` and has no OpenTelemetry requirement.
- `log/otel` imports `github.com/jgolang/log` and requires v1.5.0 or later, so adding it raises `log` past the versions that still contained the package and Go never sees two modules providing `github.com/jgolang/log/otel`.
- Tags: root `vX.Y.Z`, nested `otel/vX.Y.Z`. Release order: root `v1.5.0` first, then `otel/v1.5.0`.

## Consequences

- Existing `log/otel` users upgrade with `go get github.com/jgolang/log@v1.5.0 github.com/jgolang/log/otel@v1.5.0`; raising only `log` fails with "no required module provides package".
- Releases run checks in both modules (RELEASE.md).
