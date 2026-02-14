package http

import (
	"open-website-defender/internal/adapter/controller/http/handler"
	"open-website-defender/internal/adapter/controller/http/middleware"
	"open-website-defender/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Setup(appConfig *config.AppConfig) *gin.Engine {
	router := gin.Default()

	api := router.Group(appConfig.RootPath)
	{
		// Standalone auth check endpoint
		api.GET("/auth", handler.Auth)

		// Login with optional rate limiting
		loginHandlers := []gin.HandlerFunc{}
		if viper.GetBool("rate-limit.enabled") {
			loginRPM := viper.GetInt("rate-limit.login.requests-per-minute")
			if loginRPM <= 0 {
				loginRPM = 5
			}
			lockoutDuration := viper.GetInt("rate-limit.login.lockout-duration")
			loginHandlers = append(loginHandlers, middleware.LoginRateLimiter(loginRPM, lockoutDuration))
		}
		loginHandlers = append(loginHandlers, handler.Login)
		api.POST("/login", loginHandlers...)

		authorized := api.Group("")
		// Middleware for route protection
		authorized.Use(handler.AuthMiddleware)
		{
			authorized.POST("/users", handler.CreateUser)
			authorized.GET("/users", handler.ListUser)
			authorized.GET("/users/:id", handler.GetUser)
			authorized.PUT("/users/:id", handler.UpdateUser)
			authorized.DELETE("/users/:id", handler.DeleteUser)

			// IP Blacklist
			authorized.POST("/ip-black-list", handler.CreateIpBlackList)
			authorized.GET("/ip-black-list", handler.ListIpBlackList)
			authorized.DELETE("/ip-black-list/:id", handler.DeleteIpBlackList)

			// IP Whitelist
			authorized.POST("/ip-white-list", handler.CreateIpWhiteList)
			authorized.GET("/ip-white-list", handler.ListIpWhiteList)
			authorized.DELETE("/ip-white-list/:id", handler.DeleteIpWhiteList)

			// WAF Rules
			authorized.POST("/waf-rules", handler.CreateWafRule)
			authorized.GET("/waf-rules", handler.ListWafRules)
			authorized.PUT("/waf-rules/:id", handler.UpdateWafRule)
			authorized.DELETE("/waf-rules/:id", handler.DeleteWafRule)

			// Access Logs
			authorized.GET("/access-logs", handler.ListAccessLogs)
			authorized.GET("/access-logs/stats", handler.GetAccessLogStats)

			// Geo-blocking
			authorized.POST("/geo-block-rules", handler.CreateGeoBlockRule)
			authorized.GET("/geo-block-rules", handler.ListGeoBlockRules)
			authorized.DELETE("/geo-block-rules/:id", handler.DeleteGeoBlockRule)

			// Licenses
			authorized.POST("/licenses", handler.CreateLicense)
			authorized.GET("/licenses", handler.ListLicenses)
			authorized.DELETE("/licenses/:id", handler.DeleteLicense)

			// Dashboard & System
			authorized.GET("/dashboard/stats", handler.GetDashboardStats)
			authorized.GET("/system/settings", handler.GetSystemSettings)
			authorized.PUT("/system/settings", handler.UpdateSystemSettings)
			authorized.POST("/system/reload", handler.ReloadConfig)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
