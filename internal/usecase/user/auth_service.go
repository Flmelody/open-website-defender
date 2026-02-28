package user

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"strconv"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	domainError "open-website-defender/internal/domain/error"
	"open-website-defender/internal/infrastructure/cache"
	"open-website-defender/internal/infrastructure/config"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"
)

type AuthService struct {
	userRepo           _interface.UserRepository
	trustedDeviceRepo  _interface.TrustedDeviceRepository
}

var (
	authService *AuthService
	authOnce    sync.Once
)

func GetAuthService() *AuthService {
	authOnce.Do(func() {
		authService = &AuthService{
			userRepo:          repository.NewUserRepository(database.DB),
			trustedDeviceRepo: repository.NewTrustedDeviceRepository(database.DB),
		}
	})
	return authService
}

func NewAuthService(userRepo _interface.UserRepository, trustedDeviceRepo _interface.TrustedDeviceRepository) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		trustedDeviceRepo: trustedDeviceRepo,
	}
}

func (s *AuthService) Login(input *LoginInputDTO) (*LoginOutputDTO, error) {
	if input.Username == "" || input.Password == "" {
		return nil, domainError.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByUsernameAndPassword(input.Username, input.Password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domainError.ErrInvalidCredentials
	}

	if !user.Enabled {
		return nil, domainError.ErrUserDisabled
	}

	token, err := pkg.GenerateToken(user.Username, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutputDTO{
		Token: token,
		User: &UserInfoDTO{
			ID:       user.ID,
			Username: user.Username,
			IsAdmin:  user.IsAdmin,
			Enabled:  user.Enabled,
			Email:    user.Email,
		},
	}, nil
}

func (s *AuthService) GuardLogin(input *LoginInputDTO) (*GuardLoginOutputDTO, error) {
	if input.Username == "" || input.Password == "" {
		return nil, domainError.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByUsernameAndPassword(input.Username, input.Password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domainError.ErrInvalidCredentials
	}

	if !user.Enabled {
		return nil, domainError.ErrUserDisabled
	}

	userInfo := &UserInfoDTO{
		ID:          user.ID,
		Username:    user.Username,
		IsAdmin:     user.IsAdmin,
		Enabled:     user.Enabled,
		Email:       user.Email,
		TotpEnabled: user.TotpEnabled,
	}

	if user.TotpEnabled {
		// Check trusted device cookie
		if input.TrustedDeviceToken != "" {
			if s.CheckTrustedDevice(user.ID, input.TrustedDeviceToken) {
				token, err := pkg.GenerateToken(user.Username, user.ID)
				if err != nil {
					return nil, err
				}
				return &GuardLoginOutputDTO{
					RequiresTwoFA: false,
					Token:         token,
					User:          userInfo,
				}, nil
			}
		}

		challengeToken, err := pkg.Generate2FAToken(user.Username, user.ID)
		if err != nil {
			return nil, err
		}
		return &GuardLoginOutputDTO{
			RequiresTwoFA:  true,
			ChallengeToken: challengeToken,
			User:           userInfo,
		}, nil
	}

	token, err := pkg.GenerateToken(user.Username, user.ID)
	if err != nil {
		return nil, err
	}

	return &GuardLoginOutputDTO{
		RequiresTwoFA: false,
		Token:         token,
		User:          userInfo,
	}, nil
}

func (s *AuthService) AdminLogin(input *LoginInputDTO) (*AdminLoginOutputDTO, error) {
	if input.Username == "" || input.Password == "" {
		return nil, domainError.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByUsernameAndPassword(input.Username, input.Password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domainError.ErrInvalidCredentials
	}

	if !user.Enabled {
		return nil, domainError.ErrUserDisabled
	}

	if !user.IsAdmin {
		return nil, domainError.ErrAdminRequired
	}

	userInfo := &UserInfoDTO{
		ID:          user.ID,
		Username:    user.Username,
		IsAdmin:     user.IsAdmin,
		Enabled:     user.Enabled,
		Email:       user.Email,
		TotpEnabled: user.TotpEnabled,
	}

	if user.TotpEnabled {
		// Check trusted device cookie
		if input.TrustedDeviceToken != "" {
			if s.CheckTrustedDevice(user.ID, input.TrustedDeviceToken) {
				token, err := pkg.GenerateToken(user.Username, user.ID)
				if err != nil {
					return nil, err
				}
				return &AdminLoginOutputDTO{
					RequiresTwoFA: false,
					Token:         token,
					User:          userInfo,
				}, nil
			}
		}

		challengeToken, err := pkg.Generate2FAToken(user.Username, user.ID)
		if err != nil {
			return nil, err
		}
		return &AdminLoginOutputDTO{
			RequiresTwoFA:  true,
			ChallengeToken: challengeToken,
			User:           userInfo,
		}, nil
	}

	token, err := pkg.GenerateToken(user.Username, user.ID)
	if err != nil {
		return nil, err
	}

	return &AdminLoginOutputDTO{
		RequiresTwoFA: false,
		Token:         token,
		User:          userInfo,
	}, nil
}

func (s *AuthService) Verify2FALogin(input *TwoFALoginInputDTO) (*LoginOutputDTO, error) {
	claims, err := pkg.Parse2FAToken(input.ChallengeToken)
	if err != nil {
		return nil, domainError.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByID(fmt.Sprintf("%d", claims.UserID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domainError.ErrUserNotFound
	}

	if !user.Enabled {
		return nil, domainError.ErrUserDisabled
	}

	if !user.TotpEnabled || user.TotpSecret == "" {
		return nil, domainError.ErrTotpNotEnabled
	}

	if !totp.Validate(input.Code, user.TotpSecret) {
		return nil, domainError.ErrTotpInvalidCode
	}

	token, err := pkg.GenerateToken(user.Username, user.ID)
	if err != nil {
		return nil, err
	}

	output := &LoginOutputDTO{
		Token: token,
		User: &UserInfoDTO{
			ID:          user.ID,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			Enabled:     user.Enabled,
			Email:       user.Email,
			TotpEnabled: user.TotpEnabled,
		},
	}

	// Generate trusted device token if requested
	if input.TrustDevice {
		days := config.Get().Security.TrustedDeviceDays
		if days > 0 {
			deviceToken, err := s.createTrustedDevice(user.ID, days)
			if err == nil {
				output.TrustedDeviceToken = deviceToken
			}
		}
	}

	return output, nil
}

func (s *AuthService) CheckTrustedDevice(userID uint, token string) bool {
	if token == "" {
		return false
	}
	device, err := s.trustedDeviceRepo.FindValidByToken(token)
	if err != nil || device == nil {
		return false
	}
	return device.UserID == userID
}

func (s *AuthService) createTrustedDevice(userID uint, days int) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	device := &entity.TrustedDevice{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour),
	}
	if err := s.trustedDeviceRepo.Create(device); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) SetupTotp(userID uint) (*TotpSetupOutputDTO, error) {
	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(userID), 10))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domainError.ErrUserNotFound
	}
	if user.TotpEnabled {
		return nil, domainError.ErrTotpAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OpenWebsiteDefender",
		AccountName: user.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	user.TotpSecret = key.Secret()
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	img, err := key.Image(256, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	return &TotpSetupOutputDTO{
		Secret:        key.Secret(),
		QRCodeDataURI: dataURI,
	}, nil
}

func (s *AuthService) ConfirmTotp(userID uint, code string) error {
	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(userID), 10))
	if err != nil {
		return err
	}
	if user == nil {
		return domainError.ErrUserNotFound
	}
	if user.TotpSecret == "" {
		return domainError.ErrTotpNotEnabled
	}

	if !totp.Validate(code, user.TotpSecret) {
		return domainError.ErrTotpInvalidCode
	}

	user.TotpEnabled = true
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}
	return nil
}

