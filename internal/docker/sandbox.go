package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
)

const imageTag = "aienv/sandbox:latest"

const dockerfile = `FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    golang-go \
    nodejs \
    npm \
    python3 \
    python3-pip \
    pipx \
    && rm -rf /var/lib/apt/lists/*
RUN userdel -r ubuntu 2>/dev/null; useradd -m -u 1000 -s /bin/bash user
USER user
WORKDIR /workspace
ENV PATH="/home/user/.local/bin:${PATH}"
`

func Check() error {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not available: %w\nInstall Docker from https://docs.docker.com/engine/install/", err)
	}
	return nil
}

func Build() error {
	tmpDir, err := os.MkdirTemp("", "aienv-docker-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("writing Dockerfile: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Building Docker sandbox image (%s)...\n", imageTag)
	cmd := exec.Command("docker", "build", "-t", imageTag, tmpDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("building Docker image: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Docker sandbox image built successfully.\n")
	return nil
}

func ensureImage() error {
	cmd := exec.Command("docker", "image", "inspect", imageTag)
	if cmd.Run() == nil {
		return nil
	}
	return Build()
}

func Run(e *env.Env, cwd string) error {
	if err := Check(); err != nil {
		return err
	}

	if err := ensureImage(); err != nil {
		return err
	}

	args := []string{
		"run", "--rm", "-it",
	}

	args = append(args, "-v", fmt.Sprintf("%s:/workspace", cwd))

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

	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	if info, err := os.Stat(sshDir); err == nil && info.IsDir() {
		args = append(args, "-v", fmt.Sprintf("%s:/home/user/.ssh:ro", sshDir))
	}

	for _, key := range []string{"TERM", "COLORTERM", "LANG", "LC_ALL", "LC_CTYPE"} {
		if val := os.Getenv(key); val != "" {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
		}
	}

	args = append(args, "-e", "OPENCODE_CONFIG=/ai-env/opencode.json")
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

	openCodePath, err := exec.LookPath("opencode")
	if err != nil {
		return fmt.Errorf("opencode not found in PATH — install opencode first: %w", err)
	}
	args = append(args, "-v", fmt.Sprintf("%s:/usr/local/bin/opencode:ro", openCodePath))

	args = append(args, imageTag, "opencode")

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker sandbox exited with error: %w", err)
	}
	return nil
}
