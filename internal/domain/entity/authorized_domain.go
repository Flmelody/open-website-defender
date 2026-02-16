package entity

import "time"

type AuthorizedDomain struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
