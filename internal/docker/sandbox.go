package docker

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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

func Run(e *env.Env, cwd string) error {
	if err := Check(); err != nil {
		return err
	}

	if err := ensureImage(e.Agent); err != nil {
		return err
	}

	sessionID := fmt.Sprintf("aienv-%s-%d", e.Name, rand.Uint64())
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

	args := []string{
		"run", "--rm", "-it",
	}

	args = append(args, "-v", fmt.Sprintf("%s:/workspace", cwd))

	opencodeConfigDir := filepath.Join(os.Getenv("HOME"), ".config", "opencode")
	if info, err := os.Stat(opencodeConfigDir); err == nil && info.IsDir() {
		args = append(args, "-v", fmt.Sprintf("%s:/home/user/.config/opencode:ro", opencodeConfigDir))
	}

	opencodeLocalDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "opencode")
	if info, err := os.Stat(opencodeLocalDir); err == nil && info.IsDir() {
		volName := sessionID + "-share"
		if err := mountIsolatedVolume(opencodeLocalDir, volName, imageTag(e.Agent)); err != nil {
			return err
		}
		cleanupVolumes = append(cleanupVolumes, volName)
		args = append(args, "-v", fmt.Sprintf("%s:/home/user/.local/share/opencode", volName))
	}

	gitCfg := filepath.Join(os.Getenv("HOME"), ".gitconfig")
	if _, err := os.Stat(gitCfg); err == nil {
		args = append(args, "-v", fmt.Sprintf("%s:/home/user/.gitconfig:ro", gitCfg))
	}

	switch e.Agent {
	case "opencode":
		ocJSON := config.OpenCodeJSON(e.Name)
		args = append(args, "-v", fmt.Sprintf("%s:/ai-env/opencode.json:ro", ocJSON))

		for _, skill := range e.Skills {
			for _, prefix := range []string{".agents/skills", ".config/opencode/skills"} {
				dir := filepath.Join(os.Getenv("HOME"), prefix, skill.Name)
				if info, err := os.Stat(dir); err == nil && info.IsDir() {
					containerPath := fmt.Sprintf("/home/user/%s/%s", prefix, skill.Name)
					args = append(args, "-v", fmt.Sprintf("%s:%s:ro", dir, containerPath))
					break
				}
			}
		}

		args = append(args, "-e", "OPENCODE_CONFIG=/ai-env/opencode.json")

	case "claude-code":
		mcpCfg := filepath.Join(config.EnvDir(e.Name), "mcp-config.json")
		args = append(args, "-v", fmt.Sprintf("%s:/ai-env/mcp-config.json:ro", mcpCfg))

		claudeMD := filepath.Join(config.EnvDir(e.Name), "CLAUDE.md")
		args = append(args, "-v", fmt.Sprintf("%s:/ai-env/CLAUDE.md:ro", claudeMD))

		if e.Permissions != nil {
			claudeSettings := filepath.Join(config.EnvDir(e.Name), "claude-settings.json")
			if _, err := os.Stat(claudeSettings); err == nil {
				args = append(args, "-v", fmt.Sprintf("%s:/ai-env/claude-settings.json:ro", claudeSettings))
			}
		}

		claudeHome := filepath.Join(os.Getenv("HOME"), ".claude")
		if info, err := os.Stat(claudeHome); err == nil && info.IsDir() {
			volName := sessionID + "-claude"
			if err := mountIsolatedVolume(claudeHome, volName, imageTag(e.Agent)); err != nil {
				return err
			}
			cleanupVolumes = append(cleanupVolumes, volName)
			args = append(args, "-v", fmt.Sprintf("%s:/home/user/.claude", volName))
		}

		claudeJSON := filepath.Join(os.Getenv("HOME"), ".claude.json")
		if _, err := os.Stat(claudeJSON); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:/home/user/.claude.json.ro:ro", claudeJSON))
		}

	default:
		return fmt.Errorf("unsupported agent %q for Docker sandbox", e.Agent)
	}

	for _, key := range []string{"TERM", "COLORTERM", "LANG", "LC_ALL", "LC_CTYPE"} {
		if val := os.Getenv(key); val != "" {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
		}
	}

	args = append(args, "-e", "HOME=/home/user")

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

	switch e.Agent {
	case "opencode":
		args = append(args, imageTag(e.Agent), "opencode")
	case "claude-code":
		claudeArgs := []string{"claude"}
		claudeArgs = append(claudeArgs, "--mcp-config", "/ai-env/mcp-config.json")
		claudeArgs = append(claudeArgs, "--append-system-prompt-file", "/ai-env/CLAUDE.md")

		if e.Permissions != nil {
			claudeArgs = append(claudeArgs, "--settings", "/ai-env/claude-settings.json")
		}

		for _, rule := range e.Rules {
			containerPath := rule.Path
			if !filepath.IsAbs(rule.Path) {
				containerPath = filepath.Join("/workspace", rule.Path)
			} else {
				if _, err := os.Stat(rule.Path); err == nil {
					args = append(args, "-v", fmt.Sprintf("%s:%s:ro", rule.Path, rule.Path))
				}
			}
			claudeArgs = append(claudeArgs, "--append-system-prompt-file", containerPath)
		}

		claudeArgs = append(claudeArgs, "--strict-mcp-config")
		if e.Model != "" {
			claudeArgs = append(claudeArgs, "--model", e.Model)
		}
		args = append(args, imageTag(e.Agent))
		args = append(args, "sh", "-c",
			fmt.Sprintf("cp /home/user/.claude.json.ro /home/user/.claude.json 2>/dev/null; exec claude %s",
				strings.Join(claudeArgs[1:], " ")))
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker sandbox exited with error: %w", err)
	}
	return nil
}
