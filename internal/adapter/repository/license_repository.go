package repository

import (
	_interface "open-website-defender/internal/usecase/interface"

	"open-website-defender/internal/domain/entity"

	"gorm.io/gorm"
)

var _ _interface.LicenseRepository = (*LicenseRepository)(nil)

type LicenseRepository struct {
	db *gorm.DB
}

func NewLicenseRepository(db *gorm.DB) *LicenseRepository {
	return &LicenseRepository{db: db}
}

func (r *LicenseRepository) Create(license *entity.License) error {
	return r.db.Create(license).Error
}

func (r *LicenseRepository) Delete(id uint) error {
	return r.db.Delete(&entity.License{}, id).Error
}

func (r *LicenseRepository) FindByID(id uint) (*entity.License, error) {
	var license entity.License
	result := r.db.First(&license, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &license, nil
}

func (r *LicenseRepository) List(limit, offset int) ([]*entity.License, int64, error) {
	var licenses []*entity.License
	var total int64

	if err := r.db.Model(&entity.License{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Order("id DESC").Limit(limit).Offset(offset).Find(&licenses).Error; err != nil {
		return nil, 0, err
	}

	return licenses, total, nil
}

func (r *LicenseRepository) FindByTokenHash(tokenHash string) (*entity.License, error) {
	var license entity.License
	result := r.db.Where("token_hash = ?", tokenHash).First(&license)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &license, nil
}
