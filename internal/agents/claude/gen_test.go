package claude

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/kapilratnani/aienv/internal/env"
)

func TestBuildClaudeSettings(t *testing.T) {
	t.Run("nil permissions", func(t *testing.T) {
		settings := buildClaudeSettings(nil)
		if settings.Permissions != nil {
			t.Error("expected nil Permissions for nil input")
		}
	})

	t.Run("empty permissions", func(t *testing.T) {
		settings := buildClaudeSettings(&env.Permissions{})
		if settings.Permissions != nil {
			t.Error("expected nil Permissions for empty permissions")
		}
	})

	t.Run("filesystem read patterns", func(t *testing.T) {
		perms := &env.Permissions{
			Filesystem: &env.FilesystemPermissions{
				Read: map[string]string{
					"*":      "allow",
					".env":   "deny",
					"config": "ask",
				},
			},
		}
		settings := buildClaudeSettings(perms)
		if settings.Permissions == nil {
			t.Fatal("expected non-nil Permissions")
		}
		assertContains(t, settings.Permissions.Allow, "Read(*)")
		assertContains(t, settings.Permissions.Deny, "Read(.env)")
		assertContains(t, settings.Permissions.Ask, "Read(config)")
	})

	t.Run("filesystem edit patterns", func(t *testing.T) {
		perms := &env.Permissions{
			Filesystem: &env.FilesystemPermissions{
				Edit: map[string]string{
					"src/*": "allow",
					"*.env": "deny",
				},
			},
		}
		settings := buildClaudeSettings(perms)
		if settings.Permissions == nil {
			t.Fatal("expected non-nil Permissions")
		}
		assertContains(t, settings.Permissions.Allow, "Edit(src/*)")
		assertContains(t, settings.Permissions.Deny, "Edit(*.env)")
	})

	t.Run("bash patterns", func(t *testing.T) {
		perms := &env.Permissions{
			Bash: map[string]string{
				"*":        "ask",
				"git *":    "allow",
				"rm -rf *": "deny",
			},
		}
		settings := buildClaudeSettings(perms)
		if settings.Permissions == nil {
			t.Fatal("expected non-nil Permissions")
		}
		assertContains(t, settings.Permissions.Ask, "Bash(*)")
		assertContains(t, settings.Permissions.Allow, "Bash(git *)")
		assertContains(t, settings.Permissions.Deny, "Bash(rm -rf *)")
	})

	t.Run("mixed permissions", func(t *testing.T) {
		perms := &env.Permissions{
			Filesystem: &env.FilesystemPermissions{
				Read: map[string]string{"*": "allow"},
				Edit: map[string]string{"src/*": "allow"},
			},
			Bash: map[string]string{"git *": "allow"},
		}
		settings := buildClaudeSettings(perms)
		if settings.Permissions == nil {
			t.Fatal("expected non-nil Permissions")
		}
		if len(settings.Permissions.Allow) != 3 {
			t.Errorf("expected 3 allow entries, got %d", len(settings.Permissions.Allow))
		}
	})
}

func TestActivateCommand(t *testing.T) {
	t.Run("basic command includes mcp-config and claude-md", func(t *testing.T) {
		a := &agent{}
		cmd := a.ActivateCommand("/envs/test", &env.Env{Name: "test"})
		if cmd == "" {
			t.Fatal("expected non-empty command")
		}
		if !strings.Contains(cmd, "--mcp-config") {
			t.Error("expected --mcp-config flag")
		}
		if !strings.Contains(cmd, "--append-system-prompt-file") {
			t.Error("expected --append-system-prompt-file flag")
		}
		if !strings.Contains(cmd, "--strict-mcp-config") {
			t.Error("expected --strict-mcp-config flag")
		}
	})

	t.Run("with permissions adds --settings", func(t *testing.T) {
		a := &agent{}
		e := &env.Env{
			Name:        "test",
			Permissions: &env.Permissions{},
		}
		cmd := a.ActivateCommand("/envs/test", e)
		if !strings.Contains(cmd, "--settings") {
			t.Error("expected --settings flag when permissions are set")
		}
	})

	t.Run("with rules", func(t *testing.T) {
		origDir, _ := os.Getwd()
		t.Cleanup(func() { os.Chdir(origDir) })
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)

		a := &agent{}
		e := &env.Env{
			Name:  "test",
			Rules: []env.Rule{{Path: "rules.md"}, {Path: "/absolute/path.md"}},
		}
		cmd := a.ActivateCommand("/envs/test", e)
		if !strings.Contains(cmd, "rules.md") {
			t.Error("expected rule file in command")
		}
		if !strings.Contains(cmd, "/absolute/path.md") {
			t.Error("expected absolute rule path in command")
		}
	})

	t.Run("with model", func(t *testing.T) {
		a := &agent{}
		e := &env.Env{
			Name:  "test",
			Model: "claude-sonnet-4-5",
		}
		cmd := a.ActivateCommand("/envs/test", e)
		if !strings.Contains(cmd, "--model") {
			t.Error("expected --model flag")
		}
		if !strings.Contains(cmd, "claude-sonnet-4-5") {
			t.Error("expected model name in command")
		}
	})
}

func assertContains(t *testing.T, slice []string, item string) {
	t.Helper()
	if slices.Contains(slice, item) {
		return
	}
	t.Errorf("expected slice to contain %q, got %v", item, slice)
}
