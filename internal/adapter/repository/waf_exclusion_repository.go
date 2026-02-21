package repository

import (
	"open-website-defender/internal/domain/entity"

	"gorm.io/gorm"
)

type WafExclusionRepository struct {
	db *gorm.DB
}

func NewWafExclusionRepository(db *gorm.DB) *WafExclusionRepository {
	return &WafExclusionRepository{db: db}
}

func (r *WafExclusionRepository) Create(exclusion *entity.WafExclusion) error {
	return r.db.Create(exclusion).Error
}

func (r *WafExclusionRepository) Delete(id uint) error {
	return r.db.Delete(&entity.WafExclusion{}, id).Error
}

func (r *WafExclusionRepository) FindByID(id uint) (*entity.WafExclusion, error) {
	var exclusion entity.WafExclusion
	err := r.db.First(&exclusion, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &exclusion, err
}

func (r *WafExclusionRepository) List(limit, offset int) ([]*entity.WafExclusion, int64, error) {
	var exclusions []*entity.WafExclusion
	var total int64

	if err := r.db.Model(&entity.WafExclusion{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("id ASC").Limit(limit).Offset(offset).Find(&exclusions).Error
	return exclusions, total, err
}

func (r *WafExclusionRepository) FindAllEnabled() ([]*entity.WafExclusion, error) {
	var exclusions []*entity.WafExclusion
	err := r.db.Where("enabled = ?", true).Find(&exclusions).Error
	return exclusions, err
}
