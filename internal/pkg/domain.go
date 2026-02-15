package pkg

import (
	"net"
	"strings"
)

// MatchDomain checks if a domain matches a pattern.
// Supports exact match and wildcard patterns like "*.example.com".
// Comparison is case-insensitive.
func MatchDomain(pattern, domain string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	domain = strings.ToLower(strings.TrimSpace(domain))

	if pattern == "" || domain == "" {
		return false
	}

	if pattern == domain {
		return true
	}

	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // e.g. ".example.com"
		return strings.HasSuffix(domain, suffix)
	}

	return false
}

// StripPort removes the port from a host string (e.g. "example.com:8080" -> "example.com").
func StripPort(host string) string {
	if host == "" {
		return ""
	}
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		// No port present or IPv6 without port
		return host
	}
	return h
}

// CheckDomainScope checks if a domain is allowed by a comma-separated list of scope patterns.
// Empty scopes means unrestricted access (allow all).
func CheckDomainScope(scopes, domain string) bool {
	scopes = strings.TrimSpace(scopes)
	if scopes == "" {
		return true
	}

	domain = StripPort(domain)
	if domain == "" {
		return false
	}

	patterns := strings.Split(scopes, ",")
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if MatchDomain(pattern, domain) {
			return true
		}
	}

	return false
}
