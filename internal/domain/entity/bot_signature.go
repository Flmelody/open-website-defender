package entity

import "time"

type BotSignature struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Pattern     string    `gorm:"type:varchar(500);not null"`              // regex pattern
	MatchTarget string    `gorm:"type:varchar(20);not null;default:ua"`    // ua, header, behavior
	Category    string    `gorm:"type:varchar(30);not null"`               // malicious, search_engine, good_bot
	Action      string    `gorm:"type:varchar(20);not null;default:block"` // allow, block, challenge, monitor
	Enabled     *bool     `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
