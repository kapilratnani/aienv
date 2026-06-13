package claude

import (
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

func TestDockerConfig(t *testing.T) {
	t.Run("basic config includes required flags", func(t *testing.T) {
		a := &agent{}
		e := &env.Env{Name: "test", Workdir: "/tmp"}
		cfg, err := a.DockerConfig("/envs/test", e, "sess-1")
		if err != nil {
			t.Fatal(err)
		}
		entry := strings.Join(cfg.Entrypoint, " ")
		if !strings.Contains(entry, "--mcp-config") {
			t.Error("expected --mcp-config flag")
		}
		if !strings.Contains(entry, "/workspace/.aienv/mcp-config.json") {
			t.Error("expected mcp-config pointing to /workspace/.aienv/")
		}
		if !strings.Contains(entry, "--append-system-prompt-file") {
			t.Error("expected --append-system-prompt-file flag")
		}
		if !strings.Contains(entry, "--strict-mcp-config") {
			t.Error("expected --strict-mcp-config flag")
		}
		if len(cfg.Mounts) == 0 {
			t.Error("expected at least one mount")
		}
	})

	t.Run("with permissions adds --settings", func(t *testing.T) {
		a := &agent{}
		e := &env.Env{
			Name:        "test",
			Workdir:     "/tmp",
			Permissions: &env.Permissions{},
		}
		cfg, err := a.DockerConfig("/envs/test", e, "sess-1")
		if err != nil {
			t.Fatal(err)
		}
		entry := strings.Join(cfg.Entrypoint, " ")
		if !strings.Contains(entry, "--settings") {
			t.Error("expected --settings flag when permissions are set")
		}
		if !strings.Contains(entry, "/workspace/.aienv/claude-settings.json") {
			t.Error("expected settings pointing to /workspace/.aienv/")
		}
	})

	t.Run("with model", func(t *testing.T) {
		a := &agent{}
		e := &env.Env{
			Name:    "test",
			Workdir: "/tmp",
			Model:   "claude-sonnet-4-5",
		}
		cfg, err := a.DockerConfig("/envs/test", e, "sess-1")
		if err != nil {
			t.Fatal(err)
		}
		entry := strings.Join(cfg.Entrypoint, " ")
		if !strings.Contains(entry, "--model") {
			t.Error("expected --model flag")
		}
		if !strings.Contains(entry, "claude-sonnet-4-5") {
			t.Error("expected model name in command")
		}
	})

	t.Run("env dir mount", func(t *testing.T) {
		a := &agent{}
		e := &env.Env{Name: "test", Workdir: "/tmp"}
		cfg, err := a.DockerConfig("/envs/test", e, "sess-1")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, m := range cfg.Mounts {
			if m == "/envs/test:/workspace/.aienv:ro" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected env dir mount, got %v", cfg.Mounts)
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
