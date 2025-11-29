package logging

import "go.uber.org/zap"

import (
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func InitLogger() error {
	return InitLoggerWithEnv("production")
}

func InitLoggerWithEnv(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout", "./logs/app.log"}
	config.ErrorOutputPaths = []string{"stderr", "./logs/error.log"}

	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	Sugar = Logger.Sugar()
	return nil
}
