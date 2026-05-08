# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

modulectl is a **CLI tool for Kyma module developers**. It automates the two most common tasks when building a Kyma module:

1. **`scaffold`** — generate boilerplate files (module config, manifest, default CR, security config) for a new module
2. **`create`** — package a module into a `ModuleTemplate` CR and component constructor YAML for release

It is not a Kubernetes operator — there is no controller-runtime, no reconciler, no CRDs.

## Module & language

- Module: `github.com/kyma-project/modulectl` (Go 1.26.1)
- CLI framework: `github.com/spf13/cobra` v1.10.2
- Key dependencies: `go-git/v5` (git operations), `go-containerregistry` (image inspection), `lifecycle-manager/api` (ModuleTemplate types), `Masterminds/semver/v3` (version validation)

## Make targets

| Target | What it does |
|---|---|
| `make build` | Cross-compile for all 4 platforms (darwin/linux × amd64/arm64) into `bin/` |
| `make build-darwin` | macOS amd64 only → `bin/modulectl-darwin` |
| `make build-linux` | Linux amd64 only → `bin/modulectl-linux` |
| `make test` | Unit tests with race detector, excludes e2e (fast, no external deps) |
| `make lint` | golangci-lint |
| `make docs` | Regenerate CLI documentation in `docs/gen-docs/` from cobra commands |
| `make validate-docs` | Verify generated docs are up to date (CI gate) |

**Version is injected at build time** from git: `<branch>-<short-sha>`. Pass `VERSION=v1.2.3` to `make build` for release builds.

### Running a single test

```sh
go test -run TestFoo ./internal/service/create/...
```

E2E tests require a local Docker registry and are run separately:
```sh
./scripts/re-create-test-registry.sh   # one-time setup
./scripts/build-modulectl.sh
./scripts/run-e2e-test.sh --cmd=create    # or --cmd=scaffold
```

## Architecture

### Command → Service → Tools

```
cmd/modulectl/
  cmd.go            ← dependency injection / wiring (Cobra root command)
  create/cmd.go     ← flag parsing, calls internal/service/create
  scaffold/cmd.go   ← flag parsing, calls internal/service/scaffold

internal/service/
  create/           ← orchestrates module packaging
  scaffold/         ← orchestrates file generation
  moduleconfig/     ← parses/validates module-config.yaml
  componentdescriptor/ ← builds OCM component descriptors
  manifestparser/   ← extracts images from Kubernetes manifests
  crdparser/        ← schema-validates default CRs against CRD definitions
  git/              ← extracts commit info for component descriptors
  image/            ← inspects OCI images via go-containerregistry
  filegenerator/    ← writes output files
  fileresolver/     ← resolves local paths and remote URLs to content

tools/
  filesystem/       ← file I/O utilities
  yaml/             ← YAML marshalling helpers
  io/               ← output formatting
```

`cmd/modulectl/cmd.go` is the **composition root** — all services are wired here via constructor injection. Keep business logic out of command files; they only parse flags and call services.

### Command descriptions

Each command's `Use`, `Short`, `Long`, and `Example` strings live in embedded `.txt` files inside the command directory (`use.txt`, `short.txt`, `long.txt`, `example.txt`). Edit those files, not the Go strings directly.

## Code conventions

- **All builds use `CGO_ENABLED=0`** — static binaries, no C dependencies
- **Interface-driven design** throughout `internal/service/` — every external dependency (filesystem, git, image registry) is injected as an interface, enabling unit testing without real external systems
- **Error types** live in `internal/common/errors/` — use typed errors, not `fmt.Errorf` with sentinel strings
- **Generated docs** (`docs/gen-docs/`) must be kept in sync with cobra commands — `make validate-docs` fails CI if they drift. Always run `make docs` after changing command flags or descriptions.

## Testing

- Unit tests: co-located with source (`*_test.go` alongside each file), use `testify` and `gomega`
- E2E tests: `tests/e2e/` with Ginkgo, test the full CLI binary against a real local registry
- `make test` runs only unit tests — safe to run anytime with no external setup
- See `docs/contributor/local-test-setup.md` for full e2e environment setup

## Code conventions

Go import aliases, nolint policy, and static-binary constraints load automatically when editing `.go` files — see [`.claude/rules/go-conventions.md`](.claude/rules/go-conventions.md).

## CVE triage

Two scanners run against this repo (`sec-scanners-config.yaml`): **Checkmarx One** (SAST) and **Mend** (Go module SCA). No BDBA — modulectl has no container image. When triaging a CVE finding, see [`.claude/cve-triage/context.md`](.claude/cve-triage/context.md).

## Model usage

Follow the Kyma team's Claude Code workflow:

- **Planning complex tasks** — switch to Opus: `/model claude-opus-4-7`
- **Implementation** — use the default Sonnet: `/model claude-sonnet-4-6`

Use Opus when you need to understand an unfamiliar subsystem, design a non-trivial change, or reason about cross-cutting impacts. Switch back to Sonnet once the approach is clear and you are writing code.
