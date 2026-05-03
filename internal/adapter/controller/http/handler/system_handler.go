package handler

import (
	"castellum/internal/adapter/controller/http/request"
	"castellum/internal/adapter/controller/http/response"
	"castellum/internal/infrastructure/cache"
	"castellum/internal/infrastructure/logging"
	"castellum/internal/usecase/system"

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

		BotManagementEnabled: req.BotManagementEnabled,
		ChallengeEscalation:  req.ChallengeEscalation,
		CaptchaProvider:      req.CaptchaProvider,
		CaptchaSiteKey:       req.CaptchaSiteKey,
		CaptchaSecretKey:     req.CaptchaSecretKey,
		CaptchaCookieTTL:     req.CaptchaCookieTTL,

		CacheSyncInterval:      req.CacheSyncInterval,
		AccessLogRetentionDays: req.AccessLogRetentionDays,

		SemanticAnalysisEnabled: req.SemanticAnalysisEnabled,
	}

	if err := service.UpdateSettings(input); err != nil {
		logging.Sugar.Errorf("Failed to update system settings: %v", err)
		response.InternalServerError(c, "Failed to update system settings")
		return
	}

	response.SuccessWithMessage(c, "System settings updated", nil)
}

// ReloadConfig flushes in-memory caches so database-backed settings reload on the next request.
// File-based configuration still requires a process restart.
func ReloadConfig(c *gin.Context) {
	cache.Store().Clear()
	logging.Sugar.Info("Cache cleared via admin reload endpoint")
	response.SuccessWithMessage(c, "Caches cleared; database-backed settings will reload on next request", nil)
}
