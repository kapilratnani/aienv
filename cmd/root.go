package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aienv",
	Short: "AI environment manager — sandbox, permissions, and audit for coding agents",
	Long: `aienv manages isolated Docker sandboxes for AI coding agents.

It builds reproducible environments with the agent of your choice,
enforces network permissions, and provides audit trails. Agents are
black boxes — aienv never generates or modifies agent config files.

  aienv create my-env    Create a new environment (interactive)
  aienv my-env           Activate an environment (positional shortcut)
  aienv list              List all environments
  aienv show my-env      Show environment details
  aienv edit my-env      Edit environment config
  aienv delete my-env    Delete an environment
  aienv build my-env     Force rebuild environment image
  aienv shell my-env     Launch interactive shell in sandbox (for debugging)
  aienv clean            Remove orphaned images and audit data`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return runEnv(args[0])
		}
		return cmd.Help()
	},
	SilenceErrors: true,
}

func Execute() {
	args := os.Args[1:]
	if len(args) > 0 {
		_, _, err := rootCmd.Find(args)
		if err != nil {
			if e := runEnv(args[0]); e != nil {
				fmt.Fprintln(os.Stderr, e)
				os.Exit(1)
			}
			return
		}
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
