package entity

import "time"

type WafRule struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Pattern   string    `gorm:"type:varchar(500);not null"`
	Category  string    `gorm:"type:varchar(50);not null"`               // sqli, xss, traversal, custom
	Action    string    `gorm:"type:varchar(20);not null;default:block"` // block, log
	Enabled   *bool     `gorm:"default:true"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
