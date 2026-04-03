package accesslog

import (
	"context"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/logging"
	"testing"
	"time"
)

type stubAccessLogRepository struct {
	createCalls int
}

func (s *stubAccessLogRepository) Create(log *entity.AccessLog) error {
	s.createCalls++
	return nil
}

func (s *stubAccessLogRepository) CreateWithContext(ctx context.Context, log *entity.AccessLog) error {
	s.createCalls++
	return nil
}

func (s *stubAccessLogRepository) BatchCreate(logs []*entity.AccessLog) error {
	return nil
}

func (s *stubAccessLogRepository) BatchCreateWithContext(ctx context.Context, logs []*entity.AccessLog) error {
	return nil
}

func (s *stubAccessLogRepository) List(limit, offset int, filters map[string]interface{}) ([]*entity.AccessLog, int64, error) {
	return nil, 0, nil
}

func (s *stubAccessLogRepository) DeleteBefore(before time.Time) (int64, error) {
	return 0, nil
}

func (s *stubAccessLogRepository) DeleteAll() (int64, error) {
	return 0, nil
}

func (s *stubAccessLogRepository) GetStats() (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (s *stubAccessLogRepository) GetTopBlockedIPs(limit int) ([]repository.TopBlockedIP, error) {
	return nil, nil
}

func (s *stubAccessLogRepository) GetRequestTrend(hours int) ([]repository.HourlyTrend, error) {
	return nil, nil
}

func (s *stubAccessLogRepository) GetBlockReasonBreakdown() ([]repository.BlockReasonCount, error) {
	return nil, nil
}

func TestRecordFallsBackToDirectWriteWhenBufferIsSaturated(t *testing.T) {
	if logging.Sugar == nil {
		if err := logging.InitLoggerWithEnv("dev"); err != nil {
			t.Fatalf("failed to initialize logger: %v", err)
		}
		defer logging.Sync()
	}

	oldTimeout := accessLogEnqueueTimeout
	accessLogEnqueueTimeout = 5 * time.Millisecond
	defer func() {
		accessLogEnqueueTimeout = oldTimeout
	}()

	repo := &stubAccessLogRepository{}
	svc := &AccessLogService{
		repo:   repo,
		buffer: make(chan *entity.AccessLog, 1),
		done:   make(chan struct{}),
	}
	svc.buffer <- &entity.AccessLog{}

	svc.Record(&AccessLogInput{
		ClientIP: "127.0.0.1",
		Method:   "GET",
		Path:     "/healthz",
		Action:   "allowed",
	})

	if got := repo.createCalls; got != 1 {
		t.Fatalf("expected direct write fallback, got %d create calls", got)
	}
	if got := svc.directWriteFallbacks.Load(); got != 1 {
		t.Fatalf("expected fallback counter to be 1, got %d", got)
	}
}

func TestShouldLogAccessLogStrategyWarning(t *testing.T) {
	cases := map[uint64]bool{
		1:   true,
		2:   false,
		99:  false,
		100: true,
		200: true,
	}

	for count, want := range cases {
		if got := shouldLogAccessLogStrategyWarning(count); got != want {
			t.Fatalf("count=%d: expected %v, got %v", count, want, got)
		}
	}
}
