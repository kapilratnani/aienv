package registry

import "testing"

func TestExtractPackage(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"org/repo/sub", "org/repo"},
		{"github.com/owner/repo", "github.com/owner/repo"},
		{"just-name", "just-name"},
		{"single/slash", "single/slash"},
		{"", ""},
		{"a/b/c/d", "a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := extractPackage(tt.id)
			if got != tt.want {
				t.Errorf("extractPackage(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestSplitN(t *testing.T) {
	tests := []struct {
		s     string
		sep   string
		n     int
		want  []string
	}{
		{"a/b/c", "/", 2, []string{"a", "b/c"}},
		{"a", "/", 2, []string{"a"}},
		{"a/b", "/", 1, []string{"a/b"}},
		{"a/b/c/d", "/", 3, []string{"a", "b", "c/d"}},
		{"hello-world", "-", 2, []string{"hello", "world"}},
		{"no-separator", "/", 2, []string{"no-separator"}},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := splitN(tt.s, tt.sep, tt.n)
			if len(got) != len(tt.want) {
				t.Errorf("splitN(%q, %q, %d) length = %d, want %d", tt.s, tt.sep, tt.n, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitN(%q, %q, %d)[%d] = %q, want %q", tt.s, tt.sep, tt.n, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		s   string
		sep string
		want int
	}{
		{"hello", "ll", 2},
		{"hello", "x", -1},
		{"hello", "", 0},
		{"", "x", -1},
		{"abcabc", "b", 1},
	}

	for _, tt := range tests {
		t.Run(tt.s+"-"+tt.sep, func(t *testing.T) {
			got := indexOf(tt.s, tt.sep)
			if got != tt.want {
				t.Errorf("indexOf(%q, %q) = %d, want %d", tt.s, tt.sep, got, tt.want)
			}
		})
	}
}
