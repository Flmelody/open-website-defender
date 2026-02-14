package request

type CreateWafRuleRequest struct {
	Name     string `json:"name" binding:"required"`
	Pattern  string `json:"pattern" binding:"required"`
	Category string `json:"category" binding:"required"`
	Action   string `json:"action"`
	Enabled  *bool  `json:"enabled"`
}

type UpdateWafRuleRequest struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Category string `json:"category"`
	Action   string `json:"action"`
	Enabled  *bool  `json:"enabled"`
}

type ListWafRuleRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
