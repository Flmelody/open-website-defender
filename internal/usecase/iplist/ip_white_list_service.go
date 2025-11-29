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
	"path/filepath"
	"sync"
)

type IpWhiteListService struct {
	repo _interface.IpWhiteListRepository
}

var (
	ipWhiteListService *IpWhiteListService
	ipWhiteListOnce    sync.Once
)

const (
	cacheKeyWhiteListRules = "whitelist:rules"
	cacheKeyWhiteListIP    = "whitelist:ip:"
)

func GetIpWhiteListService() *IpWhiteListService {
	ipWhiteListOnce.Do(func() {
		ipWhiteListService = &IpWhiteListService{
			repo: repository.NewIpWhiteListRepository(database.DB),
		}
	})
	return ipWhiteListService
}

func (s *IpWhiteListService) getRules() ([]string, error) {
	cache := pkg.Cacher()

	data, err := cache.Get([]byte(cacheKeyWhiteListRules))
	if err == nil {
		var rules []string
		if err := json.Unmarshal(data, &rules); err == nil {
			return rules, nil
		}
	}

	list, _, err := s.repo.List(10000, 0)
	if err != nil {
		return nil, err
	}

	rules := make([]string, 0, len(list))
	for _, item := range list {
		rules = append(rules, item.Ip)
	}

	data, err = json.Marshal(rules)
	if err == nil {
		cache.Set([]byte(cacheKeyWhiteListRules), data, 3600)
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
		Ip:     input.Ip,
		Domain: input.Domain,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create whitelist item: %w", err)
	}

	pkg.Cacher().Del([]byte(cacheKeyWhiteListRules))

	return &IpWhiteListDto{
		ID:        item.ID,
		Ip:        item.Ip,
		Domain:    item.Domain,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *IpWhiteListService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	pkg.Cacher().Del([]byte(cacheKeyWhiteListRules))
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
			CreatedAt: item.CreatedAt,
		})
	}
	return dtos, total, nil
}

func (s *IpWhiteListService) FindByIP(ip string) (*IpWhiteListDto, error) {
	cache := pkg.Cacher()
	cacheKey := []byte(cacheKeyWhiteListIP + ip)

	if val, err := cache.Get(cacheKey); err == nil {
		if len(val) == 0 {
			return nil, nil
		}
		return &IpWhiteListDto{Ip: string(val)}, nil
	}

	rules, err := s.getRules()
	if err != nil {
		logging.Sugar.Errorf("Failed to get whitelist rules: %v", err)
		return nil, err
	}

	for _, rule := range rules {
		matched, err := filepath.Match(rule, ip)
		if err == nil && matched {
			cache.Set(cacheKey, []byte(rule), 600)
			return &IpWhiteListDto{Ip: rule}, nil
		}
		if rule == ip {
			cache.Set(cacheKey, []byte(rule), 600)
			return &IpWhiteListDto{Ip: rule}, nil
		}
	}

	cache.Set(cacheKey, []byte{}, 600)

	return nil, nil
}
