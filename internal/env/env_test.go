package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tilde alone", "~", home},
		{"tilde with path", "~/foo/bar", filepath.Join(home, "foo/bar")},
		{"tilde with subdir only", "~/", filepath.Join(home, "/")},
		{"not tilde", "~other", "~other"},
		{"absolute path", "/usr/local", "/usr/local"},
		{"relative path", "relative/path", "relative/path"},
		{"empty string", "", ""},
		{"just dot", ".", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandTilde(tt.input)
			if got != tt.want {
				t.Errorf("ExpandTilde(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
