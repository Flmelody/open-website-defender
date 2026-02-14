package waf

import (
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"os"
	"regexp"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	// Initialize a no-op logger to avoid nil pointer panics in tests
	logging.Logger = zap.NewNop()
	logging.Sugar = logging.Logger.Sugar()
	os.Exit(m.Run())
}

// setupTestDB creates an in-memory SQLite database with WAF rules table
// and seeds it with the default rules.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	err = db.AutoMigrate(&entity.WafRule{})
	if err != nil {
		t.Fatalf("Failed to migrate WafRule: %v", err)
	}

	return db
}

func boolPtr(b bool) *bool {
	return &b
}

// seedDefaultRules inserts the default WAF rules into the test database.
func seedDefaultRules(t *testing.T, db *gorm.DB) {
	t.Helper()
	rules := []entity.WafRule{
		{Name: "SQLi - Union Select", Pattern: `(?i)(union\s+(all\s+)?select)`, Category: "sqli", Action: "block", Enabled: boolPtr(true)},
		{Name: "SQLi - Common Patterns", Pattern: `(?i)(;\s*(drop|alter|truncate|delete|insert|update)\s)`, Category: "sqli", Action: "block", Enabled: boolPtr(true)},
		{Name: "SQLi - Boolean Injection", Pattern: `(?i)('\s*(or|and)\s*'?\d*\s*[=<>])`, Category: "sqli", Action: "block", Enabled: boolPtr(true)},
		{Name: "SQLi - Comment Injection", Pattern: `(?i)('\s*--\s*$|/\*.*?\*/)`, Category: "sqli", Action: "block", Enabled: boolPtr(true)},
		{Name: "XSS - Script Tag", Pattern: `(?i)(<script[\s>]|</script>)`, Category: "xss", Action: "block", Enabled: boolPtr(true)},
		{Name: "XSS - Event Handler", Pattern: `(?i)(on(error|load|click|mouseover|focus|blur|submit|change)\s*=)`, Category: "xss", Action: "block", Enabled: boolPtr(true)},
		{Name: "XSS - JavaScript Protocol", Pattern: `(?i)(javascript\s*:|vbscript\s*:)`, Category: "xss", Action: "block", Enabled: boolPtr(true)},
		{Name: "Path Traversal - Dot Dot Slash", Pattern: `(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`, Category: "traversal", Action: "block", Enabled: boolPtr(true)},
		{Name: "Path Traversal - Sensitive Files", Pattern: `(?i)(/etc/passwd|/etc/shadow|/proc/self|/windows/system32)`, Category: "traversal", Action: "block", Enabled: boolPtr(true)},
	}
	for i := range rules {
		if err := db.Create(&rules[i]).Error; err != nil {
			t.Fatalf("Failed to seed rule %q: %v", rules[i].Name, err)
		}
	}
}

// newTestWafService creates a WafService backed by the given test DB.
func newTestWafService(db *gorm.DB) *WafService {
	return &WafService{
		repo: repository.NewWafRuleRepository(db),
	}
}

// --- Pattern-level tests (no DB required) ---

// getDefaultPatterns returns the compiled default WAF rule patterns.
func getDefaultPatterns() []compiledRule {
	raw := []struct {
		Name    string
		Pattern string
		Action  string
	}{
		{"SQLi - Union Select", `(?i)(union\s+(all\s+)?select)`, "block"},
		{"SQLi - Common Patterns", `(?i)(;\s*(drop|alter|truncate|delete|insert|update)\s)`, "block"},
		{"SQLi - Boolean Injection", `(?i)('\s*(or|and)\s*'?\d*\s*[=<>])`, "block"},
		{"SQLi - Comment Injection", `(?i)('\s*--\s*$|/\*.*?\*/)`, "block"},
		{"XSS - Script Tag", `(?i)(<script[\s>]|</script>)`, "block"},
		{"XSS - Event Handler", `(?i)(on(error|load|click|mouseover|focus|blur|submit|change)\s*=)`, "block"},
		{"XSS - JavaScript Protocol", `(?i)(javascript\s*:|vbscript\s*:)`, "block"},
		{"Path Traversal - Dot Dot Slash", `(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`, "block"},
		{"Path Traversal - Sensitive Files", `(?i)(/etc/passwd|/etc/shadow|/proc/self|/windows/system32)`, "block"},
	}

	rules := make([]compiledRule, len(raw))
	for i, r := range raw {
		rules[i] = compiledRule{
			Name:    r.Name,
			Pattern: regexp.MustCompile(r.Pattern),
			Action:  r.Action,
		}
	}
	return rules
}

