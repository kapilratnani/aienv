package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var (
	worktreeFlag     string
	worktreeBaseFlag string
	worktreeKeepFlag bool
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

	// Worktree setup
	if worktreeFlag != "" {
		if len(e.Agent.Mounts) == 0 {
			return fmt.Errorf("cannot create worktree: no workdir mount configured")
		}

		mounts, cleanup, err := env.SetupWorktree(&env.WorktreeConfig{
			RepoPath:   e.Agent.Mounts[0].Source,
			Branch:     worktreeFlag,
			BaseBranch: worktreeBaseFlag,
			Keep:       worktreeKeepFlag,
		})
		if err != nil {
			return err
		}

		e.Agent.Mounts[0].Source = mounts[0].Source
		e.Agent.Mounts[0].Writable = mounts[0].Writable
		e.Agent.Mounts = append(e.Agent.Mounts, mounts[1])

		if cleanup != nil {
			defer cleanup()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				cleanup()
				os.Exit(1)
			}()
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

	upCmd.Flags().StringVarP(&worktreeFlag, "worktree", "w", "", "Create a git worktree for the given branch and activate into it")
	upCmd.Flags().StringVar(&worktreeBaseFlag, "worktree-base", "", "Base branch for the worktree (default: auto-detect from remote)")
	upCmd.Flags().BoolVar(&worktreeKeepFlag, "worktree-keep", false, "Keep worktree after session exit")
}
