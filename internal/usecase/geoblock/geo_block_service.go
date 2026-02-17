package geoblock

import (
	"encoding/json"
	"errors"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/pkg"
	"strings"
	"sync"
	"time"
)

type GeoBlockService struct {
	repo *repository.GeoBlockRepository
}

var (
	geoBlockService *GeoBlockService
	geoBlockOnce    sync.Once
)

func GetGeoBlockService() *GeoBlockService {
	geoBlockOnce.Do(func() {
		geoBlockService = &GeoBlockService{
			repo: repository.NewGeoBlockRepository(database.DB),
		}
	})
	return geoBlockService
}

// IsBlocked checks if the given IP's country is in the blocked list.
func (s *GeoBlockService) IsBlocked(ip string) (bool, string) {
	country := pkg.LookupCountry(ip)
	if country == "" {
		return false, ""
	}

	codes, err := s.getBlockedCodes()
	if err != nil || len(codes) == 0 {
		return false, country
	}

	for _, code := range codes {
		if strings.EqualFold(code, country) {
			return true, country
		}
	}
	return false, country
}

func (s *GeoBlockService) getBlockedCodes() ([]string, error) {
	c := pkg.Cacher()

	data, err := c.Get([]byte(cache.KeyGeoBlockCodes))
	if err == nil {
		var codes []string
		if err := json.Unmarshal(data, &codes); err == nil {
			return codes, nil
		}
	}

	codes, err := s.repo.FindAllCodes()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(codes)
	c.Set([]byte(cache.KeyGeoBlockCodes), jsonData, 3600)

	return codes, nil
}

type GeoBlockRuleDto struct {
	ID          uint      `json:"id"`
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *GeoBlockService) Create(countryCode, countryName string) (*GeoBlockRuleDto, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode == "" {
		return nil, errors.New("country code is required")
	}

	existing, err := s.repo.FindByCode(countryCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("country already blocked")
	}

	rule := &entity.GeoBlockRule{
		CountryCode: countryCode,
		CountryName: countryName,
	}

	if err := s.repo.Create(rule); err != nil {
		return nil, err
	}

	event.Bus().Publish(event.GeoBlockChanged)

	return &GeoBlockRuleDto{
		ID:          rule.ID,
		CountryCode: rule.CountryCode,
		CountryName: rule.CountryName,
		CreatedAt:   rule.CreatedAt,
	}, nil
}

func (s *GeoBlockService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	event.Bus().Publish(event.GeoBlockChanged)
	return nil
}

func (s *GeoBlockService) List(page, size int) ([]*GeoBlockRuleDto, int64, error) {
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

	dtos := make([]*GeoBlockRuleDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, &GeoBlockRuleDto{
			ID:          item.ID,
			CountryCode: item.CountryCode,
			CountryName: item.CountryName,
			CreatedAt:   item.CreatedAt,
		})
	}
	return dtos, total, nil
}
