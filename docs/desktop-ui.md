# Desktop UI — Design & Plan

Last updated: 2026-06-07
Status: Design Phase

## Quick Resume

Desktop UI for `aienv` built with [Wails](https://wails.io) (Go backend, Vue 3 frontend). Full replacement of the CLI for common tasks: env CRUD, embedded terminal activation, permissions visual editor, and audit viewer. Separate binary at `cmd/wails/` reusing `internal/` packages.

## Architecture

```
cmd/wails/
├── main.go                  # wails.Run() with App{} bindings
├── app.go                   # App struct: all runtime bindings
├── terminal.go              # PTY manager (creack/pty) — shared Terminal interface
├── docker.go                # GUI-aware docker.Run, build monitor
├── wails.json               # Wails project config
└── frontend/                # Vue 3 + TypeScript + Vite
    ├── src/
    │   ├── main.ts
    │   ├── App.vue                       # Layout shell
    │   ├── router/index.ts               # 7 routes
    │   ├── stores/envs.ts                # Pinia: env list, current env, CRUD
    │   ├── stores/terminal.ts            # Pinia: session state, output buffer
    │   ├── stores/docker.ts              # Pinia: daemon status, build state
    │   ├── composables/useTerminal.ts    # xterm.js ↔ Go event bridge
    │   ├── composables/useCurated.ts     # Load curated MCPs/skills
    │   ├── views/
    │   │   ├── DashboardView.vue         # Env grid + docker status + quick launch
    │   │   ├── EnvCreateView.vue         # 6-step visual wizard
    │   │   ├── EnvDetailView.vue         # Tabs: MCPs | Skills | Rules | Permissions
    │   │   ├── EnvEditView.vue           # Form-based edit (reuses wizard step components)
    │   │   ├── PermissionsView.vue       # Visual glob → action editor
    │   │   ├── TerminalView.vue          # Full-window xterm.js
    │   │   └── AuditView.vue             # Session list + detail panel
    │   ├── components/
    │   │   ├── layout/AppSidebar.vue     # Nav + env quick-list + docker indicator
    │   │   ├── env/EnvCard.vue           # Summary card: name, agent icon, MCP count
    │   │   ├── env/EnvStatusBadge.vue    # Trust status, ready indicator
    │   │   ├── curated/MCPPicker.vue     # Search + grid of curated MCPs + custom
    │   │   ├── curated/SkillPicker.vue   # Same for skills
    │   │   ├── curated/RegistrySearch.vue# Online registry lookup modal
    │   │   ├── permissions/PermissionRow.vue  # Glob + action dropdown row
    │   │   ├── permissions/NetworkDomain.vue  # Domain chip input
    │   │   ├── terminal/TerminalPanel.vue     # xterm.js wrapper
    │   │   └── audit/SessionListItem.vue      # Session timestamp + summary
    │   └── assets/styles/main.css        # Tailwind
    ├── index.html
    ├── package.json
    └── vite.config.ts
```

## Tech Stack

### Backend (Go)
- **Wails v2** — Go native desktop framework, reusable bindings
- **creack/pty** — PTY allocation for embedded terminal
- **Existing `internal/` packages** — env, docker, agents, permissions, audit, registry, skills, config, assets

### Frontend (Vue)
- **Vue 3** + Composition API + TypeScript
- **Vite** — dev server + build, hot-reload via `wails dev`
- **Vue Router** — 7 routes (Dashboard, Create, Detail, Edit, Permissions, Terminal, Audit)
- **Pinia** — state management: envs, terminal, docker stores
- **Tailwind CSS** — utility-first styling, dark mode via `class` toggle
- **xterm.js** — embedded terminal emulator

## Activation Model

All activation uses a `Terminal` interface backed by Docker.

### Terminal Interface

```go
package terminal

type Terminal interface {
    Stdin() io.WriteCloser
    Stdout() io.ReadCloser
    Resize(cols, rows uint16) error
    Close() error
    Done() <-chan struct{}
    ExitCode() int
}
```

Two implementations:

| Implementation | Used By | Backend |
|---|---|---|
| `OSTerminal` | CLI (unchanged) | Wraps `os.Stdin`/`os.Stdout` |
| `PTYTerminal` | GUI (new) | Pipes docker stdio to xterm.js via Wails events |

### Terminal Handling

- `docker run -i` (no `-t`), pipe stdio via `io.Pipe()`. `Resize()` is a no-op. No host-side PTY — Docker handles terminal internally.

### GUI Activation Flow

```
User clicks "Activate" on env card
        │
        ▼
1. Pre-checks (same as CLI): trust, skills verify, MCP env vars
2. If Docker image missing → auto-build (indeterminate spinner, non-blocking)
3. Navigate to TerminalView
4. Backend:
   a. Docker: allocate io.Pipe(), set cmd.Dir = workdir, run docker -i
5. Stream: Stdout → events.Emit("terminal:output", chunk)
   Frontend: xterm.js writes chunk
6. Input: xterm.js onKey → SendTerminalInput(data) → Stdin
7. Resize: xterm.js onResize → ResizeTerminal(cols, rows) → PTY.Resize (no-op for Docker)
8. Exit: Done() → events.Emit("terminal:exited", code)
   Frontend: show exit dialog → return to dashboard
```

### Close While Session Active

Confirm dialog → SIGTERM → 5s wait → SIGKILL.

### Ctrl+C Signal Chain

Default PTY behavior. Docker forwards SIGINT naturally. No special "Stop Session" button needed.

## Core Refactoring: `internal/terminal/`

New shared package. Extracts terminal abstraction from `internal/docker/sandbox.go`.

### Changes to Existing Packages

| Package | Change |
|---|---|
| **NEW** `internal/terminal/` | `Terminal` interface + `OSTerminal` + `PTYTerminal` |
| `internal/docker/sandbox.go` | `Run()` refactored to accept `terminal.Terminal` |
| `internal/agents/agent.go` | `DockerConfig()` used to construct docker run arguments |

### What Stays the Same

All other `internal/` packages — env, config, registry, skills, assets, agents/opencode, agents/claude, permissions, providers, shell — imported as-is with zero changes.

## Module Structure

- Separate `go.mod` at `cmd/wails/go.mod` with `replace` directive pointing to parent module.
- Wails dependency tree isolated from CLI binary.
- Separate binaries for CLI and desktop UI.
- Node/npm required for frontend development (documented in README).

## Secure Storage

- **OS-native keychain** via `zalando/go-keyring` (macOS Keychain, Linux Secret Service).
- **Fallback**: encrypted file at `~/.config/aienv/secrets.json.aes` (AES-256-GCM, Argon2id-derived key) if D-Bus unavailable.
- **Global flat keys**: `GITHUB_TOKEN` not `myenv.GITHUB_TOKEN`. One canonical value per key.
- **Docker mode**: encrypted file `:ro` mounted into container at the path the env var expects, or `-e KEY=value` injection.

## Create Flow — Visual Wizard

Replaces the 741-line CLI wizard with a 6-step stepper form (true wizard, back/next, one step at a time):

| Step | Component | Description |
|---|---|---|
| 1. Basics | `BasicsForm` | Name (validated), agent dropdown (opencode/claude-code), description, model |
| 2. MCPs | `MCPPicker` + `RegistrySearch` | Curated grid with toggle, online search modal, custom entry form |
| — Env Vars | `EnvVarForm` | Sub-step after MCP selection: fill in required env vars, store to keychain immediately |
| 3. Skills | `SkillPicker` + `RegistrySearch` | Same pattern for skills |
| 4. Rules | `RuleForm` | File path list + native OS file picker dialog |
| 5. Advanced | `AdvancedForm` | Default prompt textarea, workdir with `~` expansion |
| 6. Review | `ReviewSummary` | Full env summary with edit-any-step + Save button |

Backend: `CreateEnv(env)` → `env.Save()` + `ag.GenerateFiles()` + `skills.Install()` — same logic as `create_cmd.go`.

## Env Detail View

Overview header section (name, agent, model, description, workdir, trust badge) + 5 actionable tabs:

| Tab | Description |
|---|---|
| **MCPs** | List of MCP servers with toggle switches, inline add/remove |
| **Skills** | List of skills with badges, inline add/remove |
| **Rules** | List of rule file paths, inline add/remove |
| **Permissions** | Visual permissions editor (read/edit/bash/network) |

**Edit wizard**: "Edit" button opens the same 6-step wizard pre-filled. For quick changes, each tab has inline add/remove controls. No YAML raw editor.

## Permissions View

Visual editor on env detail page (not global nav):

- **Filesystem Read**: rows of [glob pattern] × [allow/ask/deny dropdown] + add/remove
- **Filesystem Edit**: same pattern
- **Bash**: same pattern for command globs
- **Network**: chip input (tag-like) for allow domains + deny domains
- **Provider Endpoints**: auto-detect on view mount (spinner), table with checkboxes to include
- **Trust Status**: badge + one-click trust/reject

## Sidebar Navigation

```
┌─────────────────────┐
│  ▲ aienv            │  ← app logo/name
│                     │
│  ☰ Environments    │  ← Dashboard (env grid)
│  ＋ New            │  ← Create wizard shortcut
│                     │
│  📋 Audit           │  ← Top-level audit viewer
│                     │
│  🐳 Docker          │  ← Docker info
│  ● Running          │  ← green/red dot (daemon status)
│                     │
│  ❓ Help            │  ← docs link / about
└─────────────────────┘
```

## Audit Viewer

- **Docker-only** for v1. Non-Docker audit deferred to v2.
- **Top-level nav item** — not an env detail tab.
- **Layout**: left panel (session list), right panel (session detail).
- Network tab: unique hostnames table (hostname, request count).
- Cost tab: total tokens, cost estimate (extracted via wrapper script post-exit).

## Docker Build UX

- **Image status badge**: "Not built" / "Built" on env card and env detail page.
- **Build button**: on env detail page, triggers `docker build`. Auto-trigger if user activates and image is missing.
- **Build progress**: collapsible side panel with real-time `docker build` log streamed via Wails events. Non-blocking — user can navigate away.
- **Progress**: indeterminate spinner for v1. Future: streaming log output per layer.
- **Custom images**: future feature — users can create Dockerfiles inheriting from base images.

## Docker Daemon Status

- Green/red dot in sidebar next to Docker icon.
- Checked on app start via `docker info`.
- When daemon is down: Docker "Activate" buttons disabled with tooltip. CRUD operations unaffected.

## First Launch / Empty State

- Welcome hero: "Get started by creating your first environment" → big "Create Environment" button.
- Passive Docker status indicator in sidebar (no forced setup).
- Empty env grid with subtle placeholder.

## Trust Prompt

- Full-screen modal with 3 sections: permissions summary, MCP list, skill list.
- Action buttons: Trust / Review (opens PermissionsView) / Reject.

## Error Handling

- **Toasts** (top-right, auto-dismiss 4s) for non-blocking errors (save failed, docker check failed).
- **Inline validation** for form fields.
- **Modals** for critical errors (activate failed, image missing on build).

## Dark Mode

Full dark mode from day one. Dark terminal theme + dark app UI. Tailwind `darkMode: 'class'` toggle.

## Cross-Platform

- **v1**: Linux + macOS only.
- **Windows**: deferred to v2 (ConPTY instead of creack/pty).

## Build & Dev

- **Dev**: `wails dev` — Vite dev server + hot-reload + Go backend with live-reload.
- **Build**: `wails build` — compiles Vue → `frontend/dist/`, embeds via `//go:embed`.
- **Single binary** — no external files at runtime.

## Implementation Phases

### Phase 1: Shell & Env CRUD
1. `wails init` scaffold, minimal App binding
2. Env CRUD Go bindings (ListEnvs, GetEnv, CreateEnv, UpdateEnv, DeleteEnv, GetEnvYAML)
3. Dashboard view (env grid + docker status)
4. Create wizard (6-step form)
5. Detail view (tabs: MCPs, Skills, Rules)
6. Edit view (form-based, reuses step components)

### Phase 2: Permissions & Trust
1. Secure storage: OS keychain bindings (GetSecret, SetSecret, DeleteSecret)
2. Permissions Go bindings (GetPermissions, UpdatePermissions, DetectProviderEndpoints)
3. PermissionsView (glob × action rows, network chips)
4. Trust Go bindings (CheckTrust, SetTrust)
5. Trust badge + trust prompt modal on first activation

### Phase 3: Embedded Terminal
1. `Terminal` interface + `OSTerminal` extraction
2. `PTYTerminal` with `io.Pipe()` variant for Docker stdio
3. Refactor `docker.Run()` to accept `terminal.Terminal`
4. xterm.js integration (TerminalPanel, useTerminal composable)
5. Activation flow in GUI (Docker-only)
6. Session lifecycle (start, stream, resize, exit)

### Phase 4: Audit Viewer
1. Extend proxy.go with session-aware JSONL logging
2. Audit Go bindings (ListAuditSessions, GetAuditReport)
3. Session list + detail views
4. Network tab + cost tab rendering

## Grill Session Decisions (2026-06-07)

| # | Question | Decision |
|---|---|---|
| Q1 | PTY for Docker mode? | No — Docker uses `io.Pipe()` on stdio. |
| Q2 | Shell function bypass? | Yes — GUI constructs activation commands internally, sets `cmd.Dir = workdir`. |
| Q3 | Windows? | Deferred to v2 (Linux + macOS only). |
| Q4 | Docker build UX? | Build indicator badge + manual "Build" button on env detail. Auto-trigger on activate. Spinner, no detailed progress v1. |
| Q5 | Close while session active? | Confirm dialog → SIGTERM → wait → SIGKILL. |
| Q6 | Provider endpoint detection? | Auto-detect on view mount with loading spinner. |
| Q7 | Non-Docker audit? | Disabled — Docker-only for v1. |
| Q8 | Dark mode? | Full dark mode from day one. |
| Q9 | Trust prompt UX? | Full-screen modal with 3 sections + Trust/Review/Reject. |
| Q10 | Repo-local envs? | Registered only (`~/.ai-envs/`) for v1. |
| Q11 | Proxy logging? | Direct JSONL writes from Proxy struct. |
| Q12 | Env editing UX? | Detail page inline controls for MCPs/Skills/Rules + wizard for structural edits. |
| Q13 | MCP env vars at create? | Post-MCP-picker sub-step with input fields, stored to OS keychain immediately. |
| Q14 | Secure storage scope? | Global flat keys (`GITHUB_TOKEN` not `myenv.GITHUB_TOKEN`). |
| Q15 | First launch/empty state? | Welcome hero → big "Create" button. Passive Docker status indicator. |
| Q16 | Docker build progress? | Collapsible side panel with real-time build logs, non-blocking. |
| Q17 | Module structure? | Separate `go.mod` at `cmd/wails/` with `replace` directive. |
| Q18 | Node/npm? | Accept as tooling dependency, document in README. Node/npm required for frontend dev. |
| Q19 | Docker daemon status? | Green/red dot in sidebar, disable Docker buttons when daemon down. |
| Q20 | Nav structure? | Permissions on env detail page (not global nav). Audit = top-level nav item. |
| Q21 | Env detail tabs? | Overview (header), MCPs, Skills, Rules, Permissions. No YAML tab. |
| Q22 | When to prompt env vars? | Sub-step after MCP selection in create wizard. Store to keychain immediately. |
| Q23 | Wizard layout? | True stepper (back/next), one step at a time. |
| Q24 | Error handling? | Toast for non-blocking, inline for form validation, modal for critical. |
| Q25 | Ctrl+C signal chain? | Default PTY behavior. Docker forwards SIGINT naturally. No special "Stop" button needed. |
| Q26 | Secure storage lib? | `zalando/go-keyring` primary, encrypted file fallback. |
