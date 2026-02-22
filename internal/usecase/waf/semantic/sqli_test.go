package semantic

import "testing"

func TestIsSQLi_KnownAttacks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// UNION-based injection
		{"union select basic", "1 UNION SELECT * FROM users", true},
		{"union all select", "1 UNION ALL SELECT password FROM users", true},
		{"union select mixed case", "1 Union Select username from users", true},
		{"union select with number", "1 UNION SELECT 1,2,3", true},
		{"union select function", "1 UNION SELECT version()", true},
		{"union select variable", "1 UNION SELECT @@version", true},
		{"string union select", "' UNION SELECT * FROM users--", true},

		// Boolean injection
		{"or 1=1", "' OR 1=1", true},
		{"or 1=1 with comment", "' OR 1=1--", true},
		{"and 1=1", "' AND 1=1", true},
		{"or string equals", "' OR 'a'='a", true},

		// Stacked queries
		{"drop table", "; DROP TABLE users", true},
		{"alter table", "; ALTER TABLE users", true},
		{"truncate table", "; TRUNCATE TABLE users", true},
		{"semi select", "; SELECT * FROM users", true},

		// Comment injection
		{"string comment", "'--", true},
		{"string space comment", "' -- ", true},

		// Function-based
		{"sleep function", "SLEEP(5)", true},
		{"benchmark function", "BENCHMARK(1000000,MD5('test'))", true},

		// Tautology with comment
		{"tautology comment", "1=1--", true},

		// DDL attacks
		{"drop users", "DROP TABLE users", true},
		{"drop keyword", "DROP TABLE", true},

		// SELECT info
		{"select variable", "SELECT @@version", true},
		{"select from", "SELECT * FROM users", true},
		{"select bareword", "SELECT username FROM", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, fp := IsSQLi(tt.input)
			if got != tt.want {
				t.Errorf("IsSQLi(%q) = (%v, %q), want %v", tt.input, got, fp, tt.want)
			}
		})
	}
}

func TestIsSQLi_FalsePositives(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Known false positives from the plan
		{"Accept header", "Accept: */*"},
		{"Accept header full", "image/avif,image/webp,image/svg+xml,image/*,*/*;q=0.8"},
		{"Accept html", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
		{"credit union select plan", "credit union select plan"},
		{"normal sentence", "please select your options from the menu"},

		// Common legitimate inputs
		{"normal URL", "/api/users/123"},
		{"normal query", "page=1&size=20&sort=name"},
		{"normal JSON", `{"username": "admin", "password": "test123"}`},
		{"email address", "user@example.com"},
		{"normal text", "Hello world, how are you?"},
		{"user agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
		{"number only", "42"},
		{"empty string", ""},
		{"single word", "hello"},
		{"content type", "application/json; charset=utf-8"},
		{"authorization header", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0"},
		{"cookie value", "session_id=abc123def456; theme=dark"},
		{"form data", "name=John+Doe&age=30&city=New+York"},
		{"search query", "golang web framework tutorial 2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, fp := IsSQLi(tt.input)
			if got {
				t.Errorf("IsSQLi(%q) = true (fp=%q), expected false (false positive)", tt.input, fp)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantType []byte // expected token types
	}{
		{
			name:     "simple number",
			input:    "42",
			wantLen:  1,
			wantType: []byte{tokenTypeNumber},
		},
		{
			name:     "string literal",
			input:    "'hello'",
			wantLen:  1,
			wantType: []byte{tokenTypeString},
		},
		{
			name:     "union select",
			input:    "UNION SELECT",
			wantLen:  2,
			wantType: []byte{tokenTypeUnion, tokenTypeSelect},
		},
		{
			name:     "bareword",
			input:    "users",
			wantLen:  1,
			wantType: []byte{tokenTypeBareword},
		},
		{
			name:     "variable",
			input:    "@@version",
			wantLen:  1,
			wantType: []byte{tokenTypeVariable},
		},
		{
			name:     "line comment",
			input:    "-- comment",
			wantLen:  1,
			wantType: []byte{tokenTypeComment},
		},
		{
			name:     "block comment",
			input:    "/* comment */",
			wantLen:  1,
			wantType: []byte{tokenTypeComment},
		},
		{
			name:     "function call",
			input:    "COUNT(",
			wantLen:  1,
			wantType: []byte{tokenTypeFunction},
		},
		{
			name:     "operator",
			input:    "=",
			wantLen:  1,
			wantType: []byte{tokenTypeOperator},
		},
		{
			name:     "semicolon",
			input:    ";",
			wantLen:  1,
			wantType: []byte{tokenTypeSemicolon},
		},
		{
			name:     "hex number",
			input:    "0xFF",
			wantLen:  1,
			wantType: []byte{tokenTypeNumber},
		},
		{
			name:     "float number",
			input:    "3.14",
			wantLen:  1,
			wantType: []byte{tokenTypeNumber},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			if len(tokens) != tt.wantLen {
				t.Errorf("tokenize(%q): got %d tokens, want %d", tt.input, len(tokens), tt.wantLen)
				for i, tok := range tokens {
					t.Logf("  token[%d]: type=%c value=%q", i, tok.ttype, tok.value)
				}
				return
			}
			for i, wantType := range tt.wantType {
				if tokens[i].ttype != wantType {
					t.Errorf("tokenize(%q)[%d]: got type=%c, want type=%c", tt.input, i, tokens[i].ttype, wantType)
				}
			}
		})
	}
}

func TestFold(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantFP string
	}{
		{
			name:   "union all select folds UNION ALL",
			input:  "1 UNION ALL SELECT",
			wantFP: "1UE",
		},
		{
			name:   "arithmetic folding",
			input:  "1+2*3",
			wantFP: "1",
		},
		{
			name:   "unary minus",
			input:  "-1",
			wantFP: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			folded := fold(tokens)
			fp := fingerprint(folded)
			if fp != tt.wantFP {
				t.Errorf("fold(%q): got fingerprint %q, want %q", tt.input, fp, tt.wantFP)
				for i, tok := range folded {
					t.Logf("  folded[%d]: type=%c value=%q", i, tok.ttype, tok.value)
				}
			}
		})
	}
}

func TestFingerprint(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantFP string
	}{
		{"union select", "1 UNION SELECT * FROM users", "1UEok"},
		{"boolean injection", "' OR 1=1", "s&1"},
		{"stacked drop", "; DROP TABLE users", ";Tkn"},
		{"comment", "'--", "sc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			folded := fold(tokens)
			fp := fingerprint(folded)
			if fp != tt.wantFP {
				t.Errorf("fingerprint for %q: got %q, want %q", tt.input, fp, tt.wantFP)
				for i, tok := range folded {
					t.Logf("  folded[%d]: type=%c value=%q", i, tok.ttype, tok.value)
				}
			}
		})
	}
}
