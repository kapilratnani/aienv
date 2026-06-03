# Priority 3: Permission Policies & Trust — Design Doc

Last updated: 2026-06-03
Status: Implemented

## Quick Resume

To restart this conversation from here, paste this doc into the context and say "implement Priority 3".

## Schema

```yaml
permissions:
  filesystem:
    read:
      "*": "allow"
      ".env": "deny"
    edit:
      "*": "ask"
      "src/": "allow"
  bash:
    "*": "ask"
    "git *": "allow"
    "rm -rf *": "deny"
  network:
    allow: ["api.github.com"]
    deny: ["*"]
```

- Use OpenCode's `allow` / `ask` / `deny` model for `filesystem` and `bash` → maps 1:1 to OpenCode's `permission.read`, `permission.edit`, `permission.bash`
- `network` is domain allowlist + deny list, Docker-only enforcement
- Empty/nil permissions → pass through user's global agent config unchanged

## Agents

### OpenCode
- `mergePermission()` extended to emit `filesystem.read`, `filesystem.edit`, `bash` patterns
- Read global `permission` from `~/.config/opencode/opencode.json`, deep-merge env overrides, emit full merged object
- `network` not translated to agent config

### Claude Code
- Translate to `.claude/settings.json` format:
  - `filesystem.read` → `Read(pattern)` in `permissions.allow/ask/deny`
  - `filesystem.edit` → `Edit(pattern)`
  - `bash` → `Bash(pattern)`
- Write `<envDir>/claude-settings.json`
- Pass via `--settings <path>` in `ActivateCommand()`
- No Claude native sandbox settings in this iteration
- `network` not translated to agent config

## Subcommand: `aienv permissions <name>`

Interactive wizard on existing env:

1. Prompt for `filesystem.read` patterns (glob + action), repeat until "done"
2. Prompt for `filesystem.edit` patterns
3. Prompt for `bash` command patterns
4. Prompt for `network` allow/deny domains
5. Auto-detect provider API endpoints:
   - OpenCode: runs `opencode providers list`, cross-references display names against
     `~/.cache/opencode/models.json` for API base URLs; falls back to hardcoded SDK
     defaults for native SDK providers (Groq, Google, etc.); also reads custom
     provider `baseURL` from `~/.config/opencode/opencode.json`
   - Claude Code: always allows `api.anthropic.com`
   - Both: checks `OPENAI_BASE_URL`, `ANTHROPIC_BASE_URL`, `AZURE_OPENAI_ENDPOINT` env vars
   - Loopback hosts (localhost, 127.0.0.1, etc.) filtered out — handled by `NO_PROXY`
6. Write `permissions:` to `ai-env.yaml`
7. Regenerate agent config
8. Invalidate trust cache

## Docker Enforcement

- **filesystem**: not enforced at Docker level — agent config handles path-level allow/ask/deny
- **network**: embedded Go HTTP/HTTPS proxy, injected as `HTTP_PROXY`/`HTTPS_PROXY`
  - Proxy binds to `0.0.0.0` (all host interfaces) on a random port
  - Container resolves the host via `--add-host host.docker.internal:host-gateway` (`host.docker.internal`)
  - Container uses default Docker bridge network (no `--network host`)
  - Domain allowlist/denylist enforced by the proxy
  - Provider API base URLs auto-allowlisted via `DetectProviderEndpoints()`:
     - OpenCode: parses `opencode providers list` output, maps names via
       `~/.cache/opencode/models.json`, falls back to SDK defaults for native providers
     - Claude Code: always allows `api.anthropic.com`
     - Safety net: also detected at proxy setup in `internal/docker/sandbox.go` for
       envs configured before this feature
  - HTTP/HTTPS only — SSH remotes, raw TCP not covered (documented gap)
  - No `network` = default Docker bridge network

## Trust System

- **Cache path**: `~/.config/aienv/trust/<sha256-of-ai-env-yaml>.json`
- **Cache shape**:
  ```json
  {
    "status": "trusted",
    "reviewed_at": "2026-06-02T10:00:00Z",
    "permissions_snapshot": { "...": "..." },
    "env_name": "myenv"
  }
  ```
- **Triggers**: `aienv activate <name>` and `aienv up` — compare current YAML hash vs cache
  - No cache → trust prompt
  - Hash mismatch → "Permissions changed" prompt with snapshot diff
  - Cache hit → skip
- **Trust prompt**: permissions summary, MCP servers, skills. Options: Trust, Review (`$EDITOR`), Reject
- **Invalidation**: any YAML modification (via `aienv permissions`, `aienv edit`, manual edit)

## Implementation Order

1. `Permissions` struct + YAML marshal/unmarshal in `internal/env/env.go`
2. `aienv permissions` subcommand wizard
3. OpenCode `mergePermission()` extension in `internal/agents/opencode/gen.go`
4. Claude Code settings generation + `--settings` injection in `internal/agents/claude/gen.go`
5. Trust cache read/write + hash computation in `internal/config/paths.go`
6. Trust/review prompt in activation flow (both `activate` and `up`)
7. Embedded Go proxy for Docker network enforcement in `internal/docker/`
8. Regeneration hooks on activation
