package middleware

import (
	"open-website-defender/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		headersCfg := config.Get().Security.Headers
		frameOptions := headersCfg.FrameOptions
		if frameOptions == "" {
			frameOptions = "DENY"
		}
		c.Header("X-Frame-Options", frameOptions)

		if headersCfg.HSTSEnabled {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
