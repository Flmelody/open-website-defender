package handler

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/threat"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ListSecurityEvents(c *gin.Context) {
	service := threat.GetSecurityEventService()

	page := 1
	size := 20
	if p, ok := c.GetQuery("page"); ok {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if s, ok := c.GetQuery("size"); ok {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			size = v
		}
	}

	filters := make(map[string]interface{})
	if eventType := c.Query("event_type"); eventType != "" {
		filters["event_type"] = eventType
	}
	if clientIP := c.Query("client_ip"); clientIP != "" {
		filters["client_ip"] = clientIP
	}
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filters["start_time"] = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filters["end_time"] = t
		}
	}

	list, total, err := service.List(page, size, filters)
	if err != nil {
		logging.Sugar.Errorf("Failed to list security events: %v", err)
		response.InternalServerError(c, "Failed to list security events")
		return
	}

	response.PageSuccess(c, list, total, page, size)
}

func GetThreatScore(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		response.BadRequest(c, "ip parameter is required")
		return
	}

	td := threat.GetThreatDetector()
	score := td.GetThreatScore(ip)

	response.Success(c, gin.H{
		"ip":           ip,
		"threat_score": score,
	})
}

func GetSecurityEventStats(c *gin.Context) {
	service := threat.GetSecurityEventService()

	stats, err := service.GetStats()
	if err != nil {
		logging.Sugar.Errorf("Failed to get security event stats: %v", err)
		response.InternalServerError(c, "Failed to get security event stats")
		return
	}

	response.Success(c, stats)
}
