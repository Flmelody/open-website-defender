package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"open-website-defender/internal/usecase/system"

	"github.com/gin-gonic/gin"
)

func GetSystemSettings(c *gin.Context) {
	service := system.GetSystemService()

	settings, err := service.GetSettings()
	if err != nil {
		logging.Sugar.Errorf("Failed to get system settings: %v", err)
		response.InternalServerError(c, "Failed to get system settings")
		return
	}

	response.Success(c, settings)
}

func UpdateSystemSettings(c *gin.Context) {
	service := system.GetSystemService()

	var req request.UpdateSystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &system.UpdateSystemSettingsDTO{
		GitTokenHeader:        req.GitTokenHeader,
		LicenseHeader:         req.LicenseHeader,
		JSChallengeEnabled:    req.JSChallengeEnabled,
		JSChallengeMode:       req.JSChallengeMode,
		JSChallengeDifficulty: req.JSChallengeDifficulty,
		WebhookURL:            req.WebhookURL,
	}

	if err := service.UpdateSettings(input); err != nil {
		logging.Sugar.Errorf("Failed to update system settings: %v", err)
		response.InternalServerError(c, "Failed to update system settings")
		return
	}

	response.SuccessWithMessage(c, "System settings updated", nil)
}

// ReloadConfig flushes all caches, forcing services to reload from database.
func ReloadConfig(c *gin.Context) {
	pkg.Cacher().Clear()
	logging.Sugar.Info("Cache cleared via admin reload endpoint")
	response.SuccessWithMessage(c, "Configuration reloaded, caches cleared", nil)
}
