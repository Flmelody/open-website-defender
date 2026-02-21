package middleware

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/logging"

	"github.com/gin-gonic/gin"
)

// RateLimiter returns a middleware that limits requests per IP.
// keyPrefix differentiates between global and endpoint-specific limiters.
// requestsPerMinute is the max requests allowed per minute per IP.
func RateLimiter(keyPrefix string, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requestsPerMinute <= 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := keyPrefix + ":" + clientIP

		count, err := cache.Store().Incr(key, 60)
		if err != nil {
			logging.Sugar.Errorf("Rate limiter cache error: %v", err)
			c.Next()
			return
		}

		if count > int64(requestsPerMinute) {
			logging.Sugar.Warnf("Rate limit exceeded for IP %s (prefix=%s, count=%d, limit=%d)",
				clientIP, keyPrefix, count, requestsPerMinute)
			response.TooManyRequests(c, "rate limit exceeded, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimiter is a stricter rate limiter for login endpoints.
// lockoutDuration is in seconds — after exceeding the limit, the IP is locked out for this duration.
func LoginRateLimiter(requestsPerMinute int, lockoutDuration int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requestsPerMinute <= 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		store := cache.Store()

		// Check lockout first
		lockoutKey := "lockout:" + clientIP
		if _, err := store.Get(lockoutKey); err == nil {
			logging.Sugar.Warnf("Login locked out for IP %s", clientIP)
			response.TooManyRequests(c, "too many login attempts, please try again later")
			c.Abort()
			return
		}

		// Check rate
		rateKey := "login:" + clientIP
		count, err := store.Incr(rateKey, 60)
		if err != nil {
			logging.Sugar.Errorf("Login rate limiter cache error: %v", err)
			c.Next()
			return
		}

		if count > int64(requestsPerMinute) {
			// Set lockout
			ttl := lockoutDuration
			if ttl <= 0 {
				ttl = 300 // default 5 minutes
			}
			_ = store.Set(lockoutKey, []byte{1}, ttl)
			logging.Sugar.Warnf("Login rate limit exceeded for IP %s, locked out for %ds", clientIP, ttl)
			response.TooManyRequests(c, "too many login attempts, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
