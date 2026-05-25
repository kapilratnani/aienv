package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		envsDir := config.AIEnvsDir()
		entries, err := os.ReadDir(envsDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No environments found. Run 'aienv create <name>'.")
				return nil
			}
			return fmt.Errorf("reading %s: %w", envsDir, err)
		}

		found := false
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			yamlPath := filepath.Join(envsDir, name, "ai-env.yaml")
			if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
				continue
			}

			e, err := env.Load(name)
			if err != nil {
				fmt.Printf("  %-25s (error: %v)\n", name, err)
				continue
			}

			desc := e.Description
			if desc == "" {
				desc = "(no description)"
			}
			fmt.Printf("  %-25s %s\n", name, desc)
			found = true
		}

		if !found {
			fmt.Println("No environments found. Run 'aienv create <name>'.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
