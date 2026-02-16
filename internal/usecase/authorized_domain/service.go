package authorized_domain

import (
	"errors"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	_interface "open-website-defender/internal/usecase/interface"
	"sync"
)

var (
	authorizedDomainService *AuthorizedDomainService
	authorizedDomainOnce    sync.Once
)

type AuthorizedDomainService struct {
	repo            _interface.AuthorizedDomainRepository
	ipWhiteListRepo _interface.IpWhiteListRepository
	userRepo        _interface.UserRepository
}

func GetAuthorizedDomainService() *AuthorizedDomainService {
	authorizedDomainOnce.Do(func() {
		authorizedDomainService = &AuthorizedDomainService{
			repo:            repository.NewAuthorizedDomainRepository(database.DB),
			ipWhiteListRepo: repository.NewIpWhiteListRepository(database.DB),
			userRepo:        repository.NewUserRepository(database.DB),
		}
	})
	return authorizedDomainService
}

func (s *AuthorizedDomainService) Create(input *CreateAuthorizedDomainDTO) (*AuthorizedDomainDTO, error) {
	existing, err := s.repo.FindByName(input.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("domain already exists")
	}

	item := &entity.AuthorizedDomain{
		Name: input.Name,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return &AuthorizedDomainDTO{
		ID:        item.ID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *AuthorizedDomainService) Delete(id uint) error {
	domain, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if domain == nil {
		return errors.New("domain not found")
	}

	// Cascade: remove IP whitelist entries with this domain
	if err := s.ipWhiteListRepo.DeleteByDomain(domain.Name); err != nil {
		logging.Sugar.Errorf("Failed to cascade delete IP whitelist for domain %s: %v", domain.Name, err)
	}

	// Cascade: remove this domain from user scopes
	if err := s.userRepo.RemoveScopeFromAll(domain.Name); err != nil {
		logging.Sugar.Errorf("Failed to cascade remove scope %s from users: %v", domain.Name, err)
	}

	return s.repo.Delete(id)
}

func (s *AuthorizedDomainService) List(page, size int) ([]*AuthorizedDomainDTO, int64, error) {
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

	dtos := make([]*AuthorizedDomainDTO, len(list))
	for i, item := range list {
		dtos[i] = &AuthorizedDomainDTO{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		}
	}

	return dtos, total, nil
}

func (s *AuthorizedDomainService) ListAll() ([]*AuthorizedDomainDTO, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}

	dtos := make([]*AuthorizedDomainDTO, len(list))
	for i, item := range list {
		dtos[i] = &AuthorizedDomainDTO{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		}
	}

	return dtos, nil
}