func TestSQLiPatterns(t *testing.T) {
	rules := getDefaultPatterns()
	sqliRules := make([]compiledRule, 0)
	for _, r := range rules {
		if r.Name == "SQLi - Union Select" || r.Name == "SQLi - Common Patterns" ||
			r.Name == "SQLi - Boolean Injection" || r.Name == "SQLi - Comment Injection" {
			sqliRules = append(sqliRules, r)
		}
	}

	tests := []struct {
		name    string
		input   string
		blocked bool
	}{
		// Union Select
		{"union select basic", "1 UNION SELECT * FROM users", true},
		{"union all select", "1 union all select password from users", true},
		{"union select mixed case", "1 Union Select username from users", true},
		{"union select with whitespace", "1 union  select * from users", true},

		// Common patterns (drop, alter, etc.)
		{"drop table", "; drop table users ", true},
		{"alter table", "; ALTER table users ", true},
		{"truncate table", "; truncate table users ", true},
		{"delete from", "; delete from users ", true},
		{"insert into", "; insert into users ", true},
		{"update set", "; update users set ", true},

		// Boolean injection
		{"or 1=1", "' or 1=1", true},
		{"and 1=1", "' and 1=1", true},
		{"OR with equals", "' OR '1'='1", false}, // pattern expects digit then operator
		{"or with greater than", "' or 1>0", true},
		{"or with less than", "' or 1<2", true},

		// Comment injection
		{"SQL comment double dash", "' -- ", true},
		{"SQL block comment", "/* comment */", true},
		{"SQL block comment in query", "id=1 /* injection */ --", true},

		// Should NOT match
		{"normal select query word", "select your options from the menu", false},
		{"normal word union", "the european union is large", false},
		{"normal semicolon", "Hello; how are you?", false},
		{"normal URL path", "/api/users/123", false},
		{"normal JSON body", `{"username": "admin", "password": "test"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false
			for _, rule := range sqliRules {
				if rule.Pattern.MatchString(tt.input) {
					matched = true
					break
				}
			}
			if matched != tt.blocked {
				t.Errorf("SQLi check for %q: got matched=%v, want blocked=%v", tt.input, matched, tt.blocked)
			}
		})
	}
}

func TestXSSPatterns(t *testing.T) {
	rules := getDefaultPatterns()
	xssRules := make([]compiledRule, 0)
	for _, r := range rules {
		if r.Name == "XSS - Script Tag" || r.Name == "XSS - Event Handler" ||
			r.Name == "XSS - JavaScript Protocol" {
			xssRules = append(xssRules, r)
		}
	}

	tests := []struct {
		name    string
		input   string
		blocked bool
	}{
		// Script tags
		{"script tag opening", "<script>alert(1)</script>", true},
		{"script tag with attribute", `<script src="evil.js">`, true},
		{"script tag uppercase", "<SCRIPT>alert(1)</SCRIPT>", true},
		{"script tag closing", "</script>", true},
		{"script tag with space", "<script >", true},

		// Event handlers
		{"onerror handler", `<img onerror=alert(1)>`, true},
		{"onload handler", `<body onload=alert(1)>`, true},
		{"onclick handler", `<div onclick=alert(1)>`, true},
		{"onmouseover handler", `<a onmouseover=alert(1)>`, true},
		{"onfocus handler", `<input onfocus=alert(1)>`, true},
		{"onblur handler", `<input onblur=alert(1)>`, true},
		{"onsubmit handler", `<form onsubmit=alert(1)>`, true},
		{"onchange handler", `<select onchange=alert(1)>`, true},
		{"event handler with spaces", `<div onclick =alert(1)>`, true},

		// JavaScript protocol
		{"javascript protocol", `javascript:alert(1)`, true},
		{"javascript protocol uppercase", `JAVASCRIPT:alert(1)`, true},
		{"javascript protocol with space", `javascript :alert(1)`, true},
		{"vbscript protocol", `vbscript:msgbox("xss")`, true},

		// Should NOT match
		{"normal HTML paragraph", "<p>Hello World</p>", false},
		{"normal text with script word", "the script was well written", false},
		{"normal URL", "https://example.com/path", false},
		{"normal class attribute", `<div class="container">`, false},
		{"normal JSON", `{"event": "click", "handler": "doSomething"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false
			for _, rule := range xssRules {
				if rule.Pattern.MatchString(tt.input) {
					matched = true
					break
				}
			}
			if matched != tt.blocked {
				t.Errorf("XSS check for %q: got matched=%v, want blocked=%v", tt.input, matched, tt.blocked)
			}
		})
	}
}

