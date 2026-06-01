package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

func Run(e *env.Env, cwd string) error {
	if err := Check(); err != nil {
		return err
	}

	if err := ensureImage(e.Agent); err != nil {
		return err
	}

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
		args = append(args, "--mount", fmt.Sprintf("type=bind,source=%s,target=/home/user/.local/share/opencode,bind-propagation=private", opencodeLocalDir))
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

		claudeCfgDir := filepath.Join(config.EnvDir(e.Name), "claude-config")
		if info, err := os.Stat(claudeCfgDir); err == nil && info.IsDir() {
			args = append(args, "-v", fmt.Sprintf("%s:/home/user/.claude:ro", claudeCfgDir))
		}

		for _, skill := range e.Skills {
			dir := filepath.Join(os.Getenv("HOME"), ".claude", "skills", skill.Name)
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				args = append(args, "-v", fmt.Sprintf("%s:/home/user/.claude/skills/%s:ro", dir, skill.Name))
			}
		}

		args = append(args, "-e", "OPENCODE_CONFIG=/ai-env/mcp-config.json")

	default:
		return fmt.Errorf("unsupported agent %q for Docker sandbox", e.Agent)
	}

	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	if info, err := os.Stat(sshDir); err == nil && info.IsDir() {
		args = append(args, "-v", fmt.Sprintf("%s:/home/user/.ssh:ro", sshDir))
	}

	for _, key := range []string{"TERM", "COLORTERM", "LANG", "LC_ALL", "LC_CTYPE"} {
		if val := os.Getenv(key); val != "" {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
		}
	}

	args = append(args, "-e", "HOME=/home/user")
	args = append(args, "--network", "host")

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
		args = append(args, claudeArgs...)
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
