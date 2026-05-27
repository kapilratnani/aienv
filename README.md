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

Before launching, aienv validates all referenced environment variables and warns if any are missing:

```
  Warning: MCP server "github" may not work — set GITHUB_TOKEN
```

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

- **Curated list**: 20 popular MCPs and 20 popular skills shown inline. Data comes from `curated/mcps.yaml` and `curated/skills.yaml` at the project root.
- **User overrides**: Drop YAML files in `~/.config/aienv/curated/*.yaml` to add or override MCPs/skills. Overridden entries are tagged `(user override)` in the menu.
- **Env var prompts**: When selecting a curated MCP that needs credentials, aienv prints which environment variables to set.
- **Online search**: Search the [Official MCP Registry](https://registry.modelcontextprotocol.io) or [skills.sh](https://skills.sh) directly from the prompts. Registry results are enriched with curated metadata when available.
- **Custom entry**: Type any MCP server or skill manually.

### Environment Variables (Activation-Time Validation)

When you activate an environment, aienv checks all `env:KEY` references in MCP configurations against your shell environment. Missing variables produce a warning:

```
  Warning: MCP server "brave-search" may not work — set BRAVE_API_KEY
  Warning: MCP server "datadog" may not work — set DATADOG_API_KEY, DATADOG_APP_KEY
```

This helps catch missing credentials before the agent starts.

## Curated MCP Servers

12 of the 20 curated MCPs require environment variables. During `aienv create`, selecting one shows the required variables.

| MCP | Required Env Vars |
|-----|------------------|
| GitHub | `GITHUB_TOKEN` |
| Sentry | `SENTRY_TOKEN`, `SENTRY_ORG` |
| Brave Search | `BRAVE_API_KEY` |
| Slack | `SLACK_TOKEN` |
| Stripe | `STRIPE_API_KEY` |
| Linear | `LINEAR_API_KEY` |
| Notion | `NOTION_TOKEN` |
| Figma | `FIGMA_ACCESS_TOKEN` |
| Supabase | `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY` |
| PagerDuty | `PAGERDUTY_API_KEY` |
| Datadog | `DATADOG_API_KEY`, `DATADOG_APP_KEY` |
| Gmail | `GMAIL_OAUTH_PATH` |
| Jira | `JIRA_HOST`, `JIRA_EMAIL`, `JIRA_API_TOKEN` |

## Registry MCP Search

When searching the Official MCP Registry, aienv now parses the `packages[]` response to determine the correct runtime command:

- **npm** packages → `npx -y <package>`
- **PyPI** packages → `uvx <package>`
- **Go** packages → `go run <package>`

Results are matched against the curated list to provide env var metadata when available.

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
| `aienv edit <name>` | Edit environment in `$EDITOR` (fallback `vi`) |
| `aienv delete <name>` | Delete environment with confirmation |