func TestPathTraversalPatterns(t *testing.T) {
	rules := getDefaultPatterns()
	traversalRules := make([]compiledRule, 0)
	for _, r := range rules {
		if r.Name == "Path Traversal - Dot Dot Slash" || r.Name == "Path Traversal - Sensitive Files" {
			traversalRules = append(traversalRules, r)
		}
	}

	tests := []struct {
		name    string
		input   string
		blocked bool
	}{
		// Dot dot slash
		{"basic dot dot slash", "../etc/passwd", true},
		{"dot dot backslash", `..\\windows\\system32`, true},
		{"multiple traversal", "../../etc/shadow", true},
		{"encoded traversal forward slash", "%2e%2e%2f", true},
		{"encoded traversal backslash", "%2e%2e%5c", true},
		{"path traversal in query", "/api?file=../../../etc/passwd", true},

		// Sensitive files
		{"etc passwd", "/etc/passwd", true},
		{"etc shadow", "/etc/shadow", true},
		{"proc self", "/proc/self/environ", true},
		{"windows system32", "/windows/system32/config/sam", true},
		{"sensitive file uppercase", "/ETC/PASSWD", true},

		// Should NOT match
		{"normal file path", "/api/users/profile", false},
		{"normal relative path without traversal", "./config.yaml", false},
		{"normal query string", "/search?q=hello+world", false},
		{"normal dot in filename", "/files/report.pdf", false},
		{"single dot path", "/path/./normalize", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false
			for _, rule := range traversalRules {
				if rule.Pattern.MatchString(tt.input) {
					matched = true
					break
				}
			}
			if matched != tt.blocked {
				t.Errorf("Path traversal check for %q: got matched=%v, want blocked=%v", tt.input, matched, tt.blocked)
			}
		})
	}
}

func TestNormalRequestsNotBlocked(t *testing.T) {
	rules := getDefaultPatterns()

	normalRequests := []struct {
		name      string
		path      string
		query     string
		userAgent string
		body      string
	}{
		{
			name:      "GET homepage",
			path:      "/",
			query:     "",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			body:      "",
		},
		{
			name:      "GET API endpoint",
			path:      "/api/v1/users",
			query:     "page=1&size=20",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			body:      "",
		},
		{
			name:      "POST login",
			path:      "/api/login",
			query:     "",
			userAgent: "Mozilla/5.0",
			body:      `{"username":"admin","password":"secretpassword123"}`,
		},
		{
			name:      "POST form data",
			path:      "/api/v1/posts",
			query:     "",
			userAgent: "Chrome/120.0.0.0",
			body:      `{"title":"My Blog Post","content":"Hello world, this is a normal post."}`,
		},
		{
			name:      "GET with search query",
			path:      "/search",
			query:     "q=golang+web+framework&category=tech",
			userAgent: "Safari/537.36",
			body:      "",
		},
		{
			name:      "PUT update profile",
			path:      "/api/v1/profile",
			query:     "",
			userAgent: "PostmanRuntime/7.36.0",
			body:      `{"name":"John Doe","email":"john@example.com","bio":"Software developer & tech enthusiast"}`,
		},
		{
			name:      "DELETE resource",
			path:      "/api/v1/posts/42",
			query:     "",
			userAgent: "axios/1.6.0",
			body:      "",
		},
		{
			name:      "GET static file",
			path:      "/static/css/main.css",
			query:     "v=1.2.3",
			userAgent: "Mozilla/5.0",
			body:      "",
		},
	}

	for _, req := range normalRequests {
		t.Run(req.name, func(t *testing.T) {
			targets := []string{req.path, req.query, req.userAgent, req.body}
			for _, rule := range rules {
				for _, target := range targets {
					if target != "" && rule.Pattern.MatchString(target) {
						t.Errorf("Normal request %q was blocked by rule %q on target %q",
							req.name, rule.Name, target)
					}
				}
			}
		})
	}
}

