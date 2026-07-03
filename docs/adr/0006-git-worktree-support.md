# Git Worktree Support for Task Isolation

Let aienv create a git worktree from the env's workdir, mount it into the sandbox as the agent's working directory, and clean it up after the session.

## Status

Accepted

## Context

Developers frequently work on multiple concurrent tasks (bugfixes, features) against the same repository. Without worktree support, the user must either:

- Run the agent directly in the repo's main working tree (risks dirty state, conflicting changes)
- Manually create worktrees and configure mounts before each activation
- Point the env's workdir to a different directory per task

The env model has a single `workdir` path — it cannot express "mount this branch's working tree". The agent needs a clean, task-scoped workspace that it can commit and push from, without contaminating the user's main checkout.

## Decision

Add `--worktree <branch>`, `--worktree-base <base>`, and `--worktree-keep` flags to `aienv up`. When `--worktree` is given:

1. aienv checks if the branch exists (local then remote). If remote-only, fetches it. If neither, creates it from base.
2. `git worktree add` creates a worktree directory beside the repo (e.g. `myrepo-fix-api-bug`).
3. The env's workdir mount is replaced with the worktree directory, mounted `:rw` at `/workspace`.
4. The shared `.git` directory from the parent repo is mounted `:rw` at the same host path so git commands resolve inside the container.
5. On normal exit, `git worktree remove --force` removes the worktree. On SIGINT/SIGTERM, the same cleanup runs.
6. If `--worktree-keep` is given, the worktree is left on disk after exit.

## Considered Options

- **In-container git operations only** — Mount the main repo and let the agent manage branches internally. Rejected because the agent's git workflow is unpredictable and may conflict with the user's working tree.

- **Docker volume per worktree** — Create a Docker named volume, clone the repo fresh, run the agent. Rejected because cloning large repos wastes time/bandwidth, and SSH/GH auth forwarding adds complexity.

- **Host-side worktree (chosen)** — Uses `git worktree add` on the host, which is instant (no clone), shares objects/refs, and is the git-native way to get independent working trees. The worktree is just a bind mount into the container.

## Consequences

### Positive
- Zero-cost branch switching — worktrees share Git objects, no clone needed
- Agent gets a clean working tree with full git capabilities (commit, push, status)
- User's main working tree is never touched
- Standard git tools remain usable on the host during agent runs

### Negative
- `git` CLI becomes a runtime dependency for host-side operations (only when `--worktree` is used)
- The shared `.git` directory is mounted writable — a malicious agent could corrupt the repo's git history (mitigated by Docker sandbox isolation, same as any mount)
- Stale worktrees may leak if `aienv` is SIGKILL'd (mitigated by pre-flight reuse/recreate logic)
