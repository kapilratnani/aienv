package opencode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kapilratnani/aienv/internal/env"
)

func TestBuildMCP(t *testing.T) {
	t.Run("no global config, one server", func(t *testing.T) {
		servers := map[string]env.MCPServer{
			"postgres": {Type: "local", Command: []string{"npx", "-y", "@pg/mcp"}},
		}
		mcp := buildMCP(servers, nil)
		if len(mcp) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(mcp))
		}
		if mcp["postgres"].Type != "local" {
			t.Errorf("Type = %q, want %q", mcp["postgres"].Type, "local")
		}
	})

	t.Run("global config disables missing servers", func(t *testing.T) {
		servers := map[string]env.MCPServer{
			"postgres": {Type: "local"},
		}
		global := map[string]any{
			"mcp": map[string]any{
				"github": map[string]any{},
				"filesystem": map[string]any{},
			},
		}
		mcp := buildMCP(servers, global)
		if len(mcp) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(mcp))
		}
		if mcp["postgres"].Enabled != nil && *mcp["postgres"].Enabled {
			t.Error("postgres should not have enabled=false")
		}
		if mcp["github"].Enabled == nil || *mcp["github"].Enabled {
			t.Error("github should have enabled=false")
		}
		if mcp["filesystem"].Enabled == nil || *mcp["filesystem"].Enabled {
			t.Error("filesystem should have enabled=false")
		}
	})

	t.Run("empty servers", func(t *testing.T) {
		mcp := buildMCP(nil, nil)
		if len(mcp) != 0 {
			t.Errorf("expected 0 entries, got %d", len(mcp))
		}
	})
}

func TestBuildInstructions(t *testing.T) {
	t.Run("absolute path unchanged", func(t *testing.T) {
		rules := []env.Rule{{Path: "/etc/rules.md"}, {Path: "/tmp/other.md"}}
		got := buildInstructions(rules, "/some/cwd")
		if len(got) != 2 {
			t.Fatalf("expected 2 instructions, got %d", len(got))
		}
		if got[0] != "/etc/rules.md" {
			t.Errorf("got[0] = %q, want %q", got[0], "/etc/rules.md")
		}
	})

	t.Run("relative path joined with cwd", func(t *testing.T) {
		rules := []env.Rule{{Path: "docs/rules.md"}}
		got := buildInstructions(rules, "/home/user/project")
		if len(got) != 1 {
			t.Fatalf("expected 1 instruction, got %d", len(got))
		}
		want := filepath.Join("/home/user/project", "docs/rules.md")
		if got[0] != want {
			t.Errorf("got = %q, want %q", got[0], want)
		}
	})

	t.Run("empty rules", func(t *testing.T) {
		got := buildInstructions(nil, "/cwd")
		if len(got) != 0 {
			t.Errorf("expected 0 instructions, got %d", len(got))
		}
	})
}

