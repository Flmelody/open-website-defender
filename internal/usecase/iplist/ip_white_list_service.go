package iplist

import (
	"encoding/json"
	"errors"
	"fmt"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"time"

	"open-website-defender/internal/adapter/repository"
	_interface "open-website-defender/internal/usecase/interface"
	"sync"
)

type IpWhiteListService struct {
	repo _interface.IpWhiteListRepository
}

var (
	ipWhiteListService *IpWhiteListService
	ipWhiteListOnce    sync.Once
)

func GetIpWhiteListService() *IpWhiteListService {
	ipWhiteListOnce.Do(func() {
		svc := &IpWhiteListService{
			repo: repository.NewIpWhiteListRepository(database.DB),
		}
		// go svc.cleanupLoop() // expired entries are kept, expiration is checked at query time
		ipWhiteListService = svc
	})
	return ipWhiteListService
}

// cleanupLoop periodically removes expired whitelist entries.
func (s *IpWhiteListService) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		deleted, err := s.repo.DeleteExpired()
		if err != nil {
			logging.Sugar.Errorf("Failed to cleanup expired whitelist entries: %v", err)
			continue
		}
		if deleted > 0 {
			logging.Sugar.Infof("Cleaned up %d expired whitelist entries", deleted)
			event.Bus().Publish(event.WhiteListChanged)
		}
	}
}

func (s *IpWhiteListService) getRules() ([]*whiteListRule, error) {
	store := cache.Store()

	data, err := store.Get(cache.KeyWhiteListRules)
	if err == nil {
		var rules []*whiteListRule
		if err := json.Unmarshal(data, &rules); err == nil {
			return rules, nil
		}
	}

	list, _, err := s.repo.List(10000, 0)
	if err != nil {
		return nil, err
	}

	rules := make([]*whiteListRule, 0, len(list))
	for _, item := range list {
		rules = append(rules, &whiteListRule{IP: item.Ip, Domain: item.Domain, ExpiresAt: item.ExpiresAt})
	}

	data, err = json.Marshal(rules)
	if err == nil {
		store.Set(cache.KeyWhiteListRules, data, 3600)
	}

	return rules, nil
}

func (s *IpWhiteListService) Create(input *CreateIpWhiteListDto) (*IpWhiteListDto, error) {
	if input.Ip == "" {
		return nil, errors.New("ip is required")
	}

	existing, err := s.repo.FindByIP(input.Ip)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("ip already exists in whitelist")
	}

	item := &entity.IpWhiteList{
		Ip:        input.Ip,
		Domain:    input.Domain,
		Remark:    input.Remark,
		ExpiresAt: input.ExpiresAt,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create whitelist item: %w", err)
	}

	event.Bus().Publish(event.WhiteListChanged)

	return &IpWhiteListDto{
		ID:        item.ID,
		Ip:        item.Ip,
		Domain:    item.Domain,
		Remark:    item.Remark,
		Starred:   item.Starred,
		ExpiresAt: item.ExpiresAt,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *IpWhiteListService) Update(id uint, input *UpdateIpWhiteListDto) (*IpWhiteListDto, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("whitelist item not found")
	}

	if input.Ip != "" {
		// Check for duplicate IP (but allow same record)
		existing, err := s.repo.FindByIP(input.Ip)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("ip already exists in whitelist")
		}
		item.Ip = input.Ip
	}
	item.Domain = input.Domain
	item.Remark = input.Remark
	if input.Starred != nil {
		item.Starred = *input.Starred
	}
	item.ExpiresAt = input.ExpiresAt

	if err := s.repo.Update(item); err != nil {
		return nil, fmt.Errorf("failed to update whitelist item: %w", err)
	}

	event.Bus().Publish(event.WhiteListChanged)

	return &IpWhiteListDto{
		ID:        item.ID,
		Ip:        item.Ip,
		Domain:    item.Domain,
		Remark:    item.Remark,
		Starred:   item.Starred,
		ExpiresAt: item.ExpiresAt,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *IpWhiteListService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	event.Bus().Publish(event.WhiteListChanged)
	return nil
}

func (s *IpWhiteListService) List(page, size int) ([]*IpWhiteListDto, int64, error) {
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

	dtos := make([]*IpWhiteListDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, &IpWhiteListDto{
			ID:        item.ID,
			Ip:        item.Ip,
			Domain:    item.Domain,
			Remark:    item.Remark,
			Starred:   item.Starred,
			ExpiresAt: item.ExpiresAt,
			CreatedAt: item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *IpWhiteListService) FindByIP(ip string) (*IpWhiteListDto, error) {
	rules, err := s.getRules()
	if err != nil {
		logging.Sugar.Errorf("Failed to get whitelist rules: %v", err)
		return nil, err
	}

	now := time.Now().UTC()
	for _, rule := range rules {
		// Skip expired entries
		if rule.ExpiresAt != nil && rule.ExpiresAt.Before(now) {
			continue
		}
		if pkg.MatchIP(rule.IP, ip) {
			return &IpWhiteListDto{Ip: rule.IP, Domain: rule.Domain}, nil
		}
	}

	return nil, nil
}

// GrantTemporaryAccess creates a temporary whitelist entry with a TTL.
// If the IP already has a permanent entry covering this domain, it skips.
// If the IP has a temporary entry covering this domain, it renews the expiry.
func (s *IpWhiteListService) GrantTemporaryAccess(ip, domain string, ttlSeconds int, remark string) error {
	existing, err := s.repo.FindByIP(ip)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.Domain == "" || existing.Domain == domain || pkg.MatchDomain(existing.Domain, domain) {
			if existing.ExpiresAt == nil {
				// Permanent entry already covers this domain, skip
				return nil
			}
			// Temporary entry — renew expiry
			expiresAt := time.Now().UTC().Add(time.Duration(ttlSeconds) * time.Second)
			existing.ExpiresAt = &expiresAt
			if err := s.repo.Update(existing); err != nil {
				return fmt.Errorf("failed to renew temporary whitelist: %w", err)
			}
			event.Bus().Publish(event.WhiteListChanged)
			return nil
		}
	}

	expiresAt := time.Now().UTC().Add(time.Duration(ttlSeconds) * time.Second)
	item := &entity.IpWhiteList{
		Ip:        ip,
		Domain:    domain,
		Remark:    remark,
		ExpiresAt: &expiresAt,
	}

	if err := s.repo.Create(item); err != nil {
		return fmt.Errorf("failed to grant temporary whitelist access: %w", err)
	}

	logging.Sugar.Infof("Granted temporary whitelist access for IP %s domain %s (%ds): %s", ip, domain, ttlSeconds, remark)
	event.Bus().Publish(event.WhiteListChanged)
	return nil
}
