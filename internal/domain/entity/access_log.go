package entity

import "time"

type AccessLog struct {
	ID         uint      `gorm:"primarykey"`
	ClientIP   string    `gorm:"type:varchar(45);index"`
	Method     string    `gorm:"type:varchar(10)"`
	Path       string    `gorm:"type:varchar(500)"`
	StatusCode int       `gorm:"default:0"`
	Latency    int64     `gorm:"default:0"` // microseconds
	UserAgent  string    `gorm:"type:varchar(500)"`
	Action     string    `gorm:"type:varchar(30)"` // allowed, blocked_blacklist, blocked_ratelimit, blocked_waf
	RuleName   string    `gorm:"type:varchar(100)"`
	CreatedAt  time.Time `gorm:"type:timestamp;index;default:CURRENT_TIMESTAMP"`
}
