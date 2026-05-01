package http

import (
	"open-website-defender/internal/adapter/controller/http/handler"
	"open-website-defender/internal/adapter/controller/http/middleware"
	"open-website-defender/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, appConfig *config.AppConfig) {
	api := router.Group(appConfig.RootPath)
	{
		// Standalone auth check endpoint (skip JS Challenge — nginx auth_request subrequest)
		middleware.JSChallengeSkipRoute(appConfig.RootPath + "/auth")
		api.GET("/auth", handler.Auth)

		// Login with optional rate limiting
		var loginHandlers []gin.HandlerFunc
		rlCfg := config.Get().RateLimit
		if rlCfg.Enabled {
			loginRPM := rlCfg.Login.RequestsPerMinute
			if loginRPM <= 0 {
				loginRPM = 5
			}
			lockoutDuration := rlCfg.Login.LockoutDuration
			loginHandlers = append(loginHandlers, middleware.LoginRateLimiter(loginRPM, lockoutDuration))
		}
		loginHandlers = append(loginHandlers, handler.Login)
		api.POST("/login", loginHandlers...)

		// Guard 2FA verification (shares same rate limiter as login)
		login2FAHandlers := make([]gin.HandlerFunc, len(loginHandlers)-1, len(loginHandlers))
		copy(login2FAHandlers, loginHandlers[:len(loginHandlers)-1])
		login2FAHandlers = append(login2FAHandlers, handler.Verify2FA)
		api.POST("/login/2fa", login2FAHandlers...)

		// Admin login with admin privilege check
		adminLoginHandlers := make([]gin.HandlerFunc, len(loginHandlers)-1, len(loginHandlers))
		copy(adminLoginHandlers, loginHandlers[:len(loginHandlers)-1])
		adminLoginHandlers = append(adminLoginHandlers, handler.AdminLogin)
		api.POST("/admin-login", adminLoginHandlers...)

		// 2FA verification (shares same rate limiter as admin-login)
		twoFAHandlers := make([]gin.HandlerFunc, len(adminLoginHandlers)-1, len(adminLoginHandlers))
		copy(twoFAHandlers, adminLoginHandlers[:len(adminLoginHandlers)-1])
		twoFAHandlers = append(twoFAHandlers, handler.Verify2FA)
		api.POST("/admin-login/2fa", twoFAHandlers...)

		// Admin 2FA recovery (unauthenticated, rate-limited like login)
		recoverHandlers := make([]gin.HandlerFunc, len(loginHandlers)-1, len(loginHandlers))
		copy(recoverHandlers, loginHandlers[:len(loginHandlers)-1])
		recoverHandlers = append(recoverHandlers, handler.AdminRecover2FA)
		api.POST("/admin-recover-2fa", recoverHandlers...)
		api.POST("/logout", handler.Logout)

		// OIDC Discovery (public, no auth)
		api.GET("/.well-known/openid-configuration", handler.OIDCDiscovery)
		api.GET("/.well-known/jwks.json", handler.JWKS)

		// OAuth2/OIDC endpoints (authenticated via OWD session cookie, not admin middleware)
		if config.Get().OAuth.Enabled {
			api.GET("/oauth/authorize", handler.OAuthAuthorize)
			api.POST("/oauth/consent", handler.OAuthConsent)
			api.POST("/oauth/token", handler.OAuthToken)
			api.GET("/oauth/userinfo", handler.OAuthUserInfo)
			api.POST("/oauth/userinfo", handler.OAuthUserInfo)
			api.POST("/oauth/revoke", handler.OAuthRevoke)
		}

		// CAPTCHA endpoints (public, skip challenge rendering so captcha page can load its own resources)
		middleware.ChallengeSkipRoute(appConfig.RootPath+"/captcha/generate", appConfig.RootPath+"/captcha/verify")
		api.GET("/captcha/generate", handler.GenerateCaptcha)
		api.POST("/captcha/verify", handler.VerifyCaptcha)

		authenticated := api.Group("")
		authenticated.Use(handler.AuthMiddleware)
		{
			// Self-service routes: user can only access their own resources; admins can manage all.
			authenticated.GET("/users/:id/oauth-authorizations", handler.ListUserOAuthAuthorizations)
			authenticated.DELETE("/users/:id/oauth-authorizations/:clientId", handler.RevokeUserOAuthAuthorization)
			authenticated.POST("/users/:id/totp/setup", handler.SetupTotp)
			authenticated.POST("/users/:id/totp/confirm", handler.ConfirmTotp)
			authenticated.DELETE("/users/:id/totp", handler.DisableTotp)

			admin := authenticated.Group("")
			admin.Use(handler.AdminMiddleware)
			{
				admin.GET("/admin-session", handler.AdminSession)

				admin.POST("/users", handler.CreateUser)
				admin.GET("/users", handler.ListUser)
				admin.GET("/users/:id", handler.GetUser)
				admin.PUT("/users/:id", handler.UpdateUser)
				admin.DELETE("/users/:id", handler.DeleteUser)

				// IP Blacklist
				admin.POST("/ip-black-list", handler.CreateIpBlackList)
				admin.GET("/ip-black-list", handler.ListIpBlackList)
				admin.PUT("/ip-black-list/:id", handler.UpdateIpBlackList)
				admin.DELETE("/ip-black-list/:id", handler.DeleteIpBlackList)

				// IP Whitelist
				admin.POST("/ip-white-list", handler.CreateIpWhiteList)
				admin.GET("/ip-white-list", handler.ListIpWhiteList)
				admin.PUT("/ip-white-list/:id", handler.UpdateIpWhiteList)
				admin.DELETE("/ip-white-list/:id", handler.DeleteIpWhiteList)

				// WAF Rules
				admin.POST("/waf-rules", handler.CreateWafRule)
				admin.GET("/waf-rules", handler.ListWafRules)
				admin.PUT("/waf-rules/:id", handler.UpdateWafRule)
				admin.DELETE("/waf-rules/:id", handler.DeleteWafRule)
				admin.PUT("/waf-rules/group/:name/enable", handler.BatchEnableWafGroup)
				admin.PUT("/waf-rules/group/:name/disable", handler.BatchDisableWafGroup)

				// WAF Exclusions
				admin.POST("/waf-exclusions", handler.CreateWafExclusion)
				admin.GET("/waf-exclusions", handler.ListWafExclusions)
				admin.DELETE("/waf-exclusions/:id", handler.DeleteWafExclusion)

				// Bot Signatures
				admin.POST("/bot-signatures", handler.CreateBotSignature)
				admin.GET("/bot-signatures", handler.ListBotSignatures)
				admin.PUT("/bot-signatures/:id", handler.UpdateBotSignature)
				admin.DELETE("/bot-signatures/:id", handler.DeleteBotSignature)

				// Access Logs
				admin.GET("/access-logs", handler.ListAccessLogs)
				admin.GET("/access-logs/stats", handler.GetAccessLogStats)
				admin.DELETE("/access-logs", handler.ClearAccessLogs)

				// Geo-blocking
				admin.POST("/geo-block-rules", handler.CreateGeoBlockRule)
				admin.GET("/geo-block-rules", handler.ListGeoBlockRules)
				admin.DELETE("/geo-block-rules/:id", handler.DeleteGeoBlockRule)

				// Licenses
				admin.POST("/licenses", handler.CreateLicense)
				admin.GET("/licenses", handler.ListLicenses)
				admin.DELETE("/licenses/:id", handler.DeleteLicense)

				// Authorized Domains
				admin.POST("/authorized-domains", handler.CreateAuthorizedDomain)
				admin.GET("/authorized-domains", handler.ListAuthorizedDomains)
				admin.DELETE("/authorized-domains/:id", handler.DeleteAuthorizedDomain)

				// OAuth Clients
				admin.POST("/oauth-clients", handler.CreateOAuthClient)
				admin.GET("/oauth-clients", handler.ListOAuthClients)
				admin.GET("/oauth-clients/:id", handler.GetOAuthClient)
				admin.PUT("/oauth-clients/:id", handler.UpdateOAuthClient)
				admin.DELETE("/oauth-clients/:id", handler.DeleteOAuthClient)

				// Security Events
				admin.GET("/security-events", handler.ListSecurityEvents)
				admin.GET("/security-events/stats", handler.GetSecurityEventStats)
				admin.GET("/security-events/threat-score", handler.GetThreatScore)

				// Dashboard & System
				admin.GET("/dashboard/stats", handler.GetDashboardStats)
				admin.GET("/system/settings", handler.GetSystemSettings)
				admin.PUT("/system/settings", handler.UpdateSystemSettings)
				admin.POST("/system/reload", handler.ReloadConfig)
			}
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
