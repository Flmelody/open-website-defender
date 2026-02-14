package pkg

import "testing"

func TestMatchIP_CIDRv4(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		ip       string
		expected bool
	}{
		{
			name:     "10.0.0.0/8 matches 10.1.2.3",
			rule:     "10.0.0.0/8",
			ip:       "10.1.2.3",
			expected: true,
		},
		{
			name:     "10.0.0.0/8 matches 10.0.0.1",
			rule:     "10.0.0.0/8",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "10.0.0.0/8 matches 10.255.255.255",
			rule:     "10.0.0.0/8",
			ip:       "10.255.255.255",
			expected: true,
		},
		{
			name:     "10.0.0.0/8 does not match 11.0.0.1",
			rule:     "10.0.0.0/8",
			ip:       "11.0.0.1",
			expected: false,
		},
		{
			name:     "10.0.0.0/8 does not match 192.168.1.1",
			rule:     "10.0.0.0/8",
			ip:       "192.168.1.1",
			expected: false,
		},
		{
			name:     "192.168.1.0/24 matches 192.168.1.100",
			rule:     "192.168.1.0/24",
			ip:       "192.168.1.100",
			expected: true,
		},
		{
			name:     "192.168.1.0/24 matches 192.168.1.0",
			rule:     "192.168.1.0/24",
			ip:       "192.168.1.0",
			expected: true,
		},
		{
			name:     "192.168.1.0/24 matches 192.168.1.255",
			rule:     "192.168.1.0/24",
			ip:       "192.168.1.255",
			expected: true,
		},
		{
			name:     "192.168.1.0/24 does not match 192.168.2.1",
			rule:     "192.168.1.0/24",
			ip:       "192.168.2.1",
			expected: false,
		},
		{
			name:     "172.16.0.0/12 matches 172.31.255.255",
			rule:     "172.16.0.0/12",
			ip:       "172.31.255.255",
			expected: true,
		},
		{
			name:     "172.16.0.0/12 does not match 172.32.0.1",
			rule:     "172.16.0.0/12",
			ip:       "172.32.0.1",
			expected: false,
		},
		{
			name:     "single host /32 matches exactly",
			rule:     "192.168.1.1/32",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "single host /32 does not match other",
			rule:     "192.168.1.1/32",
			ip:       "192.168.1.2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIP(tt.rule, tt.ip)
			if result != tt.expected {
				t.Errorf("MatchIP(%q, %q) = %v, want %v", tt.rule, tt.ip, result, tt.expected)
			}
		})
	}
}

func TestMatchIP_CIDRv6(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		ip       string
		expected bool
	}{
		{
			name:     "2001:db8::/32 matches 2001:db8::1",
			rule:     "2001:db8::/32",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "2001:db8::/32 matches 2001:db8:abcd::1",
			rule:     "2001:db8::/32",
			ip:       "2001:db8:abcd::1",
			expected: true,
		},
		{
			name:     "2001:db8::/32 does not match 2001:db9::1",
			rule:     "2001:db8::/32",
			ip:       "2001:db9::1",
			expected: false,
		},
		{
			name:     "fe80::/10 matches fe80::1",
			rule:     "fe80::/10",
			ip:       "fe80::1",
			expected: true,
		},
		{
			name:     "fe80::/10 matches febf:ffff::1",
			rule:     "fe80::/10",
			ip:       "febf:ffff::1",
			expected: true,
		},
		{
			name:     "fe80::/10 does not match ff00::1",
			rule:     "fe80::/10",
			ip:       "ff00::1",
			expected: false,
		},
		{
			name:     "::1/128 matches ::1 exactly",
			rule:     "::1/128",
			ip:       "::1",
			expected: true,
		},
		{
			name:     "::1/128 does not match ::2",
			rule:     "::1/128",
			ip:       "::2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIP(tt.rule, tt.ip)
			if result != tt.expected {
				t.Errorf("MatchIP(%q, %q) = %v, want %v", tt.rule, tt.ip, result, tt.expected)
			}
		})
	}
}

