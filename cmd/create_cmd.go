package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/spf13/cobra"
)

var inputReader = bufio.NewScanner(os.Stdin)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.IsValidName(name); err != nil {
			return err
		}

		if _, err := env.Load(name); err == nil {
			return fmt.Errorf("environment %q already exists", name)
		}

		e := &env.Env{}
		e.Meta.Name = name

		if err := promptAgentInstall(e); err != nil {
			return err
		}
		if err := promptAgentCommand(e); err != nil {
			return err
		}
		if err := promptDescription(e); err != nil {
			return err
		}
		if err := promptWorkdir(e); err != nil {
			return err
		}
		if err := promptAdditionalMounts(e); err != nil {
			return err
		}
		if err := promptDeps(e); err != nil {
			return err
		}
		if err := promptNetworkPerms(e); err != nil {
			return err
		}
		if err := promptAudit(e); err != nil {
			return err
		}
		if err := promptConfirm(e); err != nil {
			return err
		}

		if err := e.Save(); err != nil {
			return fmt.Errorf("saving environment: %w", err)
		}

		fmt.Printf("\nCreated environment %q\n", name)
		fmt.Printf("  Config: %s\n", config.EnvYAML(name))
		fmt.Printf("  Run:    aienv %s\n", name)
		return nil
	},
}

func promptAgentInstall(e *env.Env) error {
	fmt.Print("Agent install command (e.g. 'npm install -g opencode-ai'): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input != "" {
		e.Agent.Install = []string{input}
	}
	fmt.Print("Additional install commands? (comma-separated, or blank to skip): ")
	input, err = readLine()
	if err != nil {
		return err
	}
	for _, cmd := range strings.Split(strings.TrimSpace(input), ",") {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			e.Agent.Install = append(e.Agent.Install, cmd)
		}
	}
	return nil
}

func promptAgentCommand(e *env.Env) error {
	fmt.Print("Agent command (e.g. 'opencode'): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("agent command is required")
	}
	e.Agent.Command = strings.Fields(input)

	fmt.Print("Default arguments (comma-separated, e.g. --model claude-sonnet-4-5, or blank): ")
	input, err = readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input != "" {
		for _, arg := range strings.Split(input, ",") {
			arg = strings.TrimSpace(arg)
			if arg != "" {
				e.Agent.Args = append(e.Agent.Args, strings.Fields(arg)...)
			}
		}
	}
	return nil
}

func promptDescription(e *env.Env) error {
	fmt.Print("Description (optional): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	e.Meta.Description = strings.TrimSpace(input)
	return nil
}

func promptWorkdir(e *env.Env) error {
	for {
		fmt.Print("Working directory (absolute path, or '.' for current): ")
		input, err := readLine()
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("  Workdir is required.")
			continue
		}
		expanded := env.ExpandTilde(input)
		absPath, err := filepath.Abs(expanded)
		if err != nil {
			fmt.Printf("  Error resolving path: %v\n", err)
			continue
		}
		if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
			fmt.Printf("  Directory %q does not exist.\n", absPath)
			continue
		}
		e.Agent.Mounts = append(e.Agent.Mounts, env.Mount{
			Source: absPath,
			Target: "/workspace",
		})
		return nil
	}
}

func promptAdditionalMounts(e *env.Env) error {
	for {
		fmt.Println("\nAdditional mounts (config files, skill dirs, etc.)")
		fmt.Println("  Currently configured mounts:")
		for i, m := range e.Agent.Mounts {
			rw := "ro"
			if m.Writable {
				rw = "rw"
			}
			fmt.Printf("    %d. %s → %s (%s)\n", i+1, m.Source, m.Target, rw)
		}
		fmt.Print("  Add mount? (source:target[:rw] or blank to skip): ")
		input, err := readLine()
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return nil
		}
		parts := strings.Split(input, ":")
		if len(parts) < 2 {
			fmt.Println("  Format: source:target[:rw]")
			continue
		}
		m := env.Mount{
			Source: parts[0],
			Target: parts[1],
		}
		if len(parts) >= 3 && parts[2] == "rw" {
			m.Writable = true
		}
		e.Agent.Mounts = append(e.Agent.Mounts, m)
		fmt.Printf("  Added: %s → %s\n", m.Source, m.Target)
	}
}

func promptDeps(e *env.Env) error {
	fmt.Print("System packages to install (comma-separated, e.g. golang,nodejs,python3): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	for _, pkg := range strings.Split(strings.TrimSpace(input), ",") {
		pkg = strings.TrimSpace(pkg)
		if pkg != "" {
			e.Deps.Packages = append(e.Deps.Packages, pkg)
		}
	}
	fmt.Print("Custom install commands (comma-separated, e.g. go install foo/bar): ")
	input, err = readLine()
	if err != nil {
		return err
	}
	for _, cmd := range strings.Split(strings.TrimSpace(input), ",") {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			e.Deps.Custom = append(e.Deps.Custom, cmd)
		}
	}
	return nil
}

func promptNetworkPerms(e *env.Env) error {
	fmt.Print("\nNetwork permissions (optional)\n")
	fmt.Print("Allow domains (comma-separated, e.g. api.github.com,*.anthropic.com, or blank for learn mode): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	netPerms := &env.NetworkPermissions{}
	for _, d := range strings.Split(input, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			netPerms.Allow = append(netPerms.Allow, d)
		}
	}

	if e.Permissions == nil {
		e.Permissions = &env.Permissions{}
	}
	e.Permissions.Network = netPerms
	return nil
}

func promptAudit(e *env.Env) error {
	fmt.Print("\nEnable audit logging? Network requests will be recorded. (y/N): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "y" || input == "Y" {
		e.Audit.Persist = true
		e.Audit.Capture = []string{"network"}
	}
	return nil
}

func promptConfirm(e *env.Env) error {
	fmt.Println("\n--- Summary ---")
	fmt.Printf("  Name:        %s\n", e.Meta.Name)
	if e.Meta.Description != "" {
		fmt.Printf("  Description: %s\n", e.Meta.Description)
	}
	fmt.Printf("  Agent:       %s\n", strings.Join(e.Agent.Command, " "))
	if len(e.Agent.Args) > 0 {
		fmt.Printf("  Args:        %s\n", strings.Join(e.Agent.Args, " "))
	}
	for i, m := range e.Agent.Mounts {
		rw := "ro"
		if m.Writable {
			rw = "rw"
		}
		if i == 0 {
			fmt.Printf("  Mounts:      %s → %s (%s)\n", m.Source, m.Target, rw)
		} else {
			fmt.Printf("               %s → %s (%s)\n", m.Source, m.Target, rw)
		}
	}
	if e.Permissions != nil && e.Permissions.Network != nil && len(e.Permissions.Network.Allow) > 0 {
		fmt.Printf("  Network:     allow: %s\n", strings.Join(e.Permissions.Network.Allow, ", "))
	}
	if e.Audit.Persist {
		fmt.Printf("  Audit:      enabled\n")
	}
	fmt.Print("\nCreate this environment? (Y/n): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "n" || input == "N" {
		return fmt.Errorf("cancelled")
	}
	return nil
}

func readLine() (string, error) {
	if inputReader.Scan() {
		return inputReader.Text(), inputReader.Err()
	}
	return "", inputReader.Err()
}

func init() {
	rootCmd.AddCommand(createCmd)
}
