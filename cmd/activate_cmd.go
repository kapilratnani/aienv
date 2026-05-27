package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/kapilratnani/aienv/internal/opencode"
	"github.com/kapilratnani/aienv/internal/skills"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Print shell commands to activate an environment",
	Long: `Prints shell commands that set up the environment for the current session.
Used by the 'aienv' shell function with 'eval':

  eval "$(aienv activate frontend-design)"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.IsValidName(name); err != nil {
			return err
		}

		e, err := env.Load(name)
		if err != nil {
			return fmt.Errorf("loading environment %q: %w", name, err)
		}

		if modelOverride != "" {
			e.Model = modelOverride
		}

		missing := skills.VerifyAll(e.Skills)
		if len(missing) > 0 {
			fmt.Fprintf(os.Stderr, "Installing missing skills...\n")
			if err := skills.InstallAll(missing); err != nil {
				return fmt.Errorf("installing skills: %w", err)
			}
		}

		checkMCPEnvVars(e)

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting cwd: %w", err)
		}

		cfg, err := opencode.Generate(e, cwd)
		if err != nil {
			return fmt.Errorf("generating opencode config: %w", err)
		}

		ocPath := config.OpenCodeJSON(name)
		if err := os.WriteFile(ocPath, cfg, 0644); err != nil {
			return fmt.Errorf("writing opencode config: %w", err)
		}

		if dockerMode {
			return docker.Run(e, cwd)
		}

		absPath, _ := filepath.Abs(ocPath)

		shell := filepath.Base(os.Getenv("SHELL"))

		switch shell {
		case "fish":
			fmt.Printf("set -x OPENCODE_CONFIG %s;\n", absPath)
			fmt.Println("opencode")
			fmt.Printf("set -e OPENCODE_CONFIG;\n")
		default:
			fmt.Printf("export OPENCODE_CONFIG=%s\n", absPath)
			fmt.Println("opencode")
			fmt.Println("unset OPENCODE_CONFIG")
		}

		return nil
	},
}

func checkMCPEnvVars(e *env.Env) {
	for name, srv := range e.MCPServers {
		if len(srv.Env) == 0 {
			continue
		}
		var missingVars []string
		for key, val := range srv.Env {
			if len(val) > 4 && val[:4] == "env:" {
				envKey := val[4:]
				if os.Getenv(envKey) == "" {
					missingVars = append(missingVars, envKey)
				}
			} else if os.Getenv(key) == "" {
				missingVars = append(missingVars, key)
			}
		}
		if len(missingVars) > 0 {
			fmt.Fprintf(os.Stderr, "  Warning: MCP server %q may not work — set %s\n",
				name, strings.Join(missingVars, ", "))
		}
	}
}

func init() {
	rootCmd.AddCommand(activateCmd)
}
