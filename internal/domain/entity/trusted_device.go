package entity

import "time"

type TrustedDevice struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"index;not null"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
