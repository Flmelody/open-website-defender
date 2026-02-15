package user

import (
	"fmt"
	"sync"

	"open-website-defender/internal/adapter/repository"
	domainError "open-website-defender/internal/domain/error"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"
)

type AuthService struct {
	userRepo _interface.UserRepository
}

var (
	authService *AuthService
	authOnce    sync.Once
)

func GetAuthService() *AuthService {
	authOnce.Do(func() {
		authService = &AuthService{
			userRepo: repository.NewUserRepository(database.DB),
		}
	})
	return authService
}

func NewAuthService(userRepo _interface.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
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

	token, err := pkg.GenerateToken(user.Username, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutputDTO{
		Token: token,
		User: &UserInfoDTO{
			ID:       user.ID,
			Username: user.Username,
		},
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*UserInfoDTO, error) {
	claims, err := pkg.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(fmt.Sprintf("%d", claims.UserID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domainError.ErrUserNotFound
	}

	return &UserInfoDTO{
		ID:       user.ID,
		Username: user.Username,
		Scopes:   user.Scopes,
		IsAdmin:  user.IsAdmin,
	}, nil
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

	if user.GitToken == "" || !pkg.CheckPassword(user.GitToken, token) {
		return nil, domainError.ErrInvalidCredentials
	}

	return &UserInfoDTO{
		ID:       user.ID,
		Username: user.Username,
		Scopes:   user.Scopes,
		IsAdmin:  user.IsAdmin,
	}, nil
}
