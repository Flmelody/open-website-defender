package pkg

import (
	"net"
	"path/filepath"
	"strings"
)

// MatchIP checks if an IP address matches a rule.
// Supports three formats:
//   - CIDR notation: "192.168.1.0/24", "10.0.0.0/8", "2001:db8::/32"
//   - Exact match: "192.168.1.1"
//   - Glob pattern: "192.168.1.*" (backward compatible)
func MatchIP(rule, ip string) bool {
	// Case 1: CIDR notation
	if strings.Contains(rule, "/") {
		_, network, err := net.ParseCIDR(rule)
		if err == nil {
			parsedIP := net.ParseIP(ip)
			return parsedIP != nil && network.Contains(parsedIP)
		}
	}

	// Case 2: Exact match
	if rule == ip {
		return true
	}

	// Case 3: Glob pattern (backward compatibility)
	matched, err := filepath.Match(rule, ip)
	return err == nil && matched
}
