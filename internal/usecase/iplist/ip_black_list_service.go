package iplist

import (
	"encoding/json"
	"errors"
	"fmt"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"
	"sync"
	"time"
)

type IpBlackListService struct {
	repo _interface.IpBlackListRepository
}

var (
	ipBlackListService *IpBlackListService
	ipBlackListOnce    sync.Once
)

func GetIpBlackListService() *IpBlackListService {
	ipBlackListOnce.Do(func() {
		svc := &IpBlackListService{
			repo: repository.NewIpBlackListRepository(database.DB),
		}
		// go svc.cleanupLoop() // expired entries are kept, expiration is checked at query time
		ipBlackListService = svc
	})
	return ipBlackListService
}

// cleanupLoop periodically removes expired blacklist entries.
func (s *IpBlackListService) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		deleted, err := s.repo.DeleteExpired()
		if err != nil {
			logging.Sugar.Errorf("Failed to cleanup expired blacklist entries: %v", err)
			continue
		}
		if deleted > 0 {
			logging.Sugar.Infof("Cleaned up %d expired blacklist entries", deleted)
			event.Bus().Publish(event.BlackListChanged)
		}
	}
}

type blacklistRule struct {
	Ip        string     `json:"ip"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func (s *IpBlackListService) getRules() ([]blacklistRule, error) {
	store := cache.Store()

	data, err := store.Get(cache.KeyBlackListRules)
	if err == nil {
		var rules []blacklistRule
		if err := json.Unmarshal(data, &rules); err == nil {
			return rules, nil
		}
	}

	list, _, err := s.repo.List(10000, 0)
	if err != nil {
		return nil, err
	}

	rules := make([]blacklistRule, 0, len(list))
	for _, item := range list {
		rules = append(rules, blacklistRule{
			Ip:        item.Ip,
			ExpiresAt: item.ExpiresAt,
		})
	}

	data, err = json.Marshal(rules)
	if err == nil {
		store.Set(cache.KeyBlackListRules, data, 3600)
	}

	return rules, nil
}

func (s *IpBlackListService) Create(input *CreateIpBlackListDto) (*IpBlackListDto, error) {
	if input.Ip == "" {
		return nil, errors.New("ip is required")
	}

	existing, err := s.repo.FindByIP(input.Ip)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("ip already exists in blacklist")
	}

	item := &entity.IpBlackList{
		Ip:        input.Ip,
		Remark:    input.Remark,
		ExpiresAt: input.ExpiresAt,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create blacklist item: %w", err)
	}

	event.Bus().Publish(event.BlackListChanged)

	return &IpBlackListDto{
		ID:        item.ID,
		Ip:        item.Ip,
		Remark:    item.Remark,
		ExpiresAt: item.ExpiresAt,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *IpBlackListService) Update(id uint, input *UpdateIpBlackListDto) (*IpBlackListDto, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("blacklist item not found")
	}

	item.Remark = input.Remark
	item.ExpiresAt = input.ExpiresAt

	if err := s.repo.Update(item); err != nil {
		return nil, fmt.Errorf("failed to update blacklist item: %w", err)
	}

	event.Bus().Publish(event.BlackListChanged)

	return &IpBlackListDto{
		ID:        item.ID,
		Ip:        item.Ip,
		Remark:    item.Remark,
		ExpiresAt: item.ExpiresAt,
		CreatedAt: item.CreatedAt,
	}, nil
}

// CreateAutoBlacklist adds an IP to the blacklist with an automatic expiry duration and remark.
// Returns (true, nil) if a new ban was created, (false, nil) if already banned.
// If the existing entry is expired, it will be overwritten with the new ban.
func (s *IpBlackListService) CreateAutoBlacklist(ip, remark string, duration time.Duration) (bool, error) {
	existing, err := s.repo.FindByIP(ip)
	if err != nil {
		return false, err
	}

	expiresAt := time.Now().UTC().Add(duration)

	if existing != nil {
		// If the existing entry is expired, overwrite it
		if existing.ExpiresAt != nil && existing.ExpiresAt.Before(time.Now().UTC()) {
			existing.Remark = remark
			existing.ExpiresAt = &expiresAt
			if err := s.repo.Update(existing); err != nil {
				return false, fmt.Errorf("failed to update expired blacklist entry: %w", err)
			}
			logging.Sugar.Infof("Auto-blacklisted IP %s for %v (overwriting expired entry): %s", ip, duration, remark)
			event.Bus().Publish(event.BlackListChanged)
			return true, nil
		}
		// Already blacklisted and not expired, skip
		return false, nil
	}

	item := &entity.IpBlackList{
		Ip:        ip,
		Remark:    remark,
		ExpiresAt: &expiresAt,
	}

	if err := s.repo.Create(item); err != nil {
		return false, fmt.Errorf("failed to auto-blacklist IP: %w", err)
	}

	logging.Sugar.Infof("Auto-blacklisted IP %s for %v: %s", ip, duration, remark)
	event.Bus().Publish(event.BlackListChanged)
	return true, nil
}

func (s *IpBlackListService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	event.Bus().Publish(event.BlackListChanged)
	return nil
}

func (s *IpBlackListService) List(page, size int) ([]*IpBlackListDto, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	offset := (page - 1) * size
	list, total, err := s.repo.List(size, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*IpBlackListDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, &IpBlackListDto{
			ID:        item.ID,
			Ip:        item.Ip,
			Remark:    item.Remark,
			ExpiresAt: item.ExpiresAt,
			CreatedAt: item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *IpBlackListService) FindByIP(ip string) (*IpBlackListDto, error) {
	rules, err := s.getRules()
	if err != nil {
		logging.Sugar.Errorf("Failed to get blacklist rules: %v", err)
		return nil, err
	}

	now := time.Now().UTC()
	for _, rule := range rules {
		// Skip expired entries
		if rule.ExpiresAt != nil && rule.ExpiresAt.Before(now) {
			continue
		}
		if pkg.MatchIP(rule.Ip, ip) {
			return &IpBlackListDto{Ip: rule.Ip}, nil
		}
	}

	return nil, nil
}