func (s *AuthService) DisableTotp(userID uint) error {
	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(userID), 10))
	if err != nil {
		return err
	}
	if user == nil {
		return domainError.ErrUserNotFound
	}
	if !user.TotpEnabled {
		return domainError.ErrTotpNotEnabled
	}

	user.TotpSecret = ""
	user.TotpEnabled = false
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to disable TOTP: %w", err)
	}

	// Invalidate all trusted devices for this user
	_ = s.trustedDeviceRepo.DeleteByUserID(userID)

	return nil
}

func (s *AuthService) ValidateToken(tokenString string) (*UserInfoDTO, error) {
	claims, err := pkg.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check user info cache
	store := cache.Store()
	cacheKey := fmt.Sprintf("%s%d", cache.KeyUserInfo, claims.UserID)
	if data, err := store.Get(cacheKey); err == nil {
		var userInfo UserInfoDTO
		if json.Unmarshal(data, &userInfo) == nil {
			if !userInfo.Enabled {
				return nil, domainError.ErrUserDisabled
			}
			return &userInfo, nil
		}
	}

	user, err := s.userRepo.FindByID(fmt.Sprintf("%d", claims.UserID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domainError.ErrUserNotFound
	}

	if !user.Enabled {
		return nil, domainError.ErrUserDisabled
	}

	userInfo := &UserInfoDTO{
		ID:       user.ID,
		Username: user.Username,
		Scopes:   user.Scopes,
		IsAdmin:  user.IsAdmin,
		Enabled:  user.Enabled,
		Email:    user.Email,
	}

	// Cache user info (10 minutes)
	if data, err := json.Marshal(userInfo); err == nil {
		_ = store.Set(cacheKey, data, 600)
	}

	return userInfo, nil
}

func (s *AuthService) RecoverAdmin2FA(username, password, recoveryKey string) error {
	configuredKey := config.Get().Security.AdminRecoveryKey
	if configuredKey == "" {
		return domainError.ErrRecoveryDisabled
	}

	if recoveryKey != configuredKey {
		return domainError.ErrRecoveryKeyInvalid
	}

	user, err := s.userRepo.FindByUsernameAndPassword(username, password)
	if err != nil {
		return err
	}
	if user == nil {
		return domainError.ErrInvalidCredentials
	}

	if !user.IsAdmin {
		return domainError.ErrAdminRequired
	}

	if !user.TotpEnabled {
		return domainError.ErrTotpNotEnabled
	}

	user.TotpSecret = ""
	user.TotpEnabled = false
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to reset 2FA: %w", err)
	}

	// Invalidate all trusted devices for this user
	_ = s.trustedDeviceRepo.DeleteByUserID(user.ID)

	return nil
}

func (s *AuthService) ValidateGitToken(username, token string) (*UserInfoDTO, error) {
	if username == "" || token == "" {
		return nil, domainError.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domainError.ErrInvalidCredentials
	}

	if !user.Enabled {
		return nil, domainError.ErrUserDisabled
	}

	if user.GitToken == "" || !pkg.CheckPassword(user.GitToken, token) {
		return nil, domainError.ErrInvalidCredentials
	}

	return &UserInfoDTO{
		ID:       user.ID,
		Username: user.Username,
		Scopes:   user.Scopes,
		IsAdmin:  user.IsAdmin,
		Enabled:  user.Enabled,
		Email:    user.Email,
	}, nil
}
