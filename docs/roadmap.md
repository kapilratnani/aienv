# Roadmap

## V1 Complete: Docker Sandbox + Black Box Agents

- [x] Docker-only isolation (no shell function, no eval)
- [x] Black box agents — agents are YAML, no per-agent Go code
- [x] Build-time images from env YAML, cached by content hash
- [x] Interactive create flow with mount/deps/permissions prompts
- [x] Network proxy with allow/deny enforcement
- [x] Learn mode (proxy logs all hosts, suggests allowlist on exit)
- [x] Audit logging (JSONL network records per session)
- [x] Trust prompt on first activation, cached by env YAML hash
- [x] Session ID generation, audit dir per activation
- [x] `aienv clean` — orphaned images/audit/trust cleanup
- [x] `aienv shell` — interactive debug shell in sandbox
- [x] XDG-compliant paths (`~/.local/share/aienv/`, `~/.config/aienv/trust/`)

## V2: Hardening

- [ ] Cost extraction via wrapper script (post-session)
- [ ] `aienv audit` subcommand (aggregate reports from JSONL)
- [ ] Commands audit (auditd inside container, execve logging)
- [ ] `aienv export` / `aienv import` (share env YAML)
- [ ] Repo-local `.aienv.yaml` discovery (`aienv up` in project dirs)
- [ ] Permission signing via GPG/Sigstore

## V3: Desktop UI

- Wails-based desktop app at `cmd/wails/` — separate binary, reuses `internal/` packages
- Vue 3 + TypeScript + Vite frontend with xterm.js embedded terminal
- Full replacement of CLI for common tasks: env CRUD, activation, audit
- Phase 1: Shell & Env CRUD (dashboard, create wizard, detail/edit views)
- Phase 2: Permissions & Trust (visual editor for network rules)
- Phase 3: Embedded Terminal (Terminal interface, creack/pty, xterm.js, Docker session launch)
- Phase 4: Audit Viewer (network JSONL, cost extraction, session list/detail)
- Separate `go.mod` for Wails — no dependency bloat on CLI binary
- Linux + macOS only (Windows deferred to ConPTY)
- Design doc: `docs/desktop-ui.md`

## Previously: Black Box Agent Refactor

The original architecture had per-agent Go implementations (OpenCode, Claude Code) that generated agent config files and managed MCPs/skills. This was deleted in favor of treating agents as black boxes specified entirely through YAML. See `docs/adr/0005-black-box-agents.md`.

### Deleted / Superseded

- Per-agent Go code (`internal/agents/`, `internal/skills/`, `internal/registry/`, `internal/assets/`)
- Curated MCP/skill YAML lists (`curated/`)
- MCP/skill registry search
- `filesystem.read/edit`/`bash` permission schema (never implemented)
- Repo-local `.aienv.yaml` + lockfile (deferred)
- Agent expansion framework (Cursor, Copilot, etc.) — obviated by black-box design
- Custom MCP/skill repositories
