package docker

import "testing"

func TestDomainMatch(t *testing.T) {
	tests := []struct {
		pattern string
		host    string
		want    bool
	}{
		{"*", "anything.here.com", true},
		{"*", "", true},
		{"api.github.com", "api.github.com", true},
		{"api.github.com", "evil.com", false},
		{"*.example.com", "api.example.com", true},
		{"*.example.com", "sub.api.example.com", true},
		{"*.example.com", "example.com", true},
		{"*.example.com", "evil.com", false},
		{"*.example.com", "notexample.com", false},
		{"*.github.com", "github.com", true},
		{"*.github.com", "api.github.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.host, func(t *testing.T) {
			got := domainMatch(tt.pattern, tt.host)
			if got != tt.want {
				t.Errorf("domainMatch(%q, %q) = %v, want %v", tt.pattern, tt.host, got, tt.want)
			}
		})
	}
}

func TestProxyIsAllowed(t *testing.T) {
	t.Run("empty lists allows all", func(t *testing.T) {
		p := &Proxy{}
		if !p.isAllowed("any.host.com") {
			t.Error("expected empty allow/deny to allow any host")
		}
	})

	t.Run("allowlist allows matching", func(t *testing.T) {
		p := &Proxy{allowlist: []string{"*.good.com", "api.example.com"}}
		if !p.isAllowed("sub.good.com") {
			t.Error("expected sub.good.com to be allowed")
		}
		if !p.isAllowed("api.example.com") {
			t.Error("expected api.example.com to be allowed")
		}
	})

	t.Run("allowlist denies non-matching", func(t *testing.T) {
		p := &Proxy{allowlist: []string{"*.good.com"}}
		if p.isAllowed("evil.com") {
			t.Error("expected evil.com to be denied by allowlist-only")
		}
	})

	t.Run("denylist blocks matching", func(t *testing.T) {
		p := &Proxy{denylist: []string{"*.evil.com"}}
		if p.isAllowed("api.evil.com") {
			t.Error("expected api.evil.com to be denied")
		}
	})

	t.Run("denylist allows non-matching", func(t *testing.T) {
		p := &Proxy{denylist: []string{"*.evil.com"}}
		if !p.isAllowed("good.com") {
			t.Error("expected good.com to be allowed")
		}
	})

	t.Run("allowlist takes precedence over denylist", func(t *testing.T) {
		p := &Proxy{
			allowlist: []string{"api.github.com"},
			denylist:  []string{"*"},
		}
		if !p.isAllowed("api.github.com") {
			t.Error("expected allowlist to take precedence over denylist")
		}
		if p.isAllowed("evil.com") {
			t.Error("expected evil.com to be denied when allowlist is set")
		}
	})
}
