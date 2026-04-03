package accesslog

import (
	"context"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/system"
	"sync"
	"sync/atomic"
	"time"
)

var (
	accessLogBufferSize        = 4096
	accessLogBatchSize         = 200
	accessLogFlushInterval     = 1 * time.Second
	accessLogEnqueueTimeout    = 250 * time.Millisecond
	accessLogWriteTimeout      = 2 * time.Second
	accessLogBatchWriteTimeout = 5 * time.Second
)

type accessLogRepository interface {
	Create(log *entity.AccessLog) error
	CreateWithContext(ctx context.Context, log *entity.AccessLog) error
	BatchCreate(logs []*entity.AccessLog) error
	BatchCreateWithContext(ctx context.Context, logs []*entity.AccessLog) error
	List(limit, offset int, filters map[string]interface{}) ([]*entity.AccessLog, int64, error)
	DeleteBefore(before time.Time) (int64, error)
	DeleteAll() (int64, error)
	GetStats() (map[string]int64, error)
	GetTopBlockedIPs(limit int) ([]repository.TopBlockedIP, error)
	GetRequestTrend(hours int) ([]repository.HourlyTrend, error)
	GetBlockReasonBreakdown() ([]repository.BlockReasonCount, error)
}

type AccessLogService struct {
	repo                 accessLogRepository
	buffer               chan *entity.AccessLog
	done                 chan struct{}
	stopOnce             sync.Once
	directWriteFallbacks atomic.Uint64
	failedWrites         atomic.Uint64
}

var (
	accessLogService *AccessLogService
	accessLogOnce    sync.Once
)

func GetAccessLogService() *AccessLogService {
	accessLogOnce.Do(func() {
		svc := &AccessLogService{
			repo:   repository.NewAccessLogRepository(database.DB),
			buffer: make(chan *entity.AccessLog, accessLogBufferSize),
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
		ClientIP:       input.ClientIP,
		Method:         input.Method,
		Host:           input.Host,
		Scheme:         input.Scheme,
		Path:           input.Path,
		QueryString:    input.QueryString,
		ContentType:    input.ContentType,
		ContentLength:  input.ContentLength,
		Referer:        input.Referer,
		RequestHeaders: input.RequestHeaders,
		RequestBody:    input.RequestBody,
		StatusCode:     input.StatusCode,
		ResponseSize:   input.ResponseSize,
		Latency:        input.Latency,
		UserAgent:      input.UserAgent,
		Action:         input.Action,
		RuleName:       input.RuleName,
		CreatedAt:      time.Now().UTC(),
	}

	select {
	case s.buffer <- entry:
		return
	case <-s.done:
		s.persistDirect(entry, "service_stopping")
		return
	default:
	}

	timer := time.NewTimer(accessLogEnqueueTimeout)
	defer timer.Stop()

	select {
	case s.buffer <- entry:
	case <-s.done:
		s.persistDirect(entry, "service_stopping")
	case <-timer.C:
		s.persistDirect(entry, "buffer_saturated")
	}
}

func (s *AccessLogService) flushLoop() {
	ticker := time.NewTicker(accessLogFlushInterval)
	defer ticker.Stop()

	batch := make([]*entity.AccessLog, 0, accessLogBatchSize)

	for {
		select {
		case entry := <-s.buffer:
			batch = append(batch, entry)
			if len(batch) >= accessLogBatchSize {
				s.flush(batch)
				batch = make([]*entity.AccessLog, 0, accessLogBatchSize)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(batch)
				batch = make([]*entity.AccessLog, 0, accessLogBatchSize)
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
	ctx, cancel := context.WithTimeout(context.Background(), accessLogBatchWriteTimeout)
	err := s.repo.BatchCreateWithContext(ctx, batch)
	cancel()
	if err == nil {
		return
	}

	logging.Sugar.Errorf("Failed to flush access logs in batch, retrying individually: %v", err)
	for _, entry := range batch {
		s.persistEntry(entry, "batch_retry")
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
	retentionDays := resolveAccessLogRetentionDays()
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

func (s *AccessLogService) persistDirect(entry *entity.AccessLog, reason string) {
	count := s.directWriteFallbacks.Add(1)
	if shouldLogAccessLogStrategyWarning(count) {
		logging.Sugar.Warnf("Access log buffer saturated, writing directly to database (reason=%s, fallback_count=%d)", reason, count)
	}
	s.persistEntry(entry, reason)
}

func (s *AccessLogService) persistEntry(entry *entity.AccessLog, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), accessLogWriteTimeout)
	err := s.repo.CreateWithContext(ctx, entry)
	cancel()
	if err == nil {
		return
	}

	failures := s.failedWrites.Add(1)
	if shouldLogAccessLogStrategyWarning(failures) {
		logging.Sugar.Errorf("Failed to persist access log entry (reason=%s, failed_count=%d): %v", reason, failures, err)
	}
}

func shouldLogAccessLogStrategyWarning(count uint64) bool {
	return count == 1 || count%100 == 0
}

func resolveAccessLogRetentionDays() int {
	settings, err := system.GetSystemService().GetSettings()
	if err != nil || settings == nil || settings.AccessLogRetentionDays <= 0 {
		return 30
	}
	return settings.AccessLogRetentionDays
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
			ID:             item.ID,
			ClientIP:       item.ClientIP,
			Method:         item.Method,
			Host:           item.Host,
			Scheme:         item.Scheme,
			Path:           item.Path,
			QueryString:    item.QueryString,
			ContentType:    item.ContentType,
			ContentLength:  item.ContentLength,
			Referer:        item.Referer,
			RequestHeaders: item.RequestHeaders,
			RequestBody:    item.RequestBody,
			StatusCode:     item.StatusCode,
			ResponseSize:   item.ResponseSize,
			Latency:        item.Latency,
			UserAgent:      item.UserAgent,
			Action:         item.Action,
			RuleName:       item.RuleName,
			CreatedAt:      item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *AccessLogService) GetStats() (map[string]int64, error) {
	stats, err := s.repo.GetStats()
	if err != nil {
		return nil, err
	}
	stats["buffer_capacity"] = int64(cap(s.buffer))
	stats["buffer_pending"] = int64(len(s.buffer))
	stats["direct_write_fallbacks"] = int64(s.directWriteFallbacks.Load())
	stats["failed_persist_count"] = int64(s.failedWrites.Load())
	stats["retention_days"] = int64(resolveAccessLogRetentionDays())
	return stats, nil
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
	s.stopOnce.Do(func() {
		close(s.done)
	})
}

func StopAccessLogService() {
	if accessLogService != nil {
		accessLogService.Stop()
	}
}
