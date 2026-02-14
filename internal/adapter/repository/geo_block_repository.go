package repository

import (
	"open-website-defender/internal/domain/entity"

	"gorm.io/gorm"
)

type GeoBlockRepository struct {
	db *gorm.DB
}

func NewGeoBlockRepository(db *gorm.DB) *GeoBlockRepository {
	return &GeoBlockRepository{db: db}
}

func (r *GeoBlockRepository) Create(rule *entity.GeoBlockRule) error {
	return r.db.Create(rule).Error
}

func (r *GeoBlockRepository) Delete(id uint) error {
	return r.db.Delete(&entity.GeoBlockRule{}, id).Error
}

func (r *GeoBlockRepository) List(limit, offset int) ([]*entity.GeoBlockRule, int64, error) {
	var rules []*entity.GeoBlockRule
	var total int64

	if err := r.db.Model(&entity.GeoBlockRule{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("id ASC").Limit(limit).Offset(offset).Find(&rules).Error
	return rules, total, err
}

func (r *GeoBlockRepository) FindAllCodes() ([]string, error) {
	var codes []string
	err := r.db.Model(&entity.GeoBlockRule{}).Pluck("country_code", &codes).Error
	return codes, err
}

func (r *GeoBlockRepository) FindByCode(code string) (*entity.GeoBlockRule, error) {
	var rule entity.GeoBlockRule
	err := r.db.Where("country_code = ?", code).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &rule, err
}
