package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/docker"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/kapilratnani/aienv/internal/skills"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var activateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an environment",
	Long: `Activates an environment by launching the specified agent in a Docker container.
All execution happens in an isolated Docker sandbox.`,
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

		if err := checkTrust(e, name); err != nil {
			return err
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

		if e.Workdir == "" {
			return fmt.Errorf("workdir is not set for environment %q — recreate with 'aienv create %s' and provide a workdir", name, name)
		}
		workdir := env.ExpandTilde(e.Workdir)
		if info, err := os.Stat(workdir); err != nil || !info.IsDir() {
			return fmt.Errorf("workdir %q does not exist", workdir)
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

		// Generate session ID for this activation
		sessionID := fmt.Sprintf("aienv-%s-%d", e.Name, rand.Uint64())

		// Run the agent in Docker container
		return docker.Run(e, workdir, sessionID)
	},
}

func checkTrust(e *env.Env, name string) error {
	for {
		yamlBytes, err := yaml.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshalling env for trust check: %w", err)
		}
		hash := config.ComputeHash(yamlBytes)

		entry, err := config.ReadTrustCache(hash)
		if err == nil && entry != nil && entry.Status == "trusted" {
			return nil
		}

		fmt.Fprintf(os.Stderr, "\n--- Trust Check for %q ---\n", name)
		showMCPs(e)
		showSkills(e)
		showPermissions(e)

		fmt.Fprint(os.Stderr, "\n  Options: (t)rust, (r)eview, re(j)ect: ")
		input, err := readLine()
		if err != nil {
			return err
		}
		switch strings.TrimSpace(input) {
		case "t", "trust":
			if err := config.WriteTrustCache(hash, name); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: failed to write trust cache: %v\n", err)
			}
			return nil
		case "r", "review":
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			yamlPath := config.EnvYAML(name)
			editorCmd := exec.Command(editor, yamlPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			if err := editorCmd.Run(); err != nil {
				return fmt.Errorf("running editor: %w", err)
			}
			loaded, err := env.Load(name)
			if err != nil {
				return fmt.Errorf("re-loading env after edit: %w", err)
			}
			loaded.Model = e.Model
			*e = *loaded
			continue
		default:
			return fmt.Errorf("trust rejected for environment %q", name)
		}
	}
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

func showMCPs(e *env.Env) {
	if len(e.MCPServers) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "  MCP Servers:\n")
	for name, srv := range e.MCPServers {
		if srv.URL != "" {
			fmt.Fprintf(os.Stderr, "    - %s (remote: %s)\n", name, srv.URL)
		} else {
			fmt.Fprintf(os.Stderr, "    - %s\n", name)
		}
	}
}

func showSkills(e *env.Env) {
	if len(e.Skills) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "  Skills:\n")
	for _, sk := range e.Skills {
		fmt.Fprintf(os.Stderr, "    - %s\n", sk.Name)
	}
}

func showPermissions(e *env.Env) {
	if e.Permissions == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "  Permissions:\n")
	if e.Permissions.Filesystem != nil {
		for pattern, action := range e.Permissions.Filesystem.Read {
			fmt.Fprintf(os.Stderr, "    filesystem.read: %s → %s\n", pattern, action)
		}
		for pattern, action := range e.Permissions.Filesystem.Edit {
			fmt.Fprintf(os.Stderr, "    filesystem.edit: %s → %s\n", pattern, action)
		}
	}
	for pattern, action := range e.Permissions.Bash {
		fmt.Fprintf(os.Stderr, "    bash: %s → %s\n", pattern, action)
	}
	if e.Permissions.Network != nil {
		if len(e.Permissions.Network.Allow) > 0 {
			fmt.Fprintf(os.Stderr, "    network.allow: %s\n", strings.Join(e.Permissions.Network.Allow, ", "))
		}
		if len(e.Permissions.Network.Deny) > 0 {
			fmt.Fprintf(os.Stderr, "    network.deny: %s\n", strings.Join(e.Permissions.Network.Deny, ", "))
		}
	}
}

func init() {
	rootCmd.AddCommand(activateCmd)
}
