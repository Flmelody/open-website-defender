package semantic

// sqliFingerprints is the set of known SQLi attack fingerprints.
// Each fingerprint is a string of token-type characters (max 5).
// These are derived from analysis of common SQL injection attack patterns.
var sqliFingerprints = map[string]bool{
	// UNION-based injection patterns
	"1UE":   true, // NUM UNION SELECT  (1 UNION SELECT)
	"1UEok": true, // NUM UNION SELECT OP KEYWORD
	"1UEon": true, // NUM UNION SELECT OP BAREWORD
	"1UEo1": true, // NUM UNION SELECT OP NUM
	"1UE1":  true, // NUM UNION SELECT NUM
	"1UEn":  true, // NUM UNION SELECT BAREWORD
	"1UEk":  true, // NUM UNION SELECT KEYWORD
	"1UEf":  true, // NUM UNION SELECT FUNC
	"1UEv":  true, // NUM UNION SELECT VARIABLE
	"1UEs":  true, // NUM UNION SELECT STRING
	"sUE":   true, // STR UNION SELECT
	"sUEok": true, // STR UNION SELECT OP KEYWORD
	"sUEon": true, // STR UNION SELECT OP BAREWORD
	"sUEo1": true, // STR UNION SELECT OP NUM
	"sUE1":  true, // STR UNION SELECT NUM
	"sUEn":  true, // STR UNION SELECT BAREWORD
	"sUEk":  true, // STR UNION SELECT KEYWORD
	"sUEf":  true, // STR UNION SELECT FUNC
	"sUEs":  true, // STR UNION SELECT STRING
	"nUE":   true, // BARE UNION SELECT
	"nUEok": true, // BARE UNION SELECT OP KEYWORD
	"nUEon": true, // BARE UNION SELECT OP BAREWORD
	"nUEo1": true, // BARE UNION SELECT OP NUM
	"nUE1":  true, // BARE UNION SELECT NUM
	"nUEn":  true, // BARE UNION SELECT BAREWORD
	"nUEf":  true, // BARE UNION SELECT FUNC
	"nUEs":  true, // BARE UNION SELECT STRING
	")UE":   true, // RPAREN UNION SELECT
	")UEok": true, // RPAREN UNION SELECT OP KEYWORD
	")UEon": true, // RPAREN UNION SELECT OP BAREWORD
	")UE1":  true, // RPAREN UNION SELECT NUM
	")UEn":  true, // RPAREN UNION SELECT BAREWORD
	")UEf":  true, // RPAREN UNION SELECT FUNC
	"UEokn": true, // UNION SELECT OP KEYWORD BAREWORD (no leading context)
	"UEok":  true, // UNION SELECT OP KEYWORD
	"UEon":  true, // UNION SELECT OP BAREWORD
	"UEo1":  true, // UNION SELECT OP NUM
	"UE1":   true, // UNION SELECT NUM
	"UEn":   true, // UNION SELECT BAREWORD
	"UEk":   true, // UNION SELECT KEYWORD
	"UEf":   true, // UNION SELECT FUNC
	"UEs":   true, // UNION SELECT STRING
	"UE":    true, // UNION SELECT (minimal)
	"UEnk":  true, // UNION SELECT BAREWORD KEYWORD
	"UE1,":  true, // UNION SELECT NUM COMMA
	"UEn,":  true, // UNION SELECT BAREWORD COMMA
	"UEf)":  true, // UNION SELECT FUNC RPAREN
	"1UEnk": true, // NUM UNION SELECT BAREWORD KEYWORD
	"sUEnk": true, // STR UNION SELECT BAREWORD KEYWORD
	"1UE1,": true, // NUM UNION SELECT NUM COMMA
	"1UEf)": true, // NUM UNION SELECT FUNC RPAREN
	"sUEf)": true, // STR UNION SELECT FUNC RPAREN

	// Boolean injection patterns
	"s&1":    true, // STR AND/OR NUM  (' OR 1)
	"s&1o1":  true, // STR AND/OR NUM OP NUM  (' OR 1=1)
	"s&1c":   true, // STR AND/OR NUM COMMENT  (' OR 1--)
	"s&1o1c": true, // STR AND/OR NUM OP NUM COMMENT
	"s&s":    true, // STR AND/OR STR  (' OR '')
	"s&so1":  true, // STR AND/OR STR OP NUM
	"s&sos":  true, // STR AND/OR STR OP STR
	"snsn":   true, // STR BARE STR BARE  (' OR 'a'='a pattern with embedded quotes)
	"sns":    true, // STR BARE STR
	"snso":   true, // STR BARE STR OP
	"s&n":    true, // STR AND/OR BAREWORD
	"s&nk":   true, // STR AND/OR BAREWORD KEYWORD
	"s&v":    true, // STR AND/OR VARIABLE
	"s&f":    true, // STR AND/OR FUNC
	"1&1":    true, // NUM AND/OR NUM
	"1&1o1":  true, // NUM AND/OR NUM OP NUM
	"1&1c":   true, // NUM AND/OR NUM COMMENT
	"n&1o1":  true, // BARE AND/OR NUM OP NUM
	"n&1":    true, // BARE AND/OR NUM

	// Stacked queries / semicolon patterns
	";T":    true, // SEMI DDL  (; DROP)
	";Tn":   true, // SEMI DDL BAREWORD  (; DROP table)
	";Tnk":  true, // SEMI DDL BAREWORD KEYWORD
	";Tkn":  true, // SEMI DDL KEYWORD BAREWORD  (; DROP TABLE users)
	";k":    true, // SEMI KEYWORD  (; DELETE)
	";kn":   true, // SEMI KEYWORD BAREWORD  (; DELETE FROM users)
	";knk":  true, // SEMI KEYWORD BAREWORD KEYWORD
	";E":    true, // SEMI SELECT
	";Eo":   true, // SEMI SELECT OP
	";En":   true, // SEMI SELECT BAREWORD
	";Ek":   true, // SEMI SELECT KEYWORD
	";Ef":   true, // SEMI SELECT FUNC
	";Eokn": true, // SEMI SELECT OP KEYWORD BAREWORD  (; SELECT * FROM users)
	";Eok":  true, // SEMI SELECT OP KEYWORD

	// Comment injection patterns
	"sc":  true, // STR COMMENT  ('-- )
	"1c":  true, // NUM COMMENT
	"nc":  true, // BAREWORD COMMENT  (admin'--)
	"s;c": true, // STR SEMI COMMENT
	"1;c": true, // NUM SEMI COMMENT
	")c":  true, // RPAREN COMMENT
	"s)c": true, // STR RPAREN COMMENT

	// Function-based injection
	"f1)":   true, // FUNC NUM RPAREN  (SLEEP(5))
	"f1)c":  true, // FUNC NUM RPAREN COMMENT
	"fn)":   true, // FUNC BAREWORD RPAREN
	"fn)c":  true, // FUNC BAREWORD RPAREN COMMENT
	"fs)":   true, // FUNC STR RPAREN
	"fs)c":  true, // FUNC STR RPAREN COMMENT
	"ff":    true, // FUNC FUNC  (nested functions)
	"fv)":   true, // FUNC VARIABLE RPAREN
	"f1,":   true, // FUNC NUM COMMA (beginning of function args)
	"f1,fs": true, // FUNC NUM COMMA FUNC STR  (BENCHMARK(n,MD5('x')))
	"f1,f":  true, // FUNC NUM COMMA FUNC

	// DDL patterns
	"Tn":  true, // DDL BAREWORD  (DROP users)
	"Tnk": true, // DDL BAREWORD KEYWORD
	"Tkn": true, // DDL KEYWORD BAREWORD  (DROP TABLE users)
	"Tk":  true, // DDL KEYWORD

	// Variable access patterns (info gathering)
	"Ev":   true, // SELECT VARIABLE
	"Evk":  true, // SELECT VARIABLE KEYWORD
	"Ev,":  true, // SELECT VARIABLE COMMA
	"Efk":  true, // SELECT FUNC KEYWORD
	"Ef":   true, // SELECT FUNC
	"E1k":  true, // SELECT NUM KEYWORD
	"Esk":  true, // SELECT STRING KEYWORD
	"Enk":  true, // SELECT BAREWORD KEYWORD
	"En":   true, // SELECT BAREWORD
	"E1":   true, // SELECT NUM
	"Es":   true, // SELECT STRING
	"Eon":  true, // SELECT OP BAREWORD  (SELECT * FROM)
	"Eok":  true, // SELECT OP KEYWORD
	"Eokn": true, // SELECT OP KEYWORD BAREWORD  (SELECT * FROM users)
	"Eo1":  true, // SELECT OP NUM
	"E1,":  true, // SELECT NUM COMMA
	"En,":  true, // SELECT BAREWORD COMMA
	"Ef,":  true, // SELECT FUNC COMMA

	// Tautology patterns
	"1o1&": true, // NUM OP NUM AND/OR (1=1 AND)
	"1o1;": true, // NUM OP NUM SEMI
	"1o1c": true, // NUM OP NUM COMMENT
	"so1&": true, // STR OP NUM AND/OR
	"sos&": true, // STR OP STR AND/OR
	"so1c": true, // STR OP NUM COMMENT
	"sosc": true, // STR OP STR COMMENT
	"sos;": true, // STR OP STR SEMI

	// Operator-based patterns (common injections)
	"1okn": true, // NUM OP KEYWORD BAREWORD (1=1 FROM users)
	"sokn": true, // STR OP KEYWORD BAREWORD
	"1oE":  true, // NUM OP SELECT (subquery injection)
	"soE":  true, // STR OP SELECT
	"noE":  true, // BARE OP SELECT

	// HAVING / ORDER BY injection
	"Bno1": true, // ORDERBY BAREWORD OP NUM
	"Bnc":  true, // ORDERBY BAREWORD COMMENT
	"Bn;":  true, // ORDERBY BAREWORD SEMI

	// INTO OUTFILE / LOAD patterns
	"ksk": true, // KEYWORD STR KEYWORD (INTO OUTFILE '/tmp/x')
	"ks":  true, // KEYWORD STR

	// Misc dangerous patterns
	"Xn":  true, // EVIL BAREWORD
	"X":   true, // EVIL standalone
	"co1": true, // COMMENT OP NUM (comment then injection)
	"cUE": true, // COMMENT UNION SELECT
	"c;T": true, // COMMENT SEMI DDL
	"cos": true, // COMMENT OP STRING
}

