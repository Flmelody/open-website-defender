package accesslog

import (
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"sync"
	"time"
)

type AccessLogService struct {
	repo   *repository.AccessLogRepository
	buffer chan *entity.AccessLog
	done   chan struct{}
}

var (
	accessLogService *AccessLogService
	accessLogOnce    sync.Once
)

func GetAccessLogService() *AccessLogService {
	accessLogOnce.Do(func() {
		svc := &AccessLogService{
			repo:   repository.NewAccessLogRepository(database.DB),
			buffer: make(chan *entity.AccessLog, 1000),
			done:   make(chan struct{}),
		}
		go svc.flushLoop()
		go svc.retentionLoop()
		accessLogService = svc
	})
	return accessLogService
}

// Record adds an access log entry to the async buffer.
func (s *AccessLogService) Record(input *AccessLogInput) {
	entry := &entity.AccessLog{
		ClientIP:   input.ClientIP,
		Method:     input.Method,
		Path:       input.Path,
		StatusCode: input.StatusCode,
		Latency:    input.Latency,
		UserAgent:  input.UserAgent,
		Action:     input.Action,
		RuleName:   input.RuleName,
		CreatedAt:  time.Now().UTC(),
	}

	select {
	case s.buffer <- entry:
	default:
		// Buffer full, drop the entry to avoid blocking request processing
		logging.Sugar.Warn("Access log buffer full, dropping entry")
	}
}

func (s *AccessLogService) flushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	batch := make([]*entity.AccessLog, 0, 100)

	for {
		select {
		case entry := <-s.buffer:
			batch = append(batch, entry)
			if len(batch) >= 100 {
				s.flush(batch)
				batch = make([]*entity.AccessLog, 0, 100)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(batch)
				batch = make([]*entity.AccessLog, 0, 100)
			}
		case <-s.done:
			// Flush remaining
			for {
				select {
				case entry := <-s.buffer:
					batch = append(batch, entry)
				default:
					if len(batch) > 0 {
						s.flush(batch)
					}
					return
				}
			}
		}
	}
}

func (s *AccessLogService) flush(batch []*entity.AccessLog) {
	if err := s.repo.BatchCreate(batch); err != nil {
		logging.Sugar.Errorf("Failed to flush access logs: %v", err)
	}
}

func (s *AccessLogService) retentionLoop() {
	// Run retention cleanup daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Also run once on startup after a short delay
	time.AfterFunc(1*time.Minute, func() {
		s.cleanupOldLogs()
	})

	for {
		select {
		case <-ticker.C:
			s.cleanupOldLogs()
		case <-s.done:
			return
		}
	}
}

func (s *AccessLogService) cleanupOldLogs() {
	retentionDays := 30
	before := time.Now().UTC().AddDate(0, 0, -retentionDays)
	deleted, err := s.repo.DeleteBefore(before)
	if err != nil {
		logging.Sugar.Errorf("Failed to cleanup old access logs: %v", err)
		return
	}
	if deleted > 0 {
		logging.Sugar.Infof("Cleaned up %d access logs older than %d days", deleted, retentionDays)
	}
}

func (s *AccessLogService) List(page, size int, filters map[string]interface{}) ([]*AccessLogDto, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size
	list, total, err := s.repo.List(size, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*AccessLogDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, &AccessLogDto{
			ID:         item.ID,
			ClientIP:   item.ClientIP,
			Method:     item.Method,
			Path:       item.Path,
			StatusCode: item.StatusCode,
			Latency:    item.Latency,
			UserAgent:  item.UserAgent,
			Action:     item.Action,
			RuleName:   item.RuleName,
			CreatedAt:  item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *AccessLogService) GetStats() (map[string]int64, error) {
	return s.repo.GetStats()
}

func (s *AccessLogService) GetTopBlockedIPs(limit int) ([]repository.TopBlockedIP, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetTopBlockedIPs(limit)
}

func (s *AccessLogService) GetRequestTrend(hours int) ([]repository.HourlyTrend, error) {
	if hours <= 0 {
		hours = 24
	}
	return s.repo.GetRequestTrend(hours)
}

func (s *AccessLogService) GetBlockReasonBreakdown() ([]repository.BlockReasonCount, error) {
	return s.repo.GetBlockReasonBreakdown()
}

func (s *AccessLogService) ClearAll() (int64, error) {
	return s.repo.DeleteAll()
}

// Stop gracefully stops the service by flushing remaining logs.
func (s *AccessLogService) Stop() {
	close(s.done)
}
