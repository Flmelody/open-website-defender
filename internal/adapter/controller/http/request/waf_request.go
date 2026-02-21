package request

type CreateWafRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Pattern     string `json:"pattern" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Action      string `json:"action"`
	Operator    string `json:"operator"`
	Target      string `json:"target"`
	Priority    int    `json:"priority"`
	GroupName   string `json:"group_name"`
	RedirectURL string `json:"redirect_url"`
	RateLimit   int    `json:"rate_limit"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type UpdateWafRuleRequest struct {
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

type ListWafRuleRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

type CreateWafExclusionRequest struct {
	RuleID   uint   `json:"rule_id"`
	Path     string `json:"path" binding:"required"`
	Operator string `json:"operator"`
	Enabled  *bool  `json:"enabled"`
}

type ListWafExclusionRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
