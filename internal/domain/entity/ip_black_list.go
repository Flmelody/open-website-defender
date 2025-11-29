package entity

import "time"

type IpBlackList struct {
	ID        uint      `gorm:"primarykey"`
	Ip        string    `gorm:"type:varchar(100);not null"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
