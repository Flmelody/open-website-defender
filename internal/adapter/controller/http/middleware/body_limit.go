package middleware

import (
	"net/http"
	"open-website-defender/internal/adapter/controller/http/response"

	"github.com/gin-gonic/gin"
)

// BodyLimit returns a middleware that limits request body size.
// maxBytes is the maximum allowed body size in bytes.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
		// Check if the body was too large (MaxBytesReader sets this)
		if c.Errors.Last() != nil {
			for _, e := range c.Errors {
				if e.Error() == "http: request body too large" {
					response.Error(c, http.StatusRequestEntityTooLarge, 413, "request entity too large", "request body exceeds maximum allowed size")
					return
				}
			}
		}
	}
}
