package entity

import "time"

type OAuthAuthorizationCode struct {
	ID                  uint      `gorm:"primarykey"`
	Code                string    `gorm:"type:varchar(128);uniqueIndex;not null"`
	ClientID            string    `gorm:"type:varchar(64);index;not null"`
	UserID              uint      `gorm:"index;not null"`
	RedirectURI         string    `gorm:"type:varchar(2048);not null"`
	Scope               string    `gorm:"type:varchar(1000)"`
	Nonce               string    `gorm:"type:varchar(256)"`
	CodeChallenge       string    `gorm:"type:varchar(256)"`
	CodeChallengeMethod string    `gorm:"type:varchar(10)"`
	ExpiresAt           time.Time `gorm:"not null"`
	Used                bool      `gorm:"default:false"`
	CreatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
