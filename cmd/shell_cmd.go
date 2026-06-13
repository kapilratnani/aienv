package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell <name>",
	Short: "Launch interactive shell in sandbox (for debugging)",
	Long: `Builds the environment image and starts an interactive shell inside
the sandbox with the same mounts and permissions as a normal activation.
Useful for debugging agent installation or configuration issues.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, cmdArgs []string) error {
		name := cmdArgs[0]

		if err := config.IsValidName(name); err != nil {
			return fmt.Errorf("invalid environment name %q: %w", name, err)
		}

		e, err := env.Load(name)
		if err != nil {
			return fmt.Errorf("loading environment %q: %w", name, err)
		}

		for i := range e.Agent.Mounts {
			e.Agent.Mounts[i].Source = e.Agent.Mounts[i].ResolveSource()
			if _, err := os.Stat(e.Agent.Mounts[i].Source); err != nil {
				return fmt.Errorf("mount source %q does not exist: %w", e.Agent.Mounts[i].Source, err)
			}
		}

		if err := docker.Check(); err != nil {
			return err
		}

		if err := docker.EnsureImage(e); err != nil {
			return err
		}

		tag := docker.ImageTag(e)

		dockerArgs := []string{
			"run", "--rm", "-it",
			"--entrypoint", "/bin/bash",
		}

		for _, m := range e.Agent.Mounts {
			source := m.ResolveSource()
			target := m.ResolveTarget()
			mount := fmt.Sprintf("%s:%s", source, target)
			if !m.Writable {
				mount += ":ro"
			}
			dockerArgs = append(dockerArgs, "-v", mount)
		}

		home, _ := os.UserHomeDir()
		if home != "" {
			gcPath := filepath.Join(home, ".gitconfig")
			if _, err := os.Stat(gcPath); err == nil {
				dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/home/agent/.gitconfig:ro", gcPath))
			}
		}

		for key, val := range e.Agent.Env {
			if len(val) > 4 && val[:4] == "env:" {
				envKey := val[4:]
				if os.Getenv(envKey) != "" {
					dockerArgs = append(dockerArgs, "-e", envKey)
				}
			} else {
				dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", key, val))
			}
		}

		for _, key := range []string{"TERM", "COLORTERM", "LANG", "LC_ALL", "LC_CTYPE"} {
			if val := os.Getenv(key); val != "" {
				dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", key, val))
			}
		}
		dockerArgs = append(dockerArgs, "-e", "HOME=/home/agent")
		dockerArgs = append(dockerArgs, tag)

		c := exec.Command("docker", dockerArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
