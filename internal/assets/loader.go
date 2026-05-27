package assets

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kapilratnani/aienv/internal/registry"
	"gopkg.in/yaml.v3"
)

var bundledFS fs.FS

func Init(fsys fs.FS) {
	bundledFS = fsys
}

type curatedMCPsFile struct {
	MCPs []curatedMCPEntry `yaml:"mcps"`
}

type curatedMCPEntry struct {
	Name         string              `yaml:"name"`
	DisplayName  string              `yaml:"displayName"`
	Command      string              `yaml:"command"`
	Type         string              `yaml:"type"`
	URL          string              `yaml:"url"`
	Description  string              `yaml:"description"`
	Package      string              `yaml:"package"`
	RegistryType string              `yaml:"registryType"`
	Env          []registry.EnvVarItem `yaml:"env"`
	Categories   []string            `yaml:"categories"`
}

type curatedSkillsFile struct {
	Skills []curatedSkillEntry `yaml:"skills"`
}

type curatedSkillEntry struct {
	Name        string `yaml:"name"`
	Package     string `yaml:"package"`
	Description string `yaml:"description"`
	Installs    int    `yaml:"installs"`
}

func GetCuratedMCPs() ([]registry.MCPServerItem, error) {
	var all []registry.MCPServerItem

	entries, err := loadBundledMCPs()
	if err != nil {
		return nil, fmt.Errorf("loading bundled MCPs: %w", err)
	}
	for _, e := range entries {
		item := entryToMCPServerItem(e)
		item.Source = "aienv"
		all = append(all, item)
	}

	userEntries, err := loadUserMCPs()
	if err != nil {
		return nil, fmt.Errorf("loading user MCP overrides: %w", err)
	}
	for _, e := range userEntries {
		item := entryToMCPServerItem(e)
		item.Source = "user"
		all = append(all, item)
	}

	merged := mergeMCPServerItems(all)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Name < merged[j].Name
	})

	return merged, nil
}

func GetCuratedSkills() ([]registry.SkillItem, error) {
	var all []registry.SkillItem

	entries, err := loadBundledSkills()
	if err != nil {
		return nil, fmt.Errorf("loading bundled skills: %w", err)
	}
	for _, e := range entries {
		item := entryToSkillItem(e)
		item.Source = "aienv"
		all = append(all, item)
	}

	userEntries, err := loadUserSkills()
	if err != nil {
		return nil, fmt.Errorf("loading user skill overrides: %w", err)
	}
	for _, e := range userEntries {
		item := entryToSkillItem(e)
		item.Source = "user"
		all = append(all, item)
	}

	merged := mergeSkillItems(all)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Name < merged[j].Name
	})

	return merged, nil
}

func LookupCuratedMCP(name string) *registry.MCPServerItem {
	all, err := GetCuratedMCPs()
	if err != nil {
		return nil
	}
	for _, m := range all {
		if m.Name == name {
			return &m
		}
	}
	return nil
}

func entryToMCPServerItem(e curatedMCPEntry) registry.MCPServerItem {
	item := registry.MCPServerItem{
		Name:         e.Name,
		DisplayName:  e.DisplayName,
		Description:  e.Description,
		Type:         e.Type,
		Command:      e.Command,
		URL:          e.URL,
		Package:      e.Package,
		RegistryType: e.RegistryType,
		EnvVars:      e.Env,
	}
	if item.DisplayName == "" {
		item.DisplayName = e.Name
	}
	if item.Type == "" {
		item.Type = "local"
	}
	return item
}

func entryToSkillItem(e curatedSkillEntry) registry.SkillItem {
	return registry.SkillItem{
		Name:        e.Name,
		Package:     e.Package,
		Description: e.Description,
		Installs:    e.Installs,
	}
}

func mergeMCPServerItems(items []registry.MCPServerItem) []registry.MCPServerItem {
	seen := make(map[string]int)
	var result []registry.MCPServerItem
	for _, item := range items {
		if idx, ok := seen[item.Name]; ok {
			result[idx] = item
		} else {
			seen[item.Name] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func mergeSkillItems(items []registry.SkillItem) []registry.SkillItem {
	seen := make(map[string]int)
	var result []registry.SkillItem
	for _, item := range items {
		if idx, ok := seen[item.Name]; ok {
			result[idx] = item
		} else {
			seen[item.Name] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func loadBundledMCPs() ([]curatedMCPEntry, error) {
	if bundledFS == nil {
		return nil, nil
	}
	data, err := fs.ReadFile(bundledFS, "curated/mcps.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading bundled mcps.yaml: %w", err)
	}
	var file curatedMCPsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing bundled mcps.yaml: %w", err)
	}
	return file.MCPs, nil
}

func loadBundledSkills() ([]curatedSkillEntry, error) {
	if bundledFS == nil {
		return nil, nil
	}
	data, err := fs.ReadFile(bundledFS, "curated/skills.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading bundled skills.yaml: %w", err)
	}
	var file curatedSkillsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing bundled skills.yaml: %w", err)
	}
	return file.Skills, nil
}

func userCuratedDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aienv", "curated")
}

func loadUserMCPs() ([]curatedMCPEntry, error) {
	dir := userCuratedDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var all []curatedMCPEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var file curatedMCPsFile
		if err := yaml.Unmarshal(data, &file); err != nil {
			continue
		}
		all = append(all, file.MCPs...)
	}
	return all, nil
}

func loadUserSkills() ([]curatedSkillEntry, error) {
	dir := userCuratedDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var all []curatedSkillEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var file curatedSkillsFile
		if err := yaml.Unmarshal(data, &file); err != nil {
			continue
		}
		all = append(all, file.Skills...)
	}
	return all, nil
}