func TestMergePermission(t *testing.T) {
	t.Run("nil skills and nil perms", func(t *testing.T) {
		got := mergePermission(nil, nil, nil)
		if got != nil {
			t.Error("expected nil for nil skills and nil perms")
		}
	})

	t.Run("empty skills and nil perms", func(t *testing.T) {
		got := mergePermission(nil, []env.Skill{}, nil)
		if got != nil {
			t.Error("expected nil for empty skills and nil perms")
		}
	})

	t.Run("skills only", func(t *testing.T) {
		skills := []env.Skill{{Name: "tdd"}, {Name: "caveman"}}
		got := mergePermission(nil, skills, nil)
		if got == nil {
			t.Fatal("expected non-nil permission")
		}
		skill, ok := got["skill"].(map[string]string)
		if !ok {
			t.Fatalf("expected skill key in permission, got %T", got["skill"])
		}
		if skill["*"] != "deny" {
			t.Errorf("skill[*] = %q, want %q", skill["*"], "deny")
		}
		if skill["tdd"] != "allow" {
			t.Errorf("skill[tdd] = %q, want %q", skill["tdd"], "allow")
		}
		if skill["caveman"] != "allow" {
			t.Errorf("skill[caveman] = %q, want %q", skill["caveman"], "allow")
		}
	})

	t.Run("filesystem permissions only", func(t *testing.T) {
		perms := &env.Permissions{
			Filesystem: &env.FilesystemPermissions{
				Read: map[string]string{"*": "allow", ".env": "deny"},
				Edit: map[string]string{"src/*": "allow"},
			},
		}
		got := mergePermission(nil, nil, perms)
		if got == nil {
			t.Fatal("expected non-nil permission")
		}
		read, ok := got["read"].(map[string]any)
		if !ok {
			t.Fatal("expected read key")
		}
		if read["*"] != "allow" {
			t.Errorf("read[*] = %q, want %q", read["*"], "allow")
		}
		if read[".env"] != "deny" {
			t.Errorf("read[.env] = %q, want %q", read[".env"], "deny")
		}
		edit, ok := got["edit"].(map[string]any)
		if !ok {
			t.Fatal("expected edit key")
		}
		if edit["src/*"] != "allow" {
			t.Errorf("edit[src/*] = %q, want %q", edit["src/*"], "allow")
		}
	})

	t.Run("bash permissions", func(t *testing.T) {
		perms := &env.Permissions{
			Bash: map[string]string{"*": "ask", "git *": "allow"},
		}
		got := mergePermission(nil, nil, perms)
		bash, ok := got["bash"].(map[string]any)
		if !ok {
			t.Fatal("expected bash key")
		}
		if bash["*"] != "ask" {
			t.Errorf("bash[*] = %q, want %q", bash["*"], "ask")
		}
		if bash["git *"] != "allow" {
			t.Errorf("bash[git *] = %q, want %q", bash["git *"], "allow")
		}
	})

	t.Run("merge with global permission", func(t *testing.T) {
		global := map[string]any{
			"permission": map[string]any{
				"read": map[string]any{
					"global.md": "deny",
				},
				"bash": map[string]any{
					"npm *": "allow",
				},
			},
		}
		perms := &env.Permissions{
			Filesystem: &env.FilesystemPermissions{
				Read: map[string]string{"src/*": "allow"},
			},
		}
		skills := []env.Skill{{Name: "tdd"}}
		got := mergePermission(global, skills, perms)

		read, ok := got["read"].(map[string]any)
		if !ok {
			t.Fatal("expected read key")
		}
		if read["global.md"] != "deny" {
			t.Errorf("global read should be preserved")
		}
		if read["src/*"] != "allow" {
			t.Errorf("env read should be merged")
		}
		if len(read) != 2 {
			t.Errorf("expected 2 read entries, got %d", len(read))
		}

		bash, ok := got["bash"].(map[string]any)
		if !ok {
			t.Fatal("expected bash key from global")
		}
		if bash["npm *"] != "allow" {
			t.Errorf("global bash should be preserved")
		}

		skill, ok := got["skill"].(map[string]string)
		if !ok {
			t.Fatalf("expected skill key, got %T", got["skill"])
		}
		if skill["tdd"] != "allow" {
			t.Errorf("skill[tdd] = %q, want %q", skill["tdd"], "allow")
		}
	})
}

func TestActivateCommand(t *testing.T) {
	t.Run("bash default", func(t *testing.T) {
		origShell := os.Getenv("SHELL")
		os.Setenv("SHELL", "/bin/bash")
		t.Cleanup(func() { os.Setenv("SHELL", origShell) })

		a := &agent{}
		cmd := a.ActivateCommand("/some/env/dir", &env.Env{Name: "test"})
		if cmd == "" {
			t.Fatal("expected non-empty command")
		}
	})

	t.Run("fish shell", func(t *testing.T) {
		origShell := os.Getenv("SHELL")
		os.Setenv("SHELL", "/usr/bin/fish")
		t.Cleanup(func() { os.Setenv("SHELL", origShell) })

		a := &agent{}
		cmd := a.ActivateCommand("/some/env/dir", &env.Env{Name: "test"})
		if cmd == "" {
			t.Fatal("expected non-empty command")
		}
	})
}
