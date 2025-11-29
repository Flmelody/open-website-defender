package config

type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Whitelist WhitelistConfig `mapstructure:"whitelist" json:"-"`
	Blacklist BlacklistConfig `mapstructure:"blacklist" json:"-"`
	Wall      AppConfig       `mapstructure:"wall"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	FilePath string `mapstructure:"file-path"`
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
