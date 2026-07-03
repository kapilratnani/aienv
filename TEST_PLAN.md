# Worktree Feature — Test Plan

## Feature Overview

`aienv up <env> --worktree <branch>` creates a git worktree from the env's first mount (repo),
mounts it into the sandbox as the agent's working directory, and cleans up on exit.

## Branch Resolution (in `WorktreeAdd`)

| Case | Action |
|------|--------|
| Branch exists locally | `git worktree add <wt> <branch>` |
| Branch exists on remote only | `git fetch origin <branch>` → `git branch <branch> origin/<branch>` → `git worktree add <wt> <branch>` |
| Branch does not exist | `git worktree add -b <branch> <wt> <baseBranch>` (baseBranch from `--worktree-base` or auto-detect) |

## Use Cases

### 1. New branch from default branch

```
aienv up my-env --worktree fix-login-bug
```

Precondition: repo has `main` (auto-detected). Worktree created at `<parent>/myapp-fix-login-bug`,
mounted writable, shared `.git` mounted. Cleaned up on exit.

### 2. New branch from explicit base

```
aienv up my-env --worktree epic-feature --worktree-base develop
```

Precondition: `develop` branch exists locally.

### 3. Existing local branch

```
aienv up my-env --worktree feature-x
```

Precondition: `feature-x` exists locally.

### 4. Existing remote-only branch

```
aienv up my-env --worktree feature-y
```

Precondition: `feature-y` exists on `origin` only. Fetched, local tracking branch created, then worktree added.

### 5. Keep worktree after exit

```
aienv up my-env --worktree debug-session --worktree-keep
```

Worktree persists on disk after agent exits or Ctrl+C.

### 6. Reuse stale worktree

```
aienv up my-env --worktree fix-login-bug
```

Precondition: worktree dir exists from prior session (e.g., SIGKILL). Detect and reuse.

### 7. No workdir mount

```
aienv up misconfigured-env --worktree some-branch
```

Expected: error `"cannot create worktree: no workdir mount configured"`

### 8. Git not available

```
aienv up my-env --worktree some-branch
```

Precondition: `git` CLI not on PATH. Expected: error `"git is required for worktree support"`

### 9. Path collision — not a worktree

```
aienv up my-env --worktree static-dir
```

Precondition: `<parent>/myapp-static-dir` exists but is not a git worktree.
Expected: error `"path exists but is not a git worktree; remove it manually"`

### 10. Cannot determine base branch

```
aienv up my-env --worktree new-branch
```

Precondition: no `origin/HEAD`, `main`, or `master`. No `--worktree-base` given.
Expected: error `"cannot determine base branch; specify with --worktree-base"`

### 11. WorktreeAdd succeeds, ResolveGitDir fails (rollback)

```
aienv up my-env --worktree fragile-branch
```

Expected: worktree created, then cleaned up on ResolveGitDir failure. Error propagated.

### 12. Signal-based cleanup (SIGINT)

```
aienv up my-env --worktree cleanup-test
```

Expected: Ctrl+C triggers worktree removal, exit code 1.

### 13. Positional activation ignores worktree

```
aienv my-env --worktree ignored
```

`--worktree` only registered on `up` subcommand. Positional shortcut bypasses flag parsing.

---

## Existing Unit Tests (in `internal/env/git_test.go`)

| Test | Status |
|------|--------|
| `TestIsGitWorktree` | ✅ |
| `TestResolveGitDir` | ✅ |
| `TestRepoName` | ✅ |
| `TestWorktreePath` | ✅ |
| `TestBranchExistsLocally` | ✅ |
| `TestBranchExistsOnRemote` | ✅ |
| `TestGetDefaultBranch` | ✅ |
| `TestGitAvailable` | ✅ |
| `TestWorktreeAddAndRemove` | ✅ |

## Missing Tests

### Unit tests to add in `git_test.go`

- `TestWorktreeAddRemoteOnly` — set up a repo with a remote, push a branch, verify WorktreeAdd fetches and creates it
- `TestWorktreeAddNewBranch` — verify `git worktree add -b <branch> <wt> <base>` is invoked for new branches
- `TestWorktreeAddRemoteOnly_NoFetchNeeded` — remote branch that's already fetched locally

### Activation-level tests (new file: `cmd/wt_test.go` or extract to `internal/...`)

To test the worktree block in `activate_cmd.go` without Docker, the code needs to be extracted into a testable function (currently it's inline in `runEnv`). Options:

**Option A**: Extract a function `func setupWorktree(e *env.Env, branch, baseBranch string, keep bool) (cleanup func(), err error)` that returns a cleanup closure. This can be unit-tested without involving `docker.Run()`.

**Option B**: Write an integration-style test that creates a temp env + temp repo and calls `runEnv` directly, using a mock agent command (like `echo done`) so Docker exits immediately.

### Proposed extraction

Extract worktree logic from `activate_cmd.go` into a function on `env` package or a new package:

```go
// in internal/env/worktree.go
type WorktreeConfig struct {
    RepoPath   string
    Branch     string
    BaseBranch string
    Keep       bool
}

func SetupWorktree(cfg *WorktreeConfig) (mounts []Mount, cleanup func(), err error)
```

This makes the worktree logic independently testable.
