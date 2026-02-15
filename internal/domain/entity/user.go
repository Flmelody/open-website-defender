package entity

import "time"

type User struct {
	ID        uint      `gorm:"primarykey"`
	Username  string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password  string    `gorm:"type:varchar(255);not null"`
	GitToken  string    `gorm:"type:varchar(300);not null"`
	IsAdmin   bool      `gorm:"type:boolean;default:false"`
	Scopes    string    `gorm:"type:varchar(1000);default:''"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
