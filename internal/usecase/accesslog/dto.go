package accesslog

import "time"

type AccessLogDto struct {
	ID         uint      `json:"id"`
	ClientIP   string    `json:"client_ip"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	Latency    int64     `json:"latency_us"` // microseconds
	UserAgent  string    `json:"user_agent"`
	Action     string    `json:"action"`
	RuleName   string    `json:"rule_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AccessLogInput struct {
	ClientIP   string
	Method     string
	Path       string
	StatusCode int
	Latency    int64 // microseconds
	UserAgent  string
	Action     string // allowed, blocked_blacklist, blocked_ratelimit, blocked_waf
	RuleName   string
}
