package entity

import "time"

type License struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	TokenHash string    `gorm:"type:varchar(64);not null;uniqueIndex"` // SHA-256 hex
	Remark    string    `gorm:"type:varchar(500);default:''"`
	Active    bool      `gorm:"type:boolean;default:true"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
