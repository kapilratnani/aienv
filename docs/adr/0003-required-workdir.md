# Required Workdir Field

## Status
Accepted

## Context
Currently, the `workdir` field in the environment schema is optional. When not specified, the system falls back to using the current working directory (CWD). This creates complexity in the activation logic and inconsistent behavior.

In Docker-only mode, we need a consistent way to mount the user's code into the container. Making workdir required ensures that every environment explicitly defines where the user's code resides.

## Decision
Make the `workdir` field required in the environment schema. Remove CWD fallback behavior. The workdir will always be mounted as `/workspace` inside the agent container.

## Consequences

### Positive
- Simplified activation logic (no need to check for empty workdir)
- Consistent behavior across all environments
- Explicit documentation of where code resides for each environment
- Better integration with Docker workflows

### Negative
- Breaking change for existing environments without workdir
- Users must specify workdir when creating new environments
- Migration required for existing environments

## Implementation Plan
1. Update `internal/env/env.go` to validate workdir is non-empty in `Save()`
2. Update `cmd/create_cmd.go` to require workdir input (remove optional note)
3. Remove workdir fallback logic in `cmd/activate_cmd.go`
4. Update documentation to reflect required workdir
5. Provide guidance for migrating existing environments

## Migration Strategy
For existing environments without workdir:
1. During activation, detect missing workdir
2. Prompt user to specify workdir for the environment
3. Save the workdir to the environment configuration
4. Continue with activation using the specified workdir

Alternative: Provide a migration command to add workdir to existing environments.