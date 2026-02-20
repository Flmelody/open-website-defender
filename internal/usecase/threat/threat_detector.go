package threat

import (
	"encoding/binary"
	"fmt"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"open-website-defender/internal/usecase/iplist"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/spf13/viper"
)

type ThreatDetector struct {
	blacklistSvc *iplist.IpBlackListService
}

var (
	threatDetector     *ThreatDetector
	threatDetectorOnce sync.Once
)

func GetThreatDetector() *ThreatDetector {
	threatDetectorOnce.Do(func() {
		threatDetector = &ThreatDetector{
			blacklistSvc: iplist.GetIpBlackListService(),
		}
	})
	return threatDetector
}

// RecordRequest is called from access_log middleware after each request.
// It tracks 4xx responses and rate limit hits per IP for anomaly detection.
func (td *ThreatDetector) RecordRequest(ip string, statusCode int, wasRateLimited bool) {
	if !viper.GetBool("threat-detection.enabled") {
		return
	}

	// Skip threat counting for already-banned IPs to avoid feedback loop:
	// banned IP → 403 → counts as 4xx → re-triggers ban → duplicate event
	if dto, _ := td.blacklistSvc.FindByIP(ip); dto != nil {
		return
	}

	cache := pkg.Cacher()

	// Track 4xx responses
	if statusCode >= 400 && statusCode < 500 {
		threshold := viper.GetInt("threat-detection.status-code-threshold")
		if threshold <= 0 {
			threshold = 20
		}
		window := viper.GetInt("threat-detection.status-code-window")
		if window <= 0 {
			window = 60
		}

		key := []byte(fmt.Sprintf("threat:4xx:%s", ip))
		count := td.incrementCounter(cache, key, window)

		if count >= int64(threshold) {
			banDuration := viper.GetInt("threat-detection.auto-ban-duration")
			if banDuration <= 0 {
				banDuration = 3600
			}
			td.checkAndBan(ip, "excessive 4xx responses", time.Duration(banDuration)*time.Second)
			// Reset counter after banning to avoid repeated bans
			cache.Del(key)
		}
	}

	// Track 404s for scan detection (Batch 2)
	if statusCode == 404 {
		threshold := viper.GetInt("threat-detection.scan-threshold")
		if threshold <= 0 {
			threshold = 10
		}
		window := viper.GetInt("threat-detection.scan-window")
		if window <= 0 {
			window = 300
		}

		key := []byte(fmt.Sprintf("threat:scan:%s", ip))
		count := td.incrementCounter(cache, key, window)

		if count >= int64(threshold) {
			banDuration := viper.GetInt("threat-detection.scan-ban-duration")
			if banDuration <= 0 {
				banDuration = 14400
			}
			td.checkAndBan(ip, "path scanning", time.Duration(banDuration)*time.Second)
			cache.Del(key)
		}
	}

	// Track rate limit abuse
	if wasRateLimited {
		threshold := viper.GetInt("threat-detection.rate-limit-abuse-threshold")
		if threshold <= 0 {
			threshold = 5
		}
		window := viper.GetInt("threat-detection.rate-limit-abuse-window")
		if window <= 0 {
			window = 300
		}

		key := []byte(fmt.Sprintf("threat:ratelimit:%s", ip))
		count := td.incrementCounter(cache, key, window)

		if count >= int64(threshold) {
			banDuration := viper.GetInt("threat-detection.auto-ban-duration")
			if banDuration <= 0 {
				banDuration = 3600
			}
			td.checkAndBan(ip, "rate limit abuse", time.Duration(banDuration*2)*time.Second)
			cache.Del(key)
		}
	}
}

// RecordFailedLogin tracks failed login attempts per IP for brute force detection.
func (td *ThreatDetector) RecordFailedLogin(ip string) {
	if !viper.GetBool("threat-detection.enabled") {
		return
	}

	cache := pkg.Cacher()
	threshold := viper.GetInt("threat-detection.brute-force-threshold")
	if threshold <= 0 {
		threshold = 10
	}
	window := viper.GetInt("threat-detection.brute-force-window")
	if window <= 0 {
		window = 600
	}

	key := []byte(fmt.Sprintf("threat:bruteforce:%s", ip))
	count := td.incrementCounter(cache, key, window)

	if count >= int64(threshold) {
		banDuration := viper.GetInt("threat-detection.brute-force-ban-duration")
		if banDuration <= 0 {
			banDuration = 3600
		}
		td.checkAndBan(ip, "brute force", time.Duration(banDuration)*time.Second)
		cache.Del(key)
	}
}

// GetThreatScore returns the current threat score for an IP.
// Higher score = more suspicious activity.
func (td *ThreatDetector) GetThreatScore(ip string) int {
	cache := pkg.Cacher()
	key := []byte(fmt.Sprintf("threat:score:%s", ip))
	val, err := cache.Get(key)
	if err != nil || len(val) < 8 {
		return 0
	}
	return int(binary.BigEndian.Uint64(val))
}

// AddThreatScore adds points to an IP's threat score.
func (td *ThreatDetector) AddThreatScore(ip string, points int) {
	cache := pkg.Cacher()
	key := []byte(fmt.Sprintf("threat:score:%s", ip))

	val, err := cache.Get(key)
	var score int64
	if err == nil && len(val) == 8 {
		score = int64(binary.BigEndian.Uint64(val))
	}
	score += int64(points)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(score))
	_ = cache.Set(key, buf, 3600) // 1 hour TTL, score decays naturally
}

func (td *ThreatDetector) incrementCounter(cache *freecache.Cache, key []byte, ttlSeconds int) int64 {
	val, err := cache.Get(key)
	var count int64
	if err == nil && len(val) == 8 {
		count = int64(binary.BigEndian.Uint64(val))
	}

	count++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(count))
	_ = cache.Set(key, buf, ttlSeconds)

	return count
}

func (td *ThreatDetector) checkAndBan(ip, reason string, duration time.Duration) {
	created, err := td.blacklistSvc.CreateAutoBlacklist(ip, reason, duration)
	if err != nil {
		logging.Sugar.Errorf("Failed to auto-ban IP %s (%s): %v", ip, reason, err)
		return
	}
	if !created {
		// Already banned, don't record duplicate event
		return
	}
	logging.Sugar.Warnf("Auto-banned IP %s for %v: %s", ip, duration, reason)

	// Emit security event
	eventType := "auto_ban"
	if reason == "brute force" {
		eventType = "brute_force"
	} else if reason == "path scanning" {
		eventType = "scan_detected"
	}
	detail := fmt.Sprintf("IP %s auto-banned for %v: %s", ip, duration, reason)
	GetSecurityEventService().Record(eventType, ip, detail)

	// Send webhook notification
	go sendWebhookNotification(eventType, ip, reason, duration)
}
