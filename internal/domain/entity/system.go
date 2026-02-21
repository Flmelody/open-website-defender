package entity

import "time"

type System struct {
	ID            uint                  `gorm:"primarykey"`
	Security      Security              `json:"security" gorm:"serializer:json;column:security"`
	BotManagement BotManagementSettings `json:"bot_management" gorm:"serializer:json;column:bot_management"`
	CacheSettings CacheSettings         `json:"cache_settings" gorm:"serializer:json;column:cache_settings"`
}

type Security struct {
	GitTokenHeader        string `json:"git_token_header"`        // default "Defender-Git-Token"
	LicenseHeader         string `json:"license_header"`          // default "Defender-License"
	JSChallengeEnabled    *bool  `json:"js_challenge_enabled"`    // nil = use config file
	JSChallengeMode       string `json:"js_challenge_mode"`       // off, suspicious, all
	JSChallengeDifficulty int    `json:"js_challenge_difficulty"` // 1-6, 0 = use config file
	WebhookURL            string `json:"webhook_url"`
}

type BotManagementSettings struct {
	Enabled             *bool  `json:"enabled"`
	ChallengeEscalation *bool  `json:"challenge_escalation"`
	CaptchaProvider     string `json:"captcha_provider"`
	CaptchaSiteKey      string `json:"captcha_site_key"`
	CaptchaSecretKey    string `json:"captcha_secret_key"`
	CaptchaCookieTTL    int    `json:"captcha_cookie_ttl"`
}

type CacheSettings struct {
	SyncInterval *int `json:"sync_interval"` // nil = use config, seconds
}

type GeoBlockRule struct {
	ID          uint      `gorm:"primarykey"`
	CountryCode string    `gorm:"type:varchar(10);uniqueIndex;not null"`
	CountryName string    `gorm:"type:varchar(100)"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
