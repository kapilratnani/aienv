# Agent Interface Refactor: DockerConfig

## Status
Accepted

## Context
The current Agent interface includes an `ActivateCommand()` method that returns a shell string for non-Docker execution. This method is only used in non-Docker mode and represents dead code after migrating to Docker-only.

Additionally, `docker/sandbox.go` contains a giant switch statement that handles agent-specific Docker configuration (mounts, environment variables, entrypoint), violating the Open/Closed Principle.

## Decision
Replace the `ActivateCommand()` method with a `DockerConfig()` method that returns structured Docker configuration. Each agent will implement this method to return its specific mount points, environment variables, and entrypoint. The sandbox implementation will then apply generic configuration (workdir, gitconfig, TTY vars, etc.) and execute the container.

## Agent Interface Changes
```go
type DockerConfig struct {
    Mounts     []string // "-v" arguments (config files, skills, etc.)
    EnvVars    []string // "-e" arguments
    Entrypoint []string // ["image:tag", "command", "--flag", ...]
}

type Agent interface {
    Name() string
    GenerateFiles(e *env.Env, cwd string) ([]AgentFile, error)
    DockerConfig(envDir string, e *env.Env, sessionID string) (*DockerConfig, error)
}
```

## Sandbox.Run() Changes
The `Run()` function will:
1. Validate Docker availability and ensure agent image exists
2. Generate session ID
3. Call `agent.DockerConfig()` to get agent-specific configuration
4. Append generic mounts (workdir, gitconfig, global config dirs)
5. Append environment variables (TTY vars, MCP env vars, proxy setup)
6. Append entrypoint
7. Execute `docker run` with all combined arguments

## Consequences

### Positive
- Eliminates dead code (ActivateCommand implementations)
- Removes giant switch statement in sandbox.go
- Follows Open/Closed Principle (easy to add new agents)
- Clear separation between agent-specific and generic Docker configuration
- Easier to test agent Docker configuration logic
- Consistent container creation flow

### Negative
- Requires updating all existing agent implementations
- Slightly more complex return type (struct vs string)
- Need to handle errors from DockerConfig() method

## Implementation Details
Each agent's DockerConfig() method should:
- Return agent-specific volume mounts (config files, etc.)
- Return agent-specific environment variables
- Return the container entrypoint (image and command)
- Not include generic mounts like workdir or gitconfig (added by sandbox)
- Not include TTY environment variables or proxy setup (added by sandbox)