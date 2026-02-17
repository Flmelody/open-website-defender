package middleware

import (
	"bytes"
	"io"
	"net/http"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/waf"

	"github.com/gin-gonic/gin"
)

const maxBodyRead = 10 * 1024 // 10KB max for WAF inspection

// WAF returns a middleware that inspects requests for malicious patterns.
func WAF() gin.HandlerFunc {
	return func(c *gin.Context) {
		service := waf.GetWafService()

		path := c.Request.URL.Path
		queryString := c.Request.URL.RawQuery
		userAgent := c.GetHeader("User-Agent")

		// Read body for inspection (only for methods that carry a body)
		var bodyStr string
		if c.Request.Body != nil && c.Request.ContentLength > 0 &&
			c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead &&
			c.Request.Method != http.MethodOptions {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyRead))
			if err == nil {
				bodyStr = string(bodyBytes)
				// Restore the body so downstream handlers can read it
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		result := service.CheckRequest(c.Request.Method, path, queryString, userAgent, bodyStr)
		if result != nil {
			logging.Sugar.Warnf("WAF rule matched: %s (action=%s, path=%s, ip=%s)",
				result.RuleName, result.Action, path, c.ClientIP())

			// Store for access logging
			c.Set("waf_action", "blocked_waf")
			c.Set("waf_rule", result.RuleName)

			if result.Blocked {
				response.Forbidden(c, "request blocked by WAF")
				c.Abort()
				return
			}
			// Action is "log" — let it through but mark it
		}

		c.Next()
	}
}
