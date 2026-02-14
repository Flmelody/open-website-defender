package middleware

import (
	"encoding/binary"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"

	"github.com/gin-gonic/gin"
)

// RateLimiter returns a middleware that limits requests per IP using freecache.
// keyPrefix differentiates between global and endpoint-specific limiters.
// requestsPerMinute is the max requests allowed per minute per IP.
func RateLimiter(keyPrefix string, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requestsPerMinute <= 0 {
			c.Next()
			return
		}

		clientIP := pkg.GetClientIP(c)
		key := []byte(keyPrefix + ":" + clientIP)
		cache := pkg.Cacher()

		val, err := cache.Get(key)
		var count int64
		if err == nil && len(val) == 8 {
			count = int64(binary.BigEndian.Uint64(val))
		}

		if count >= int64(requestsPerMinute) {
			logging.Sugar.Warnf("Rate limit exceeded for IP %s (prefix=%s, count=%d, limit=%d)",
				clientIP, keyPrefix, count, requestsPerMinute)
			response.TooManyRequests(c, "rate limit exceeded, please try again later")
			c.Abort()
			return
		}

		count++
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(count))
		_ = cache.Set(key, buf, 60) // 60 seconds TTL

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

		clientIP := pkg.GetClientIP(c)
		cache := pkg.Cacher()

		// Check lockout first
		lockoutKey := []byte("lockout:" + clientIP)
		if _, err := cache.Get(lockoutKey); err == nil {
			logging.Sugar.Warnf("Login locked out for IP %s", clientIP)
			response.TooManyRequests(c, "too many login attempts, please try again later")
			c.Abort()
			return
		}

		// Check rate
		rateKey := []byte("login:" + clientIP)
		val, err := cache.Get(rateKey)
		var count int64
		if err == nil && len(val) == 8 {
			count = int64(binary.BigEndian.Uint64(val))
		}

		if count >= int64(requestsPerMinute) {
			// Set lockout
			ttl := lockoutDuration
			if ttl <= 0 {
				ttl = 300 // default 5 minutes
			}
			_ = cache.Set(lockoutKey, []byte{1}, ttl)
			logging.Sugar.Warnf("Login rate limit exceeded for IP %s, locked out for %ds", clientIP, ttl)
			response.TooManyRequests(c, "too many login attempts, please try again later")
			c.Abort()
			return
		}

		count++
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(count))
		_ = cache.Set(rateKey, buf, 60) // 60 seconds TTL

		c.Next()
	}
}
