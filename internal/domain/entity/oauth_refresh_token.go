package entity

import "time"

type OAuthRefreshToken struct {
	ID        uint      `gorm:"primarykey"`
	Token     string    `gorm:"type:varchar(128);uniqueIndex;not null"`
	ClientID  string    `gorm:"type:varchar(64);index;not null"`
	UserID    uint      `gorm:"index;not null"`
	Scope     string    `gorm:"type:varchar(1000)"`
	Revoked   bool      `gorm:"default:false"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
