package middleware

import (
	"castellum/internal/adapter/controller/http/response"
	"castellum/internal/infrastructure/logging"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logging.Sugar.Errorf("Panic recovered:\nError: %v\nStack: %s", err, stack)
				response.InternalServerError(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