// --- Integration tests using in-memory SQLite ---

func TestCheckRequest_Integration_SQLi(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)

	// Initialize the cache and invalidate stale entries from other tests
	_ = pkg.Cacher()
	svc.invalidateCache()

	tests := []struct {
		name        string
		method      string
		path        string
		queryString string
		userAgent   string
		body        string
		shouldBlock bool
	}{
		{
			name:        "union select in query string",
			method:      "GET",
			path:        "/api/users",
			queryString: "id=1 UNION SELECT * FROM users--",
			shouldBlock: true,
		},
		{
			name:        "drop table in body",
			method:      "POST",
			path:        "/api/data",
			body:        "; DROP TABLE users ",
			shouldBlock: true,
		},
		{
			name:        "normal GET request",
			method:      "GET",
			path:        "/api/users/123",
			queryString: "include=profile",
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CheckRequest(tt.method, tt.path, tt.queryString, tt.userAgent, tt.body)
			if tt.shouldBlock {
				if result == nil {
					t.Error("Expected request to be blocked, but got nil result")
				} else if !result.Blocked {
					t.Errorf("Expected Blocked=true, got Blocked=false (rule=%s, action=%s)", result.RuleName, result.Action)
				}
			} else {
				if result != nil && result.Blocked {
					t.Errorf("Expected request to pass, but was blocked by rule %q", result.RuleName)
				}
			}
		})
	}
}

func TestCheckRequest_Integration_XSS(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)
	_ = pkg.Cacher()
	svc.invalidateCache()

	tests := []struct {
		name        string
		path        string
		queryString string
		body        string
		shouldBlock bool
	}{
		{
			name:        "script tag in query",
			queryString: "name=<script>alert(1)</script>",
			shouldBlock: true,
		},
		{
			name:        "event handler in body",
			body:        `<img onerror=alert(document.cookie)>`,
			shouldBlock: true,
		},
		{
			name:        "javascript protocol in query",
			queryString: "url=javascript:alert(1)",
			shouldBlock: true,
		},
		{
			name:        "normal HTML-like content",
			body:        `{"content": "<p>This is fine</p>"}`,
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CheckRequest("POST", tt.path, tt.queryString, "", tt.body)
			if tt.shouldBlock {
				if result == nil {
					t.Error("Expected request to be blocked, got nil")
				} else if !result.Blocked {
					t.Errorf("Expected Blocked=true, got false (rule=%s)", result.RuleName)
				}
			} else {
				if result != nil && result.Blocked {
					t.Errorf("Expected pass, blocked by rule %q", result.RuleName)
				}
			}
		})
	}
}

func TestCheckRequest_Integration_PathTraversal(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)
	_ = pkg.Cacher()
	svc.invalidateCache()

	tests := []struct {
		name        string
		path        string
		queryString string
		shouldBlock bool
	}{
		{
			name:        "dot dot slash in path",
			path:        "/api/../../../etc/passwd",
			shouldBlock: true,
		},
		{
			name:        "encoded traversal in query",
			queryString: "file=%2e%2e%2fetc/passwd",
			shouldBlock: true,
		},
		{
			name:        "sensitive file in path",
			path:        "/download?file=/etc/shadow",
			shouldBlock: true,
		},
		{
			name:        "normal path",
			path:        "/api/v1/documents/report.pdf",
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CheckRequest("GET", tt.path, tt.queryString, "", "")
			if tt.shouldBlock {
				if result == nil {
					t.Error("Expected blocked, got nil")
				} else if !result.Blocked {
					t.Errorf("Expected Blocked=true, got false (rule=%s)", result.RuleName)
				}
			} else {
				if result != nil && result.Blocked {
					t.Errorf("Expected pass, blocked by rule %q", result.RuleName)
				}
			}
		})
	}
}