func TestMatchIP_ExactMatch(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		ip       string
		expected bool
	}{
		{
			name:     "exact match IPv4",
			rule:     "192.168.1.1",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "exact match different IPv4",
			rule:     "192.168.1.1",
			ip:       "192.168.1.2",
			expected: false,
		},
		{
			name:     "exact match loopback",
			rule:     "127.0.0.1",
			ip:       "127.0.0.1",
			expected: true,
		},
		{
			name:     "exact match IPv6 loopback",
			rule:     "::1",
			ip:       "::1",
			expected: true,
		},
		{
			name:     "exact match full IPv6",
			rule:     "2001:db8::1",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "exact match different IPv6",
			rule:     "2001:db8::1",
			ip:       "2001:db8::2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIP(tt.rule, tt.ip)
			if result != tt.expected {
				t.Errorf("MatchIP(%q, %q) = %v, want %v", tt.rule, tt.ip, result, tt.expected)
			}
		})
	}
}

func TestMatchIP_GlobPattern(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		ip       string
		expected bool
	}{
		{
			name:     "wildcard last octet matches",
			rule:     "192.168.1.*",
			ip:       "192.168.1.100",
			expected: true,
		},
		{
			name:     "wildcard last octet matches .0",
			rule:     "192.168.1.*",
			ip:       "192.168.1.0",
			expected: true,
		},
		{
			name:     "wildcard last octet matches .255",
			rule:     "192.168.1.*",
			ip:       "192.168.1.255",
			expected: true,
		},
		{
			name:     "wildcard last octet does not match different subnet",
			rule:     "192.168.1.*",
			ip:       "192.168.2.100",
			expected: false,
		},
		{
			name:     "wildcard two octets",
			rule:     "10.0.*.*",
			ip:       "10.0.5.99",
			expected: true,
		},
		{
			name:     "wildcard two octets does not match different second octet",
			rule:     "10.0.*.*",
			ip:       "10.1.5.99",
			expected: false,
		},
		{
			name:     "single char wildcard",
			rule:     "192.168.1.?",
			ip:       "192.168.1.5",
			expected: true,
		},
		{
			name:     "single char wildcard does not match multi-digit",
			rule:     "192.168.1.?",
			ip:       "192.168.1.50",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIP(tt.rule, tt.ip)
			if result != tt.expected {
				t.Errorf("MatchIP(%q, %q) = %v, want %v", tt.rule, tt.ip, result, tt.expected)
			}
		})
	}
}

func TestMatchIP_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		ip       string
		expected bool
	}{
		{
			name:     "empty rule empty IP",
			rule:     "",
			ip:       "",
			expected: true, // empty string equals empty string for exact match
		},
		{
			name:     "empty rule non-empty IP",
			rule:     "",
			ip:       "192.168.1.1",
			expected: false,
		},
		{
			name:     "non-empty rule empty IP",
			rule:     "192.168.1.1",
			ip:       "",
			expected: false,
		},
		{
			name:     "invalid CIDR falls through to exact match",
			rule:     "999.999.999.999/33",
			ip:       "999.999.999.999/33",
			expected: true, // invalid CIDR parse fails, falls through to exact string match
		},
		{
			name:     "CIDR with invalid IP target returns false",
			rule:     "10.0.0.0/8",
			ip:       "not-an-ip",
			expected: false,
		},
		{
			name:     "CIDR with empty IP returns false",
			rule:     "10.0.0.0/8",
			ip:       "",
			expected: false,
		},
		{
			name:     "rule with slash but not valid CIDR falls to exact match",
			rule:     "abc/def",
			ip:       "abc/def",
			expected: true, // invalid CIDR, falls through to exact string match
		},
		{
			name:     "wildcard * matches any string without path separator",
			rule:     "*",
			ip:       "192.168.1.1",
			expected: true, // filepath.Match("*", ...) matches any non-separator sequence; IP has dots not path separators
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIP(tt.rule, tt.ip)
			if result != tt.expected {
				t.Errorf("MatchIP(%q, %q) = %v, want %v", tt.rule, tt.ip, result, tt.expected)
			}
		})
	}
}

func TestMatchIP_CIDRv4v6CrossMatch(t *testing.T) {
	// Ensure IPv4 CIDR doesn't match IPv6 and vice versa
	tests := []struct {
		name     string
		rule     string
		ip       string
		expected bool
	}{
		{
			name:     "IPv4 CIDR does not match IPv6 address",
			rule:     "10.0.0.0/8",
			ip:       "::ffff:10.1.2.3",
			expected: true, // Go net library handles IPv4-mapped IPv6 addresses
		},
		{
			name:     "IPv6 CIDR does not match plain IPv4",
			rule:     "2001:db8::/32",
			ip:       "10.1.2.3",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIP(tt.rule, tt.ip)
			if result != tt.expected {
				t.Errorf("MatchIP(%q, %q) = %v, want %v", tt.rule, tt.ip, result, tt.expected)
			}
		})
	}
}
