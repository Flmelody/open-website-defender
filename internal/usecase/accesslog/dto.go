package accesslog

import "time"

type AccessLogDto struct {
	ID             uint      `json:"id"`
	ClientIP       string    `json:"client_ip"`
	Method         string    `json:"method"`
	Host           string    `json:"host,omitempty"`
	Scheme         string    `json:"scheme,omitempty"`
	Path           string    `json:"path"`
	QueryString    string    `json:"query_string,omitempty"`
	ContentType    string    `json:"content_type,omitempty"`
	ContentLength  int64     `json:"content_length"`
	Referer        string    `json:"referer,omitempty"`
	RequestHeaders string    `json:"request_headers,omitempty"`
	RequestBody    string    `json:"request_body,omitempty"`
	StatusCode     int       `json:"status_code"`
	ResponseSize   int       `json:"response_size"`
	Latency        int64     `json:"latency_us"`
	UserAgent      string    `json:"user_agent"`
	Action         string    `json:"action"`
	RuleName       string    `json:"rule_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AccessLogInput struct {
	ClientIP       string
	Method         string
	Host           string
	Scheme         string
	Path           string
	QueryString    string
	ContentType    string
	ContentLength  int64
	Referer        string
	RequestHeaders string // JSON-encoded
	RequestBody    string // truncated
	StatusCode     int
	ResponseSize   int
	Latency        int64
	UserAgent      string
	Action         string
	RuleName       string
}
