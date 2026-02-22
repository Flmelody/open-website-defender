package waf

import "time"

type CreateWafRuleDto struct {
	Name        string `json:"name" binding:"required"`
	Pattern     string `json:"pattern" binding:"required"`
	Category    string `json:"category" binding:"required"` // sqli, xss, traversal, custom
	Action      string `json:"action"`                      // block, log, redirect, challenge, rate-limit (default: block)
	Operator    string `json:"operator"`                    // regex, contains, prefix, suffix, equals, gt, lt (default: regex)
	Target      string `json:"target"`                      // all, url, headers, body, cookies, query (default: all)
	Priority    int    `json:"priority"`
	GroupName   string `json:"group_name"`
	RedirectURL string `json:"redirect_url"`
	RateLimit   int    `json:"rate_limit"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type UpdateWafRuleDto struct {
	Name        string  `json:"name"`
	Pattern     string  `json:"pattern"`
	Category    string  `json:"category"`
	Action      string  `json:"action"`
	Operator    string  `json:"operator"`
	Target      string  `json:"target"`
	Priority    *int    `json:"priority"`
	GroupName   *string `json:"group_name"`
	RedirectURL *string `json:"redirect_url"`
	RateLimit   *int    `json:"rate_limit"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

type WafRuleDto struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Pattern     string    `json:"pattern"`
	Category    string    `json:"category"`
	Action      string    `json:"action"`
	Operator    string    `json:"operator"`
	Target      string    `json:"target"`
	Priority    int       `json:"priority"`
	GroupName   string    `json:"group_name"`
	RedirectURL string    `json:"redirect_url"`
	RateLimit   int       `json:"rate_limit"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type WafCheckResult struct {
	Blocked             bool
	RuleName            string
	Action              string // block, log, redirect, challenge, rate-limit
	RedirectURL         string
	RateLimit           int
	SemanticConfirmed   bool
	SemanticFingerprint string
}

// RequestContext provides all request data for WAF inspection.
type RequestContext struct {
	Method   string
	Path     string
	Query    string
	UA       string
	Body     string
	Headers  map[string]string
	Cookies  map[string]string
	ClientIP string
}

type CreateWafExclusionDto struct {
	RuleID   uint   `json:"rule_id"`
	Path     string `json:"path" binding:"required"`
	Operator string `json:"operator"` // prefix, exact, regex
	Enabled  *bool  `json:"enabled"`
}

type WafExclusionDto struct {
	ID        uint      `json:"id"`
	RuleID    uint      `json:"rule_id"`
	Path      string    `json:"path"`
	Operator  string    `json:"operator"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}
