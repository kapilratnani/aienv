# aienv — AI Environment Manager

Task-specific MCP servers, agent skills, and rules for AI coding agents — like Python's `virtualenv` for AI.

## Install

```bash
go install github.com/kapilratnani/aienv@latest
```

Then install the shell function:

```bash
aienv init
source ~/.zshrc  # or restart your shell
```

## Quick Start

```bash
# Create an environment
aienv create backend-api

# Activate it
aienv backend-api
# → opens opencode with MCPs, skills, and rules loaded
# → unset OPENCODE_CONFIG on exit

# List all environments
aienv list

# Show environment details
aienv show backend-api

# Edit environment config
aienv edit backend-api

# Delete an environment
aienv delete backend-api
```

The `aienv` shell function handles activation transparently — just run `aienv <name>` to start a session.

## How It Works

Each environment is stored in `~/.ai-envs/<name>/`:

```
~/.ai-envs/
├── backend-api/
│   ├── ai-env.yaml          # your environment definition
│   └── opencode.json        # generated OpenCode config
├── frontend-design/
│   └── ai-env.yaml
└── incident-response/
    └── ai-env.yaml
```

Activation sets `OPENCODE_CONFIG` to point at the generated `opencode.json`, spawns `opencode`, and unsets the variable on exit.

## Environment Format

```yaml
name: backend-api
agent: opencode
model: claude-sonnet-4-5
description: Backend API development environment

mcp:
  github:
    type: local
    command: ["npx", "-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "env:GITHUB_TOKEN"
  postgres:
    type: remote
    url: "https://mcp.example.com/postgres"

skills:
  - name: api-design
    source: registry
    package: vercel-labs/agent-skills

rules:
  - path: ./docs/backend-standards.md
  - path: ./AGENTS.md
```

## Create Flow

- **Curated list**: 20 popular MCPs and 20 popular skills shown inline
- **Online search**: Search the [Official MCP Registry](https://registry.modelcontextprotocol.io) or [skills.sh](https://skills.sh) directly from the prompts
- **Custom entry**: Type any MCP server or skill manually

## Supported Agents

Currently targets **OpenCode** via `OPENCODE_CONFIG` env var injection. Other agents (Claude Code, Cursor) planned.

## Commands

| Command | Description |
|---------|-------------|
| `aienv create <name>` | Interactive environment creation |
| `aienv <name>` | Activate environment (requires shell function) |
| `aienv list` | List all environments |
| `aienv show <name>` | Show environment details |
| `aienv init` | Install shell function to `.bashrc`/`.zshrc` |
| `aienv edit <name>` | Edit environment in `$EDITOR` |
| `aienv delete <name>` | Delete environment with confirmation |
