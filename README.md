# aienv

> The permission and isolation layer for AI coding agents.

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
| `aienv list` | List all environments |
| `aienv show <name>` | Show environment details |
| `aienv edit <name>` | Edit env YAML in $EDITOR |
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
  prompt_flag: "-p"          # optional: flag to use when sending prompt via aienv up -p "..."
  mounts:
    - source: ~/project
      target: /workspace
    - source: ~/.config/opencode
      target: /home/agent/.config/opencode
      writable: true

deps:
  packages: [golang, python3]
  custom: [go install foo/bar]

permissions:
  network:
    allow: [api.github.com, api.anthropic.com]
    deny: ["*"]

audit:
  persist: true
  capture: [network]
```

## Features

- **Black box agents** — any agent binary, defined entirely in YAML. No per-agent Go code.
- **Build-time images** — auto-generated from env YAML, cached by content hash.
- **Network proxy** — HTTP/HTTPS proxy with allow/deny/learn modes, runs on host.
- **Audit logging** — JSONL session records, network requests logged per activation.
- **Trust system** — first activation shows mounts + network rules, cached by env hash.
- **Session isolation** — each `aienv up` spawns a separate container with unique session ID.
- **XDG-compliant** — config at `~/.local/share/aienv/`, trust cache at `~/.config/aienv/trust/`.

## Contributing

PRs, issues, and ideas welcome. Open a discussion for larger changes first.

See [architecture](docs/architecture.md), [roadmap](docs/roadmap.md), and [docs/](docs/) for details.

## License

MIT
