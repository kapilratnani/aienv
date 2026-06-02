package cmd

import (
	"fmt"
	"strings"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show environment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.IsValidName(name); err != nil {
			return err
		}

		e, err := env.Load(name)
		if err != nil {
			return fmt.Errorf("loading environment %q: %w", name, err)
		}

		mcpNames := make([]string, 0, len(e.MCPServers))
		for name := range e.MCPServers {
			mcpNames = append(mcpNames, name)
		}

		skillNames := make([]string, len(e.Skills))
		for i, s := range e.Skills {
			skillNames[i] = s.Name
		}

		rules := make([]string, len(e.Rules))
		for i, r := range e.Rules {
			rules[i] = r.Path
		}

		fmt.Printf("Name:        %s\n", e.Name)
		fmt.Printf("Agent:       %s\n", e.Agent)
		if e.Model != "" {
			fmt.Printf("Model:       %s\n", e.Model)
		}
		if e.Description != "" {
			fmt.Printf("Description: %s\n", e.Description)
		}
		if e.Workdir != "" {
			fmt.Printf("Workdir:     %s\n", e.Workdir)
		}
		fmt.Printf("MCPs:        %s\n", joinOrNone(mcpNames))
		fmt.Printf("Skills:      %s\n", joinOrNone(skillNames))
		fmt.Printf("Rules:       %s\n", joinOrNone(rules))
		return nil
	},
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return strings.Join(items, ", ")
}

func init() {
	rootCmd.AddCommand(showCmd)
}
