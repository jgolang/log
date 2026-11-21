# 0002 — log is licensed under MIT

- Status: accepted (2026-11-21, owner decisions Q23 and the epic 8 retrospective follow-up)
- Release: log v1.7.1, otel/v1.6.1 (#41)

## Context

log and log/otel were GPL-3.0, while api, errors and helpers are MIT. Any distributed program importing log had to be offered under the GPL, which kept the libraries from being used together under one license.

## Decision

Relicense log and log/otel from GPL-3.0 to MIT, from v1.7.1 and otel/v1.6.1. The root `LICENSE` covers the otel module.

The repository has one author. The epic 8 retrospective found six commits from 2019 (`a93cbb4` to `b9fdd9b`, the initial module setup) recorded with a work e-mail domain. The owner confirmed that this code is the owner's own work, so the owner can relicense all of it.

## Consequences

- Releases before v1.7.1 and otel/v1.6.1 stay GPL-3.0; the license applies per release.
- All active jgolang libraries are MIT.
