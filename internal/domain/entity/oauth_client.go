package entity

import "time"

type OAuthClient struct {
	ID           uint      `gorm:"primarykey"`
	ClientID     string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	ClientSecret string    `gorm:"type:varchar(255);not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	RedirectURIs string    `gorm:"type:text;not null"`
	Scopes       string    `gorm:"type:varchar(1000);default:'openid profile email'"`
	Trusted      bool      `gorm:"type:boolean;default:false"`
	Active       bool      `gorm:"type:boolean;default:true"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
