package entity

import "time"

type WafRule struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Pattern     string    `gorm:"type:varchar(500);not null"`
	Category    string    `gorm:"type:varchar(50);not null"`               // sqli, xss, traversal, custom
	Action      string    `gorm:"type:varchar(20);not null;default:block"` // block, log, redirect, challenge, rate-limit
	Operator    string    `gorm:"type:varchar(20);not null;default:regex"` // regex, contains, prefix, suffix, equals, gt, lt
	Target      string    `gorm:"type:varchar(30);not null;default:all"`   // all, url, headers, body, cookies, query, response_body, response_headers
	Priority    int       `gorm:"default:100"`                             // lower = higher priority
	GroupName   string    `gorm:"type:varchar(100)"`
	RedirectURL string    `gorm:"type:varchar(500)"` // for action=redirect
	RateLimit   int       `gorm:"default:0"`         // for action=rate-limit, requests/min
	Description string    `gorm:"type:varchar(500)"`
	Enabled     *bool     `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
