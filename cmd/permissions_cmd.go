package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var permissionsCmd = &cobra.Command{
	Use:   "permissions <name>",
	Short: "Configure permissions for an environment",
	Long: `Interactive wizard to set filesystem, bash, and network permissions
for an existing environment. These are translated to the agent's native
permission format on activation.`,
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

		perms := &env.Permissions{}
		fs := &env.FilesystemPermissions{}
		net := &env.NetworkPermissions{}

		if err := promptFSRead(fs); err != nil {
			return err
		}
		if err := promptFSEdit(fs); err != nil {
			return err
		}
		if len(fs.Read) > 0 || len(fs.Edit) > 0 {
			perms.Filesystem = fs
		}

		if err := promptBash(perms); err != nil {
			return err
		}
		if err := promptNetwork(net); err != nil {
			return err
		}
		autoDetectNetwork(net, e)

		if len(net.Allow) > 0 || len(net.Deny) > 0 {
			perms.Network = net
		}

		if perms.Filesystem == nil && len(perms.Bash) == 0 && perms.Network == nil {
			fmt.Println("  No permissions configured, clearing any existing permissions.")
			e.Permissions = nil
		} else {
			e.Permissions = perms
		}

		if err := e.Save(); err != nil {
			return fmt.Errorf("saving environment: %w", err)
		}

		ag, err := agents.Get(e.Agent)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting cwd: %w", err)
		}

		files, err := ag.GenerateFiles(e, cwd)
		if err != nil {
			return fmt.Errorf("regenerating agent config: %w", err)
		}

		for _, f := range files {
			path := config.AgentConfigPath(name, f.Path)
			if err := os.WriteFile(path, f.Content, 0644); err != nil {
				return fmt.Errorf("writing %s: %w", f.Path, err)
			}
		}

		if err := config.InvalidateTrustCache(name); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to invalidate trust cache: %v\n", err)
		}

		fmt.Printf("✓ Permissions updated for environment %q\n", name)
		return nil
	},
}

func promptFSRead(fs *env.FilesystemPermissions) error {
	fmt.Println("\nFilesystem Read Permissions")
	fmt.Println("--------------------------")
	fmt.Println("  Controls which files/directories the agent can read.")
	fmt.Println("  Patterns are globs (e.g. *, *.env, src/).")
	fmt.Println("  Actions: allow, ask, deny")
	fmt.Println("  Leave pattern blank to finish.")
	for {
		fmt.Print("  Read pattern (e.g. *): ")
		pattern, err := readLine()
		if err != nil {
			return err
		}
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			return nil
		}

		action := promptAction()
		if action == "" {
			continue
		}

		if fs.Read == nil {
			fs.Read = make(map[string]string)
		}
		fs.Read[pattern] = action
		fmt.Printf("    Added: %s → %s\n", pattern, action)
	}
}

func promptFSEdit(fs *env.FilesystemPermissions) error {
	fmt.Println("\nFilesystem Edit Permissions")
	fmt.Println("--------------------------")
	fmt.Println("  Controls which files/directories the agent can modify.")
	fmt.Println("  Leave pattern blank to finish.")
	for {
		fmt.Print("  Edit pattern (e.g. src/): ")
		pattern, err := readLine()
		if err != nil {
			return err
		}
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			return nil
		}

		action := promptAction()
		if action == "" {
			continue
		}

		if fs.Edit == nil {
			fs.Edit = make(map[string]string)
		}
		fs.Edit[pattern] = action
		fmt.Printf("    Added: %s → %s\n", pattern, action)
	}
}

func promptBash(perms *env.Permissions) error {
	fmt.Println("\nBash Command Permissions")
	fmt.Println("-----------------------")
	fmt.Println("  Controls which shell commands the agent can run.")
	fmt.Println("  Patterns are command globs (e.g. *, git *, rm -rf *).")
	fmt.Println("  Leave pattern blank to finish.")
	for {
		fmt.Print("  Command pattern (e.g. git *): ")
		pattern, err := readLine()
		if err != nil {
			return err
		}
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			return nil
		}

		action := promptAction()
		if action == "" {
			continue
		}

		if perms.Bash == nil {
			perms.Bash = make(map[string]string)
		}
		perms.Bash[pattern] = action
		fmt.Printf("    Added: %s → %s\n", pattern, action)
	}
}

func promptAction() string {
	for {
		fmt.Print("  Action (allow/ask/deny): ")
		input, err := readLine()
		if err != nil {
			return ""
		}
		input = strings.TrimSpace(input)
		switch input {
		case "allow", "ask", "deny":
			return input
		default:
			fmt.Printf("    Invalid action %q, use allow, ask, or deny\n", input)
		}
	}
}

func promptNetwork(net *env.NetworkPermissions) error {
	fmt.Println("\nNetwork Permissions")
	fmt.Println("------------------")
	fmt.Println("  Controls external network access (Docker only).")
	fmt.Println("  Comma-separated domain list, or leave blank to skip.")

	fmt.Print("  Allow domains (e.g. api.github.com, *.docker.com): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input != "" {
		for d := range strings.SplitSeq(input, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				net.Allow = append(net.Allow, d)
			}
		}
	}

	fmt.Print("  Deny domains (e.g. *.example.com): ")
	input, err = readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input != "" {
		for d := range strings.SplitSeq(input, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				net.Deny = append(net.Deny, d)
			}
		}
	}

	return nil
}

func autoDetectNetwork(net *env.NetworkPermissions, e *env.Env) {
	var detected []string
	for _, srv := range e.MCPServers {
		if srv.URL != "" {
			host := extractHost(srv.URL)
			if host != "" {
				detected = append(detected, host)
			}
		}
	}

	providerHosts := env.DetectProviderEndpoints(e.Agent)
	detected = append(detected, providerHosts...)

	if len(detected) > 0 {
		fmt.Printf("  Auto-detected endpoints: %s\n", strings.Join(detected, ", "))
		net.Allow = append(net.Allow, detected...)
	}
}

func extractHost(rawURL string) string {
	idx := strings.Index(rawURL, "://")
	if idx < 0 {
		return ""
	}
	rest := rawURL[idx+3:]
	slash := strings.Index(rest, "/")
	if slash >= 0 {
		rest = rest[:slash]
	}
	colon := strings.Index(rest, ":")
	if colon >= 0 {
		rest = rest[:colon]
	}
	return rest
}

func init() {
	rootCmd.AddCommand(permissionsCmd)
}
