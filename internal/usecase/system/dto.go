package system

type SystemSettingsDTO struct {
	Mode                  string `json:"mode"`
	GitTokenHeader        string `json:"git_token_header"`
	LicenseHeader         string `json:"license_header"`
	JSChallengeEnabled    bool   `json:"js_challenge_enabled"`
	JSChallengeMode       string `json:"js_challenge_mode"`
	JSChallengeDifficulty int    `json:"js_challenge_difficulty"`
	WebhookURL            string `json:"webhook_url"`

	// Bot Management
	BotManagementEnabled bool   `json:"bot_management_enabled"`
	ChallengeEscalation  bool   `json:"challenge_escalation"`
	CaptchaProvider      string `json:"captcha_provider"`
	CaptchaSiteKey       string `json:"captcha_site_key"`
	CaptchaSecretKey     string `json:"captcha_secret_key"`
	CaptchaCookieTTL     int    `json:"captcha_cookie_ttl"`

	// Cache
	CacheSyncInterval int `json:"cache_sync_interval"`

	// Semantic Analysis
	SemanticAnalysisEnabled bool `json:"semantic_analysis_enabled"`
}

type UpdateSystemSettingsDTO struct {
	GitTokenHeader        string `json:"git_token_header" binding:"required"`
	LicenseHeader         string `json:"license_header" binding:"required"`
	JSChallengeEnabled    bool   `json:"js_challenge_enabled"`
	JSChallengeMode       string `json:"js_challenge_mode"`
	JSChallengeDifficulty int    `json:"js_challenge_difficulty"`
	WebhookURL            string `json:"webhook_url"`

	// Bot Management
	BotManagementEnabled bool   `json:"bot_management_enabled"`
	ChallengeEscalation  bool   `json:"challenge_escalation"`
	CaptchaProvider      string `json:"captcha_provider"`
	CaptchaSiteKey       string `json:"captcha_site_key"`
	CaptchaSecretKey     string `json:"captcha_secret_key"`
	CaptchaCookieTTL     int    `json:"captcha_cookie_ttl"`

	// Cache
	CacheSyncInterval int `json:"cache_sync_interval"`

	// Semantic Analysis
	SemanticAnalysisEnabled bool `json:"semantic_analysis_enabled"`
}
