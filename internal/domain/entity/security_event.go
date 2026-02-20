package entity

import "time"

type SecurityEvent struct {
	ID        uint      `gorm:"primaryKey"`
	EventType string    `gorm:"type:varchar(50);index"`
	ClientIP  string    `gorm:"type:varchar(45);index"`
	Detail    string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"index"`
}
