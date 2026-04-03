package logging

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func InitLogger() error {
	return InitLoggerWithEnvAndLevel("production", "")
}

func InitLoggerWithEnv(env string) error {
	return InitLoggerWithEnvAndLevel(env, "")
}

func InitLoggerWithEnvAndLevel(env string, level string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.Level = zap.NewAtomicLevelAt(parseLevel(env, level))

	os.MkdirAll("./logs", 0755)
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

func parseLevel(env string, level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}

	if env == "production" {
		return zapcore.InfoLevel
	}
	return zapcore.DebugLevel
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
