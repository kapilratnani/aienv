# aienv Domain Context

AI agent sandboxing platform that provides isolation, dependency management, and comprehensive audit trails for coding agents.

## Language

**Black Box Agent**:
An AI coding assistant (OpenCode, Claude Code, Cursor, etc.) treated as an opaque binary. aienv does not generate config files, install MCPs, or understand agent-specific formats. The user provides the agent binary, its config, and any tools it needs.
_Avoid_: Agent integration, agent support

**Environment**:
A named, self-contained configuration at `~/.local/share/aienv/<name>/env.yaml` that declares which agent to run, what to mount, what dependencies to install, and what security rules to enforce. Analogous to a Docker Compose service definition but for AI agents.
_Avoid_: Virtualenv, workspace

**Build-time Image**:
A Docker image generated from an Environment definition. The base image (`aienv/sandbox:latest`) contains governance scripts; the user's env adds agent installation, system packages, and custom tools via a generated Dockerfile. Cached by hash of the env.yaml content.

**Governance Scripts**:
The entrypoint wrapper and supporting tools in the base image that provide audit logging, network enforcement, and session management without modifying the agent binary.

**Mount**:
A host-to-container bind mount declared in the Environment. All mounts are read-only by default; the `writable` flag makes a mount read-write for persistent agent state (auth tokens, sessions, cache).

**Permissions**:
Sandbox-level security rules enforced by the network proxy (allow/deny/learn for HTTP/S hosts) and Docker run flags (`--read-only`, allowed filesystem paths).

**Audit Log**:
Append-only JSONL event stream capturing commands executed and network requests made inside the sandbox. Written to `~/.local/share/aienv/<name>/audit/<session-id>/events.jsonl`.

**Learn Mode**:
A permissions mode where the network proxy logs all hosts contacted and suggests an allowlist on exit. Default when no `permissions.network.allow` is configured.

**Session**:
A single invocation of `aienv up`. Each activation creates a new container with a unique session ID. Concurrent activations produce separate containers.

**Worktree**:
A git worktree — a linked checkout of a Git repository that shares the repository's `.git` directory with the parent working tree. aienv can create a worktree from the env's `workdir`, mount it into the sandbox as the agent's working directory, run the agent, and clean it up on session exit. Used for isolating agent work to a task-specific branch.
