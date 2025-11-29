package database

import (
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	var dialector gorm.Dialector
	var err error

	dbType := viper.GetString("database.type")
	if len(dbType) == 0 {
		dbType = "sqlite"
	}
	dbPath := viper.GetString("database.file-path")
	if len(dbPath) == 0 {
		dbPath = "./data/app.db"
	}

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logging.Sugar.Warnf("Failed to create database directory: %v", err)
	}

	logging.Sugar.Infof("Initializing database: %s", dbPath)

	dialector = sqlite.Open(dbPath)

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
	err = DB.AutoMigrate(&entity.User{}, &entity.IpWhiteList{}, &entity.IpBlackList{})
	if err != nil {
		return err
	}

	err = initDefaultUser()
	if err != nil {
		logging.Sugar.Warnf("Failed to initialize default user: %v", err)
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

	defaultUser := &entity.User{
		Username: defaultUsername,
		Password: pkg.MD5Hash(defaultPassword),
		IsAdmin:  true,
	}

	if err := DB.Create(defaultUser).Error; err != nil {
		return err
	}

	logging.Sugar.Infof("Default user created successfully: username=%s, password=%s (MD5 encrypted)", defaultUsername, defaultPassword)
	return nil
}
