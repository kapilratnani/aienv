package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
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

		missing := skills.VerifyAll(e.Skills, e.Agent)
		if len(missing) > 0 {
			fmt.Fprintf(os.Stderr, "Installing missing skills...\n")
			if err := skills.InstallAll(missing, e.Agent); err != nil {
				return fmt.Errorf("installing skills: %w", err)
			}
		}

		checkMCPEnvVars(e)

		promptText := promptOverride
		if promptText == "" {
			promptText = e.Prompt
		}
		if promptText != "" {
			promptPath := filepath.Join(config.EnvDir(name), "starter-prompt.md")
			if err := os.WriteFile(promptPath, []byte(promptText+"\n"), 0644); err != nil {
				return fmt.Errorf("writing starter prompt: %w", err)
			}
			e.Rules = append([]env.Rule{{Path: promptPath}}, e.Rules...)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting cwd: %w", err)
		}

		workdir := cwd
		if e.Workdir != "" {
			wd := env.ExpandTilde(e.Workdir)
			if info, err := os.Stat(wd); err == nil && info.IsDir() {
				workdir = wd
			} else {
				fmt.Fprintf(os.Stderr, "  Warning: workdir %q not found, using current directory\n", e.Workdir)
			}
		}

		ag, err := agents.Get(e.Agent)
		if err != nil {
			return err
		}

		files, err := ag.GenerateFiles(e, workdir)
		if err != nil {
			return fmt.Errorf("generating agent config: %w", err)
		}

		for _, f := range files {
			path := config.AgentConfigPath(name, f.Path)
			if err := os.WriteFile(path, f.Content, 0644); err != nil {
				return fmt.Errorf("writing %s: %w", f.Path, err)
			}
		}

		if dockerMode {
			return docker.Run(e, workdir)
		}

		activateCmd := ag.ActivateCommand(config.EnvDir(name), e)
		if workdir != cwd {
			activateCmd = fmt.Sprintf("cd %s\n%s", workdir, activateCmd)
		}
		fmt.Print(activateCmd)

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
