# Permissions & Trust

Last updated: 2026-06-14
Status: Implemented

## Schema

The current permissions schema is network-only:

```yaml
permissions:
  network:
    allow: [api.github.com, api.anthropic.com]
    deny: ["*"]
    learn: false  # default: true when no allow/deny rules
```

- `allow` — domain patterns permitted (e.g. `api.github.com`, `*.example.com`, `*`)
- `deny` — domain patterns blocked (checked only when no allowlist is set)
- `learn` — when no allow/deny rules, proxy records all hosts and suggests an allowlist on exit
- Empty/nil permissions → no proxy (default Docker bridge network)
- HTTP/HTTPS only — SSH remotes, raw TCP not covered (documented gap)

## Network Proxy (Docker Enforcement)

Embedded Go HTTP/HTTPS proxy, injected as `HTTP_PROXY`/`HTTPS_PROXY`:

1. Proxy binds to `0.0.0.0` (all host interfaces) on a random port
2. Container resolves the host via `--add-host host.docker.internal:host-gateway` (`host.docker.internal`)
3. Container receives `HTTP_PROXY`/`HTTPS_PROXY` pointing to `host.docker.internal:<port>`
4. Container uses default Docker bridge network (no `--network host`)
5. Domain allowlist/denylist enforced by the proxy
6. Proxy also started when `audit.persist: true` (no allow/deny, just logging)
7. Learn mode: no allow/deny rules → pass-through + hostname recording
8. Proxy prints learned hosts to stderr on `Close()` with suggested `permissions.network.allow` entries

### Changes from earlier design

Previous design docs referenced `filesystem.read/edit` and `bash` permission
schemas. These were never implemented. The current architecture treats agents as
black boxes — aienv only enforces network-level permissions via the Docker proxy.
Filesystem and command permissions are the agent's responsibility.

## Trust System

- **Cache path**: `~/.config/aienv/trust/<sha256-of-env-yaml>.json`
- **Cache shape**:
  ```json
  {
    "env_name": "my-env",
    "status": "trusted",
    "reviewed_at": "2026-06-14T10:00:00Z"
  }
  ```
- **Triggers**: `aienv up` / `aienv <name>` — compare current YAML hash vs cache
  - No cache → trust prompt
  - Hash mismatch → re-prompt (env config changed)
  - Cache hit → skip
- **Trust prompt**: shows mounts + network rules. Options: Trust (y) or Reject (N)
- **Invalidation**: any YAML modification changes the hash, triggering re-prompt on next activation
