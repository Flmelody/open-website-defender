package user

import (
	"errors"
	"fmt"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	_interface "open-website-defender/internal/usecase/interface"
	"strconv"
	"sync"
)

type UserService struct {
	userRepo _interface.UserRepository
}

var (
	userService *UserService
	userOnce    sync.Once
)

func GetUserService() *UserService {
	userOnce.Do(func() {
		userService = &UserService{
			userRepo: repository.NewUserRepository(database.DB),
		}
	})
	return userService
}

func NewUserService(userRepo _interface.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func maskGitToken(token string) string {
	if token != "" {
		return "***"
	}
	return ""
}

func (s *UserService) CreateUser(input *CreateUserDTO) (*UserDTO, error) {
	if input.Username == "" || input.Password == "" {
		return nil, errors.New("username and password are required")
	}

	user := &entity.User{
		Username: input.Username,
		Password: input.Password,
		GitToken: input.GitToken,
		IsAdmin:  input.IsAdmin,
		Scopes:   input.Scopes,
		Email:    input.Email,
	}

	if err := s.userRepo.Save(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		GitToken: maskGitToken(input.GitToken),
		IsAdmin:  user.IsAdmin,
		Scopes:   user.Scopes,
		Email:    user.Email,
	}, nil
}

func (s *UserService) GetUser(id uint) (*UserDTO, error) {
	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(id), 10))
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		GitToken: maskGitToken(user.GitToken),
		IsAdmin:  user.IsAdmin,
		Scopes:   user.Scopes,
		Email:    user.Email,
	}, nil
}

func (s *UserService) UpdateUser(id uint, input *UpdateUserDTO) (*UserDTO, error) {
	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(id), 10))
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Password != "" {
		user.Password = input.Password
	}
	if input.GitToken != "" {
		user.GitToken = input.GitToken
	}
	user.IsAdmin = input.IsAdmin
	user.Scopes = input.Scopes
	user.Email = input.Email

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		GitToken: maskGitToken(user.GitToken),
		IsAdmin:  user.IsAdmin,
		Scopes:   user.Scopes,
		Email:    user.Email,
	}, nil
}

func (s *UserService) DeleteUser(id uint) error {
	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(id), 10))
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.IsAdmin {
		return errors.New("cannot delete admin user")
	}

	if err := s.userRepo.Delete(strconv.FormatUint(uint64(id), 10)); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) ListUsers(page, size int) ([]*UserDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	offset := (page - 1) * size
	users, total, err := s.userRepo.List(size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	userDtos := make([]*UserDTO, 0, len(users))
	for _, user := range users {
		userDtos = append(userDtos, &UserDTO{
			ID:       user.ID,
			Username: user.Username,
			GitToken: maskGitToken(user.GitToken),
			IsAdmin:  user.IsAdmin,
			Scopes:   user.Scopes,
			Email:    user.Email,
		})
	}

	return userDtos, total, nil
}
