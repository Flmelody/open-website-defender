package entity

import "time"

type User struct {
	ID       uint   `gorm:"primarykey"`
	Username string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password string `gorm:"type:varchar(255);not null"`
	GitToken string `gorm:"type:varchar(300);not null"`
	IsAdmin  bool   `gorm:"type:boolean;default:false"`
	Email    string `gorm:"type:varchar(255)"`

	Scopes      string    `gorm:"type:varchar(1000);default:''"`
	TotpSecret  string    `gorm:"type:varchar(255);default:''"`
	TotpEnabled bool      `gorm:"type:boolean;default:false"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
