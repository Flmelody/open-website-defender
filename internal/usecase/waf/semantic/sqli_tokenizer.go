package semantic

// sqlToken represents a single token from SQL lexical analysis.
type sqlToken struct {
	ttype byte   // token type character (see constants below)
	value string // original text of the token
}

// Token type constants
const (
	tokenTypeKeyword   byte = 'k' // SQL keyword (FROM, WHERE, INTO, etc.)
	tokenTypeUnion     byte = 'U' // UNION keyword
	tokenTypeSelect    byte = 'E' // SELECT, EXECUTE, etc.
	tokenTypeGroupBy   byte = 'B' // GROUP BY, ORDER BY
	tokenTypeDDL       byte = 'T' // DROP, CREATE, ALTER, TRUNCATE
	tokenTypeString    byte = 's' // string literal
	tokenTypeNumber    byte = '1' // numeric literal
	tokenTypeBareword  byte = 'n' // bareword / identifier
	tokenTypeVariable  byte = 'v' // SQL variable (@var, @@global)
	tokenTypeFunction  byte = 'f' // function call (name followed by open paren)
	tokenTypeOperator  byte = 'o' // operator (=, <>, *, +, -, etc.)
	tokenTypeLogic     byte = '&' // logical operator (AND, OR, NOT, XOR)
	tokenTypeComment   byte = 'c' // comment (--, /* */, #)
	tokenTypeSemicolon byte = ';' // semicolon
	tokenTypeLParen    byte = '(' // left parenthesis
	tokenTypeRParen    byte = ')' // right parenthesis
	tokenTypeComma     byte = ',' // comma
	tokenTypeSQLType   byte = 't' // SQL type (INT, VARCHAR, etc.)
	tokenTypeEvil      byte = 'X' // detected malicious construct
)

// tokenize performs lexical analysis on the input string, producing a token stream.
func tokenize(input string) []sqlToken {
	tokens := make([]sqlToken, 0, 16)
	i := 0
	n := len(input)

	for i < n {
		ch := input[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}

		switch {
		case ch == '\'' || ch == '"':
			tok, end := parseString(input, i)
			tokens = append(tokens, tok)
			i = end

		case ch >= '0' && ch <= '9':
			tok, end := parseNumber(input, i)
			tokens = append(tokens, tok)
			i = end

		case ch == '.' && i+1 < n && input[i+1] >= '0' && input[i+1] <= '9':
			tok, end := parseNumber(input, i)
			tokens = append(tokens, tok)
			i = end

		case isWordStart(ch):
			tok, end := parseWord(input, i)
			// Check if this is a function call (word immediately followed by '(')
			if end < n && input[end] == '(' {
				tok.ttype = tokenTypeFunction
				tok.value = tok.value + "("
				end++ // consume the '('
			}
			tokens = append(tokens, tok)
			i = end

		case ch == '@':
			tok, end := parseVariable(input, i)
			tokens = append(tokens, tok)
			i = end

		case ch == '-':
			if i+1 < n && input[i+1] == '-' {
				tok, end := parseLineComment(input, i)
				tokens = append(tokens, tok)
				i = end
			} else {
				tokens = append(tokens, sqlToken{ttype: tokenTypeOperator, value: "-"})
				i++
			}

		case ch == '/':
			if i+1 < n && input[i+1] == '*' {
				tok, end := parseBlockComment(input, i)
				tokens = append(tokens, tok)
				i = end
			} else {
				tokens = append(tokens, sqlToken{ttype: tokenTypeOperator, value: "/"})
				i++
			}

		case ch == '#':
			tok, end := parseLineComment(input, i)
			tokens = append(tokens, tok)
			i = end

		case ch == ';':
			tokens = append(tokens, sqlToken{ttype: tokenTypeSemicolon, value: ";"})
			i++

		case ch == '(':
			tokens = append(tokens, sqlToken{ttype: tokenTypeLParen, value: "("})
			i++

		case ch == ')':
			tokens = append(tokens, sqlToken{ttype: tokenTypeRParen, value: ")"})
			i++

		case ch == ',':
			tokens = append(tokens, sqlToken{ttype: tokenTypeComma, value: ","})
			i++

		case ch == '`':
			// backtick-quoted identifier
			tok, end := parseBacktickIdent(input, i)
			tokens = append(tokens, tok)
			i = end

		case ch == '~':
			tokens = append(tokens, sqlToken{ttype: tokenTypeOperator, value: "~"})
			i++

		case isOperatorChar(ch):
			tok, end := parseOperator(input, i)
			tokens = append(tokens, tok)
			i = end

		default:
			// Unknown character, skip it
			i++
		}
	}

	return tokens
}

