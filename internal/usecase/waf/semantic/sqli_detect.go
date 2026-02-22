package semantic

// detectSQLi analyzes the input string for SQL injection.
// Returns (isSQLi, fingerprint).
func detectSQLi(input string) (bool, string) {
	if len(input) == 0 {
		return false, ""
	}

	// Tokenize the input
	tokens := tokenize(input)
	if len(tokens) == 0 {
		return false, ""
	}

	// Quick check: if all tokens are barewords or all are numbers, skip
	allSafe := true
	for _, t := range tokens {
		if t.ttype != tokenTypeBareword && t.ttype != tokenTypeNumber && t.ttype != tokenTypeComma && t.ttype != tokenTypeOperator {
			allSafe = false
			break
		}
	}
	// A string of only barewords and numbers with no SQL semantics is safe
	if allSafe {
		hasSQLSemantics := false
		for _, t := range tokens {
			if t.ttype == tokenTypeOperator {
				hasSQLSemantics = true
				break
			}
		}
		if !hasSQLSemantics {
			return false, ""
		}
	}

	// Fold the token stream
	folded := fold(tokens)
	if len(folded) == 0 {
		return false, ""
	}

	// Generate fingerprint
	fp := fingerprint(folded)

	// Check against known attack fingerprints
	if !isKnownFingerprint(fp) {
		return false, ""
	}

	// Apply whitelist (false positive reduction)
	if isWhitelisted(fp, tokens) {
		return false, ""
	}

	return true, fp
}

// fingerprint generates a fingerprint string from a token sequence.
// Each token contributes its type character.
func fingerprint(tokens []sqlToken) string {
	b := make([]byte, len(tokens))
	for i, t := range tokens {
		b[i] = t.ttype
	}
	return string(b)
}
