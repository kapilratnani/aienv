package cmd

import (
	"fmt"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := config.ListEnvNames()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Println("No environments found.")
			return nil
		}
		for _, name := range names {
			e, err := env.Load(name)
			if err != nil {
				fmt.Printf("  %s (error loading: %v)\n", name, err)
				continue
			}
			desc := e.Meta.Description
			if desc != "" {
				desc = " — " + desc
			}
			fmt.Printf("  %s%s\n", name, desc)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