// parseString parses a quoted string starting at position i.
// For SQL injection detection, unterminated strings are treated as
// single-character string tokens so the rest of the input continues
// to be tokenized (attackers deliberately leave quotes unterminated).
func parseString(input string, i int) (sqlToken, int) {
	quote := input[i]
	j := i + 1
	n := len(input)

	for j < n {
		if input[j] == '\\' {
			j += 2 // skip escaped character
			if j > n {
				j = n
			}
			continue
		}
		if input[j] == quote {
			if j+1 < n && input[j+1] == quote {
				// doubled quote as escape
				j += 2
				continue
			}
			j++ // consume closing quote
			return sqlToken{ttype: tokenTypeString, value: input[i:j]}, j
		}
		j++
	}

	// Unterminated string — emit just the opening quote as a string token.
	// This allows the rest of the input to be tokenized normally,
	// which is critical for detecting SQLi patterns like: ' OR 1=1--
	return sqlToken{ttype: tokenTypeString, value: string(quote)}, i + 1
}

// parseNumber parses a numeric literal starting at position i.
func parseNumber(input string, i int) (sqlToken, int) {
	j := i
	n := len(input)

	// Handle 0x (hex) and 0b (binary) prefixes
	if j < n && input[j] == '0' && j+1 < n {
		if input[j+1] == 'x' || input[j+1] == 'X' {
			j += 2
			for j < n && isHexDigit(input[j]) {
				j++
			}
			return sqlToken{ttype: tokenTypeNumber, value: input[i:j]}, j
		}
		if input[j+1] == 'b' || input[j+1] == 'B' {
			j += 2
			for j < n && (input[j] == '0' || input[j] == '1') {
				j++
			}
			return sqlToken{ttype: tokenTypeNumber, value: input[i:j]}, j
		}
	}

	// Integer part
	for j < n && input[j] >= '0' && input[j] <= '9' {
		j++
	}

	// Decimal point
	if j < n && input[j] == '.' && j+1 < n && input[j+1] >= '0' && input[j+1] <= '9' {
		j++
		for j < n && input[j] >= '0' && input[j] <= '9' {
			j++
		}
	}

	// Exponent
	if j < n && (input[j] == 'e' || input[j] == 'E') {
		j++
		if j < n && (input[j] == '+' || input[j] == '-') {
			j++
		}
		for j < n && input[j] >= '0' && input[j] <= '9' {
			j++
		}
	}

	return sqlToken{ttype: tokenTypeNumber, value: input[i:j]}, j
}

// parseWord parses a keyword or identifier starting at position i.
func parseWord(input string, i int) (sqlToken, int) {
	j := i
	n := len(input)

	for j < n && isWordChar(input[j]) {
		j++
	}

	word := input[i:j]
	ttype := lookupKeyword(word)

	return sqlToken{ttype: ttype, value: word}, j
}

// parseVariable parses a SQL variable starting at '@'.
func parseVariable(input string, i int) (sqlToken, int) {
	j := i + 1 // skip initial '@'
	n := len(input)

	// Handle @@ for global variables
	if j < n && input[j] == '@' {
		j++
	}

	for j < n && isWordChar(input[j]) {
		j++
	}

	return sqlToken{ttype: tokenTypeVariable, value: input[i:j]}, j
}

// parseLineComment parses a line comment (-- or #) starting at position i.
func parseLineComment(input string, i int) (sqlToken, int) {
	j := i
	n := len(input)

	for j < n && input[j] != '\n' {
		j++
	}

	return sqlToken{ttype: tokenTypeComment, value: input[i:j]}, j
}

// parseBlockComment parses a /* ... */ comment starting at position i.
func parseBlockComment(input string, i int) (sqlToken, int) {
	j := i + 2 // skip /*
	n := len(input)

	for j < n-1 {
		if input[j] == '*' && input[j+1] == '/' {
			j += 2
			return sqlToken{ttype: tokenTypeComment, value: input[i:j]}, j
		}
		j++
	}

	// Unterminated comment
	return sqlToken{ttype: tokenTypeComment, value: input[i:]}, n
}

// parseBacktickIdent parses a backtick-quoted identifier.
func parseBacktickIdent(input string, i int) (sqlToken, int) {
	j := i + 1
	n := len(input)

	for j < n && input[j] != '`' {
		j++
	}
	if j < n {
		j++ // consume closing backtick
	}

	return sqlToken{ttype: tokenTypeBareword, value: input[i:j]}, j
}

// parseOperator parses an operator token.
func parseOperator(input string, i int) (sqlToken, int) {
	n := len(input)
	ch := input[i]

	// Multi-character operators
	if i+1 < n {
		two := input[i : i+2]
		switch two {
		case "<=", ">=", "<>", "!=", "||", "&&", "<<", ">>", ":=":
			return sqlToken{ttype: tokenTypeOperator, value: two}, i + 2
		}
	}

	return sqlToken{ttype: tokenTypeOperator, value: string(ch)}, i + 1
}

// isWordStart checks if a character can start an identifier/keyword.
func isWordStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isWordChar checks if a character can appear in an identifier/keyword.
func isWordChar(ch byte) bool {
	return isWordStart(ch) || (ch >= '0' && ch <= '9') || ch == '.'
}

// isHexDigit checks if a character is a hexadecimal digit.
func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// isOperatorChar checks if a character is an operator character.
func isOperatorChar(ch byte) bool {
	switch ch {
	case '=', '<', '>', '!', '+', '-', '*', '%', '&', '|', '^', ':', '?':
		return true
	}
	return false
}
