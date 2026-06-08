package docker

import "testing"

func TestReadDockerfile(t *testing.T) {
	t.Run("opencode", func(t *testing.T) {
		data, err := readDockerfile("opencode")
		if err != nil {
			t.Fatalf("readDockerfile('opencode') unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("readDockerfile('opencode') returned empty bytes")
		}
	})

	t.Run("claude-code", func(t *testing.T) {
		data, err := readDockerfile("claude-code")
		if err != nil {
			t.Fatalf("readDockerfile('claude-code') unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("readDockerfile('claude-code') returned empty bytes")
		}
	})

	t.Run("unknown agent", func(t *testing.T) {
		_, err := readDockerfile("unknown-agent")
		if err == nil {
			t.Fatal("readDockerfile('unknown-agent') expected error")
		}
	})
}
