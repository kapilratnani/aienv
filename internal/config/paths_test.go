package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"foo", false},
		{"foo-bar", false},
		{"foo123", false},
		{"a", false},
		{"123", false},
		{"multi-word-with-hyphens", false},
		{"Foo", true},
		{"foo_bar", true},
		{"foo bar", true},
		{"-foo", true},
		{"foo-", true},
		{"", true},
		{"FOO", true},
		{"foo--bar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidName(%q) error = %v, wantErr = %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestComputeHash(t *testing.T) {
	h1 := ComputeHash([]byte("hello"))
	h2 := ComputeHash([]byte("hello"))
	h3 := ComputeHash([]byte("world"))

	if h1 != h2 {
		t.Error("ComputeHash should be deterministic for same input")
	}
	if h1 == h3 {
		t.Error("ComputeHash should differ for different inputs")
	}
	if len(h1) != 64 {
		t.Errorf("ComputeHash length = %d, want 64", len(h1))
	}
}

func TestPathFunctions(t *testing.T) {
	home, _ := os.UserHomeDir()

	t.Run("AIEnvsDir", func(t *testing.T) {
		got := AIEnvsDir()
		want := filepath.Join(home, ".ai-envs")
		if got != want {
			t.Errorf("AIEnvsDir() = %q, want %q", got, want)
		}
	})

	t.Run("EnvDir", func(t *testing.T) {
		got := EnvDir("test-env")
		want := filepath.Join(home, ".ai-envs", "test-env")
		if got != want {
			t.Errorf("EnvDir() = %q, want %q", got, want)
		}
	})

	t.Run("EnvYAML", func(t *testing.T) {
		got := EnvYAML("test-env")
		want := filepath.Join(home, ".ai-envs", "test-env", "ai-env.yaml")
		if got != want {
			t.Errorf("EnvYAML() = %q, want %q", got, want)
		}
	})

	t.Run("OpenCodeJSON", func(t *testing.T) {
		got := OpenCodeJSON("test-env")
		want := filepath.Join(home, ".ai-envs", "test-env", "opencode.json")
		if got != want {
			t.Errorf("OpenCodeJSON() = %q, want %q", got, want)
		}
	})

	t.Run("AgentConfigPath", func(t *testing.T) {
		got := AgentConfigPath("test-env", "foo.json")
		want := filepath.Join(home, ".ai-envs", "test-env", "foo.json")
		if got != want {
			t.Errorf("AgentConfigPath() = %q, want %q", got, want)
		}
	})

	t.Run("TrustDir", func(t *testing.T) {
		got := TrustDir()
		want := filepath.Join(home, ".config", "aienv", "trust")
		if got != want {
			t.Errorf("TrustDir() = %q, want %q", got, want)
		}
	})

	t.Run("TrustCachePath", func(t *testing.T) {
		got := TrustCachePath("abc123")
		want := filepath.Join(home, ".config", "aienv", "trust", "abc123.json")
		if got != want {
			t.Errorf("TrustCachePath() = %q, want %q", got, want)
		}
	})
}

func TestTrustCacheRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	hash := ComputeHash([]byte("test-data"))
	envName := "test-env"

	if err := WriteTrustCache(hash, envName); err != nil {
		t.Fatalf("WriteTrustCache() unexpected error: %v", err)
	}

	entry, err := ReadTrustCache(hash)
	if err != nil {
		t.Fatalf("ReadTrustCache() unexpected error: %v", err)
	}

	if entry.Status != "trusted" {
		t.Errorf("Status = %q, want %q", entry.Status, "trusted")
	}
	if entry.EnvName != envName {
		t.Errorf("EnvName = %q, want %q", entry.EnvName, envName)
	}
	if entry.YAMLHash != hash {
		t.Errorf("YAMLHash = %q, want %q", entry.YAMLHash, hash)
	}
	if entry.ReviewedAt == "" {
		t.Error("ReviewedAt should not be empty")
	}
}

func TestReadTrustCacheMissing(t *testing.T) {
	_, err := ReadTrustCache("nonexistent-hash")
	if err == nil {
		t.Fatal("ReadTrustCache expected error for missing hash")
	}
}

func TestInvalidateTrustCache(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	hash := ComputeHash([]byte("data-1"))
	hash2 := ComputeHash([]byte("data-2"))

	if err := WriteTrustCache(hash, "env-a"); err != nil {
		t.Fatal(err)
	}
	if err := WriteTrustCache(hash2, "env-b"); err != nil {
		t.Fatal(err)
	}

	if err := InvalidateTrustCache("env-a"); err != nil {
		t.Fatalf("InvalidateTrustCache() unexpected error: %v", err)
	}

	if _, err := ReadTrustCache(hash); err == nil {
		t.Error("expected env-a trust entry to be removed")
	}
	if _, err := ReadTrustCache(hash2); err != nil {
		t.Error("expected env-b trust entry to remain")
	}
}

func TestInvalidateTrustCacheNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	last := strings.TrimSuffix
	_ = last
	if err := InvalidateTrustCache("nonexistent-env"); err != nil {
		t.Errorf("InvalidateTrustCache on non-existent env should not error: %v", err)
	}
}
