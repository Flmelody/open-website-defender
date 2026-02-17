package repository

import (
	"open-website-defender/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type AccessLogRepository struct {
	db *gorm.DB
}

func NewAccessLogRepository(db *gorm.DB) *AccessLogRepository {
	return &AccessLogRepository{db: db}
}

func (r *AccessLogRepository) Create(log *entity.AccessLog) error {
	return r.db.Create(log).Error
}

func (r *AccessLogRepository) BatchCreate(logs []*entity.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(logs, 100).Error
}

func (r *AccessLogRepository) List(limit, offset int, filters map[string]interface{}) ([]*entity.AccessLog, int64, error) {
	var logs []*entity.AccessLog
	var total int64

	query := r.db.Model(&entity.AccessLog{})

	if ip, ok := filters["client_ip"]; ok && ip != "" {
		query = query.Where("client_ip = ?", ip)
	}
	if action, ok := filters["action"]; ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if statusCode, ok := filters["status_code"]; ok && statusCode != 0 {
		query = query.Where("status_code = ?", statusCode)
	}
	if startTime, ok := filters["start_time"]; ok {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["end_time"]; ok {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *AccessLogRepository) DeleteBefore(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&entity.AccessLog{})
	return result.RowsAffected, result.Error
}

func (r *AccessLogRepository) DeleteAll() (int64, error) {
	result := r.db.Where("1 = 1").Delete(&entity.AccessLog{})
	return result.RowsAffected, result.Error
}

func (r *AccessLogRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	var total int64
	r.db.Model(&entity.AccessLog{}).Count(&total)
	stats["total"] = total

	var blocked int64
	r.db.Model(&entity.AccessLog{}).Where("action LIKE 'blocked_%'").Count(&blocked)
	stats["blocked"] = blocked

	return stats, nil
}

type TopBlockedIP struct {
	ClientIP string `json:"client_ip"`
	Count    int64  `json:"count"`
}

func (r *AccessLogRepository) GetTopBlockedIPs(limit int) ([]TopBlockedIP, error) {
	var results []TopBlockedIP
	err := r.db.Model(&entity.AccessLog{}).
		Select("client_ip, count(*) as count").
		Where("action LIKE 'blocked_%'").
		Group("client_ip").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}
