package license

import (
	"encoding/json"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"
	"sync"
)

var (
	licenseService *LicenseService
	licenseOnce    sync.Once
)

type LicenseService struct {
	licenseRepo _interface.LicenseRepository
}

func GetLicenseService() *LicenseService {
	licenseOnce.Do(func() {
		licenseService = &LicenseService{
			licenseRepo: repository.NewLicenseRepository(database.DB),
		}
	})
	return licenseService
}

func NewLicenseService(licenseRepo _interface.LicenseRepository) *LicenseService {
	return &LicenseService{licenseRepo: licenseRepo}
}

func (s *LicenseService) Create(input *CreateLicenseDTO) (*LicenseCreatedDTO, error) {
	token := pkg.GenerateRandomToken(32) // 32 bytes = 64 hex chars
	tokenHash := pkg.SHA256Hash(token)

	lic := &entity.License{
		Name:      input.Name,
		Remark:    input.Remark,
		TokenHash: tokenHash,
		Active:    true,
	}

	if err := s.licenseRepo.Create(lic); err != nil {
		return nil, err
	}

	return &LicenseCreatedDTO{
		LicenseDTO: LicenseDTO{
			ID:        lic.ID,
			Name:      lic.Name,
			Remark:    lic.Remark,
			Active:    lic.Active,
			CreatedAt: lic.CreatedAt,
		},
		Token: token,
	}, nil
}

func (s *LicenseService) Delete(id uint) error {
	lic, _ := s.licenseRepo.FindByID(id)

	if err := s.licenseRepo.Delete(id); err != nil {
		return err
	}

	if lic != nil {
		event.Bus().Publish(event.LicenseChanged, lic.TokenHash)
	}
	return nil
}

func (s *LicenseService) List(page, size int) ([]*LicenseDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	licenses, total, err := s.licenseRepo.List(size, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*LicenseDTO, len(licenses))
	for i, lic := range licenses {
		dtos[i] = &LicenseDTO{
			ID:        lic.ID,
			Name:      lic.Name,
			Remark:    lic.Remark,
			Active:    lic.Active,
			CreatedAt: lic.CreatedAt,
		}
	}

	return dtos, total, nil
}

func (s *LicenseService) ValidateToken(token string) (bool, error) {
	tokenHash := pkg.SHA256Hash(token)
	cacheKey := cache.KeyLicenseToken + tokenHash

	store := cache.Store()

	// Check cache
	if cached, err := store.Get(cacheKey); err == nil {
		if len(cached) == 0 {
			return false, nil // cached "not found"
		}
		var lic entity.License
		if json.Unmarshal(cached, &lic) == nil {
			return lic.Active, nil
		}
	}

	// Query database
	lic, err := s.licenseRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return false, err
	}

	if lic == nil {
		// Cache negative result for 60 seconds
		_ = store.Set(cacheKey, []byte{}, 60)
		return false, nil
	}

	// Cache positive result for 5 minutes
	data, _ := json.Marshal(lic)
	_ = store.Set(cacheKey, data, 300)

	return lic.Active, nil
}
