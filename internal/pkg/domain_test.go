package pkg

import "testing"

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		pattern string
		domain  string
		want    bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "Example.Com", true},
		{"Example.Com", "example.com", true},
		{"example.com", "other.com", false},
		{"*.example.com", "sub.example.com", true},
		{"*.example.com", "deep.sub.example.com", true},
		{"*.example.com", "example.com", false},
		{"*.example.com", "other.com", false},
		{"", "example.com", false},
		{"example.com", "", false},
	}

	for _, tt := range tests {
		got := MatchDomain(tt.pattern, tt.domain)
		if got != tt.want {
			t.Errorf("MatchDomain(%q, %q) = %v, want %v", tt.pattern, tt.domain, got, tt.want)
		}
	}
}

func TestStripPort(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"example.com:8080", "example.com"},
		{"example.com", "example.com"},
		{"example.com:443", "example.com"},
		{"", ""},
		{"[::1]:8080", "::1"},
	}

	for _, tt := range tests {
		got := StripPort(tt.input)
		if got != tt.want {
			t.Errorf("StripPort(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCheckDomainScope(t *testing.T) {
	tests := []struct {
		scopes string
		domain string
		want   bool
	}{
		// Empty scopes = unrestricted
		{"", "anything.com", true},
		{"  ", "anything.com", true},

		// Single exact match
		{"gitea.com", "gitea.com", true},
		{"gitea.com", "gitlab.com", false},

		// Multiple patterns
		{"gitea.com,gitlab.com", "gitea.com", true},
		{"gitea.com,gitlab.com", "gitlab.com", true},
		{"gitea.com,gitlab.com", "github.com", false},

		// Wildcard
		{"*.example.com", "sub.example.com", true},
		{"*.example.com", "example.com", false},

		// Mixed
		{"gitea.com, *.internal.org", "gitea.com", true},
		{"gitea.com, *.internal.org", "app.internal.org", true},
		{"gitea.com, *.internal.org", "external.com", false},

		// Domain with port
		{"gitea.com", "gitea.com:3000", true},
		{"gitea.com", "other.com:3000", false},

		// Empty domain
		{"gitea.com", "", false},

		// Spaces in patterns
		{" gitea.com , gitlab.com ", "gitea.com", true},
	}

	for _, tt := range tests {
		got := CheckDomainScope(tt.scopes, tt.domain)
		if got != tt.want {
			t.Errorf("CheckDomainScope(%q, %q) = %v, want %v", tt.scopes, tt.domain, got, tt.want)
		}
	}
}
