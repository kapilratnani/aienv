package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var modelOverride string
var dockerMode bool
var promptOverride string

var rootCmd = &cobra.Command{
	Use:   "aienv",
	Short: "AI environment manager — task-specific MCPs, skills, and rules",
	Long: `aienv manages AI coding agent environments.

An environment encapsulates the MCP servers, agent skills, and rules
needed for a specific task, similar to Python's virtualenv.

  aienv create frontend-design    Create a new environment
  aienv frontend-design           Activate an environment
  aienv list                       List all environments
  aienv show frontend-design      Show environment details
  aienv edit frontend-design      Edit environment config
  aienv delete frontend-design    Delete an environment
  aienv init                       Install shell function
  aienv docker build [agent]       Build Docker sandbox image (or specify agent)
  aienv docker check               Verify Docker availability`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(&modelOverride, "model", "", "Override model for the environment")
	rootCmd.PersistentFlags().BoolVar(&dockerMode, "docker", false, "Run agent in Docker sandbox")
	rootCmd.PersistentFlags().StringVar(&promptOverride, "prompt", "", "Starter prompt for the session (overrides env default)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
