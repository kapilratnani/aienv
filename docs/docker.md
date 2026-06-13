# Docker Design

Last updated: 2026-06-14
Status: Current

## Architecture

- One Docker image per environment, tagged `aienv/env:<name>`
- Dockerfile auto-generated from env YAML
- Image cached by content hash of YAML
- Container launched with writable bind mounts (no `--read-only`, no tmpfs)

## Mount Model

Mounts defined in `env.yaml`:

```yaml
agent:
  mounts:
    - source: ~/project
      target: /workspace
    - source: ~/.config/opencode
      target: ~/.config/opencode
      writable: true
```

- `~` in source is expanded to host home directory
- `~` in target is resolved to `/home/agent/` (container agent user home)
- Agent state directories (`~/.config`, `~/.local/share`, `~/.cache`) are writable bind mounts
- All mounts use `-v type=bind` with either `:ro` or `:rw` suffix
- Audit dir at `<envDir>/audit/<session-id>/` mounted at `/aienv/audit:rw`

## Container Runtime

```bash
docker run \
  -v host_src:/workspace \
  -v host_agent_config:/home/agent/.config \
  -v host_agent_local:/home/agent/.local/share \
  -v host_agent_cache:/home/agent/.cache \
  -v audit_dir:/aienv/audit:rw \
  -e HTTP_PROXY=http://host.docker.internal:PORT \
  -e HTTPS_PROXY=http://host.docker.internal:PORT \
  -e HOME=/home/agent \
  --add-host host.docker.internal:host-gateway \
  -it --rm \
  aienv/env:<name>
```

- Agent runs with uid/gid 1000 (non-root inside container)
- No `--read-only` — agent needs to write to its state directories
- No `--tmpfs` — parent-path tmpfs at e.g. `/home/agent/` shadows subdirectory bind mounts
- Default Docker bridge network (no `--network host`)
- Proxy container networking via `host.docker.internal` DNS

## Commands

| Command | Behavior |
|---------|----------|
| `aienv up <name>` | Build image if missing, trust prompt (if needed), audit dir + session, proxy start, `docker run` |
| `aienv build <name>` | Force `docker build` even if cached |
| `aienv shell <name>` | `docker run` with same mounts/env but no proxy, no audit, no session — interactive bash |

## Dockerfile

Generated from env YAML:

```dockerfile
FROM aienv/sandbox:latest
USER root
RUN apt-get update && apt-get install -y ... && rm -rf /var/lib/apt/lists/*
RUN npm install -g ...
ENV ... from agent.env
COPY files/ /home/agent/
USER agent
WORKDIR /home/agent
```

Base image (`aienv/sandbox:latest`) includes: `git`, `curl`, `ca-certificates`, `gh`, `nodejs`, `npm`, `python3`, `python3-pip`, `pipx`.

## Audit Integration

- Each activation creates `<envDir>/audit/<session-id>/`
- Proxy writes `network.jsonl` with every request
- `session.meta.json` written at start
- Learn mode prints all observed hosts on proxy close
