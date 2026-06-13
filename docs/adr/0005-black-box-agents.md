# Black Box Agent Architecture

Treat AI agents as opaque binaries that users supply and configure themselves. aienv provides the sandbox, dependency management, permissions, and audit — but never generates agent-config files, installs MCPs, or understands agent-specific protocols.

## Status

Accepted

## Context

The original aienv architecture had per-agent Go implementations: each agent (OpenCode, Claude Code) implemented an `Agent` interface with `GenerateFiles()` and `DockerConfig()` methods. Adding a new agent required a Go implementation that knew how to produce the agent's specific JSON config files and mount its specific state directories. This was a treadmill — every new agent (Cursor, Windsurf, Copilot, Codex) would need a new Go file, a new Dockerfile, and ongoing maintenance as agent config formats changed.

At the same time, the tool's real value proposition was the sandbox, audit, and security layer — not the agent-config generation, which users could easily do themselves.

## Decision

Delete the entire per-agent code path. Agents are now black boxes specified entirely through the env YAML:

```yaml
agent:
  install:
    - npm install -g @anthropic-ai/claude-code
  command: [claude]
  mounts:
    - source: ~/.claude.json
      target: /home/agent/.claude.json
```

No `internal/agents/` package. No per-agent Dockerfiles. No registry clients for MCP or skill search. The env YAML is the single source of truth for what the agent is and how it runs.

## Considered Options

- **Keep implementing per-agent adapters** — More agents would mean more Go code. Config-format changes (e.g., Claude Code switching from JSON to TOML) would require aienv updates. Rejected as unscalable.

- **Abstract agent config into a universal schema** — A hypothetical "agent config template" language that could produce opencode.json, CLAUDE.md, .cursor/mcp.json, etc. from one intermediate representation. Complex to design, fragile when agents add new config features. Rejected as over-engineering.

- **Black box** — The chosen path. Simpler, more maintainable, and keeps the core value (sandbox + audit) while eliminating the treadmill.

## Consequences

### Positive
- Zero ongoing cost to add new agents (they're just YAML)
- Users can run any agent — including custom or private agents
- No lock-in to agents we've implemented
- aienv is resilient to agent config-format changes
- Massive code deletion (~2000 lines)

### Negative
- Users must write agent config files themselves (MCPs, instructions, etc.)
- No trusted/curated MCP or skill discovery in aienv CLI
- No automated skill installation
- Slightly steeper onboarding (user needs to know their agent's install command)
