package cmd

import (
	"fmt"
	"os"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up <name>",
	Short: "Activate an environment",
	Long: `Builds (if needed) and launches the environment's sandbox container.
The agent runs inside an isolated Docker container with configured mounts
and permissions.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return runEnv(name)
	},
}

// runEnv is shared by "up" subcommand and positional activation.
func runEnv(name string) error {
	if err := config.IsValidName(name); err != nil {
		return fmt.Errorf("invalid environment name %q: %w", name, err)
	}

	e, err := env.Load(name)
	if err != nil {
		return fmt.Errorf("loading environment %q: %w", name, err)
	}

	// Resolve mount sources
	for i := range e.Agent.Mounts {
		e.Agent.Mounts[i].Source = e.Agent.Mounts[i].ResolveSource()
		if _, err := os.Stat(e.Agent.Mounts[i].Source); err != nil {
			return fmt.Errorf("mount source %q does not exist: %w", e.Agent.Mounts[i].Source, err)
		}
	}

	// Trust prompt before activation
	needed, err := docker.NeedsTrustPrompt(e)
	if err != nil {
		return fmt.Errorf("checking trust: %w", err)
	}
	if needed {
		if err := docker.TrustPrompt(e); err != nil {
			return err
		}
		if err := docker.SaveTrust(e); err != nil {
			return fmt.Errorf("saving trust: %w", err)
		}
	}

	return docker.Run(e)
}

// Positional activation: `aienv <name>`
var activateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEnv(args[0])
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(activateCmd)
}
