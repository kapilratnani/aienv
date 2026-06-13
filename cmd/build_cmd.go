package cmd

import (
	"fmt"

	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build <name>",
	Short: "Build (or rebuild) an environment's Docker image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		e, err := env.Load(name)
		if err != nil {
			return fmt.Errorf("loading environment %q: %w", name, err)
		}
		return docker.Build(e)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
