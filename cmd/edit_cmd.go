package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit environment config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		yamlPath := config.EnvYAML(name)

		if _, err := os.Stat(yamlPath); err != nil {
			return fmt.Errorf("environment %q not found", name)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, yamlPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("running editor: %w", err)
		}

		fmt.Printf("Environment %q updated.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
