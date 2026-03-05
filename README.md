# Go Backend Template

Template for Go backends using:

- DDD (rich domain model)
- CQRS-lite (read/write path separation)
- Anti-Corruption Layer (external contract isolation)

`agents.md` is the architectural source of truth for contributors and AI agents. This README is a practical, developer-focused guide aligned with those rules.

## Architecture Principles

- Keep business rules and invariants in `internal/model`.
- Keep application orchestration in `internal/service`.
- Keep transport concerns (HTTP/gRPC/consumer DTOs and mapping) in `internal/handler`.
- Keep persistence concerns in `internal/repository` with explicit storage-to-domain/query mapping.
- Keep external vendor contracts inside `internal/integration` adapters only.
- Keep dependencies flowing downward only:
	- `cmd -> handler -> service -> model`
	- `service -> repository` (via ports/interfaces)
	- `service -> integration` (via ports/interfaces)
	- `handler -> read_repository` is allowed for CQRS-lite reads

## Project Layout

```text
.
|- cmd/                     # app entry points, wiring, startup only
|- db/
|  |- migrations/           # SQL migrations
|  `- schema.sql            # schema snapshot/base schema
|- internal/
|  |- config/               # config model, load, validation
|  |- di/                   # dependency injection container
|  |- errors/               # shared sentinel errors
|  |- handler/              # transport layer
|  |- integration/          # ACL for external systems
|  |- model/                # domain entities, value objects, invariants
|  |- repository/           # storage implementations
|  `- service/              # use-case orchestration
|- logs/                    # runtime logs (local/dev)
|- Taskfile.yaml            # format/lint/test/migration tasks
`- Makefile                 # build/run helpers
```

## Prerequisites

- Development is supported on Unix-like systems only (Linux, macOS, BSD).
- Go `1.25.0` (see `go.mod`)
- `task` CLI (https://taskfile.dev)
- `make`

## Quick Start

1. Clone the repository.
2. Adjust `config.yaml` if needed.
3. Run the app:

```bash
make run
```

Equivalent command:

```bash
go run ./cmd/main.go
```

You can pass a custom config path:

```bash
go run ./cmd/main.go -config ./config.yaml
```

## Configuration

Current config root is:

- `logger`

Config loading behavior:

- Reads YAML from a path provided via CLI flag (`-config`).
- Expands environment templates using `os.ExpandEnv` (`$VAR` or `${VAR}`).
- Validates the resulting struct using `go-playground/validator`.

Important note:

- Missing environment variables are expanded to an empty string by `os.ExpandEnv`.
- Any critical settings should be validated as required in config validation tags.

Logging in this template uses `zerolog`.

- Documentation: https://github.com/rs/zerolog
- Commented logging configuration example: `config.yaml`

## Development Commands

Quality and tooling commands are defined in `Taskfile.yaml`.
The same `task` commands are also used in CI pipelines to keep local and CI checks consistent.

```bash
task format       # gofumpt + gci
task lint         # golangci-lint
task test         # go test ./...
task vuln         # govulncheck
task generate     # go generate ./...
```

Build and run commands are in `Makefile`.

```bash
make build
make build-all
make run
make clean
```

## PR CI (GitHub Actions)

Pull request workflow is defined in `.github/workflows/pr.yml`.

It runs on PR events:

- `opened`
- `synchronize`
- `reopened`
- `ready_for_review`

Checks executed for non-draft PRs:

- `lint` -> `task lint`
- `test` -> `task test`
- `quality` ->
	- `task format` + `git diff --exit-code` (no formatting drift)
	- `task generate` + `git diff --exit-code` (generated files are up to date)
- `security` -> `task vuln` (`govulncheck`)

## Release CI (GitHub Actions)

Release workflow is defined in `.github/workflows/release.yml` and starts only on tag push.

Supported tag patterns:

- `v*.*.*` (stable), for example: `v1.2.3`
- `v*.*.*-alpha.*`, for example: `v1.2.3-alpha.1`
- `v*.*.*-beta.*`, for example: `v1.2.3-beta.1`
- `v*.*.*-rc.*`, for example: `v1.2.3-rc.1`

Important behavior:

- Tags can point to commits from any branch (no `main`-only restriction).
- Any tag containing `-` is treated as prerelease.
- Stable tags additionally publish Docker tag `latest`.
- Workflow builds and pushes:
	- multi-arch Docker image (`linux/amd64`, `linux/arm64`) to `ghcr.io/<owner>/<repo>`
	- compiled binaries from `dist/*` to GitHub Release

How to cut a release:

```bash
git tag v0.0.1-beta.1
git push origin v0.0.1-beta.1
```

## Database Migrations

Set `DATABASE_URL` in Go DSN format before running migration commands.

PostgreSQL example:

```bash
export DATABASE_URL='postgres://user:password@localhost:5432/dbname?sslmode=disable'
```

Migration workflow:

```bash
task migration-new -- add_orders_table
task migration-up
task migration-down
```

Migration quality requirements:

- Each migration must include both `-- migrate:up` and `-- migrate:down`.
- `down` must be meaningfully reversible.
- Validate apply and rollback before merge.

## Module Rename Helper

If you are bootstrapping a new project from this template, `setup.sh` can replace the module path safely with a dry-run + confirmation workflow.