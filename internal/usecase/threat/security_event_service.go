package threat

import (
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"sync"
	"time"
)

type SecurityEventService struct {
	repo   *repository.SecurityEventRepository
	buffer chan *entity.SecurityEvent
	done   chan struct{}
}

var (
	securityEventService *SecurityEventService
	securityEventOnce    sync.Once
)

func GetSecurityEventService() *SecurityEventService {
	securityEventOnce.Do(func() {
		svc := &SecurityEventService{
			repo:   repository.NewSecurityEventRepository(database.DB),
			buffer: make(chan *entity.SecurityEvent, 500),
			done:   make(chan struct{}),
		}
		go svc.flushLoop()
		go svc.retentionLoop()
		securityEventService = svc
	})
	return securityEventService
}

// Record adds a security event to the async buffer.
func (s *SecurityEventService) Record(eventType, clientIP, detail string) {
	event := &entity.SecurityEvent{
		EventType: eventType,
		ClientIP:  clientIP,
		Detail:    detail,
		CreatedAt: time.Now().UTC(),
	}

	select {
	case s.buffer <- event:
	default:
		logging.Sugar.Warn("Security event buffer full, dropping entry")
	}
}

func (s *SecurityEventService) flushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	batch := make([]*entity.SecurityEvent, 0, 50)

	for {
		select {
		case event := <-s.buffer:
			batch = append(batch, event)
			if len(batch) >= 50 {
				s.flush(batch)
				batch = make([]*entity.SecurityEvent, 0, 50)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(batch)
				batch = make([]*entity.SecurityEvent, 0, 50)
			}
		case <-s.done:
			for {
				select {
				case event := <-s.buffer:
					batch = append(batch, event)
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

func (s *SecurityEventService) flush(batch []*entity.SecurityEvent) {
	if err := s.repo.BatchCreate(batch); err != nil {
		logging.Sugar.Errorf("Failed to flush security events: %v", err)
	}
}

func (s *SecurityEventService) retentionLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	time.AfterFunc(2*time.Minute, func() {
		s.cleanupOldEvents()
	})

	for {
		select {
		case <-ticker.C:
			s.cleanupOldEvents()
		case <-s.done:
			return
		}
	}
}

func (s *SecurityEventService) cleanupOldEvents() {
	retentionDays := 90
	before := time.Now().UTC().AddDate(0, 0, -retentionDays)
	deleted, err := s.repo.DeleteBefore(before)
	if err != nil {
		logging.Sugar.Errorf("Failed to cleanup old security events: %v", err)
		return
	}
	if deleted > 0 {
		logging.Sugar.Infof("Cleaned up %d security events older than %d days", deleted, retentionDays)
	}
}

type SecurityEventDto struct {
	ID        uint      `json:"id"`
	EventType string    `json:"event_type"`
	ClientIP  string    `json:"client_ip"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *SecurityEventService) List(page, size int, filters map[string]interface{}) ([]*SecurityEventDto, int64, error) {
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

	dtos := make([]*SecurityEventDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, &SecurityEventDto{
			ID:        item.ID,
			EventType: item.EventType,
			ClientIP:  item.ClientIP,
			Detail:    item.Detail,
			CreatedAt: item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *SecurityEventService) GetStats() (*repository.SecurityEventStats, error) {
	return s.repo.GetStats()
}
