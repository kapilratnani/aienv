# Audit Logs — Design & Plan

Last updated: 2026-06-06
Status: Design Phase

## Quick Resume

Audit Logs feature for `aienv`: record network requests and LLM cost per session, queryable via `aienv audit <session-id>`.

## Grill Session (2026-06-06)

### Scope Decisions

| Dimension | Priority | Approach |
|---|---|---|
| **Network** | P0 | Log hostnames only (no paths, no MITM). Proxy in `proxy.go` writes one JSONL line per unique hostname. |
| **Cost** | P0 | Post-session extraction from agent data stores via wrapper script. OpenCode: `opencode.db` SQLite. Claude Code: `~/.claude.json` snapshot. |
| **Commands** | P1 | Deferred to v2. |
| **Filesystem** | Dropped | Git covers "what changed" retroactively. |

### Key Decisions

1. **Cost extraction via wrapper script** — runs inside container after agent exits, reads ephemeral volume, writes to persistent audit mount. Abnormal exits (crash, SIGKILL) lose cost data — acceptable for v1.
2. **Scripts mounted at runtime** (`:ro`) — not baked into Dockerfile. Supports custom images without rebuild.
3. **Separate JSONL files** per session: `session-<id>.network.jsonl` (streamed), `session-<id>.cost.jsonl` (single dump).
4. **Session metadata** written host-side by `activate_cmd.go` to `~/.ai-envs/<name>/audit/<id>.meta.json` — single source of truth for session ID propagation.
5. **Hostnames only** for network audit — clean, privacy-friendly, no MITM complexity.

### Open Issue: Ephemeral Volume Data Loss

**Problem**: `~/.local/share/opencode/` and `~/.claude/` are mounted as ephemeral Docker volumes, deleted on container exit (`defer docker volume rm -f`). We lose agent session transcripts, full cost data, token usage per turn.

**Proposed wrapper approach**: Script runs inside container before volume is destroyed → extracts cost → writes to persistent `:rw` audit mount. Works for normal exits but not crashes.

**User concern**: Agents have built-in auditing (session transcripts, tool calls) that provides richer data than cost alone. Are we discarding too much?

**Resolution**: Pending — user wants to brainstorm alternative approaches to preserve agent-native audit data without sacrificing write isolation.

## Implementation Plan

### Phase 1: Core Infrastructure (V1)

1. **`internal/audit/`** — new package
   - `schema.go` — `NetworkEntry`, `CostEntry`, `SessionMeta` structs
   - `writer.go` — JSONL write helpers
   - `reader.go` — LoadSessionMeta, ListSessions, AggregateSession
   - `aggregator.go` — SessionReport with NetworkSummary, CostSummary

2. **`internal/docker/proxy.go`** — extend Proxy
   - Add `logPath string`
   - Write hostname to `<logPath>/session-<id>.network.jsonl` per request
   - Thread session ID into `NewProxy()`

3. **`internal/docker/sandbox.go`** — extend Run()
   - Create `<envDir>/audit/`
   - Write `<envDir>/audit/<id>.meta.json` (host-side)
   - Mount `<envDir>/audit/` as `/ai-env/audit/` `:rw`
   - Mount extraction scripts as `/ai-env/scripts/` `:ro`
   - Wrap entrypoint: `sh -c "<agent>; /ai-env/scripts/extract-cost.sh <session-id>"`
   - Print session ID to stderr

4. **`cmd/activate_cmd.go`** — print session ID to stderr

5. **Extraction scripts** embedded via `embed.FS`, extracted to `<envDir>/scripts/`

6. **`cmd/audit_cmd.go`** — new cobra command
   - `aienv audit <session-id>` — aggregated report
   - `--env`, `--list`, `--json`, `--since`, `--until` flags

### Phase 2: Commands Audit (Post-V1)

- auditd inside container with `execve` rules
- `--cap-add AUDIT_CONTROL --cap-add AUDIT_WRITE`
- Parse `/var/log/audit/` → JSONL → aggregate

## Open Questions

1. Ephemeral volume → data loss: is the wrapper-on-exit approach sufficient, or do we need a different mount strategy that preserves agent-native audit data?
2. (resolved — Docker-only; container lifecycle provides the hook)
