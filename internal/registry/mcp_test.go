package registry

import "testing"

func TestShortRegistryName(t *testing.T) {
	tests := []struct {
		full string
		want string
	}{
		{"org/postgres-mcp", "postgres"},
		{"org/server-postgres-mcp", "postgres"},
		{"server-foo", "foo"},
		{"foo-bar-server", "foo-bar"},
		{"bar_mcp", "bar-mcp"},
		{"simple-name", "simple-name"},
		{"org/nested/server-name-mcp", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.full, func(t *testing.T) {
			got := shortRegistryName(tt.full)
			if got != tt.want {
				t.Errorf("shortRegistryName(%q) = %q, want %q", tt.full, got, tt.want)
			}
		})
	}
}

func TestRegistryCommand(t *testing.T) {
	tests := []struct {
		registryType string
		identifier   string
		want         string
	}{
		{"npm", "@modelcontextprotocol/server-postgres", "npx -y @modelcontextprotocol/server-postgres"},
		{"pypi", "mcp-server-foo", "uvx mcp-server-foo"},
		{"go", "github.com/foo/mcp-server", "go run github.com/foo/mcp-server"},
		{"unknown", "some-package", "npx -y some-package"},
		{"", "some-package", "npx -y some-package"},
	}

	for _, tt := range tests {
		t.Run(tt.registryType+"-"+tt.identifier, func(t *testing.T) {
			got := registryCommand(tt.registryType, tt.identifier)
			if got != tt.want {
				t.Errorf("registryCommand(%q, %q) = %q, want %q", tt.registryType, tt.identifier, got, tt.want)
			}
		})
	}
}
