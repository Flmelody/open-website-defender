package middleware

import (
	"open-website-defender/internal/usecase/accesslog"
	"open-website-defender/internal/usecase/threat"
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
		statusCode := c.Writer.Status()

		action := "allowed"
		ruleName := ""

		// Check if WAF or other middleware set action
		if wafAction, exists := c.Get("waf_action"); exists {
			action = wafAction.(string)
		}
		if wafRule, exists := c.Get("waf_rule"); exists {
			ruleName = wafRule.(string)
		}

		// Check if rate limiter flagged this request
		wasRateLimited := false
		if action == "allowed" {
			switch statusCode {
			case 403:
				action = "blocked"
			case 429:
				action = "blocked_ratelimit"
				wasRateLimited = true
			}
		}
		if action == "blocked_ratelimit" {
			wasRateLimited = true
		}

		svc := accesslog.GetAccessLogService()
		svc.Record(&accesslog.AccessLogInput{
			ClientIP:   clientIP,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: statusCode,
			Latency:    latency,
			UserAgent:  c.GetHeader("User-Agent"),
			Action:     action,
			RuleName:   ruleName,
		})

		// Feed request data to threat detector for anomaly detection
		td := threat.GetThreatDetector()
		td.RecordRequest(clientIP, statusCode, wasRateLimited)

		// Update threat score based on action
		if action == "blocked_waf" {
			td.AddThreatScore(clientIP, 5)
		} else if wasRateLimited {
			td.AddThreatScore(clientIP, 3)
		} else if statusCode >= 400 && statusCode < 500 {
			td.AddThreatScore(clientIP, 1)
		}
	}
}
