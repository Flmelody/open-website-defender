package system

import (
	"encoding/json"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	_interface "open-website-defender/internal/usecase/interface"
	"sync"

	"github.com/spf13/viper"
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
	store := cache.Store()

	// Check cache
	if cached, err := store.Get(cache.KeySystemSettings); err == nil {
		var dto SystemSettingsDTO
		if json.Unmarshal(cached, &dto) == nil {
			return &dto, nil
		}
	}

	sys, err := s.systemRepo.Get()
	if err != nil {
		return nil, err
	}

	mode := viper.GetString("mode")
	if mode == "" {
		mode = "auth_request"
	}

	dto := &SystemSettingsDTO{
		Mode:                  mode,
		GitTokenHeader:        defaultGitTokenHeader,
		LicenseHeader:         defaultLicenseHeader,
		JSChallengeEnabled:    viper.GetBool("js-challenge.enabled"),
		JSChallengeMode:       viper.GetString("js-challenge.mode"),
		JSChallengeDifficulty: viper.GetInt("js-challenge.difficulty"),
		WebhookURL:            viper.GetString("webhook.url"),

		// Bot Management defaults from config
		BotManagementEnabled: viper.GetBool("bot-management.enabled"),
		ChallengeEscalation:  viper.GetBool("bot-management.challenge-escalation"),
		CaptchaProvider:      viper.GetString("bot-management.captcha.provider"),
		CaptchaSiteKey:       viper.GetString("bot-management.captcha.site-key"),
		CaptchaSecretKey:     viper.GetString("bot-management.captcha.secret-key"),
		CaptchaCookieTTL:     viper.GetInt("bot-management.captcha.cookie-ttl"),

		// Cache defaults from config
		CacheSyncInterval: viper.GetInt("cache.sync-interval"),
	}

	// Apply sensible defaults for zero values
	if dto.JSChallengeMode == "" {
		dto.JSChallengeMode = "suspicious"
	}
	if dto.JSChallengeDifficulty <= 0 {
		dto.JSChallengeDifficulty = 4
	}
	if dto.CaptchaProvider == "" {
		dto.CaptchaProvider = "hcaptcha"
	}
	if dto.CaptchaCookieTTL <= 0 {
		dto.CaptchaCookieTTL = 86400
	}

	// DB settings override config file
	if sys != nil {
		if sys.Security.GitTokenHeader != "" {
			dto.GitTokenHeader = sys.Security.GitTokenHeader
		}
		if sys.Security.LicenseHeader != "" {
			dto.LicenseHeader = sys.Security.LicenseHeader
		}
		if sys.Security.JSChallengeEnabled != nil {
			dto.JSChallengeEnabled = *sys.Security.JSChallengeEnabled
		}
		if sys.Security.JSChallengeMode != "" {
			dto.JSChallengeMode = sys.Security.JSChallengeMode
		}
		if sys.Security.JSChallengeDifficulty > 0 {
			dto.JSChallengeDifficulty = sys.Security.JSChallengeDifficulty
		}
		if sys.Security.WebhookURL != "" {
			dto.WebhookURL = sys.Security.WebhookURL
		}

		// Bot Management overrides
		if sys.BotManagement.Enabled != nil {
			dto.BotManagementEnabled = *sys.BotManagement.Enabled
		}
		if sys.BotManagement.ChallengeEscalation != nil {
			dto.ChallengeEscalation = *sys.BotManagement.ChallengeEscalation
		}
		if sys.BotManagement.CaptchaProvider != "" {
			dto.CaptchaProvider = sys.BotManagement.CaptchaProvider
		}
		if sys.BotManagement.CaptchaSiteKey != "" {
			dto.CaptchaSiteKey = sys.BotManagement.CaptchaSiteKey
		}
		if sys.BotManagement.CaptchaSecretKey != "" {
			dto.CaptchaSecretKey = sys.BotManagement.CaptchaSecretKey
		}
		if sys.BotManagement.CaptchaCookieTTL > 0 {
			dto.CaptchaCookieTTL = sys.BotManagement.CaptchaCookieTTL
		}

		// Cache overrides
		if sys.CacheSettings.SyncInterval != nil {
			dto.CacheSyncInterval = *sys.CacheSettings.SyncInterval
		}
	}

	// Cache for 10 minutes
	data, _ := json.Marshal(dto)
	_ = store.Set(cache.KeySystemSettings, data, 600)

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
	sys.Security.JSChallengeEnabled = &input.JSChallengeEnabled
	sys.Security.JSChallengeMode = input.JSChallengeMode
	sys.Security.JSChallengeDifficulty = input.JSChallengeDifficulty
	sys.Security.WebhookURL = input.WebhookURL

	// Bot Management
	sys.BotManagement = entity.BotManagementSettings{
		Enabled:             &input.BotManagementEnabled,
		ChallengeEscalation: &input.ChallengeEscalation,
		CaptchaProvider:     input.CaptchaProvider,
		CaptchaSiteKey:      input.CaptchaSiteKey,
		CaptchaSecretKey:    input.CaptchaSecretKey,
		CaptchaCookieTTL:    input.CaptchaCookieTTL,
	}

	// Cache
	sys.CacheSettings = entity.CacheSettings{
		SyncInterval: &input.CacheSyncInterval,
	}

	if err := s.systemRepo.Save(sys); err != nil {
		return err
	}

	event.Bus().Publish(event.SystemSettingsChanged)

	// Restart cache sync if interval changed
	cache.RestartSync(input.CacheSyncInterval)

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