// isKnownFingerprint checks if a fingerprint matches a known SQLi pattern.
func isKnownFingerprint(fp string) bool {
	return sqliFingerprints[fp]
}

// whitelistedContexts contains fingerprints that are benign in certain contexts.
// The key is the fingerprint, the value describes when it should be whitelisted.
// For now, the whitelist check uses token content analysis rather than
// a simple fingerprint match, so this map is used as documentation.
var whitelistedContexts = map[string]string{
	"nnnn": "natural language: 'credit union select plan'",
	"nnn":  "natural language: common phrases",
	"nn":   "natural language: two words",
}

// isWhitelisted checks if a match should be suppressed based on token content.
// This reduces false positives by examining what the tokens actually contain.
func isWhitelisted(_ string, tokens []sqlToken) bool {
	// If all original tokens are "word-like" (barewords or SQL keywords that
	// could also be English words), with no operators, strings, numbers,
	// variables, comments, semicolons, or parens, this is likely natural
	// language rather than SQL injection.
	//
	// Exception: sequences containing DDL keywords (DROP, ALTER, TRUNCATE)
	// or that start with SELECT/EXEC are not whitelisted even as pure words,
	// because these form valid SQL statements.
	allWords := true
	hasDDL := false
	startsWithSQL := false
	for i, t := range tokens {
		switch t.ttype {
		case tokenTypeBareword, tokenTypeKeyword, tokenTypeUnion,
			tokenTypeGroupBy, tokenTypeLogic, tokenTypeSQLType:
			continue
		case tokenTypeSelect:
			if i == 0 {
				startsWithSQL = true
			}
			continue
		case tokenTypeDDL:
			hasDDL = true
			continue
		default:
			allWords = false
		}
		if !allWords {
			break
		}
	}
	if allWords && !hasDDL && !startsWithSQL {
		return true
	}

	return false
}
