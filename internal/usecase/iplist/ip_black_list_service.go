package iplist

import (
	"encoding/json"
	"errors"
	"fmt"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"
	"sync"
)

type IpBlackListService struct {
	repo _interface.IpBlackListRepository
}

var (
	ipBlackListService *IpBlackListService
	ipBlackListOnce    sync.Once
)

const (
	cacheKeyBlackListRules = "blacklist:rules"
	cacheKeyBlackListIP    = "blacklist:ip:"
)

func GetIpBlackListService() *IpBlackListService {
	ipBlackListOnce.Do(func() {
		ipBlackListService = &IpBlackListService{
			repo: repository.NewIpBlackListRepository(database.DB),
		}
	})
	return ipBlackListService
}

func (s *IpBlackListService) getRules() ([]string, error) {
	cache := pkg.Cacher()

	// Try to get from cache
	data, err := cache.Get([]byte(cacheKeyBlackListRules))
	if err == nil {
		var rules []string
		if err := json.Unmarshal(data, &rules); err == nil {
			return rules, nil
		}
	}

	// Load from DB
	// Assuming reasonable count for now.
	list, _, err := s.repo.List(10000, 0)
	if err != nil {
		return nil, err
	}

	rules := make([]string, 0, len(list))
	for _, item := range list {
		rules = append(rules, item.Ip)
	}

	// Cache it
	data, err = json.Marshal(rules)
	if err == nil {
		cache.Set([]byte(cacheKeyBlackListRules), data, 3600) // 1 hour TTL for rules list
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
		Ip: input.Ip,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create blacklist item: %w", err)
	}

	// Invalidate cache
	pkg.Cacher().Del([]byte(cacheKeyBlackListRules))
	// Note: individual IP decisions might still be cached for a short time

	return &IpBlackListDto{
		ID:        item.ID,
		Ip:        item.Ip,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *IpBlackListService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	pkg.Cacher().Del([]byte(cacheKeyBlackListRules))
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
			CreatedAt: item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *IpBlackListService) FindByIP(ip string) (*IpBlackListDto, error) {
	cache := pkg.Cacher()
	cacheKey := []byte(cacheKeyBlackListIP + ip)

	// Try to get cached decision for this IP
	if val, err := cache.Get(cacheKey); err == nil {
		if len(val) == 0 {
			return nil, nil // Cached as "not found"
		}
		return &IpBlackListDto{Ip: string(val)}, nil // Cached as "found", val is the rule
	}

	// Slow path: match against rules
	rules, err := s.getRules()
	if err != nil {
		logging.Sugar.Errorf("Failed to get blacklist rules: %v", err)
		return nil, err
	}

	for _, rule := range rules {
		if pkg.MatchIP(rule, ip) {
			cache.Set(cacheKey, []byte(rule), 600) // 10 min TTL for IP check
			return &IpBlackListDto{Ip: rule}, nil
		}
	}

	// Cache negative result (empty byte slice)
	cache.Set(cacheKey, []byte{}, 600)

	return nil, nil
}
