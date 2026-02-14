package waf

import "time"

type CreateWafRuleDto struct {
	Name     string `json:"name" binding:"required"`
	Pattern  string `json:"pattern" binding:"required"`
	Category string `json:"category" binding:"required"` // sqli, xss, traversal, custom
	Action   string `json:"action"`                      // block, log (default: block)
	Enabled  *bool  `json:"enabled"`
}

type UpdateWafRuleDto struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Category string `json:"category"`
	Action   string `json:"action"`
	Enabled  *bool  `json:"enabled"`
}

type WafRuleDto struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Pattern   string    `json:"pattern"`
	Category  string    `json:"category"`
	Action    string    `json:"action"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type WafCheckResult struct {
	Blocked  bool
	RuleName string
	Action   string // block or log
}
