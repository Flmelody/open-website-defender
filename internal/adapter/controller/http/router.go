package http

import (
	"open-website-defender/internal/adapter/controller/http/handler"
	"open-website-defender/internal/adapter/controller/http/middleware"
	"open-website-defender/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Setup(router *gin.Engine, appConfig *config.AppConfig) {
	api := router.Group(appConfig.RootPath)
	{
		// Standalone auth check endpoint
		api.GET("/auth", handler.Auth)

		// Login with optional rate limiting
		var loginHandlers []gin.HandlerFunc
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

		// Admin login with admin privilege check
		adminLoginHandlers := make([]gin.HandlerFunc, len(loginHandlers)-1, len(loginHandlers))
		copy(adminLoginHandlers, loginHandlers[:len(loginHandlers)-1])
		adminLoginHandlers = append(adminLoginHandlers, handler.AdminLogin)
		api.POST("/admin-login", adminLoginHandlers...)

		// OIDC Discovery (public, no auth)
		api.GET("/.well-known/openid-configuration", handler.OIDCDiscovery)
		api.GET("/.well-known/jwks.json", handler.JWKS)

		// OAuth2/OIDC endpoints (authenticated via OWD session cookie, not admin middleware)
		if viper.GetBool("oauth.enabled") {
			api.GET("/oauth/authorize", handler.OAuthAuthorize)
			api.POST("/oauth/consent", handler.OAuthConsent)
			api.POST("/oauth/token", handler.OAuthToken)
			api.GET("/oauth/userinfo", handler.OAuthUserInfo)
			api.POST("/oauth/userinfo", handler.OAuthUserInfo)
			api.POST("/oauth/revoke", handler.OAuthRevoke)
		}

		authorized := api.Group("")
		// Middleware for route protection
		authorized.Use(handler.AuthMiddleware)
		{
			authorized.POST("/users", handler.CreateUser)
			authorized.GET("/users", handler.ListUser)
			authorized.GET("/users/:id", handler.GetUser)
			authorized.PUT("/users/:id", handler.UpdateUser)
			authorized.DELETE("/users/:id", handler.DeleteUser)

			// User OAuth Authorizations
			authorized.GET("/users/:id/oauth-authorizations", handler.ListUserOAuthAuthorizations)
			authorized.DELETE("/users/:id/oauth-authorizations/:clientId", handler.RevokeUserOAuthAuthorization)

			// IP Blacklist
			authorized.POST("/ip-black-list", handler.CreateIpBlackList)
			authorized.GET("/ip-black-list", handler.ListIpBlackList)
			authorized.DELETE("/ip-black-list/:id", handler.DeleteIpBlackList)

			// IP Whitelist
			authorized.POST("/ip-white-list", handler.CreateIpWhiteList)
			authorized.GET("/ip-white-list", handler.ListIpWhiteList)
			authorized.PUT("/ip-white-list/:id", handler.UpdateIpWhiteList)
			authorized.DELETE("/ip-white-list/:id", handler.DeleteIpWhiteList)

			// WAF Rules
			authorized.POST("/waf-rules", handler.CreateWafRule)
			authorized.GET("/waf-rules", handler.ListWafRules)
			authorized.PUT("/waf-rules/:id", handler.UpdateWafRule)
			authorized.DELETE("/waf-rules/:id", handler.DeleteWafRule)

			// Access Logs
			authorized.GET("/access-logs", handler.ListAccessLogs)
			authorized.GET("/access-logs/stats", handler.GetAccessLogStats)
			authorized.DELETE("/access-logs", handler.ClearAccessLogs)

			// Geo-blocking
			authorized.POST("/geo-block-rules", handler.CreateGeoBlockRule)
			authorized.GET("/geo-block-rules", handler.ListGeoBlockRules)
			authorized.DELETE("/geo-block-rules/:id", handler.DeleteGeoBlockRule)

			// Licenses
			authorized.POST("/licenses", handler.CreateLicense)
			authorized.GET("/licenses", handler.ListLicenses)
			authorized.DELETE("/licenses/:id", handler.DeleteLicense)

			// Authorized Domains
			authorized.POST("/authorized-domains", handler.CreateAuthorizedDomain)
			authorized.GET("/authorized-domains", handler.ListAuthorizedDomains)
			authorized.DELETE("/authorized-domains/:id", handler.DeleteAuthorizedDomain)

			// OAuth Clients (admin management)
			authorized.POST("/oauth-clients", handler.CreateOAuthClient)
			authorized.GET("/oauth-clients", handler.ListOAuthClients)
			authorized.GET("/oauth-clients/:id", handler.GetOAuthClient)
			authorized.PUT("/oauth-clients/:id", handler.UpdateOAuthClient)
			authorized.DELETE("/oauth-clients/:id", handler.DeleteOAuthClient)

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
}
