# aienv - AI Environment Manager

Go CLI tool for managing task-specific AI coding environments.

## Build

- `go build ./...`
- `go install .`

## Test

- `go test ./...`

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

## Key Design Docs

Design research and decisions are in an Obsidian vault at:
`~/gdrive/obsidian/Personal/Ideas/Research Notes/AI Env/`
