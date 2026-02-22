package semantic

// fold applies token folding rules to reduce the token stream.
// It merges compound keywords, folds arithmetic, removes unary operators,
// and compresses the result to at most 5 tokens.
func fold(tokens []sqlToken) []sqlToken {
	if len(tokens) == 0 {
		return tokens
	}

	// Phase 1: Merge compound keywords (UNION ALL, GROUP BY, ORDER BY, IS NOT, NOT IN, etc.)
	tokens = mergeCompound(tokens)

	// Phase 2: Multi-round folding until stable
	for round := 0; round < 5; round++ {
		prev := len(tokens)
		tokens = foldTwoTokens(tokens)
		tokens = foldThreeTokens(tokens)
		if len(tokens) == prev {
			break
		}
	}

	// Phase 3: Truncate to at most 5 tokens
	if len(tokens) > 5 {
		tokens = tokens[:5]
	}

	return tokens
}

// mergeCompound merges multi-word SQL keywords into single tokens.
func mergeCompound(tokens []sqlToken) []sqlToken {
	if len(tokens) < 2 {
		return tokens
	}

	result := make([]sqlToken, 0, len(tokens))
	i := 0

	for i < len(tokens) {
		if i+1 < len(tokens) {
			merged, skip := tryMerge(tokens[i], tokens[i+1])
			if skip > 0 {
				result = append(result, merged)
				i += skip + 1
				continue
			}
		}
		result = append(result, tokens[i])
		i++
	}

	return result
}

// tryMerge attempts to merge two tokens into a compound keyword.
// Returns the merged token and the number of extra tokens consumed (0 if no merge).
func tryMerge(a, b sqlToken) (sqlToken, int) {
	au := toUpper(a.value)
	bu := toUpper(b.value)

	// UNION ALL → single UNION token
	if a.ttype == tokenTypeUnion && bu == "ALL" {
		return sqlToken{ttype: tokenTypeUnion, value: a.value + " " + b.value}, 1
	}

	// GROUP BY, ORDER BY → single GroupBy token
	if a.ttype == tokenTypeGroupBy && bu == "BY" {
		return sqlToken{ttype: tokenTypeGroupBy, value: a.value + " " + b.value}, 1
	}

	// IS NULL, IS NOT → keep as keyword
	if au == "IS" && (bu == "NULL" || bu == "NOT") {
		return sqlToken{ttype: tokenTypeKeyword, value: a.value + " " + b.value}, 1
	}

	// NOT IN, NOT LIKE, NOT EXISTS, NOT BETWEEN → logic token
	if au == "NOT" && (bu == "IN" || bu == "LIKE" || bu == "EXISTS" || bu == "BETWEEN") {
		return sqlToken{ttype: tokenTypeLogic, value: a.value + " " + b.value}, 1
	}

	// INSERT INTO → keyword
	if au == "INSERT" && bu == "INTO" {
		return sqlToken{ttype: tokenTypeKeyword, value: a.value + " " + b.value}, 1
	}

	// DELETE FROM → keyword
	if au == "DELETE" && bu == "FROM" {
		return sqlToken{ttype: tokenTypeKeyword, value: a.value + " " + b.value}, 1
	}

	return sqlToken{}, 0
}

// foldTwoTokens applies 2-token folding rules in a single pass.
func foldTwoTokens(tokens []sqlToken) []sqlToken {
	if len(tokens) < 2 {
		return tokens
	}

	result := make([]sqlToken, 0, len(tokens))
	i := 0

	for i < len(tokens) {
		if i+1 < len(tokens) && applyTwoFold(result, &tokens[i], &tokens[i+1]) {
			result = append(result, tokens[i])
			i += 2
			continue
		}
		result = append(result, tokens[i])
		i++
	}

	return result
}

