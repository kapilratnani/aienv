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
	Short: "Open environment configuration in editor",
	Long: `Opens the environment's ai-env.yaml in your default editor ($EDITOR).
If EDITOR is not set, falls back to vi.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.IsValidName(name); err != nil {
			return err
		}

		yamlPath := config.EnvYAML(name)
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			return fmt.Errorf("environment %q not found at %s", name, yamlPath)
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
		fmt.Printf("  Run 'aienv %s' to use the updated config.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
