package cmd

import (
	"fmt"
	"strings"

	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show environment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		e, err := env.Load(name)
		if err != nil {
			return err
		}

		fmt.Printf("Name:        %s\n", e.Meta.Name)
		if e.Meta.Description != "" {
			fmt.Printf("Description: %s\n", e.Meta.Description)
		}
		fmt.Printf("Agent:       %s\n", strings.Join(e.Agent.Command, " "))
		if len(e.Agent.Args) > 0 {
			fmt.Printf("Args:        %s\n", strings.Join(e.Agent.Args, " "))
		}
		fmt.Printf("Mounts:\n")
		for _, m := range e.Agent.Mounts {
			rw := "ro"
			if m.Writable {
				rw = "rw"
			}
			fmt.Printf("  - %s → %s (%s)\n", m.Source, m.Target, rw)
		}
		if len(e.Deps.Packages) > 0 {
			fmt.Printf("Packages:    %s\n", strings.Join(e.Deps.Packages, ", "))
		}
		if len(e.Deps.Custom) > 0 {
			fmt.Printf("Custom deps: %s\n", strings.Join(e.Deps.Custom, ", "))
		}
		if e.Permissions != nil && e.Permissions.Network != nil {
			fmt.Printf("Network:\n")
			for _, h := range e.Permissions.Network.Allow {
				fmt.Printf("  allow: %s\n", h)
			}
			for _, h := range e.Permissions.Network.Deny {
				fmt.Printf("  deny: %s\n", h)
			}
		}
		if e.Audit.Persist {
			fmt.Printf("Audit: enabled (%v)\n", e.Audit.Capture)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
