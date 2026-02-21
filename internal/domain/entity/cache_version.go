package entity

import "time"

type CacheVersion struct {
	Key       string `gorm:"primaryKey;size:64"`
	Version   int64
	UpdatedAt time.Time
}
