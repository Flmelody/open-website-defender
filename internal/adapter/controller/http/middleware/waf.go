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

		// Collect headers
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}

		// Collect cookies
		cookies := make(map[string]string)
		for _, cookie := range c.Request.Cookies() {
			cookies[cookie.Name] = cookie.Value
		}

		// Read body for inspection (only for methods that carry a body)
		var bodyStr string
		if c.Request.Body != nil && c.Request.ContentLength > 0 &&
			c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead &&
			c.Request.Method != http.MethodOptions {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyRead))
			if err == nil {
				bodyStr = string(bodyBytes)
				remaining := c.Request.Body
				c.Request.Body = io.NopCloser(io.MultiReader(
					bytes.NewReader(bodyBytes),
					remaining,
				))
			}
		}

		ctx := &waf.RequestContext{
			Method:   c.Request.Method,
			Path:     path,
			Query:    queryString,
			UA:       userAgent,
			Body:     bodyStr,
			Headers:  headers,
			Cookies:  cookies,
			ClientIP: c.ClientIP(),
		}

		result := service.CheckRequestContext(ctx)
		if result != nil {
			if result.SemanticFingerprint != "" {
				logging.Sugar.Warnf("WAF rule matched: %s (action=%s, path=%s, ip=%s, fingerprint=%s)",
					result.RuleName, result.Action, path, c.ClientIP(), result.SemanticFingerprint)
			} else {
				logging.Sugar.Warnf("WAF rule matched: %s (action=%s, path=%s, ip=%s)",
					result.RuleName, result.Action, path, c.ClientIP())
			}

			// Store for access logging
			c.Set("waf_action", "blocked_waf")
			c.Set("waf_rule", result.RuleName)

			switch result.Action {
			case "block":
				response.Forbidden(c, "request blocked by WAF")
				c.Abort()
				return
			case "redirect":
				if result.RedirectURL != "" {
					c.Redirect(http.StatusFound, result.RedirectURL)
					c.Abort()
					return
				}
				response.Forbidden(c, "request blocked by WAF")
				c.Abort()
				return
			case "challenge":
				c.Set("waf_challenge", true)
				// Let the JS Challenge middleware handle it
			case "rate-limit":
				if result.RateLimit > 0 {
					c.Set("waf_rate_limit", result.RateLimit)
				}
				// Let the rate limiter handle it
			case "log":
				// Let it through but marked
			default:
				// Unknown action, block by default
				if result.Blocked {
					response.Forbidden(c, "request blocked by WAF")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
