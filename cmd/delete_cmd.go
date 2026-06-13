package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		envDir := config.EnvDir(name)

		if _, err := os.Stat(envDir); err != nil {
			return fmt.Errorf("environment %q not found", name)
		}

		// Remove cached Docker image
		yamlPath := config.EnvYAML(name)
		if data, err := os.ReadFile(yamlPath); err == nil {
			hash := config.ComputeHash(data)
			tag := config.ImageTag(hash)
			exec.Command("docker", "rmi", "-f", tag).Run()
		}

		// Remove env directory
		if err := os.RemoveAll(envDir); err != nil {
			return fmt.Errorf("removing environment: %w", err)
		}

		fmt.Printf("Deleted environment %q.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
