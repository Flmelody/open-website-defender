package semantic

// IsSQLi detects whether the input contains SQL injection.
// Returns (isSQLi, fingerprint).
func IsSQLi(input string) (bool, string) {
	return detectSQLi(input)
}

// IsXSS detects whether the input contains XSS attack patterns.
func IsXSS(input string) bool {
	return detectXSS(input)
}
