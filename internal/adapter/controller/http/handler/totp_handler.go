package handler

import (
	"errors"
	"strconv"

	"open-website-defender/internal/adapter/controller/http/response"
	domainError "open-website-defender/internal/domain/error"
	"open-website-defender/internal/usecase/user"

	"github.com/gin-gonic/gin"
)

func canManageTotp(c *gin.Context, targetID uint) bool {
	currentUser, exists := c.Get("user")
	if !exists {
		return false
	}
	userInfo := currentUser.(*user.UserInfoDTO)
	return userInfo.IsAdmin || userInfo.ID == targetID
}

func SetupTotp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if !canManageTotp(c, uint(id)) {
		response.Forbidden(c, "Not allowed to manage 2FA for this user")
		return
	}

	service := user.GetAuthService()
	output, err := service.SetupTotp(uint(id))
	if err != nil {
		if errors.Is(err, domainError.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, domainError.ErrTotpAlreadyEnabled) {
			response.Conflict(c, "2FA is already enabled")
			return
		}
		response.InternalServerError(c, "Failed to setup 2FA")
		return
	}

	response.Success(c, gin.H{
		"secret":  output.Secret,
		"qr_code": output.QRCodeDataURI,
	})
}

func ConfirmTotp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if !canManageTotp(c, uint(id)) {
		response.Forbidden(c, "Not allowed to manage 2FA for this user")
		return
	}

	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: code must be 6 digits")
		return
	}

	service := user.GetAuthService()
	if err := service.ConfirmTotp(uint(id), req.Code); err != nil {
		if errors.Is(err, domainError.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, domainError.ErrTotpNotEnabled) {
			response.BadRequest(c, "Please setup 2FA first")
			return
		}
		if errors.Is(err, domainError.ErrTotpInvalidCode) {
			response.BadRequest(c, "Invalid verification code")
			return
		}
		response.InternalServerError(c, "Failed to confirm 2FA")
		return
	}

	response.SuccessWithMessage(c, "2FA enabled successfully", nil)
}

func DisableTotp(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if !canManageTotp(c, uint(id)) {
		response.Forbidden(c, "Not allowed to manage 2FA for this user")
		return
	}

	service := user.GetAuthService()
	if err := service.DisableTotp(uint(id)); err != nil {
		if errors.Is(err, domainError.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, domainError.ErrTotpNotEnabled) {
			response.BadRequest(c, "2FA is not enabled")
			return
		}
		response.InternalServerError(c, "Failed to disable 2FA")
		return
	}

	response.SuccessWithMessage(c, "2FA disabled successfully", nil)
}
