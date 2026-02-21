package bot

import "time"

type CreateBotSignatureDto struct {
	Name        string `json:"name" binding:"required"`
	Pattern     string `json:"pattern" binding:"required"`
	MatchTarget string `json:"match_target"`                // ua, header, behavior (default: ua)
	Category    string `json:"category" binding:"required"` // malicious, search_engine, good_bot
	Action      string `json:"action"`                      // allow, block, challenge, monitor (default: block)
	Enabled     *bool  `json:"enabled"`
}

type UpdateBotSignatureDto struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	MatchTarget string `json:"match_target"`
	Category    string `json:"category"`
	Action      string `json:"action"`
	Enabled     *bool  `json:"enabled"`
}

type BotSignatureDto struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Pattern     string    `json:"pattern"`
	MatchTarget string    `json:"match_target"`
	Category    string    `json:"category"`
	Action      string    `json:"action"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type BotCheckResult struct {
	Matched       bool
	SignatureName string
	Category      string // malicious, search_engine, good_bot
	Action        string // allow, block, challenge, monitor
	IsVerified    bool   // true if search engine bot was DNS-verified
}
