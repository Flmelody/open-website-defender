package system

type SystemSettingsDTO struct {
	GitTokenHeader        string `json:"git_token_header"`
	LicenseHeader         string `json:"license_header"`
	JSChallengeEnabled    bool   `json:"js_challenge_enabled"`
	JSChallengeMode       string `json:"js_challenge_mode"`
	JSChallengeDifficulty int    `json:"js_challenge_difficulty"`
	WebhookURL            string `json:"webhook_url"`
}

type UpdateSystemSettingsDTO struct {
	GitTokenHeader        string `json:"git_token_header" binding:"required"`
	LicenseHeader         string `json:"license_header" binding:"required"`
	JSChallengeEnabled    bool   `json:"js_challenge_enabled"`
	JSChallengeMode       string `json:"js_challenge_mode"`
	JSChallengeDifficulty int    `json:"js_challenge_difficulty"`
	WebhookURL            string `json:"webhook_url"`
}
