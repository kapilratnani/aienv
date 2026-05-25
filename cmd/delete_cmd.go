package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.IsValidName(name); err != nil {
			return err
		}

		envDir := config.EnvDir(name)
		if _, err := os.Stat(envDir); os.IsNotExist(err) {
			return fmt.Errorf("environment %q not found", name)
		}

		fmt.Printf("Are you sure you want to delete environment %q? (y/N): ", name)
		input, err := readLine()
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		if input != "y" && input != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}

		if err := os.RemoveAll(envDir); err != nil {
			return fmt.Errorf("deleting environment %q: %w", name, err)
		}

		fmt.Printf("Deleted environment %q.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
