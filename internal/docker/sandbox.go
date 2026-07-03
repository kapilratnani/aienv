package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kapilratnani/aienv/internal/audit"
	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
)

const baseImage = "aienv/sandbox:latest"

func Check() error {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not available: %w\nInstall Docker from https://docs.docker.com/engine/install/", err)
	}
	return nil
}

func BuildBase() error {
	dfData, err := BaseDockerfile()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "aienv-base-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, dfData, 0644); err != nil {
		return fmt.Errorf("writing Dockerfile: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Building base sandbox image...\n")
	cmd := exec.Command("docker", "build", "-t", baseImage, tmpDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("building base image: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Base sandbox image built successfully.\n")
	return nil
}

func ensureBaseImage() error {
	cmd := exec.Command("docker", "image", "inspect", baseImage)
	if cmd.Run() == nil {
		return nil
	}
	return BuildBase()
}

func Build(e *env.Env) error {
	if err := Check(); err != nil {
		return err
	}
	if err := ensureBaseImage(); err != nil {
		return err
	}

	dockerfile := generateDockerfile(e)

	tmpDir, err := os.MkdirTemp("", "aienv-env-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("writing Dockerfile: %w", err)
	}

	tag := ImageTag(e)
	fmt.Fprintf(os.Stderr, "Building environment image...\n")
	cmd := exec.Command("docker", "build", "-t", tag, tmpDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("building environment image: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Environment image built successfully.\n")
	return nil
}

func ImageTag(e *env.Env) string {
	yamlPath := config.EnvYAML(e.Meta.Name)
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Sprintf("aienv/env:%s", e.Meta.Name)
	}
	return config.ImageTag(config.ComputeHash(data))
}

func EnsureImage(e *env.Env) error {
	tag := ImageTag(e)
	cmd := exec.Command("docker", "image", "inspect", tag)
	if cmd.Run() == nil {
		return nil
	}
	return Build(e)
}

func generateDockerfile(e *env.Env) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("FROM %s\n", baseImage))

	// Build as root; agent user is only for runtime
	b.WriteString("USER root\n")

	if len(e.Deps.Packages) > 0 {
		b.WriteString("RUN apt-get update && apt-get install -y --no-install-recommends \\\n")
		for i, pkg := range e.Deps.Packages {
			end := " \\\n"
			if i == len(e.Deps.Packages)-1 {
				end = " && rm -rf /var/lib/apt/lists/*\n"
			}
			b.WriteString(fmt.Sprintf("    %s%s", pkg, end))
		}
	}

	for _, cmd := range e.Agent.Install {
		b.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}

	for _, cmd := range e.Deps.Custom {
		b.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}

	// Switch to agent user for runtime
	b.WriteString("USER agent\n")

	return b.String()
}

func Run(e *env.Env, exitMode bool) error {
	if err := Check(); err != nil {
		return err
	}

	if err := EnsureImage(e); err != nil {
		return err
	}

	tag := ImageTag(e)
	sessionID := config.GenerateSessionID()

	// Set up audit directory if audit.persist is enabled
	var auditWriter *audit.Writer
	auditDir := config.AuditDir(e.Meta.Name, sessionID)

	if e.Audit.Persist {
		if err := os.MkdirAll(auditDir, 0755); err != nil {
			return fmt.Errorf("creating audit dir: %w", err)
		}

		aw := audit.NewWriter(auditDir)
		if err := aw.WriteSessionMeta(audit.SessionMeta{
			ID:           sessionID,
			EnvName:      e.Meta.Name,
			AgentCommand: strings.Join(e.Agent.Command, " "),
			StartedAt:    timeNow(),
		}); err != nil {
			return fmt.Errorf("writing session meta: %w", err)
		}
		auditWriter = aw

		fmt.Fprintf(os.Stderr, "Session: %s\n", sessionID)
		fmt.Fprintf(os.Stderr, "Audit:   %s\n", auditDir)
	}

	args := []string{"run", "--rm"}
	if exitMode {
		args = append(args, "--attach=stdout", "--attach=stderr")
	} else {
		args = append(args, "-it")
	}

	// Mounts from env spec
	for _, m := range e.Agent.Mounts {
		source := m.ResolveSource()
		target := m.ResolveTarget()
		mount := fmt.Sprintf("%s:%s", source, target)
		if !m.Writable {
			mount += ":ro"
		}
		args = append(args, "-v", mount)
	}

	// Mount audit dir as writable
	if e.Audit.Persist {
		args = append(args, "-v", fmt.Sprintf("%s:/aienv/audit:rw", auditDir))
	}

	// Gitconfig for commit authorship
	home, _ := os.UserHomeDir()
	if home != "" {
		gcPath := filepath.Join(home, ".gitconfig")
		if _, err := os.Stat(gcPath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:/home/agent/.gitconfig:ro", gcPath))
		}
	}

	// Agent environment variables
	for key, val := range e.Agent.Env {
		if len(val) > 4 && val[:4] == "env:" {
			envKey := val[4:]
			if os.Getenv(envKey) != "" {
				args = append(args, "-e", envKey)
			}
		} else {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
		}
	}

	// Network proxy setup
	proxyNeeded := e.Permissions != nil && e.Permissions.Network != nil
	auditNeedsProxy := e.Audit.Persist && (e.Permissions == nil || e.Permissions.Network == nil)

	if proxyNeeded || auditNeedsProxy {
		var allowlist, denylist []string
		if e.Permissions != nil && e.Permissions.Network != nil {
			allowlist = e.Permissions.Network.Allow
			denylist = e.Permissions.Network.Deny
		}

		p, port, err := RunProxy(allowlist, denylist, "0.0.0.0", auditWriter)
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

	if !exitMode {
		for _, key := range []string{"TERM", "COLORTERM", "LANG", "LC_ALL", "LC_CTYPE"} {
			if val := os.Getenv(key); val != "" {
				args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
			}
		}
	}
	args = append(args, "-e", "HOME=/home/agent")

	// Entrypoint: tag + command + args
	entrypoint := append([]string{tag}, e.Agent.Command...)
	entrypoint = append(entrypoint, e.Agent.Args...)
	args = append(args, entrypoint...)
	fmt.Printf("%v", args)
	cmd := exec.Command("docker", args...)
	if exitMode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sandbox exited with error: %w", err)
	}
	return nil
}

// timeNow is a variable so we can override in tests.
var timeNow = timeNowFunc

func timeNowFunc() time.Time {
	return time.Now()
}