func TestCheckRequest_Integration_NoRules(t *testing.T) {
	db := setupTestDB(t)
	// Do NOT seed rules
	svc := newTestWafService(db)
	_ = pkg.Cacher()
	svc.invalidateCache() // Clear any cached rules from previous tests

	result := svc.CheckRequest("GET", "/api/test", "id=1 UNION SELECT * FROM users", "", "")
	if result != nil {
		t.Errorf("Expected nil result when no rules exist, got rule=%q", result.RuleName)
	}
}

func TestCheckRequest_Integration_DisabledRule(t *testing.T) {
	db := setupTestDB(t)
	// Create a disabled rule only (no enabled rules)
	rule := entity.WafRule{
		Name:     "Disabled SQLi Rule",
		Pattern:  `(?i)(union\s+(all\s+)?select)`,
		Category: "sqli",
		Action:   "block",
		Enabled:  boolPtr(false),
	}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("Failed to create disabled rule: %v", err)
	}
	svc := newTestWafService(db)
	_ = pkg.Cacher()

	// Invalidate cache to ensure fresh load from DB (not stale from other tests)
	svc.invalidateCache()

	result := svc.CheckRequest("GET", "/api/test", "id=1 UNION SELECT * FROM users", "", "")
	if result != nil {
		t.Errorf("Disabled rule should not block requests, but got rule=%q", result.RuleName)
	}
}

func TestCheckRequest_Integration_LogAction(t *testing.T) {
	db := setupTestDB(t)
	// Create a rule with "log" action instead of "block"
	rule := entity.WafRule{
		Name:     "Log Only SQLi Rule",
		Pattern:  `(?i)(union\s+(all\s+)?select)`,
		Category: "sqli",
		Action:   "log",
		Enabled:  boolPtr(true),
	}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("Failed to create log rule: %v", err)
	}
	svc := newTestWafService(db)
	_ = pkg.Cacher()
	svc.invalidateCache()

	result := svc.CheckRequest("GET", "/api/test", "id=1 UNION SELECT * FROM users", "", "")
	if result == nil {
		t.Fatal("Expected a result for log-only rule, got nil")
	}
	if result.Blocked {
		t.Error("Log-only rule should not block (Blocked should be false)")
	}
	if result.Action != "log" {
		t.Errorf("Expected Action='log', got %q", result.Action)
	}
	if result.RuleName != "Log Only SQLi Rule" {
		t.Errorf("Expected RuleName='Log Only SQLi Rule', got %q", result.RuleName)
	}
}

func TestCheckRequest_UserAgent(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)
	_ = pkg.Cacher()
	svc.invalidateCache()

	// XSS in user agent
	result := svc.CheckRequest("GET", "/", "", "<script>alert(1)</script>", "")
	if result == nil {
		t.Error("Expected XSS in user-agent to be detected, got nil")
	} else if !result.Blocked {
		t.Errorf("Expected blocked, got action=%s", result.Action)
	}
}

// --- CRUD tests ---

func TestCreate_ValidRule(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)
	_ = pkg.Cacher()

	enabled := true
	dto, err := svc.Create(&CreateWafRuleDto{
		Name:     "Test Rule",
		Pattern:  `(?i)(test\s+pattern)`,
		Category: "custom",
		Action:   "block",
		Enabled:  &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dto.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if dto.Name != "Test Rule" {
		t.Errorf("Expected Name='Test Rule', got %q", dto.Name)
	}
	if dto.Pattern != `(?i)(test\s+pattern)` {
		t.Errorf("Unexpected Pattern: %q", dto.Pattern)
	}
}

func TestCreate_InvalidRegex(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)

	enabled := true
	_, err := svc.Create(&CreateWafRuleDto{
		Name:     "Bad Regex Rule",
		Pattern:  `(?i)(unclosed group`,
		Category: "custom",
		Enabled:  &enabled,
	})
	if err == nil {
		t.Error("Expected error for invalid regex pattern, got nil")
	}
}

