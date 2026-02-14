package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		frameOptions := viper.GetString("security.headers.frame-options")
		if frameOptions == "" {
			frameOptions = "DENY"
		}
		c.Header("X-Frame-Options", frameOptions)

		if viper.GetBool("security.headers.hsts-enabled") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
