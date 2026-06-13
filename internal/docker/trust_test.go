package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
)

func setTestDirs(t *testing.T) {
	t.Helper()
	td := t.TempDir()
	config.TestDataDir = filepath.Join(td, "data")
	config.TestTrustDir = filepath.Join(td, "trust")
	t.Cleanup(func() {
		config.TestDataDir = ""
		config.TestTrustDir = ""
	})
}

func writeTestEnv(t *testing.T, name, content string) string {
	t.Helper()
	dir := filepath.Join(config.DataDir(), name)
	os.MkdirAll(dir, 0755)
	yamlPath := filepath.Join(dir, "env.yaml")
	os.WriteFile(yamlPath, []byte(content), 0644)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return yamlPath
}

func TestGenerateSessionID(t *testing.T) {
	id1 := config.GenerateSessionID()
	id2 := config.GenerateSessionID()

	if id1 == id2 {
		t.Error("expected unique session IDs")
	}

	if len(id1) < 15 {
		t.Errorf("session ID too short: %q", id1)
	}
}

func TestBuild(t *testing.T) {
	if err := Check(); err != nil {
		t.Skip("Docker not available:", err)
	}

	e := &env.Env{
		Meta: env.EnvMeta{Name: "test-env"},
		Agent: env.AgentConfig{
			Install: []string{"echo agent-installed"},
			Command: []string{"echo", "hello"},
			Mounts:  []env.Mount{{Source: t.TempDir(), Target: "/workspace"}},
		},
	}

	if err := ensureBaseImage(); err != nil {
		t.Fatalf("ensureBaseImage: %v", err)
	}

	if err := Build(e); err != nil {
		t.Fatalf("Build: %v", err)
	}

	tag := ImageTag(e)
	if tag == "" {
		t.Fatal("expected non-empty image tag")
	}
}

func TestImageTagConsistency(t *testing.T) {
	e1 := &env.Env{Meta: env.EnvMeta{Name: "test-env"}}
	e2 := &env.Env{Meta: env.EnvMeta{Name: "test-env"}}

	tag1 := ImageTag(e1)
	tag2 := ImageTag(e2)

	if tag1 != tag2 {
		t.Errorf("expected same image tag for same env content, got %q vs %q", tag1, tag2)
	}
}

func TestGenerateDockerfile(t *testing.T) {
	e := &env.Env{
		Deps: env.DepsConfig{
			Packages: []string{"golang", "python3"},
			Custom:   []string{"go install foo"},
		},
		Agent: env.AgentConfig{
			Install: []string{"npm install -g opencode"},
			Command: []string{"opencode"},
		},
	}

	df := generateDockerfile(e)

	if df == "" {
		t.Fatal("expected non-empty Dockerfile")
	}

	if !strings.Contains(df, "FROM aienv/sandbox:latest") {
		t.Error("expected FROM aienv/sandbox:latest")
	}
	if !strings.Contains(df, "USER root") {
		t.Error("expected USER root")
	}
	if !strings.Contains(df, "USER agent") {
		t.Error("expected USER agent")
	}
	if !strings.Contains(df, "golang") || !strings.Contains(df, "python3") {
		t.Error("expected packages in Dockerfile")
	}
	if !strings.Contains(df, "npm install -g opencode") {
		t.Error("expected agent install command")
	}
	if !strings.Contains(df, "go install foo") {
		t.Error("expected custom dep command")
	}
}

func TestNeedsTrustPromptNoCache(t *testing.T) {
	setTestDirs(t)
	writeTestEnv(t, "test-trust", "agent:\n  command: [echo]\n  mounts:\n  - source: /\n    target: /workspace\n")

	e, err := env.Load("test-trust")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	needed, err := NeedsTrustPrompt(e)
	if err != nil {
		t.Fatalf("NeedsTrustPrompt: %v", err)
	}
	if !needed {
		t.Error("expected trust prompt needed for new env")
	}
}

func TestSaveTrust(t *testing.T) {
	setTestDirs(t)
	writeTestEnv(t, "test-save-trust", "agent:\n  command: [echo]\n  mounts:\n  - source: /\n    target: /workspace\n")

	e, err := env.Load("test-save-trust")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := SaveTrust(e); err != nil {
		t.Fatalf("SaveTrust: %v", err)
	}

	needed, err := NeedsTrustPrompt(e)
	if err != nil {
		t.Fatalf("NeedsTrustPrompt after save: %v", err)
	}
	if needed {
		t.Error("expected no trust prompt needed after saving trust")
	}
}

func TestSaveTrustInvalidatesOnChange(t *testing.T) {
	setTestDirs(t)
	writeTestEnv(t, "test-change-trust", "agent:\n  command: [echo]\n  mounts:\n  - source: /\n    target: /workspace\n")

	e, err := env.Load("test-change-trust")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Save trust for initial content
	if err := SaveTrust(e); err != nil {
		t.Fatalf("SaveTrust: %v", err)
	}

	needed, err := NeedsTrustPrompt(e)
	if err != nil {
		t.Fatalf("NeedsTrustPrompt: %v", err)
	}
	if needed {
		t.Fatal("expected no trust prompt before change")
	}

	// Change env content
	yamlPath := config.EnvYAML("test-change-trust")
	os.WriteFile(yamlPath, []byte("agent:\n  command: [echo]\n  mounts:\n  - source: /\n    target: /workspace\n  env:\n    FOO: bar\n"), 0644)

	e2, err := env.Load("test-change-trust")
	if err != nil {
		t.Fatalf("Load after change: %v", err)
	}

	needed, err = NeedsTrustPrompt(e2)
	if err != nil {
		t.Fatalf("NeedsTrustPrompt after change: %v", err)
	}
	if !needed {
		t.Error("expected trust prompt needed after content change")
	}
}
