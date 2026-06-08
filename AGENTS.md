# aienv - AI Environment Manager

Go CLI tool for managing task-specific AI coding environments.

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

- `cmd/` — cobra commands (create, activate, list, show, edit, delete, init, docker)
- `internal/` — core logic
  - `env/` — Env struct, YAML load/save/validate
  - `opencode/` — opencode.json generation
  - `skills/` — verify and install skills
  - `registry/` — MCP registry and skills.sh API clients
  - `assets/` — curated MCP/skill YAML loader
  - `docker/` — Docker sandbox container
  - `shell/` — shell function install
  - `config/` — path helpers
- `curated/` — YAML-backed curated MCP and skill lists, bundled via embed.FS

## Conventions

- Standard Go project layout with `internal/` for unexported packages
- Commands in `cmd/` are thin — logic lives in `internal/`
- Persistent flags on root (`--model`, `--docker`, `--prompt`) passed to activate

## Docs

- `docs/` — Implementation details, roadmap, architecture, use cases
- Obsidian vault: `~/gdrive/obsidian/Personal/Ideas/Research Notes/AI Env/`
