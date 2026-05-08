---
paths:
  - "**/*.go"
---

# Go code conventions — modulectl

`make lint` is the authoritative check. The full linter config is in `.golangci.yaml`.

## Import aliases

Strict aliases are enforced by `importas` — violations fail CI. The **complete alias list** is in `.golangci.yaml` under `linters-settings.importas.alias` (13 entries). Check that file before adding an import.

## nolint policy

Every `//nolint` directive **must** include an explanation:
```go
//nolint:funlen // service wiring — acceptable exception
```
Bare suppressions fail CI (`nolintlint` is enabled). Check `.golangci.yaml` before adding any.

## Static binary

All builds use `CGO_ENABLED=0` — no C dependencies. Do not add a dependency that requires cgo. Cross-compilation for all four targets (darwin/linux × amd64/arm64) must remain possible.

## Generated mocks

Mocks live in `internal/service/<name>/mocks/`. After changing an interface, run `make generate` to regenerate them. Stale mocks cause test failures that are hard to diagnose.