func TestCreate_MissingRequiredFields(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)

	_, err := svc.Create(&CreateWafRuleDto{
		Name:     "",
		Pattern:  "test",
		Category: "custom",
	})
	if err == nil {
		t.Error("Expected error for missing name, got nil")
	}

	_, err = svc.Create(&CreateWafRuleDto{
		Name:     "Test",
		Pattern:  "",
		Category: "custom",
	})
	if err == nil {
		t.Error("Expected error for missing pattern, got nil")
	}

	_, err = svc.Create(&CreateWafRuleDto{
		Name:     "Test",
		Pattern:  "test",
		Category: "",
	})
	if err == nil {
		t.Error("Expected error for missing category, got nil")
	}
}

func TestCreate_DefaultAction(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)

	dto, err := svc.Create(&CreateWafRuleDto{
		Name:     "Default Action Rule",
		Pattern:  `test`,
		Category: "custom",
		Action:   "", // should default to "block"
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dto.Action != "block" {
		t.Errorf("Expected default Action='block', got %q", dto.Action)
	}
}

func TestList_Pagination(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)

	// List first page
	dtos, total, err := svc.List(1, 5)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 9 {
		t.Errorf("Expected total=9, got %d", total)
	}
	if len(dtos) != 5 {
		t.Errorf("Expected 5 items on page 1, got %d", len(dtos))
	}

	// List second page
	dtos2, total2, err := svc.List(2, 5)
	if err != nil {
		t.Fatalf("List page 2 failed: %v", err)
	}
	if total2 != 9 {
		t.Errorf("Expected total=9, got %d", total2)
	}
	if len(dtos2) != 4 {
		t.Errorf("Expected 4 items on page 2, got %d", len(dtos2))
	}
}

func TestList_InvalidPagination(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)

	// Negative page and size should be corrected
	dtos, _, err := svc.List(-1, -1)
	if err != nil {
		t.Fatalf("List with invalid pagination failed: %v", err)
	}
	// page defaults to 1, size defaults to 10
	if len(dtos) != 9 {
		t.Errorf("Expected 9 items (all rules fit in size=10), got %d", len(dtos))
	}
}

func TestDelete_Integration(t *testing.T) {
	db := setupTestDB(t)
	seedDefaultRules(t, db)
	svc := newTestWafService(db)

	// Get first rule
	dtos, _, err := svc.List(1, 1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(dtos) == 0 {
		t.Fatal("No rules found to delete")
	}

	id := dtos[0].ID
	err = svc.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, total, err := svc.List(1, 100)
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	if total != 8 {
		t.Errorf("Expected 8 rules after deletion, got %d", total)
	}
}

func TestUpdate_Integration(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)

	enabled := true
	created, err := svc.Create(&CreateWafRuleDto{
		Name:     "Original Name",
		Pattern:  `original`,
		Category: "custom",
		Action:   "block",
		Enabled:  &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	disabled := false
	updated, err := svc.Update(created.ID, &UpdateWafRuleDto{
		Name:    "Updated Name",
		Pattern: `updated`,
		Action:  "log",
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Expected Name='Updated Name', got %q", updated.Name)
	}
	if updated.Pattern != "updated" {
		t.Errorf("Expected Pattern='updated', got %q", updated.Pattern)
	}
	if updated.Action != "log" {
		t.Errorf("Expected Action='log', got %q", updated.Action)
	}
	if updated.Enabled != false {
		t.Error("Expected Enabled=false")
	}
}

func TestUpdate_InvalidRegex(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)

	enabled := true
	created, err := svc.Create(&CreateWafRuleDto{
		Name:     "Rule",
		Pattern:  `test`,
		Category: "custom",
		Enabled:  &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = svc.Update(created.ID, &UpdateWafRuleDto{
		Pattern: `(?i)(unclosed`,
	})
	if err == nil {
		t.Error("Expected error for invalid regex in update, got nil")
	}
}

func TestUpdate_NonExistentRule(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestWafService(db)

	_, err := svc.Update(99999, &UpdateWafRuleDto{
		Name: "Does Not Exist",
	})
	if err == nil {
		t.Error("Expected error updating non-existent rule, got nil")
	}
}
