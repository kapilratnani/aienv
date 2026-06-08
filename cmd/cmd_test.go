package cmd

import (
	"testing"

	"github.com/kapilratnani/aienv/internal/env"
)

func TestJoinOrNone(t *testing.T) {
	tests := []struct {
		items []string
		want  string
	}{
		{nil, "(none)"},
		{[]string{}, "(none)"},
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a, b"},
		{[]string{"foo", "bar", "baz"}, "foo, bar, baz"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := joinOrNone(tt.items)
			if got != tt.want {
				t.Errorf("joinOrNone(%v) = %q, want %q", tt.items, got, tt.want)
			}
		})
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want []string
	}{
		{"", nil},
		{"npx", []string{"npx"}},
		{"npx -y foo", []string{"npx", "-y", "foo"}},
		{"uvx bar", []string{"uvx", "bar"}},
		{"go run .", []string{"go", "run", "."}},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := splitCommand(tt.cmd)
			if len(got) != len(tt.want) {
				t.Errorf("splitCommand(%q) length = %d, want %d", tt.cmd, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitCommand(%q)[%d] = %q, want %q", tt.cmd, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSortedMCPKeys(t *testing.T) {
	t.Run("returns sorted keys", func(t *testing.T) {
		m := map[string]env.MCPServer{
			"z": {},
			"a": {},
			"m": {},
		}
		got := sortedMCPKeys(m)
		want := []string{"a", "m", "z"}
		if len(got) != len(want) {
			t.Fatalf("length = %d, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("empty map", func(t *testing.T) {
		got := sortedMCPKeys(nil)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
}

func TestHasSkill(t *testing.T) {
	t.Run("skill present", func(t *testing.T) {
		e := &env.Env{
			Skills: []env.Skill{{Name: "tdd"}, {Name: "caveman"}},
		}
		if !hasSkill(e, "tdd") {
			t.Error("expected 'tdd' to be found")
		}
	})

	t.Run("skill not present", func(t *testing.T) {
		e := &env.Env{
			Skills: []env.Skill{{Name: "tdd"}},
		}
		if hasSkill(e, "nonexistent") {
			t.Error("expected 'nonexistent' not to be found")
		}
	})

	t.Run("empty skills", func(t *testing.T) {
		e := &env.Env{}
		if hasSkill(e, "anything") {
			t.Error("expected false for empty skills")
		}
	})
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		rawURL string
		want   string
	}{
		{"https://api.github.com/v1", "api.github.com"},
		{"http://localhost:8080/path", "localhost"},
		{"https://example.com", "example.com"},
		{"invalid", ""},
		{"", ""},
		{"https://api.anthropic.com", "api.anthropic.com"},
	}

	for _, tt := range tests {
		t.Run(tt.rawURL, func(t *testing.T) {
			got := extractHost(tt.rawURL)
			if got != tt.want {
				t.Errorf("extractHost(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}
