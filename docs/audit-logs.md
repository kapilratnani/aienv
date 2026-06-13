# Audit Logs — Design & Plan

Last updated: 2026-06-14
Status: Phase 1 Complete

## Quick Resume

Audit Logs feature for `aienv`: record network requests per session as JSONL,
stored at `~/.local/share/aienv/<name>/audit/<session-id>/`.

## What's Implemented

1. **`internal/audit/`** package:
   - `schema.go` — `NetworkEntry`, `SessionMeta` structs
   - `writer.go` — `AppendNetworkEntry()`, `WriteSessionMeta()`, `ListNetworkEntries()`

2. **Proxy integration** (`internal/docker/proxy.go`):
   - Every HTTP request and HTTPS CONNECT logged with hostname, method, allowed/denied
   - Audit writer passed to proxy from `Run()` in sandbox.go

3. **Session lifecycle** (`internal/docker/sandbox.go`):
   - Unique session ID generated per activation (`config.GenerateSessionID()`)
   - Audit dir created at `<envDir>/audit/<session-id>/`
   - `session.meta.json` written with env name, agent command, start time
   - Audit dir mounted as `/aienv/audit:rw` inside container (for future cost extraction scripts)

4. **Learn mode**: proxy records all hostnames, prints suggested allowlist on exit when no `permissions.network.allow` configured.

## Phase 2: Cost Extraction (Pending)

- Post-session extraction from agent data stores via wrapper script
- OpenCode: `opencode.db` SQLite → token usage / cost
- Claude Code: `~/.claude.json` snapshot → API cost
- Wrapper script runs inside container after agent exits, writes to persistent audit mount
- Abnormal exits (crash, SIGKILL) lose cost data — acceptable for current scope

## Phase 3: Commands Audit (Pending)

- auditd inside container with `execve` rules
- `--cap-add AUDIT_CONTROL --cap-add AUDIT_WRITE`
- Parse `/var/log/audit/` → JSONL → aggregate

## Open Questions

1. Ephemeral volume → data loss: is the wrapper-on-exit approach sufficient for cost extraction?
2. Should we add `aienv audit <session-id>` command for aggregated reports?
