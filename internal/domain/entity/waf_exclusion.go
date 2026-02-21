package entity

import "time"

type WafExclusion struct {
	ID        uint      `gorm:"primarykey"`
	RuleID    uint      `gorm:"default:0"` // 0 = applies to all rules
	Path      string    `gorm:"type:varchar(500);not null"`
	Operator  string    `gorm:"type:varchar(20);not null;default:prefix"` // prefix, exact, regex
	Enabled   *bool     `gorm:"default:true"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
