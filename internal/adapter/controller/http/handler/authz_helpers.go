package handler

import (
	"castellum/internal/adapter/controller/http/response"
	"castellum/internal/usecase/user"

	"github.com/gin-gonic/gin"
)

func currentUserInfo(c *gin.Context) (*user.UserInfoDTO, bool) {
	currentUser, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userInfo, ok := currentUser.(*user.UserInfoDTO)
	if !ok || userInfo == nil {
		return nil, false
	}

	return userInfo, true
}

func canAccessUserResource(c *gin.Context, targetID uint) bool {
	userInfo, ok := currentUserInfo(c)
	if !ok {
		return false
	}
	return userInfo.IsAdmin || userInfo.ID == targetID
}

func AdminMiddleware(c *gin.Context) {
	userInfo, ok := currentUserInfo(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		c.Abort()
		return
	}
	if !userInfo.IsAdmin {
		response.Forbidden(c, "Admin privileges required")
		c.Abort()
		return
	}
	c.Next()
}
