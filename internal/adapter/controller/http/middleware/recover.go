package middleware

import (
	"open-website-defender/internal/infrastructure/logging"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logging.Sugar.Errorf("Panic recovered:\nError: %v\nStack: %s", err, stack)
				c.Abort()
			}
		}()
		c.Next()
	}
}
