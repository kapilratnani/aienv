# Docker-Only Migration

## Status
Accepted

## Context
aienv currently supports both Docker and non-Docker modes for running AI agents. The non-Docker mode uses shell functions and `eval` to modify the current shell session, while Docker mode runs agents in isolated containers.

Maintaining both modes increases complexity:
- Code duplication in agent implementations (ActivateCommand vs Docker execution)
- Complex conditional logic in activation paths
- Inconsistent behavior between modes
- Security concerns with shell modification via eval
- Dependency on shell-specific configurations

## Decision
Remove all non-Docker code paths and make aienv a Docker-only sandbox for running terminal-based AI agents. All activation will launch a container. No shell function, no eval, no `--docker` flag — just `docker run`.

## Consequences

### Positive
- Simplified codebase with single execution path
- Consistent behavior across all environments
- Improved security (no shell modification via eval)
- Better isolation and reproducibility
- Easier maintenance and testing
- Clear separation of concerns (environment configuration vs execution)

### Negative
- Docker becomes a hard dependency
- Potential performance overhead from container startup
- Loss of ability to modify current shell session directly
- Need for users to have Docker installed and accessible

## Implementation Plan
1. Replace `ActivateCommand()` method in Agent interface with `DockerConfig()` 
2. Remove non-Docker execution path in `cmd/activate_cmd.go`
3. Remove `dockerMode` flag and related code
4. Delete `internal/shell/` package and `cmd/init_cmd.go`
5. Make `workdir` required in environment schema
6. Update `internal/docker/sandbox.go` to use agent.DockerConfig()
7. Implement DockerConfig() for OpenCode and Claude Code agents
8. Plan for Codex/Copilot DockerConfig() implementations
9. Update documentation to reflect Docker-only mode