# aienv - AI Environment Manager

Go CLI tool for managing sandboxed, permissions-controlled AI coding environments.

## Build

- `make build` — build binary
- `make install` — `go install .`

## Test

- `make test` — run all tests
- `make test-verbose` — verbose output
- `make test-race` — with race detector
- `make coverage` — test + coverage report
- `make coverage-html` — open HTML coverage

## Quality

- `make vet` / `make lint` — `go vet`
- `make fmt` — format all Go files

## Misc

- `make clean` — remove binary and coverage artifacts

## Project Structure

- `cmd/` — cobra commands (create, up/activate, list, show, edit, delete, build, clean)
- `internal/` — core logic (unexported packages)
  - `audit/` — JSONL audit log schema and writer
  - `config/` — XDG paths, hash computation, session ID generation
  - `docker/` — Docker sandbox (build, run, proxy, trust prompts)
  - `env/` — Env struct, YAML load/save/validate

## Conventions

- Standard Go project layout with `internal/` for unexported packages
- Commands in `cmd/` are thin — logic lives in `internal/`
- `config.TestDataDir` and `config.TestTrustDir` overrides for test isolation

## CLI

- `aienv create <name>` — interactive env creation
- `aienv <name>` / `aienv up <name>` — build image if needed, launch sandbox
- `aienv list` / `aienv show <name>` — inspect environments
- `aienv edit <name>` — edit env YAML in $EDITOR
- `aienv build <name>` — force rebuild Docker image
- `aienv delete <name>` — remove env + image + audit data
- `aienv clean` — remove orphaned images, audit dirs, trust cache

## Docs

- `docs/` — Implementation details, roadmap, architecture, use cases
- `CONTEXT.md` — Domain glossary (black box agent, trust, audit, learn mode)
- `docs/adr/` — Architecture Decision Records

## Agent skills

### Issue tracker

Issues live in GitHub Issues (uses `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Uses default triage label names: needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: one CONTEXT.md + docs/adr/ at repo root. See `docs/agents/domain.md`.
