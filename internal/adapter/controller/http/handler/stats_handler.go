package handler

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/accesslog"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func GetDashboardStats(c *gin.Context) {
	logService := accesslog.GetAccessLogService()

	// Access log stats
	logStats, err := logService.GetStats()
	if err != nil {
		logging.Sugar.Errorf("Failed to get access log stats: %v", err)
		logStats = map[string]int64{}
	}

	topBlocked, err := logService.GetTopBlockedIPs(10)
	if err != nil {
		logging.Sugar.Errorf("Failed to get top blocked IPs: %v", err)
	}

	// Entity counts
	var blacklistCount, whitelistCount, userCount, wafRuleCount int64
	database.DB.Model(&entity.IpBlackList{}).Count(&blacklistCount)
	database.DB.Model(&entity.IpWhiteList{}).Count(&whitelistCount)
	database.DB.Model(&entity.User{}).Count(&userCount)
	database.DB.Model(&entity.WafRule{}).Count(&wafRuleCount)

	// Uptime
	uptime := time.Since(startTime)

	response.Success(c, gin.H{
		"total_requests":   logStats["total"],
		"blocked_requests": logStats["blocked"],
		"blacklist_count":  blacklistCount,
		"whitelist_count":  whitelistCount,
		"user_count":       userCount,
		"waf_rule_count":   wafRuleCount,
		"top_blocked_ips":  topBlocked,
		"uptime_seconds":   int64(uptime.Seconds()),
	})
}
