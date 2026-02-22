package config

type Config struct {
	Mode             string                 `mapstructure:"mode"` // "auth_request" (default) | "reverse_proxy"
	Database         DatabaseConfig         `mapstructure:"database"`
	Cache            CacheConfig            `mapstructure:"cache"`
	Security         SecurityConfig         `mapstructure:"security"`
	RateLimit        RateLimitConfig        `mapstructure:"rate-limit"`
	RequestFiltering RequestFilteringConfig `mapstructure:"request-filtering"`
	Server           ServerConfig           `mapstructure:"server"`
	Whitelist        WhitelistConfig        `mapstructure:"whitelist" json:"-"`
	Blacklist        BlacklistConfig        `mapstructure:"blacklist" json:"-"`
	Wall             AppConfig              `mapstructure:"wall"`
	OAuth            OAuthConfig            `mapstructure:"oauth"`
	ThreatDetection  ThreatDetectionConfig  `mapstructure:"threat-detection"`
	JSChallenge      JSChallengeConfig      `mapstructure:"js-challenge"`
	Webhook          WebhookConfig          `mapstructure:"webhook"`
	BotManagement    BotManagementConfig    `mapstructure:"bot-management"`
}

type CacheConfig struct {
	SizeMB       int `mapstructure:"size-mb"`       // Maximum memory in MB (default: 100)
	SyncInterval int `mapstructure:"sync-interval"` // Multi-instance sync polling interval in seconds (default: 0 = disabled)
}

type OAuthConfig struct {
	Enabled                   bool   `mapstructure:"enabled"`
	Issuer                    string `mapstructure:"issuer"`
	RSAPrivateKeyPath         string `mapstructure:"rsa-private-key-path"`
	AuthorizationCodeLifetime int    `mapstructure:"authorization-code-lifetime"`
	AccessTokenLifetime       int    `mapstructure:"access-token-lifetime"`
	RefreshTokenLifetime      int    `mapstructure:"refresh-token-lifetime"`
	IDTokenLifetime           int    `mapstructure:"id-token-lifetime"`
}

type RequestFilteringConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type ServerConfig struct {
	MaxBodySizeMB int64 `mapstructure:"max-body-size-mb"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl-mode"`
	FilePath string `mapstructure:"file-path"`
}

type SecurityConfig struct {
	JWTSecret              string       `mapstructure:"jwt-secret"`
	TokenExpirationHrs     int          `mapstructure:"token-expiration-hours"`
	AdminRecoveryKey       string       `mapstructure:"admin-recovery-key"`
	AdminRecoveryLocalOnly bool         `mapstructure:"admin-recovery-local-only"`
	TrustedDeviceDays      int          `mapstructure:"trusted-device-days"`
	CORS                   CORSConfig   `mapstructure:"cors"`
	Headers                HeaderConfig `mapstructure:"headers"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed-origins"`
	AllowCredentials bool     `mapstructure:"allow-credentials"`
}

type HeaderConfig struct {
	HSTSEnabled  bool   `mapstructure:"hsts-enabled"`
	FrameOptions string `mapstructure:"frame-options"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests-per-minute"`
	Login             struct {
		RequestsPerMinute int `mapstructure:"requests-per-minute"`
		LockoutDuration   int `mapstructure:"lockout-duration"`
	} `mapstructure:"login"`
}

type WhitelistConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	IPs     []string `mapstructure:"ips"`
}

type BlacklistConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	IPs     []string `mapstructure:"ips"`
}

type AppConfig struct {
	BaseURL   string `mapstructure:"backend-host" json:"baseURL"`
	RootPath  string `mapstructure:"root-path" json:"rootPath"`
	AdminPath string `mapstructure:"admin-path" json:"adminPath"`
	GuardPath string `mapstructure:"guard-path" json:"guardPath"`
}

type ThreatDetectionConfig struct {
	Enabled                 bool `mapstructure:"enabled"`
	StatusCodeThreshold     int  `mapstructure:"status-code-threshold"`
	StatusCodeWindow        int  `mapstructure:"status-code-window"`
	RateLimitAbuseThreshold int  `mapstructure:"rate-limit-abuse-threshold"`
	RateLimitAbuseWindow    int  `mapstructure:"rate-limit-abuse-window"`
	AutoBanDuration         int  `mapstructure:"auto-ban-duration"`
	ScanThreshold           int  `mapstructure:"scan-threshold"`
	ScanWindow              int  `mapstructure:"scan-window"`
	ScanBanDuration         int  `mapstructure:"scan-ban-duration"`
	BruteForceThreshold     int  `mapstructure:"brute-force-threshold"`
	BruteForceWindow        int  `mapstructure:"brute-force-window"`
	BruteForceBanDuration   int  `mapstructure:"brute-force-ban-duration"`
}

type JSChallengeConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Mode         string `mapstructure:"mode"`
	Difficulty   int    `mapstructure:"difficulty"`
	CookieTTL    int    `mapstructure:"cookie-ttl"`
	CookieSecret string `mapstructure:"cookie-secret"`
}

type WebhookConfig struct {
	URL     string   `mapstructure:"url"`
	Timeout int      `mapstructure:"timeout"`
	Events  []string `mapstructure:"events"`
}

type BotManagementConfig struct {
	Enabled             bool          `mapstructure:"enabled"`
	ChallengeEscalation bool          `mapstructure:"challenge-escalation"`
	Captcha             CaptchaConfig `mapstructure:"captcha"`
}

type CaptchaConfig struct {
	Provider  string `mapstructure:"provider"` // hcaptcha, turnstile
	SiteKey   string `mapstructure:"site-key"`
	SecretKey string `mapstructure:"secret-key"`
	CookieTTL int    `mapstructure:"cookie-ttl"` // seconds, default 86400
}
