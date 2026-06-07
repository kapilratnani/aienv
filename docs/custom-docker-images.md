# Custom Docker Images

Last updated: 2026-06-07
Status: Planned

## Quick Resume

Allow users to provide a custom Dockerfile per environment for project-specific dependencies (JDK, Rust, CUDA, etc.). The Dockerfile is used as a base layer; aienv builds a two-stage image with the agent installed on top.

## Grill Session (2026-06-07)

### Decisions

| # | Question | Decision |
|---|---|---|
| Q1 | How is custom image specified? | `dockerfile` top-level field in env YAML, path relative to env YAML's directory |
| Q2 | Relation to agent installation? | Two-stage build: user's Dockerfile builds base, aienv adds agent layer on top |
| Q3 | Field name and scope? | `dockerfile` top-level field (not nested under `docker:`) |
| Q4 | Build process? | Hash-based: `dockerfile` content → SHA256 → tag `aienv/env/<name>:<hash>`. Auto-build on hash miss at activation. `aienv docker build <envname>` for force rebuild. |
| Q5 | Build context? | Env YAML's directory (same dir as `~/.ai-envs/<name>/` or alongside `.aienv.yaml`) |
| Q6 | Image tag and cleanup? | `aienv/env/<name>:<hash>`. No automatic cleanup — user manages via `docker image rm`. Docker Desktop shows them. |

### Key Design Points

1. `dockerfile` field is relative to the env YAML directory (portable, works with repo-local `.aienv.yaml`)
2. Hash-based tagging enables automatic change detection: edit Dockerfile → new hash → rebuild on next activation
3. Two-stage build keeps concerns separate: user owns runtime deps, aienv owns agent installation
4. No changes needed for `edit_cmd`, `delete_cmd`, `list_cmd` — field is serialized/deserialized automatically
5. `agent_import.go`, `proxy.go`, trust system — no changes needed

## Schema

```yaml
name: my-java-env
agent: opencode
dockerfile: Dockerfile.java  # relative to env YAML directory
mcp:
  # ...
```

When `dockerfile` is not set, existing behavior is preserved (embedded `aienv/sandbox:latest-<agent>`).

## Implementation

### Build Flow (Two-Stage)

```
Step 1: docker build -t aienv/env/<name>:base-<hash> \
          -f <envDir>/<dockerfile> \
          <envDir>

Step 2: Generate wrapper Dockerfile:
          FROM aienv/env/<name>:base-<hash>
          # agent install + gh CLI + user setup
          (extracted from embedded opencode/claude Dockerfile)

        docker build -t aienv/env/<name>:<hash> \
          /tmp/aienv-wrapper-xxx

Activation uses: aienv/env/<name>:<hash>
```

The `base-` tag stays cached — if the user's Dockerfile hasn't changed, only Step 2 rebuilds (fast, just agent layer).

### Hash Computation

SHA256 of Dockerfile content + agent name concatenated. This ensures:
- Different content → different hash → rebuild
- Same content, different agent → different hash → rebuild

### Files Changed

| File | Change |
|---|---|
| `internal/env/env.go` | Add `Dockerfile string \`yaml:"dockerfile,omitempty"\`` field |
| `internal/docker/sandbox.go` | `imageTag()`, `Build()`, `ensureImage()` accept `*env.Env`; hash computation; two-stage build |
| `internal/docker/embed.go` | `readDockerfileForEnv()` reads user's file or falls back to embedded |
| `cmd/docker_cmd.go` | `aienv docker build <envname>` subcommand for force rebuild |
| `cmd/create_cmd.go` | `promptDockerfile()` prompt + summary display |
| `cmd/show_cmd.go` | Display Dockerfile when set |

### What Stays the Same

- `edit_cmd.go` — raw YAML, field auto-serializes
- `delete_cmd.go`, `list_cmd.go` — no changes
- `internal/env/yaml.go` — `Dockerfile` is optional, no validation needed
- `internal/agents/` — agent interface unchanged
- `proxy.go` — network policy is runtime concern
- Trust system — Dockerfile field changes trust hash trivially via YAML marshalling

### Example: Java Env

User creates `~/.ai-envs/java-dev/ai-env.yaml`:
```yaml
name: java-dev
agent: opencode
dockerfile: Dockerfile.java
```

User writes `~/.ai-envs/java-dev/Dockerfile.java`:
```dockerfile
FROM eclipse-temurin:21-jdk
RUN apt-get update && apt-get install -y maven
```

On `aienv java-dev --docker`:
1. Hash `Dockerfile.java` content → `abc123`
2. Check `docker image inspect aienv/env/java-dev:abc123` — miss
3. Build base: `docker build -t aienv/env/java-dev:base-abc123` (pulls temurin, installs maven)
4. Build wrapper: `FROM aienv/env/java-dev:base-abc123` + install opencode/gh
5. Tag: `aienv/env/java-dev:abc123`
6. Launch container with `aienv/env/java-dev:abc123`
