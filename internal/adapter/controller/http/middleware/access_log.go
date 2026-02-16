package middleware

import (
	"open-website-defender/internal/usecase/accesslog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog returns a middleware that records request details to the access log.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start).Microseconds()
		clientIP := c.ClientIP()

		action := "allowed"
		ruleName := ""

		// Check if WAF or other middleware set action
		if wafAction, exists := c.Get("waf_action"); exists {
			action = wafAction.(string)
		}
		if wafRule, exists := c.Get("waf_rule"); exists {
			ruleName = wafRule.(string)
		}

		// Detect blocked status from response code
		if action == "allowed" {
			switch c.Writer.Status() {
			case 403:
				action = "blocked"
			case 429:
				action = "blocked_ratelimit"
			}
		}

		svc := accesslog.GetAccessLogService()
		svc.Record(&accesslog.AccessLogInput{
			ClientIP:   clientIP,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			Latency:    latency,
			UserAgent:  c.GetHeader("User-Agent"),
			Action:     action,
			RuleName:   ruleName,
		})
	}
}
