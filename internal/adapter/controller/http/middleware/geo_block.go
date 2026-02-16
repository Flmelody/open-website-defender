package middleware

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/geoblock"

	"github.com/gin-gonic/gin"
)

// GeoBlock returns a middleware that blocks requests from specific countries.
func GeoBlock() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		service := geoblock.GetGeoBlockService()

		blocked, country := service.IsBlocked(clientIP)
		if blocked {
			logging.Sugar.Warnf("Geo-blocked request from IP %s (country: %s)", clientIP, country)
			c.Set("waf_action", "blocked_geo")
			c.Set("waf_rule", "geo:"+country)
			response.Forbidden(c, "access denied from your region")
			c.Abort()
			return
		}

		c.Next()
	}
}
