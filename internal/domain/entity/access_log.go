package entity

import "time"

type AccessLog struct {
	ID             uint      `gorm:"primarykey"`
	ClientIP       string    `gorm:"type:varchar(45);index"`
	Method         string    `gorm:"type:varchar(10)"`
	Host           string    `gorm:"type:varchar(255)"`
	Scheme         string    `gorm:"type:varchar(5)"`
	Path           string    `gorm:"type:varchar(500)"`
	QueryString    string    `gorm:"type:text"`
	ContentType    string    `gorm:"type:varchar(255)"`
	ContentLength  int64     `gorm:"default:0"`
	Referer        string    `gorm:"type:varchar(500)"`
	RequestHeaders string    `gorm:"type:text"` // JSON-encoded
	RequestBody    string    `gorm:"type:text"` // truncated, non-GET only
	StatusCode     int       `gorm:"default:0"`
	ResponseSize   int       `gorm:"default:0"`
	Latency        int64     `gorm:"default:0"` // microseconds
	UserAgent      string    `gorm:"type:varchar(500)"`
	Action         string    `gorm:"type:varchar(30)"` // allowed, blocked_blacklist, blocked_ratelimit, blocked_waf
	RuleName       string    `gorm:"type:varchar(100)"`
	CreatedAt      time.Time `gorm:"type:timestamp;index;default:CURRENT_TIMESTAMP"`
}
