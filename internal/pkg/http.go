package pkg

import (
	"net"
	"open-website-defender/internal/infrastructure/logging"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if ip := net.ParseIP(clientIP); ip != nil {
				logging.Sugar.Debugf("Client IP from X-Forwarded-For: %s", clientIP)
				return clientIP
			}
		}
	}

	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			logging.Sugar.Debugf("Client IP from X-Real-IP: %s", xri)
			return xri
		}
	}

	if cfip := c.GetHeader("CF-Connecting-IP"); cfip != "" {
		if ip := net.ParseIP(cfip); ip != nil {
			logging.Sugar.Debugf("Client IP from CF-Connecting-IP: %s", cfip)
			return cfip
		}
	}

	if tcip := c.GetHeader("True-Client-IP"); tcip != "" {
		if ip := net.ParseIP(tcip); ip != nil {
			logging.Sugar.Debugf("Client IP from True-Client-IP: %s", tcip)
			return tcip
		}
	}

	remoteIP := c.RemoteIP()
	logging.Sugar.Debugf("Client IP from RemoteIP (fallback): %s", remoteIP)
	return remoteIP
}
