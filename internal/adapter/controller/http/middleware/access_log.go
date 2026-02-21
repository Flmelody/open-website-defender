package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"open-website-defender/internal/usecase/accesslog"
	"open-website-defender/internal/usecase/threat"
	"time"

	"github.com/gin-gonic/gin"
)

const maxBodyCapture = 4096 // 4KB

// AccessLog returns a middleware that records request details to the access log.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body before processing (non-GET only, truncated)
		var requestBody string
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyCapture+1))
			c.Request.Body.Close()
			if len(bodyBytes) > maxBodyCapture {
				requestBody = string(bodyBytes[:maxBodyCapture]) + "...(truncated)"
			} else {
				requestBody = string(bodyBytes)
			}
			// Restore body for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

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

		// Serialize request headers
		headersJSON := ""
		if h, err := json.Marshal(c.Request.Header); err == nil {
			headersJSON = string(h)
		}

		// Determine scheme
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		} else if c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		svc := accesslog.GetAccessLogService()
		svc.Record(&accesslog.AccessLogInput{
			ClientIP:       clientIP,
			Method:         c.Request.Method,
			Host:           c.Request.Host,
			Scheme:         scheme,
			Path:           c.Request.URL.Path,
			QueryString:    c.Request.URL.RawQuery,
			ContentType:    c.GetHeader("Content-Type"),
			ContentLength:  c.Request.ContentLength,
			Referer:        c.GetHeader("Referer"),
			RequestHeaders: headersJSON,
			RequestBody:    requestBody,
			StatusCode:     statusCode,
			ResponseSize:   c.Writer.Size(),
			Latency:        latency,
			UserAgent:      c.GetHeader("User-Agent"),
			Action:         action,
			RuleName:       ruleName,
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
