package middleware

import (
	"net/http"
	"open-website-defender/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		corsCfg := config.Get().Security.CORS
		allowedOrigins := corsCfg.AllowedOrigins
		allowCredentials := corsCfg.AllowCredentials

		if len(allowedOrigins) == 0 {
			// No origins configured: CORS is disabled for security.
			// Configure security.cors.allowed-origins in production.
			c.Next()
			return
		}

		allowed := false
		wildcard := false
		for _, o := range allowedOrigins {
			if o == "*" {
				allowed = true
				wildcard = true
				break
			}
			if o == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			c.Next()
			return
		}

		if wildcard {
			// Wildcard: no credentials allowed per spec
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			if allowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Defender-Authorization, X-Requested-With, Cookie")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
