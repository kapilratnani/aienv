# aienv

> Reproducible, isolated environments for AI coding agents.
> Think `virtualenv` for AI agents — but with MCPs, skills, and rules.

AI coding setups become chaotic fast. Different projects need different MCP servers,
prompts, skills, model providers, API credentials, and tooling. Most developers
manage this with copied config files, global installs, and README instructions —
unreproducible, hard to share, and insecure.

`aienv` fixes this with project-scoped MCPs and skills, reproducible YAML configs,
multi-agent support (OpenCode, Claude Code), and disposable Docker sandboxes.

## Quick Start

```bash
go install github.com/kapilratnani/aienv@latest
aienv init
source ~/.zshrc

aienv create backend-api    # interactive: agent, MCPs, skills
aienv backend-api            # activate (local)
aienv --docker backend-api   # activate (Docker sandbox)
```

## Permissions (Experimental)

Network and filesystem permission enforcement works to some extent — the schema and configuration wizard are in place, but runtime enforcement works for opencode.

```yaml
permissions:
  filesystem:
    read:
      "*": "allow"
    edit:
      "*": "ask"
  bash:
    "*": "ask"
  network:
    allow: ["api.github.com"]
    deny: ["*"]
```

Existing features: `aienv permissions <name>` wizard, Docker network proxy (enforces allow/deny), OpenCode config translation for `filesystem.read`/`edit` and `bash` patterns.

Planned: Docker-level filesystem isolation, trust-system review prompt, Claude Code settings generation testing.

## Contributing

### MCPs
Add curated MCPs to `curated/mcps.yaml` following the existing schema. Include `env[]` metadata for any required environment variables.

### Skills
Add curated skills to `curated/skills.yaml` with a `description` that helps the create-flow search match user intent.

### New Agents
Agent support is pluggable via `internal/agents/agent.go`. Implement the `Agent` interface (`Name()`, `GenerateFiles()`, `ActivateCommand()`) and register via blank import in `agent_import.go`.

### General
PRs, issues, and ideas welcome. Open a discussion for larger changes before submitting.

## Roadmap

- [x] Create flow with curated & registry search
- [x] Docker sandbox isolation
- [x] Starter prompts
- [x] Claude Code support
- [x] Config inheritance & Docker auth
- [x] Docker write isolation (session-unique volumes)
- [x] Claude Code config inheritance
- [x] Default environment directory
- [ ] Repo-local `.aienv.yaml` + `aienv up`
- [ ] Permission policies & trust (test in progress on OpenCode)
- [ ] Agent expansion framework (Cursor, Copilot, etc.)
- [ ] Custom MCP/skill repositories
- [ ] Environment sharing & team features

---

Detailed docs: [architecture](docs/architecture.md), [completed features](docs/completed.md), [docker sandbox](docs/docker.md), [trust & permissions](docs/trust-permissions.md), [use cases](docs/use-cases.md), [roadmap](docs/roadmap.md)

MIT License
