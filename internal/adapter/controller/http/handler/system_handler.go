package handler

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"

	"github.com/gin-gonic/gin"
)

// ReloadConfig flushes all caches, forcing services to reload from database.
func ReloadConfig(c *gin.Context) {
	pkg.Cacher().Clear()
	logging.Sugar.Info("Cache cleared via admin reload endpoint")
	response.SuccessWithMessage(c, "Configuration reloaded, caches cleared", nil)
}
