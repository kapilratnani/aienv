package env

import (
	"slices"
	"testing"
)

func TestDetectOpenCodeEndpoints(t *testing.T) {
	hosts := DetectProviderEndpoints("opencode")
	t.Logf("Detected OpenCode provider hosts: %v", hosts)
	// Should not include loopback
	for _, h := range hosts {
		if isLoopback(h) {
			t.Errorf("should not include loopback host: %s", h)
		}
	}
}

func TestDetectClaudeEndpoints(t *testing.T) {
	hosts := DetectProviderEndpoints("claude-code")
	t.Logf("Detected Claude provider hosts: %v", hosts)
	found := slices.Contains(hosts, "api.anthropic.com")
	if !found {
		t.Error("claude-code should always include api.anthropic.com")
	}
}

func TestIsLoopback(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"0.0.0.0", true},
		{"::1", true},
		{"127.0.0.2", true},
		{"api.anthropic.com", false},
		{"api.openai.com", false},
		{"example.com", false},
	}
	for _, c := range cases {
		got := isLoopback(c.host)
		if got != c.want {
			t.Errorf("isLoopback(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}
