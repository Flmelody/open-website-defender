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
		ipBlackListService = &IpBlackListService{
			repo: repository.NewIpBlackListRepository(database.DB),
		}
	})
	return ipBlackListService
}

func (s *IpBlackListService) getRules() ([]string, error) {
	c := pkg.Cacher()

	data, err := c.Get([]byte(cache.KeyBlackListRules))
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
		c.Set([]byte(cache.KeyBlackListRules), data, 3600)
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

	event.Bus().Publish(event.BlackListChanged)

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

	for _, rule := range rules {
		if pkg.MatchIP(rule, ip) {
			return &IpBlackListDto{Ip: rule}, nil
		}
	}

	return nil, nil
}
