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
	Use:   "build [opencode|claude-code]",
	Short: "Build the Docker sandbox image for an agent",
	Long: `Build the Docker sandbox image. If an agent name is provided,
only that agent's image is built. Otherwise, both opencode and claude-code images are built.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return docker.Build(args[0])
		}
		return docker.BuildAll()
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
