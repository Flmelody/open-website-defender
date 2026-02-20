package entity

import "time"

type System struct {
	ID       uint     `gorm:"primarykey"`
	Security Security `json:"security" gorm:"serializer:json;column:security"`
}

type Security struct {
	GitTokenHeader        string `json:"git_token_header"`        // default "Defender-Git-Token"
	LicenseHeader         string `json:"license_header"`          // default "Defender-License"
	JSChallengeEnabled    *bool  `json:"js_challenge_enabled"`    // nil = use config file
	JSChallengeMode       string `json:"js_challenge_mode"`       // off, suspicious, all
	JSChallengeDifficulty int    `json:"js_challenge_difficulty"` // 1-6, 0 = use config file
	WebhookURL            string `json:"webhook_url"`
}

type GeoBlockRule struct {
	ID          uint      `gorm:"primarykey"`
	CountryCode string    `gorm:"type:varchar(10);uniqueIndex;not null"`
	CountryName string    `gorm:"type:varchar(100)"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
