package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/assets"
	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
	"github.com/kapilratnani/aienv/internal/registry"
	"github.com/kapilratnani/aienv/internal/skills"
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

		e := &env.Env{
			Name: name,
		}

		if err := promptAgent(e); err != nil {
			return err
		}
		if err := promptDescription(e); err != nil {
			return err
		}
		if err := promptModel(e); err != nil {
			return err
		}
		if err := promptMCPServers(e); err != nil {
			return err
		}
		if err := promptSkills(e); err != nil {
			return err
		}
		if err := promptRules(e); err != nil {
			return err
		}
		if err := promptPrompt(e); err != nil {
			return err
		}
		if err := promptWorkdir(e); err != nil {
			return err
		}
		if err := promptConfirm(e); err != nil {
			return err
		}

		if err := e.Save(); err != nil {
			return fmt.Errorf("saving environment: %w", err)
		}

		missing := skills.VerifyAll(e.Skills, e.Agent)
		if len(missing) > 0 {
			fmt.Println("Installing skills...")
			if err := skills.InstallAll(missing, e.Agent); err != nil {
				return err
			}
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
			return fmt.Errorf("generating agent config: %w", err)
		}

		for _, f := range files {
			path := config.AgentConfigPath(name, f.Path)
			if err := os.WriteFile(path, f.Content, 0644); err != nil {
				return fmt.Errorf("writing %s: %w", f.Path, err)
			}
		}

		fmt.Printf("\n✓ Created environment %q\n", name)
		fmt.Printf("  Config:  %s\n", config.EnvYAML(name))
		fmt.Printf("  Run:     aienv %s\n", name)
		return nil
	},
}