// applyTwoFold checks and applies 2-token fold rules.
// If folded, it modifies 'a' in place and returns true (b should be skipped).
// 'preceding' provides context for determining if an operator is unary.
func applyTwoFold(preceding []sqlToken, a, b *sqlToken) bool {
	// Unary operator before number: +1, -1 → number
	// Only applies when the operator is truly unary (at start of expression
	// or after another operator/lparen/comma/keyword/semicolon)
	if a.ttype == tokenTypeOperator && (a.value == "+" || a.value == "-") && b.ttype == tokenTypeNumber {
		isUnary := len(preceding) == 0
		if !isUnary && len(preceding) > 0 {
			prev := preceding[len(preceding)-1]
			switch prev.ttype {
			case tokenTypeOperator, tokenTypeLParen, tokenTypeComma,
				tokenTypeSemicolon, tokenTypeKeyword, tokenTypeSelect,
				tokenTypeUnion, tokenTypeDDL, tokenTypeLogic:
				isUnary = true
			}
		}
		if isUnary {
			a.ttype = tokenTypeNumber
			a.value = a.value + b.value
			return true
		}
	}

	// Unary NOT before expression
	if a.ttype == tokenTypeLogic && toUpper(a.value) == "NOT" {
		if b.ttype == tokenTypeNumber || b.ttype == tokenTypeString || b.ttype == tokenTypeBareword {
			// NOT expr → expr (keep the expr)
			a.ttype = b.ttype
			a.value = b.value
			return true
		}
	}

	// Consecutive strings: 'a' 'b' → single string
	if a.ttype == tokenTypeString && b.ttype == tokenTypeString {
		a.value = a.value + b.value
		return true
	}

	// Consecutive comments: merge
	if a.ttype == tokenTypeComment && b.ttype == tokenTypeComment {
		a.value = a.value + " " + b.value
		return true
	}

	return false
}

// foldThreeTokens applies 3-token folding rules (mainly arithmetic).
func foldThreeTokens(tokens []sqlToken) []sqlToken {
	if len(tokens) < 3 {
		return tokens
	}

	result := make([]sqlToken, 0, len(tokens))
	i := 0

	for i < len(tokens) {
		if i+2 < len(tokens) && applyThreeFold(&tokens[i], &tokens[i+1], &tokens[i+2]) {
			result = append(result, tokens[i])
			i += 3
			continue
		}
		result = append(result, tokens[i])
		i++
	}

	return result
}

// applyThreeFold checks and applies 3-token fold rules.
// If folded, it modifies 'a' in place and returns true (b and c should be skipped).
func applyThreeFold(a, b, c *sqlToken) bool {
	// Number OP Number → Number (arithmetic folding)
	// e.g., 1+2, 3*4, 5/6, 7-8, 1%2
	if a.ttype == tokenTypeNumber && b.ttype == tokenTypeOperator && c.ttype == tokenTypeNumber {
		if isArithmeticOp(b.value) {
			a.value = a.value + b.value + c.value
			return true
		}
	}

	// Number COMPARISON Number → Number (comparison folding for fingerprint)
	// e.g., 1=1, 1<2, 1>0
	if a.ttype == tokenTypeNumber && b.ttype == tokenTypeOperator && c.ttype == tokenTypeNumber {
		if isComparisonOp(b.value) {
			a.value = a.value + b.value + c.value
			return true
		}
	}

	// String OP String → String
	if a.ttype == tokenTypeString && b.ttype == tokenTypeOperator && c.ttype == tokenTypeString {
		a.value = a.value + b.value + c.value
		return true
	}

	// ( expr ) → expr (unwrap parens for simple expression)
	if a.ttype == tokenTypeLParen && c.ttype == tokenTypeRParen {
		a.ttype = b.ttype
		a.value = b.value
		return true
	}

	// bareword.bareword → bareword (table.column)
	if a.ttype == tokenTypeBareword && b.ttype == tokenTypeOperator && b.value == "." && c.ttype == tokenTypeBareword {
		a.value = a.value + "." + c.value
		return true
	}

	return false
}

// isArithmeticOp returns true for arithmetic operators.
func isArithmeticOp(op string) bool {
	switch op {
	case "+", "-", "*", "/", "%", "^", "<<", ">>", "|", "&":
		return true
	}
	return false
}

// isComparisonOp returns true for comparison operators.
func isComparisonOp(op string) bool {
	switch op {
	case "=", "<", ">", "<=", ">=", "<>", "!=":
		return true
	}
	return false
}

// toUpper converts a string to uppercase (simple ASCII-only).
func toUpper(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}
