package database

import (
	"fmt"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	var dialector gorm.Dialector
	var err error

	dbDriver := viper.GetString("database.driver")
	if len(dbDriver) == 0 {
		dbDriver = "sqlite"
	}

	switch dbDriver {
	case "postgres", "postgresql":
		host := viper.GetString("database.host")
		if len(host) == 0 {
			host = "localhost"
		}
		port := viper.GetInt("database.port")
		if port == 0 {
			port = 5432
		}
		dbName := viper.GetString("database.name")
		if len(dbName) == 0 {
			dbName = "open_website_defender"
		}
		user := viper.GetString("database.user")
		if len(user) == 0 {
			user = "postgres"
		}
		password := viper.GetString("database.password")
		sslMode := viper.GetString("database.ssl-mode")
		if len(sslMode) == 0 {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbName, sslMode)
		logging.Sugar.Infof("Initializing PostgreSQL database: %s:%d/%s", host, port, dbName)
		dialector = postgres.Open(dsn)

	case "mysql":
		host := viper.GetString("database.host")
		if len(host) == 0 {
			host = "localhost"
		}
		port := viper.GetInt("database.port")
		if port == 0 {
			port = 3306
		}
		dbName := viper.GetString("database.name")
		if len(dbName) == 0 {
			dbName = "open_website_defender"
		}
		user := viper.GetString("database.user")
		if len(user) == 0 {
			user = "root"
		}
		password := viper.GetString("database.password")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, port, dbName)
		logging.Sugar.Infof("Initializing MySQL database: %s:%d/%s", host, port, dbName)
		dialector = mysql.Open(dsn)

	case "sqlite":
		dbPath := viper.GetString("database.file-path")
		if len(dbPath) == 0 {
			dbPath = "./data/app.db"
		}
		dbDir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			logging.Sugar.Warnf("Failed to create database directory: %v", err)
		}
		logging.Sugar.Infof("Initializing SQLite database: %s", dbPath)
		dialector = sqlite.Open(dbPath)

	default:
		return fmt.Errorf("unsupported database driver: %s (supported: sqlite, postgres, mysql)", dbDriver)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return err
	}

	logging.Sugar.Info("Running database migrations...")
	err = DB.AutoMigrate(
		&entity.User{},
		&entity.IpWhiteList{},
		&entity.IpBlackList{},
		&entity.WafRule{},
		&entity.AccessLog{},
		&entity.GeoBlockRule{},
	)
	if err != nil {
		return err
	}

	err = initDefaultUser()
	if err != nil {
		logging.Sugar.Warnf("Failed to initialize default user: %v", err)
	}

	err = initDefaultWafRules()
	if err != nil {
		logging.Sugar.Warnf("Failed to initialize default WAF rules: %v", err)
	}

	return nil
}

func initDefaultUser() error {
	var count int64
	if err := DB.Model(&entity.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logging.Sugar.Info("Users already exist, skipping default user creation")
		return nil
	}

	defaultUsername := viper.GetString("default-user.username")
	if len(defaultUsername) == 0 {
		defaultUsername = "defender"
	}

	defaultPassword := viper.GetString("default-user.password")
	if len(defaultPassword) == 0 {
		defaultPassword = "defender"
	}

	hashedPassword, err := pkg.HashPassword(defaultPassword)
	if err != nil {
		return err
	}

	defaultUser := &entity.User{
		Username: defaultUsername,
		Password: hashedPassword,
		IsAdmin:  true,
	}

	if err := DB.Create(defaultUser).Error; err != nil {
		return err
	}

	logging.Sugar.Infof("Default user created successfully: username=%s (password hashed with bcrypt)", defaultUsername)
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}

func initDefaultWafRules() error {
	var count int64
	if err := DB.Model(&entity.WafRule{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logging.Sugar.Info("WAF rules already exist, skipping default rule creation")
		return nil
	}

	defaultRules := []entity.WafRule{
		// SQL Injection patterns
		{
			Name:     "SQLi - Union Select",
			Pattern:  `(?i)(union\s+(all\s+)?select)`,
			Category: "sqli",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		{
			Name:     "SQLi - Common Patterns",
			Pattern:  `(?i)(;\s*(drop|alter|truncate|delete|insert|update)\s)`,
			Category: "sqli",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		{
			Name:     "SQLi - Boolean Injection",
			Pattern:  `(?i)('\s*(or|and)\s*'?\d*\s*[=<>])`,
			Category: "sqli",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		{
			Name:     "SQLi - Comment Injection",
			Pattern:  `(?i)('\s*--\s*$|/\*.*?\*/)`,
			Category: "sqli",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		// XSS patterns
		{
			Name:     "XSS - Script Tag",
			Pattern:  `(?i)(<script[\s>]|</script>)`,
			Category: "xss",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		{
			Name:     "XSS - Event Handler",
			Pattern:  `(?i)(on(error|load|click|mouseover|focus|blur|submit|change)\s*=)`,
			Category: "xss",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		{
			Name:     "XSS - JavaScript Protocol",
			Pattern:  `(?i)(javascript\s*:|vbscript\s*:)`,
			Category: "xss",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		// Path Traversal patterns
		{
			Name:     "Path Traversal - Dot Dot Slash",
			Pattern:  `(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`,
			Category: "traversal",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
		{
			Name:     "Path Traversal - Sensitive Files",
			Pattern:  `(?i)(/etc/passwd|/etc/shadow|/proc/self|/windows/system32)`,
			Category: "traversal",
			Action:   "block",
			Enabled:  boolPtr(true),
		},
	}

	for i := range defaultRules {
		if err := DB.Create(&defaultRules[i]).Error; err != nil {
			logging.Sugar.Warnf("Failed to create default WAF rule '%s': %v", defaultRules[i].Name, err)
		}
	}

	logging.Sugar.Infof("Created %d default WAF rules", len(defaultRules))
	return nil
}
