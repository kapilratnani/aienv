package assets

import (
	"testing"

	"github.com/kapilratnani/aienv/internal/registry"
)

func TestEntryToMCPServerItem(t *testing.T) {
	t.Run("empty display name defaults to name", func(t *testing.T) {
		e := curatedMCPEntry{Name: "test-server", Type: "local"}
		item := entryToMCPServerItem(e)
		if item.DisplayName != "test-server" {
			t.Errorf("DisplayName = %q, want %q", item.DisplayName, "test-server")
		}
	})

	t.Run("empty type defaults to local", func(t *testing.T) {
		e := curatedMCPEntry{Name: "test-server"}
		item := entryToMCPServerItem(e)
		if item.Type != "local" {
			t.Errorf("Type = %q, want %q", item.Type, "local")
		}
	})

	t.Run("preserves all fields when set", func(t *testing.T) {
		e := curatedMCPEntry{
			Name:         "gh",
			DisplayName:  "GitHub",
			Command:      "npx -y @github/mcp-server",
			Type:         "local",
			Description:  "GitHub API",
			Package:      "@github/mcp-server",
			RegistryType: "npm",
			Env: []registry.EnvVarItem{
				{Key: "GITHUB_TOKEN", Description: "GitHub PAT", Required: true},
			},
		}
		item := entryToMCPServerItem(e)
		if item.DisplayName != "GitHub" {
			t.Errorf("DisplayName = %q, want %q", item.DisplayName, "GitHub")
		}
		if item.Type != "local" {
			t.Errorf("Type = %q, want %q", item.Type, "local")
		}
		if len(item.EnvVars) != 1 || item.EnvVars[0].Key != "GITHUB_TOKEN" {
			t.Error("EnvVars not preserved")
		}
	})
}

func TestEntryToSkillItem(t *testing.T) {
	e := curatedSkillEntry{
		Name:        "tdd",
		Package:     "tdd-skill",
		Description: "Test-driven development",
		Installs:    100,
	}
	item := entryToSkillItem(e)
	if item.Name != "tdd" {
		t.Errorf("Name = %q, want %q", item.Name, "tdd")
	}
	if item.Package != "tdd-skill" {
		t.Errorf("Package = %q, want %q", item.Package, "tdd-skill")
	}
	if item.Installs != 100 {
		t.Errorf("Installs = %d, want %d", item.Installs, 100)
	}
}

func TestMergeMCPServerItems(t *testing.T) {
	t.Run("different names kept", func(t *testing.T) {
		items := []registry.MCPServerItem{
			{Name: "a", Source: "aienv"},
			{Name: "b", Source: "aienv"},
		}
		merged := mergeMCPServerItems(items)
		if len(merged) != 2 {
			t.Fatalf("expected 2 items, got %d", len(merged))
		}
	})

	t.Run("duplicate name last wins", func(t *testing.T) {
		items := []registry.MCPServerItem{
			{Name: "a", Source: "aienv", Description: "original"},
			{Name: "a", Source: "user", Description: "override"},
		}
		merged := mergeMCPServerItems(items)
		if len(merged) != 1 {
			t.Fatalf("expected 1 item, got %d", len(merged))
		}
		if merged[0].Source != "user" {
			t.Errorf("expected user override, got %q", merged[0].Source)
		}
		if merged[0].Description != "override" {
			t.Errorf("Description = %q, want %q", merged[0].Description, "override")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		merged := mergeMCPServerItems(nil)
		if len(merged) != 0 {
			t.Errorf("expected 0 items, got %d", len(merged))
		}
	})
}

func TestMergeSkillItems(t *testing.T) {
	t.Run("different names kept", func(t *testing.T) {
		items := []registry.SkillItem{
			{Name: "tdd", Package: "tdd-skill"},
			{Name: "caveman", Package: "caveman-skill"},
		}
		merged := mergeSkillItems(items)
		if len(merged) != 2 {
			t.Fatalf("expected 2 items, got %d", len(merged))
		}
	})

	t.Run("duplicate name last wins", func(t *testing.T) {
		items := []registry.SkillItem{
			{Name: "tdd", Package: "original", Installs: 10},
			{Name: "tdd", Package: "override", Installs: 99},
		}
		merged := mergeSkillItems(items)
		if len(merged) != 1 {
			t.Fatalf("expected 1 item, got %d", len(merged))
		}
		if merged[0].Package != "override" {
			t.Errorf("Package = %q, want %q", merged[0].Package, "override")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		merged := mergeSkillItems(nil)
		if len(merged) != 0 {
			t.Errorf("expected 0 items, got %d", len(merged))
		}
	})
}
