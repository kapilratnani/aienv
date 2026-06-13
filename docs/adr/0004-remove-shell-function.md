# Remove Shell Function

## Status
Accepted

## Context
aienv currently provides a shell function that users can eval to activate environments:
```bash
eval "$(aienv activate <name>)"
```

This approach has several drawbacks:
- Modifies the current shell session (can cause conflicts)
- Requires shell-specific implementation (fish, bash, zsh)
- Complex quoting and escaping issues
- Difficult to debug and test
- Security concerns with arbitrary code execution via eval

In Docker-only mode, we don't need to modify the shell session since all agent execution happens in isolated containers.

## Decision
Remove the shell function and associated code:
- Delete `internal/shell/` package
- Delete `cmd/init_cmd.go`
- Remove shell function installation functionality
- Remove `aienv init` command

Users will activate environments using the standard `aienv activate <name>` command, which will print the docker run command to execute, or we can modify the behavior to directly execute the docker run command.

## Consequences

### Positive
- Simplified codebase
- Eliminates shell-specific complexity
- Removes security risks associated with eval
- Consistent activation mechanism
- Easier to test and debug

### Negative
- Breaking change for users who rely on the shell function
- Slightly different activation workflow
- Need to update user documentation and habits

## Implementation Plan
1. Remove `internal/shell/` directory and all files
2. Remove `cmd/init_cmd.go`
3. Remove shell function registration from `cmd/root.go` if applicable
4. Update `cmd/activate_cmd.go` to always execute docker run (no need to return command for eval)
5. Update documentation and help text
6. Provide migration guidance for users

## Activation Workflow Changes
Before (with shell function):
```bash
# User runs in shell:
eval "$(aienv activate frontend)"
# This modifies current shell environment
```

After (shell function removed):
```bash
# User runs in shell:
aienv activate frontend
# This directly executes: docker run [...]
# Or alternatively:
eval "$(aienv activate frontend)"  # Still works if we keep command output
```

We need to decide whether `aienv activate` should:
1. Directly execute the docker run command (recommended for simplicity)
2. Continue to output the command for eval (maintains backward compatibility but adds complexity)

Given our Docker-only focus, option 1 (direct execution) is preferred.