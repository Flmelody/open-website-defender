package request

type ListAccessLogRequest struct {
	Page       int    `form:"page"`
	Size       int    `form:"size"`
	ClientIP   string `form:"client_ip"`
	Action     string `form:"action"`
	StatusCode int    `form:"status_code"`
	StartTime  string `form:"start_time"` // RFC3339 format
	EndTime    string `form:"end_time"`   // RFC3339 format
}
