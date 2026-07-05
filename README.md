# AIenv

> Bring your own AI agent. AIenv provides the secure runtime.

AI coding setups become chaotic fast. Different projects need different tools,
credentials, and configs. `aienv` fixes this with reproducible Docker sandboxes
that isolate agents from your host, enforce network permissions, and provide
audit trails. Agents are **black boxes** — you bring your own agent config,
aienv provides the sandbox.

## Quick Start

```bash
go install github.com/kapilratnani/aienv@latest

aienv create my-env     # interactive: install, command, mounts, permissions
aienv my-env             # build image + launch sandbox
```

## CLI

| Command | Function |
|---------|----------|
| `aienv create <name>` | Interactive environment creation |
| `aienv up <name>` / `aienv <name>` | Build image if needed, launch sandbox |
| `aienv up <name> -p "prompt"` | Launch sandbox and send prompt to agent |
| `aienv up <name> -p "prompt" -x` | One-shot: agent runs prompt and exits |
| `aienv up <name> -w <branch>` | Create git worktree and activate into it |
| `aienv list` | List all environments |
| `aienv show <name>` | Show environment details |
| `aienv edit <name>` | Edit env YAML in `$EDITOR` |
| `aienv build <name>` | Force rebuild Docker image |
| `aienv shell <name>` | Interactive shell in sandbox (debugging) |
| `aienv delete <name>` | Remove env + image + audit data |
| `aienv clean` | Remove orphaned images, audit dirs, trust cache |

## Schema

```yaml
env:
  name: my-env
  description: Backend API development

agent:
  install:
    - npm install -g opencode-ai
  command: [opencode]
  args: [--model, claude-sonnet-4-5]  # default arguments
  prompt_flag: "-p"                    # flag for -p "prompt" mode
  exit_subcommand: "run"               # subcommand for -x one-shot mode
  env:                                 # environment variables
    MY_KEY: "env:HOST_KEY_NAME"        # "env:NAME" = passthrough from host
    OPENCODE_CONFIG: /home/agent/.config/opencode.json
  mounts:
    - source: ~/project
      target: /workspace
    - source: ~/.config/opencode
      target: /home/agent/.config/opencode
      writable: true

deps:
  packages: [golang, python3, nodejs]
  custom: [go install foo/bar@latest]

permissions:
  network:
    allow: [api.github.com, api.anthropic.com]
    deny: ["*"]

audit:
  persist: true
  capture: [network]
```

## Recipes

Ready-to-use environment configurations for popular AI coding agents.

### OpenCode

