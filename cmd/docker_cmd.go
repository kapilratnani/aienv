package cmd

import (
	"fmt"

	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Manage Docker sandbox image",
}

var dockerBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Docker sandbox image",
	RunE: func(cmd *cobra.Command, args []string) error {
		return docker.Build()
	},
}

var dockerCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if Docker is available",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := docker.Check(); err != nil {
			return err
		}
		fmt.Println("Docker is available.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerBuildCmd)
	dockerCmd.AddCommand(dockerCheckCmd)
}
