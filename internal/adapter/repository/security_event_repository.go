package repository

import (
	"open-website-defender/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type SecurityEventRepository struct {
	db *gorm.DB
}

func NewSecurityEventRepository(db *gorm.DB) *SecurityEventRepository {
	return &SecurityEventRepository{db: db}
}

func (r *SecurityEventRepository) Create(event *entity.SecurityEvent) error {
	return r.db.Create(event).Error
}

func (r *SecurityEventRepository) BatchCreate(events []*entity.SecurityEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.CreateInBatches(events, 100).Error
}

func (r *SecurityEventRepository) List(limit, offset int, filters map[string]interface{}) ([]*entity.SecurityEvent, int64, error) {
	var list []*entity.SecurityEvent
	var total int64

	query := r.db.Model(&entity.SecurityEvent{})

	if eventType, ok := filters["event_type"].(string); ok && eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if clientIP, ok := filters["client_ip"].(string); ok && clientIP != "" {
		query = query.Where("client_ip = ?", clientIP)
	}
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

type SecurityEventStats struct {
	TotalEvents int64            `json:"total_events"`
	AutoBans24h int64            `json:"auto_bans_24h"`
	TopIPs      []TopThreatIP    `json:"top_ips"`
	TypeCounts  []EventTypeCount `json:"type_counts"`
}

type TopThreatIP struct {
	ClientIP    string `json:"client_ip"`
	Count       int64  `json:"count"`
	ThreatScore int    `json:"threat_score"`
}

type EventTypeCount struct {
	EventType string `json:"event_type"`
	Count     int64  `json:"count"`
}

func (r *SecurityEventRepository) GetStats() (*SecurityEventStats, error) {
	stats := &SecurityEventStats{}

	// Total events
	if err := r.db.Model(&entity.SecurityEvent{}).Count(&stats.TotalEvents).Error; err != nil {
		return nil, err
	}

	// Auto-bans in last 24h
	since := time.Now().UTC().Add(-24 * time.Hour)
	if err := r.db.Model(&entity.SecurityEvent{}).
		Where("event_type = ? AND created_at >= ?", "auto_ban", since).
		Count(&stats.AutoBans24h).Error; err != nil {
		return nil, err
	}

	// Top IPs (last 24h) with threat score derived from event types
	// Score weights: brute_force=10, scan_detected=5, auto_ban=3, other=1
	var topIPs []TopThreatIP
	if err := r.db.Model(&entity.SecurityEvent{}).
		Select(`client_ip, COUNT(*) as count, SUM(
			CASE event_type
				WHEN 'brute_force' THEN 10
				WHEN 'scan_detected' THEN 5
				WHEN 'auto_ban' THEN 3
				ELSE 1
			END
		) as threat_score`).
		Where("created_at >= ?", since).
		Group("client_ip").
		Order("count DESC").
		Limit(10).
		Scan(&topIPs).Error; err != nil {
		return nil, err
	}
	stats.TopIPs = topIPs

	// Event type breakdown
	var typeCounts []EventTypeCount
	if err := r.db.Model(&entity.SecurityEvent{}).
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Order("count DESC").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}
	stats.TypeCounts = typeCounts

	return stats, nil
}

func (r *SecurityEventRepository) DeleteBefore(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&entity.SecurityEvent{})
	return result.RowsAffected, result.Error
}
