package entity

import "time"

type IpWhiteList struct {
	ID        uint       `gorm:"primarykey"`
	Domain    string     `gorm:"type:varchar(100);not null"`
	Ip        string     `gorm:"type:varchar(100);not null"`
	Remark    string     `gorm:"type:varchar(255);default:''"`
	ExpiresAt *time.Time `gorm:"type:datetime;index"`
	CreatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