Sandbox [OpenCode](https://opencode.ai) — the agent you're using now.

```yaml
env:
  name: opencode-dev
  description: Develop aienv with OpenCode

agent:
  install:
    - npm install -g opencode-ai
  command: [opencode]
  prompt_flag: "-p"
  args: [--model, opencode/deepseek-v4-flash-free]
  mounts:
    - source: /home/you/projects/aienv
      target: /workspace
      writable: true
    - source: ~/.config/opencode
      target: ~/.config/opencode
      writable: true
    - source: ~/.local/share/opencode
      target: ~/.local/share/opencode
      writable: true
    - source: ~/.cache/opencode
      target: ~/.cache/opencode
      writable: true
    - source: ~/.agents/skills
      target: ~/.agents/skills

deps:
  packages: [golang-go]

audit:
  persist: true
  capture: [network]
```

### Claude Code

Sandbox [Claude Code](https://docs.anthropic.com/en/docs/claude-code) by Anthropic.

```yaml
env:
  name: claude-dev
  description: Sandboxed Claude Code for project work

agent:
  install:
    - npm install -g @anthropic-ai/claude-code
  command: [claude]
  env:
    ANTHROPIC_API_KEY: "env:ANTHROPIC_API_KEY"
  mounts:
    - source: /home/you/projects/my-app
      target: /workspace
      writable: true
    - source: ~/.claude
      target: ~/.claude
      writable: true

permissions:
  network:
    allow:
      - api.anthropic.com
      - raw.githubusercontent.com

audit:
  persist: true
  capture: [network]
```

Activate with `aienv up claude-dev` or send a prompt directly:

```bash
aienv up claude-dev -p "Refactor the auth module to use JWT"
```

### Codex CLI

Sandbox [Codex CLI](https://github.com/openai/codex) by OpenAI.

```yaml
env:
  name: codex-dev
  description: Sandboxed Codex CLI

agent:
  install:
    - npm install -g @openai/codex
  command: [codex]
  args: [--sandbox, workspace-write]
  env:
    OPENAI_API_KEY: "env:OPENAI_API_KEY"
  mounts:
    - source: /home/you/projects/my-app
      target: /workspace
      writable: true
    - source: ~/.codex
      target: ~/.codex
      writable: true

permissions:
  network:
    allow:
      - api.openai.com
      - raw.githubusercontent.com

audit:
  persist: true
  capture: [network]
```

```bash
aienv up codex-dev                              # interactive TUI
aienv up codex-dev -p "Add input validation"    # send prompt
```

### Pi

Sandbox [Pi](https://pi.dev) — the terminal AI coding agent.

```yaml
env:
  name: pi-dev
  description: Sandboxed Pi coding agent

agent:
  install:
    - npm install -g @earendil-works/pi-coding-agent
  command: [pi]
  env:
    ANTHROPIC_API_KEY: "env:ANTHROPIC_API_KEY"   # or OPENAI_API_KEY, etc.
  mounts:
    - source: /home/you/projects/my-app
      target: /workspace
      writable: true
    - source: ~/.pi
      target: ~/.pi
      writable: true

deps:
  packages: [nodejs, git, curl, ripgrep]

audit:
  persist: true
  capture: [network]
```

```bash
aienv up pi-dev
```

### Bring your own

Any CLI-based coding agent works — just change `agent.install` and
`agent.command`. The black box architecture means zero Go code changes.

## Features

- **Black box agents** — any agent binary, defined entirely in YAML. No per-agent Go code.
- **Build-time images** — auto-generated from env YAML, cached by content hash. Auto-rebuilt on config change.
- **Network proxy** — HTTP/HTTPS proxy with allow/deny/learn modes, runs on host. Learn mode suggests an allowlist on exit.
- **Audit logging** — JSONL session records, network requests logged per activation. Stored at `~/.local/share/aienv/<name>/audit/<session-id>/`.
- **Trust system** — first activation shows mounts + network rules and asks for confirmation, cached by env hash.
- **Session isolation** — each `aienv up` spawns a separate container with a unique session ID. Concurrent activations are independent.
- **Git worktree support** — `aienv up -w <branch>` creates a git worktree, mounts it into the sandbox, and cleans up on exit.
- **Debug shell** — `aienv shell <name>` drops you into `/bin/bash` inside the sandbox without the agent or proxy.
- **XDG-compliant** — config at `~/.local/share/aienv/`, trust cache at `~/.config/aienv/trust/`.

## How it works

```
┌──────────────────────────────────────────────────────┐
│  aienv up my-env                                     │
│                                                      │
│  1. Load ~/.local/share/aienv/my-env/env.yaml        │
│  2. Compute SHA-256 hash → check image cache         │
│  3. Build Docker image if missing (auto-generates     │
│     Dockerfile from env YAML)                         │
│  4. Start network proxy on random host port           │
│  5. Spawn container with:                             │
│     • Project + config mounts (ro by default)          │
│     • HTTP_PROXY pointing at host proxy                │
│     • Audit dir mounted at /aienv/audit                │
│     • Environment variables                            │
│     • Agent command as entrypoint                      │
│  6. Agent runs inside container — all network          │
│     traffic passes through proxy enforcement           │
│  7. On exit: container auto-removed, proxy             │
│     stopped, audit logs persisted                     │
└──────────────────────────────────────────────────────┘
```

## Contributing

PRs, issues, and ideas welcome. Open a discussion for larger changes first.

See [architecture](docs/architecture.md), [roadmap](docs/roadmap.md), and [docs/](docs/) for details.

## License

MIT
