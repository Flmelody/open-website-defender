package entity

import "time"

type System struct {
	ID       uint     `gorm:"primarykey"`
	Security Security `json:"security" gorm:"serializer:json;column:security"`
}

type Security struct {
	GitTokenSecret string `json:"git_token_secret"`
}

type License struct {
	HttpHeader string `json:"http_header"`
}

type GeoBlockRule struct {
	ID          uint      `gorm:"primarykey"`
	CountryCode string    `gorm:"type:varchar(10);uniqueIndex;not null"`
	CountryName string    `gorm:"type:varchar(100)"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
