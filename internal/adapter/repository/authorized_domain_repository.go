package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"

	"gorm.io/gorm"
)

type AuthorizedDomainRepository struct {
	db *gorm.DB
}

var _ _interface.AuthorizedDomainRepository = (*AuthorizedDomainRepository)(nil)

func NewAuthorizedDomainRepository(db *gorm.DB) *AuthorizedDomainRepository {
	return &AuthorizedDomainRepository{db: db}
}

func (r *AuthorizedDomainRepository) Create(domain *entity.AuthorizedDomain) error {
	return r.db.Create(domain).Error
}

func (r *AuthorizedDomainRepository) Delete(id uint) error {
	return r.db.Delete(&entity.AuthorizedDomain{}, id).Error
}

func (r *AuthorizedDomainRepository) FindByID(id uint) (*entity.AuthorizedDomain, error) {
	var item entity.AuthorizedDomain
	err := r.db.First(&item, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}

func (r *AuthorizedDomainRepository) List(limit, offset int) ([]*entity.AuthorizedDomain, int64, error) {
	var list []*entity.AuthorizedDomain
	var total int64
	if err := r.db.Model(&entity.AuthorizedDomain{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *AuthorizedDomainRepository) ListAll() ([]*entity.AuthorizedDomain, error) {
	var list []*entity.AuthorizedDomain
	err := r.db.Order("name ASC").Find(&list).Error
	return list, err
}

func (r *AuthorizedDomainRepository) FindByName(name string) (*entity.AuthorizedDomain, error) {
	var item entity.AuthorizedDomain
	err := r.db.Where("name = ?", name).First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}
