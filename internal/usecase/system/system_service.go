package system

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

const (
	defaultGitTokenHeader = "Defender-Git-Token"
	defaultLicenseHeader  = "Defender-License"
)

var (
	systemService *SystemService
	systemOnce    sync.Once
)

type SystemService struct {
	systemRepo _interface.SystemRepository
}

func GetSystemService() *SystemService {
	systemOnce.Do(func() {
		systemService = &SystemService{
			systemRepo: repository.NewSystemRepository(database.DB),
		}
	})
	return systemService
}

func NewSystemService(systemRepo _interface.SystemRepository) *SystemService {
	return &SystemService{systemRepo: systemRepo}
}

func (s *SystemService) GetSettings() (*SystemSettingsDTO, error) {
	// Check cache
	if cached, err := pkg.Cacher().Get([]byte(cache.KeySystemSettings)); err == nil {
		var dto SystemSettingsDTO
		if json.Unmarshal(cached, &dto) == nil {
			return &dto, nil
		}
	}

	sys, err := s.systemRepo.Get()
	if err != nil {
		return nil, err
	}

	dto := &SystemSettingsDTO{
		GitTokenHeader: defaultGitTokenHeader,
		LicenseHeader:  defaultLicenseHeader,
	}

	if sys != nil {
		if sys.Security.GitTokenHeader != "" {
			dto.GitTokenHeader = sys.Security.GitTokenHeader
		}
		if sys.Security.LicenseHeader != "" {
			dto.LicenseHeader = sys.Security.LicenseHeader
		}
	}

	// Cache for 10 minutes
	data, _ := json.Marshal(dto)
	_ = pkg.Cacher().Set([]byte(cache.KeySystemSettings), data, 600)

	return dto, nil
}

func (s *SystemService) UpdateSettings(input *UpdateSystemSettingsDTO) error {
	sys, err := s.systemRepo.Get()
	if err != nil {
		return err
	}

	if sys == nil {
		sys = &entity.System{}
	}

	sys.Security.GitTokenHeader = input.GitTokenHeader
	sys.Security.LicenseHeader = input.LicenseHeader

	if err := s.systemRepo.Save(sys); err != nil {
		return err
	}

	event.Bus().Publish(event.SystemSettingsChanged)

	return nil
}

// GetHeaderNames returns the configured header names for git token and license.
// Falls back to defaults if not configured.
func (s *SystemService) GetHeaderNames() (gitTokenHeader, licenseHeader string) {
	settings, err := s.GetSettings()
	if err != nil || settings == nil {
		return defaultGitTokenHeader, defaultLicenseHeader
	}
	return settings.GitTokenHeader, settings.LicenseHeader
}
