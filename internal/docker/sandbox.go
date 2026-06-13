package docker

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
)

func imageTag(agent string) string {
	switch agent {
	case "opencode":
		return "aienv/sandbox:latest-opencode"
	case "claude-code":
		return "aienv/sandbox:latest-claude"
	default:
		return "aienv/sandbox:latest"
	}
}

func Check() error {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not available: %w\nInstall Docker from https://docs.docker.com/engine/install/", err)
	}
	return nil
}

func Build(agent string) error {
	dfData, err := readDockerfile(agent)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "aienv-docker-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, dfData, 0644); err != nil {
		return fmt.Errorf("writing Dockerfile: %w", err)
	}

	tag := imageTag(agent)
	fmt.Fprintf(os.Stderr, "Building Docker sandbox image (%s)...\n", tag)
	cmd := exec.Command("docker", "build", "-t", tag, tmpDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("building Docker image: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Docker sandbox image built successfully.\n")
	return nil
}

func BuildAll() error {
	for _, agent := range []string{"opencode", "claude-code"} {
		if err := Build(agent); err != nil {
			return err
		}
	}
	return nil
}

func ensureImage(agent string) error {
	tag := imageTag(agent)
	cmd := exec.Command("docker", "image", "inspect", tag)
	if cmd.Run() == nil {
		return nil
	}
	return Build(agent)
}

func mountIsolatedVolume(hostDir, volName, imageTag string) error {
	ec := exec.Command("docker", "volume", "create", volName)
	if err := ec.Run(); err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}

	init := exec.Command("docker", "run", "--rm", "--user", "root",
		"-v", volName+":/target",
		"-v", hostDir+":/source:ro",
		imageTag,
		"sh", "-c", "cp -a /source/. /target/. && chown -R 1000:1000 /target")
	if out, err := init.CombinedOutput(); err != nil {
		exec.Command("docker", "volume", "rm", "-f", volName).Run()
		return fmt.Errorf("initializing volume from host data: %w\n%s", err, out)
	}

	return nil
}

// Run executes the agent in a Docker container with the given environment and workdir.
func Run(e *env.Env, workdir string, sessionID string) error {
	if err := Check(); err != nil {
		return err
	}

	if err := ensureImage(e.Agent); err != nil {
		return err
	}

	var cleanupVolumes []string

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh {
		}
	}()

	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
		for _, v := range cleanupVolumes {
			exec.Command("docker", "volume", "rm", "-f", v).Run()
		}
	}()

	// Get agent-specific Docker configuration
	ag, err := agents.Get(e.Agent)
	if err != nil {
		return err
	}
	dockerCfg, err := ag.DockerConfig(config.EnvDir(e.Name), e, sessionID)
	if err != nil {
		return fmt.Errorf("getting docker config for agent %q: %w", e.Agent, err)
	}

	args := []string{
		"run", "--rm", "-it",
	}

	// Mount workdir as /workspace
	args = append(args, "-v", fmt.Sprintf("%s:/workspace", workdir))
	args = append(args, "--workdir", "/workspace")

	// Mount .gitconfig for commit authorship
	home, _ := os.UserHomeDir()
	if home != "" {
		gcPath := filepath.Join(home, ".gitconfig")
		if _, err := os.Stat(gcPath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:/home/user/.gitconfig:ro", gcPath))
		}
	}

	// Agent-specific mounts
	for _, m := range dockerCfg.Mounts {
		args = append(args, "-v", m)
	}

	// Agent-specific environment variables
	for _, ev := range dockerCfg.EnvVars {
		args = append(args, "-e", ev)
	}

	// Agent-specific isolated volumes for persistent state (auth, sessions)
	switch e.Agent {
	case "opencode":
		opencodeLocalDir := filepath.Join(home, ".local", "share", "opencode")
		if info, err := os.Stat(opencodeLocalDir); err == nil && info.IsDir() {
			volName := sessionID + "-share"
			if err := mountIsolatedVolume(opencodeLocalDir, volName, imageTag(e.Agent)); err != nil {
				return err
			}
			cleanupVolumes = append(cleanupVolumes, volName)
			args = append(args, "-v", fmt.Sprintf("%s:/home/user/.local/share/opencode", volName))
		}

	case "claude-code":
		claudeHome := filepath.Join(home, ".claude")
		if info, err := os.Stat(claudeHome); err == nil && info.IsDir() {
			volName := sessionID + "-claude"
			if err := mountIsolatedVolume(claudeHome, volName, imageTag(e.Agent)); err != nil {
				return err
			}
			cleanupVolumes = append(cleanupVolumes, volName)
			args = append(args, "-v", fmt.Sprintf("%s:/home/user/.claude", volName))
		}
	}

	// Generic environment variables
	for _, key := range []string{"TERM", "COLORTERM", "LANG", "LC_ALL", "LC_CTYPE"} {
		if val := os.Getenv(key); val != "" {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
		}
	}

	args = append(args, "-e", "HOME=/home/user")

	// Network proxy setup if needed
	if e.Permissions != nil && e.Permissions.Network != nil {
		allowlist := e.Permissions.Network.Allow
		denylist := e.Permissions.Network.Deny

		providerHosts := env.DetectProviderEndpoints(e.Agent)
		allowlist = append(allowlist, providerHosts...)
		unique := map[string]bool{}
		deduped := make([]string, 0, len(allowlist))
		for _, h := range allowlist {
			if !unique[h] {
				unique[h] = true
				deduped = append(deduped, h)
			}
		}
		allowlist = deduped

		p, port, err := RunProxy(allowlist, denylist, "0.0.0.0")
		if err != nil {
			return fmt.Errorf("starting network proxy: %w", err)
		}
		defer p.Close()

		proxyURL := fmt.Sprintf("http://host.docker.internal:%d", port)
		args = append(args, "--add-host", "host.docker.internal:host-gateway")
		args = append(args, "-e", "HTTP_PROXY="+proxyURL)
		args = append(args, "-e", "HTTPS_PROXY="+proxyURL)
		args = append(args, "-e", "NO_PROXY=localhost,127.0.0.1")
	}

	// MCP server environment variables
	for _, srv := range e.MCPServers {
		for _, val := range srv.Env {
			if len(val) > 4 && val[:4] == "env:" {
				envKey := val[4:]
				if os.Getenv(envKey) != "" {
					args = append(args, "-e", envKey)
				}
			}
		}
	}

	// Entrypoint: [image:tag, "command", "--flag", ...]
	args = append(args, dockerCfg.Entrypoint...)
	fmt.Printf("%v\n", args)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker sandbox exited with error: %w", err)
	}
	return nil
}
