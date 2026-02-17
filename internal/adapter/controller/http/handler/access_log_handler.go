package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/accesslog"
	"time"

	"github.com/gin-gonic/gin"
)

func ListAccessLogs(c *gin.Context) {
	service := accesslog.GetAccessLogService()

	var req request.ListAccessLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logging.Sugar.Errorf("Invalid query parameters: %v", err)
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	filters := make(map[string]interface{})
	if req.ClientIP != "" {
		filters["client_ip"] = req.ClientIP
	}
	if req.Action != "" {
		filters["action"] = req.Action
	}
	if req.StatusCode != 0 {
		filters["status_code"] = req.StatusCode
	}
	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			filters["start_time"] = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			filters["end_time"] = t
		}
	}

	list, total, err := service.List(req.Page, req.Size, filters)
	if err != nil {
		logging.Sugar.Errorf("Failed to list access logs: %v", err)
		response.InternalServerError(c, "Failed to list access logs")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}

func ClearAccessLogs(c *gin.Context) {
	service := accesslog.GetAccessLogService()

	deleted, err := service.ClearAll()
	if err != nil {
		logging.Sugar.Errorf("Failed to clear access logs: %v", err)
		response.InternalServerError(c, "Failed to clear access logs")
		return
	}

	logging.Sugar.Infof("Cleared %d access logs", deleted)
	response.Success(c, gin.H{"deleted": deleted})
}

func GetAccessLogStats(c *gin.Context) {
	service := accesslog.GetAccessLogService()

	stats, err := service.GetStats()
	if err != nil {
		logging.Sugar.Errorf("Failed to get access log stats: %v", err)
		response.InternalServerError(c, "Failed to get stats")
		return
	}

	topBlocked, err := service.GetTopBlockedIPs(10)
	if err != nil {
		logging.Sugar.Errorf("Failed to get top blocked IPs: %v", err)
		response.InternalServerError(c, "Failed to get top blocked IPs")
		return
	}

	response.Success(c, gin.H{
		"stats":           stats,
		"top_blocked_ips": topBlocked,
	})
}
