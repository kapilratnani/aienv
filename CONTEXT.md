# aienv Domain Context

## Core Concepts

### Environment
A self-contained directory that encapsulates MCP servers, agent skills, and configuration needed for a specific AI coding task. Similar to Python's virtualenv but for AI agents.

### Agent
An AI coding assistant (like OpenCode, Claude Code) that can be launched within an environment. Each agent has specific configuration and execution requirements.

### Workdir
The working directory that gets mounted into the agent container at `/workspace`. This is where the user's code resides and where the agent operates.

### Docker Container
The execution environment for agents in Docker-only mode. All agent activation happens through `docker run` with appropriate mounts, environment variables, and entrypoint.

### MCP Server
Model Context Protocol servers that provide tools, resources, and capabilities to AI agents. Can be local (stdio) or remote (HTTP/SSE).

### Skill
Reusable functionality that can be added to environments to extend agent capabilities with specific domain knowledge or tools.

### Permission
Security boundaries that control what an agent can access (filesystem, network, bash commands) and how.

## Key Relationships

- An Environment contains zero or more MCP Servers
- An Environment contains zero or more Skills  
- An Environment specifies exactly one Agent to use
- An Environment has optional Permission constraints
- An Environment has an optional Workdir (required in Docker-only mode)
- Agents generate configuration files based on Environment specifications
- Agents run inside Docker containers with environment-specific configuration mounted in

## Operational Flow

1. User creates environment with `aienv create <name>`
2. User activates environment with `aienv activate <name>` (or just `aienv <name>`)
3. System generates agent configuration files in environment directory
4. System launches agent in Docker container with:
   - Workdir mounted at `/workspace`
   - Agent-specific configuration mounted at appropriate paths
   - Environment variables set for agent configuration
   - MCP servers connected via stdio or network
   - Skills made available to agent
   - Permissions enforced through container restrictions