func promptAgent(e *env.Env) error {
	fmt.Print("Agent (opencode/claude-code) [default: opencode]: ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		e.Agent = "opencode"
	} else {
		switch input {
		case "opencode", "claude-code":
			e.Agent = input
		default:
			fmt.Printf("  Unknown agent %q, defaulting to opencode\n", input)
			e.Agent = "opencode"
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
	e.Description = strings.TrimSpace(input)
	return nil
}

func promptModel(e *env.Env) error {
	fmt.Print("Default model (optional, e.g. claude-sonnet-4-5): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	e.Model = strings.TrimSpace(input)
	return nil
}

func promptMCPServers(e *env.Env) error {
	curatedMCPs, err := assets.GetCuratedMCPs()
	if err != nil {
		return fmt.Errorf("loading curated MCPs: %w", err)
	}

	for {
		fmt.Println("\nMCP Servers")
		fmt.Println("-----------")
		mcpList := sortedMCPKeys(e.MCPServers)
		if len(mcpList) == 0 {
			fmt.Println("  Currently selected: (none)")
		} else {
			fmt.Printf("  Currently selected: %s\n", strings.Join(mcpList, ", "))
		}

		fmt.Println("\n  Popular MCP servers:")
		for i, mcp := range curatedMCPs {
			sourceTag := ""
			if mcp.Source == "user" {
				sourceTag = " (user override)"
			}
			fmt.Printf("    %2d. %-20s %s%s\n", i+1, mcp.DisplayName, mcp.Description, sourceTag)
		}
		fmt.Println("  s.  Search online  |  c.  Custom entry  |  r.  Remove  |  d.  Done")
		fmt.Print("  Choice (e.g. 1,3, s, c, r, d): ")

		input, err := readLine()
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch input {
		case "d":
			return nil
		case "s":
			searchMCPServers(e)
		case "c":
			promptCustomMCP(e)
		case "r":
			promptRemoveMCP(e)
		default:
			for part := range strings.SplitSeq(input, ",") {
				part = strings.TrimSpace(part)
				idx := 0
				if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(curatedMCPs) {
					fmt.Printf("    Invalid: %s\n", part)
					continue
				}
				addCuratedMCP(e, curatedMCPs[idx-1])
			}
		}
	}
}

func addCuratedMCP(e *env.Env, item registry.MCPServerItem) {
	if e.MCPServers == nil {
		e.MCPServers = make(map[string]env.MCPServer)
	}
	name := item.Name
	if _, exists := e.MCPServers[name]; exists {
		fmt.Printf("    %s already selected, skipping.\n", name)
		return
	}
	srv := env.MCPServer{
		Type: item.Type,
	}
	if item.Type == "local" || item.Type == "" {
		srv.Command = splitCommand(item.Command)
	} else {
		srv.URL = item.URL
	}
	if len(item.EnvVars) > 0 {
		srv.Env = make(map[string]string, len(item.EnvVars))
		for _, ev := range item.EnvVars {
			srv.Env[ev.Key] = fmt.Sprintf("env:%s", ev.Key)
		}
	}
	e.MCPServers[name] = srv
	fmt.Printf("    Added %s\n", name)

	if len(item.EnvVars) > 0 {
		fmt.Printf("    Note: %s needs environment variables:\n", item.DisplayName)
		for _, ev := range item.EnvVars {
			req := ""
			if ev.Required {
				req = " (required)"
			}
			fmt.Printf("      - %s: %s%s\n", ev.Key, ev.Description, req)
		}
	}
}

func searchMCPServers(e *env.Env) error {
	fmt.Print("  Search MCP servers: ")
	query, err := readLine()
	if err != nil {
		return err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	fmt.Print("  Searching registry.modelcontextprotocol.io... ")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reg := registry.NewOfficialRegistry()
	results, err := reg.Search(ctx, query, 10)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil
	}
	fmt.Printf("found %d results\n", len(results))

	if len(results) == 0 {
		fmt.Println("  No results found.")
		return nil
	}

	fmt.Println("\n  Search results:")
	for i, item := range results {
		desc := item.Description
		if desc == "" {
			desc = "(no description)"
		}
		displayName := item.DisplayName
		if displayName == "" {
			displayName = item.Name
		}
		fmt.Printf("    %2d. %-20s %s\n", i+1, displayName, desc)
	}
	fmt.Print("  Select (e.g. 1,3, or blank to skip): ")

	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	for part := range strings.SplitSeq(input, ",") {
		part = strings.TrimSpace(part)
		idx := 0
		if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(results) {
			fmt.Printf("    Invalid: %s\n", part)
			continue
		}
		item := results[idx-1]

		if curated := assets.LookupCuratedMCP(item.Name); curated != nil {
			item.EnvVars = curated.EnvVars
			if item.DisplayName == "" {
				item.DisplayName = curated.DisplayName
			}
		}
		addCuratedMCP(e, item)
	}
	return nil
}

func promptCustomMCP(e *env.Env) error {
	fmt.Print("  Name: ")
	name, err := readLine()
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	fmt.Print("  Type (local/remote): ")
	typ, err := readLine()
	if err != nil {
		return err
	}
	typ = strings.TrimSpace(typ)

	var srv env.MCPServer
	switch typ {
	case "local":
		fmt.Print("  Command (e.g. npx -y @package/server): ")
		cmdLine, err := readLine()
		if err != nil {
			return err
		}
		srv = env.MCPServer{Type: "local", Command: splitCommand(strings.TrimSpace(cmdLine))}
	case "remote":
		fmt.Print("  URL (e.g. https://mcp.example.com/v1): ")
		u, err := readLine()
		if err != nil {
			return err
		}
		srv = env.MCPServer{Type: "remote", URL: strings.TrimSpace(u)}
	default:
		fmt.Println("  Invalid type, skipping.")
		return nil
	}

	if e.MCPServers == nil {
		e.MCPServers = make(map[string]env.MCPServer)
	}
	if _, exists := e.MCPServers[name]; exists {
		fmt.Printf("  %s already selected, skipping.\n", name)
		return nil
	}
	e.MCPServers[name] = srv
	fmt.Printf("  Added %s\n", name)
	return nil
}

func promptRemoveMCP(e *env.Env) error {
	if len(e.MCPServers) == 0 {
		fmt.Println("  Nothing to remove.")
		return nil
	}
	mcpList := sortedMCPKeys(e.MCPServers)
	fmt.Println("  Remove which MCP?")
	for i, name := range mcpList {
		fmt.Printf("    %d. %s\n", i+1, name)
	}
	fmt.Print("  Number (blank to cancel): ")

	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	idx := 0
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(mcpList) {
		fmt.Println("  Invalid, cancelling.")
		return nil
	}
	delete(e.MCPServers, mcpList[idx-1])
	fmt.Printf("  Removed.\n")
	return nil
}

func promptSkills(e *env.Env) error {
	curatedSkills, err := assets.GetCuratedSkills()
	if err != nil {
		return fmt.Errorf("loading curated skills: %w", err)
	}

	for {
		fmt.Println("\nAgent Skills")
		fmt.Println("-----------")
		if len(e.Skills) == 0 {
			fmt.Println("  Currently selected: (none)")
		} else {
			names := make([]string, len(e.Skills))
			for i, sk := range e.Skills {
				names[i] = sk.Name
			}
			fmt.Printf("  Currently selected: %s\n", strings.Join(names, ", "))
		}

		fmt.Println("\n  Popular skills:")
		for i, sk := range curatedSkills {
			sourceTag := ""
			if sk.Source == "user" {
				sourceTag = " (user override)"
			}
			fmt.Printf("    %2d. %-35s %s%s\n", i+1, sk.Name, sk.Package, sourceTag)
		}
		fmt.Println("  s.  Search online  |  c.  Custom entry  |  r.  Remove  |  d.  Done")
		fmt.Print("  Choice (e.g. 1,3, s, c, r, d): ")

		input, err := readLine()
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch input {
		case "d":
			return nil
		case "s":
			searchSkills(e)
		case "c":
			promptCustomSkill(e)
		case "r":
			promptRemoveSkill(e)
		default:
			for part := range strings.SplitSeq(input, ",") {
				part = strings.TrimSpace(part)
				idx := 0
				if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(curatedSkills) {
					fmt.Printf("    Invalid: %s\n", part)
					continue
				}
				item := curatedSkills[idx-1]
				if hasSkill(e, item.Name) {
					fmt.Printf("    %s already selected, skipping.\n", item.Name)
					continue
				}
				e.Skills = append(e.Skills, env.Skill{
					Name:    item.Name,
					Source:  "registry",
					Package: item.Package,
				})
				fmt.Printf("    Added %s\n", item.Name)
			}
		}
	}
}

func hasSkill(e *env.Env, name string) bool {
	for _, s := range e.Skills {
		if s.Name == name {
			return true
		}
	}
	return false
}

func searchSkills(e *env.Env) error {
	fmt.Print("  Search skills: ")
	query, err := readLine()
	if err != nil {
		return err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	fmt.Print("  Searching skills.sh... ")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sks := registry.NewSkillsDotSh()
	results, err := sks.Search(ctx, query, 10)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil
	}
	fmt.Printf("found %d results\n", len(results))

	if len(results) == 0 {
		fmt.Println("  No results found.")
		return nil
	}

	fmt.Println("\n  Search results:")
	for i, item := range results {
		fmt.Printf("    %2d. %-35s %s (installs: %d)\n", i+1, item.Name, item.Package, item.Installs)
	}
	fmt.Print("  Select (e.g. 1,3, or blank to skip): ")

	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	for part := range strings.SplitSeq(input, ",") {
		part = strings.TrimSpace(part)
		idx := 0
		if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(results) {
			fmt.Printf("    Invalid: %s\n", part)
			continue
		}
		item := results[idx-1]
		if hasSkill(e, item.Name) {
			fmt.Printf("    %s already selected, skipping.\n", item.Name)
			continue
		}
		e.Skills = append(e.Skills, env.Skill{
			Name:    item.Name,
			Source:  "registry",
			Package: item.Package,
		})
		fmt.Printf("    Added %s\n", item.Name)
	}
	return nil
}

func promptCustomSkill(e *env.Env) error {
	fmt.Print("  Skill name: ")
	name, err := readLine()
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	fmt.Print("  Package (GitHub repo, e.g. owner/repo): ")
	pkg, err := readLine()
	if err != nil {
		return err
	}
	pkg = strings.TrimSpace(pkg)

	if hasSkill(e, name) {
		fmt.Printf("  %s already selected, skipping.\n", name)
		return nil
	}
	e.Skills = append(e.Skills, env.Skill{
		Name:    name,
		Source:  "registry",
		Package: pkg,
	})
	fmt.Printf("  Added %s\n", name)
	return nil
}

func promptRemoveSkill(e *env.Env) error {
	if len(e.Skills) == 0 {
		fmt.Println("  Nothing to remove.")
		return nil
	}
	fmt.Println("  Remove which skill?")
	for i, sk := range e.Skills {
		fmt.Printf("    %d. %s\n", i+1, sk.Name)
	}
	fmt.Print("  Number (blank to cancel): ")

	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	idx := 0
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(e.Skills) {
		fmt.Println("  Invalid, cancelling.")
		return nil
	}
	i := idx - 1
	e.Skills = append(e.Skills[:i], e.Skills[i+1:]...)
	fmt.Printf("  Removed.\n")
	return nil
}

func promptRules(e *env.Env) error {
	fmt.Print("\nRule files (comma-separated paths, or 'n' for none): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" || input == "n" {
		return nil
	}
	for p := range strings.SplitSeq(input, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			e.Rules = append(e.Rules, env.Rule{Path: p})
		}
	}
	return nil
}

func promptPrompt(e *env.Env) error {
	fmt.Print("Default starter prompt (optional, e.g. 'Use caveman mode'): ")
	input, err := readLine()
	if err != nil {
		return err
	}
	e.Prompt = strings.TrimSpace(input)
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
			fmt.Println("  Workdir is required. Enter '.' to use the current directory.")
			continue
		}

		expanded := env.ExpandTilde(input)
		absPath, err := filepath.Abs(expanded)
		if err != nil {
			fmt.Printf("  Error resolving path: %v\n", err)
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Printf("  Directory %q does not exist.\n", absPath)
			continue
		}
		if !info.IsDir() {
			fmt.Printf("  %q is not a directory.\n", absPath)
			continue
		}

		e.Workdir = absPath
		return nil
	}
}

func promptConfirm(e *env.Env) error {
	fmt.Println("\n--- Summary ---")
	fmt.Printf("  Name:        %s\n", e.Name)
	fmt.Printf("  Agent:       %s\n", e.Agent)
	if e.Model != "" {
		fmt.Printf("  Model:       %s\n", e.Model)
	}
	if e.Description != "" {
		fmt.Printf("  Description: %s\n", e.Description)
	}
	if e.Prompt != "" {
		fmt.Printf("  Prompt:      %s\n", e.Prompt)
	}
	fmt.Printf("  Workdir:     %s\n", e.Workdir)
	mcpNames := sortedMCPKeys(e.MCPServers)
	fmt.Printf("  MCPs:        %s\n", joinOrNone(mcpNames))
	skillNames := make([]string, len(e.Skills))
	for i, s := range e.Skills {
		skillNames[i] = s.Name
	}
	fmt.Printf("  Skills:      %s\n", joinOrNone(skillNames))
	rulePaths := make([]string, len(e.Rules))
	for i, r := range e.Rules {
		rulePaths[i] = r.Path
	}
	fmt.Printf("  Rules:       %s\n", joinOrNone(rulePaths))

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

func sortedMCPKeys(m map[string]env.MCPServer) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func readLine() (string, error) {
	if inputReader.Scan() {
		return inputReader.Text(), inputReader.Err()
	}
	return "", inputReader.Err()
}

func splitCommand(cmd string) []string {
	if cmd == "" {
		return nil
	}
	return strings.Fields(cmd)
}

func init() {
	rootCmd.AddCommand(createCmd)
}
