package middleware

import (
	"open-website-defender/internal/infrastructure/logging"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logging.Sugar.Infof("[%3d] %-7s %-50s | %13v | %s",
			statusCode,
			c.Request.Method,
			path,
			latency,
			c.ClientIP(),
		)
	}
}
