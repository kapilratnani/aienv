# Roadmap

## Priority 2: Repo-Local `.aienv.yaml` + `aienv up`

- `.aienv.yaml` as canonical repo-level env declaration (checked into VCS)
- Discovery algorithm: walk up from CWD to git root
- Registration symlink at `~/.ai-envs/_repo_/<dirname>/`
- Shell function: `aienv up` reserved subcommand
- `aienv down` to unregister repo-local env
- `.aienv.lock` for reproducible environments (pins MCP/skill versions)
- `--frozen` mode for CI — fails if lockfile would change
- `aienv init-repo` to scaffold `.aienv.yaml` from project type detection

## Priority 3: Permission Policies & Trust

- `permissions` struct in schema: `filesystem` (writable/readonly), `network` (allow/deny), `bash` (allow/deny)
- Permission to OpenCode `opencode.json` permission config translation
- Permission to Docker sandbox enforcement (read-only mounts, network rules)
- Interactive trust/review prompt on first `aienv up` for unknown repos
- Trust cache: `~/.config/aienv/trust/<content-hash>.json`, invalidated on `.aienv.yaml` changes
- Future: permission signing via GPG/Sigstore for trustless verification

## Priority 4: Agent Expansion Framework

- Redesign agent architecture for simplicity — most agents share common patterns (MCPs as JSON, instructions as markdown rules files)
- Extract a base/default agent with overridable paths, file templates, and activation command patterns
- Deferred: actual per-agent implementations (Cursor, GitHub Copilot, Windsurf, Codex) until framework is stable

## Priority 5: Default Environment Directory

- On activation, change directory to a configured workspace path
- Could be a `workdir` field in the env schema, or default to CWD at activate time

## Priority 6: Custom MCP/Skill Repositories

- Support additional registries beyond skills.sh and modelcontextprotocol.io
- Enterprise/internal repo support via configurable registry list
- Multi-registry orchestration (merge results from all configured repos)

## Priority 7: Sharing & Team Features

- `aienv install <source>` — install environments from GitHub repos or URLs
- `aienv publish` — export environment to GitHub
- `aienv update` — pull latest version of a shared environment
- GitHub-based discovery: search for `.aienv.yaml` files

## Priority 8: Docker Image Size

- Multi-stage Dockerfiles, distroless base, smaller image
- Not a blocker for local development
