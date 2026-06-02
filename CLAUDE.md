# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

modulectl is a **CLI tool for Kyma module developers**. It provides commands to scaffold a new module structure and to package and push a module as an OCI artifact to a registry, ready for consumption by [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

It is a [Cobra](https://github.com/spf13/cobra)-based CLI written in Go. There is no operator, no controller, and no Kubernetes runtime dependency — it is a developer toolchain binary.

## Commands

| Command | What it does |
|---|---|
| `modulectl create` | Packages a module as an OCI artifact and pushes it to a registry |
| `modulectl scaffold` | Generates the files needed to start a new module (`module-config.yaml`, default CR, etc.) |
| `modulectl version` | Prints the current version |

Full flag reference: `docs/gen-docs/` (auto-generated — do not edit manually, run `make docs` to regenerate).

## Architecture

```
cmd/modulectl/          ← CLI entrypoint, Cobra command definitions
  create/               ← `create` command flags and wiring
  scaffold/             ← `scaffold` command flags and wiring
internal/service/       ← Business logic, one package per concern
  create/               ← Orchestrates module packaging
  scaffold/             ← Orchestrates scaffold generation
  componentconstructor/ ← Builds the OCM component descriptor
  contentprovider/      ← Parses module-config.yaml
  moduleconfig/         ← Reads and validates module configuration
  filegenerator/        ← Writes scaffold files
  git/                  ← Git source metadata
  verifier/             ← Validates module inputs
tools/                  ← Shared low-level utilities (filesystem, yaml, io)
```

The composition root is in `cmd/modulectl/cmd.go` — all dependencies are wired there, not inside the service packages.

## Build and test commands

modulectl builds cross-platform. There is no FIPS requirement — `CGO_ENABLED=0` with a plain `go build`.

| Target | What it does |
|---|---|
| `make build` | Build all four platform variants (darwin/linux × amd64/arm64) to `bin/` |
| `make build-darwin` | Build for macOS amd64 |
| `make build-linux` | Build for Linux amd64 |
| `make build-darwin-arm` | Build for macOS arm64 |
| `make build-linux-arm` | Build for Linux arm64 |
| `make test` | Unit tests with race detector (excludes e2e) |
| `make lint` | golangci-lint |
| `make docs` | Regenerate CLI reference docs in `docs/gen-docs/` |

### Running a single unit test

```sh
go test -run TestFoo ./internal/service/create/... -v -race
```

### E2E tests (`tests/e2e/`)

E2E tests require a local registry. Set it up first:

```sh
./scripts/re-create-test-registry.sh
./scripts/build-modulectl.sh
```

Then run individual suites:

| Target | What it does |
|---|---|
| `make test-create-cmd` | Run `create` command e2e suite |
| `make test-scaffold-cmd` | Run `scaffold` command e2e suite |

See `docs/contributor/local-test-setup.md` for the full local setup walkthrough.

## Code conventions

Go conventions load automatically when editing `.go` files — see [`.claude/rules/go-conventions.md`](.claude/rules/go-conventions.md).

Key rules from `.golangci.yaml`:
- **All linters enabled by default** — check `.golangci.yaml` before adding `//nolint`
- **`//nolint` requires explanation**: e.g., `//nolint:funlen // command wiring`
- **Import ordering** (gci): standard → third-party → project (`github.com/kyma-project/modulectl`)
- **Import aliases strictly enforced** — key ones: `iotools`, `commonerrors`, `scaffoldcmd`, `createcmd`, `moduleconfiggenerator`, `moduleconfigreader`, `ocmv1`, `ociartifacttypes`
- **Line length**: 120 chars | **Function length**: 80 lines / 50 statements | **Cyclomatic complexity**: 20
- `fmt.Print` is forbidden — use `fmt.Println` for any console output

## Commits and Pull Requests

- PRs are usually created from a **fork branch** against `main`.
- PRs are merged with **squash merge** — the PR title and description form the commit message.
- Follow [conventional commits](https://www.conventionalcommits.org/), enforced by `.github/workflows/lint-conventional-prs.yml`.
- PR title format: `<type>: <title>` where the title is one sentence explaining the reason for the changeset.
- Ask what type to use when creating a PR: `deps`, `chore`, `docs`, `feat`, `fix`, `refactor`, `test`.
- PR description should contain a short summary of the changes and, if applicable, a reference to the issue using the `closes` or `resolves` keyword.
- Never mention Claude or any AI agent in commits or PRs (no author attribution, no `Co-Authored-By`, no references in commit messages).

## Documentation

When reviewing or editing documentation in `docs/`, the SAP/Kyma technical writing styleguide loads automatically — see [`.claude/rules/documentation-style.md`](.claude/rules/documentation-style.md).

- `docs/gen-docs/` — auto-generated CLI reference (do not edit manually, use `make docs`)
- `docs/contributor/` — development and local test setup guides
- `docs/user/` — end-user guides
