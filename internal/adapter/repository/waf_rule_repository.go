package repository

import (
	"open-website-defender/internal/domain/entity"

	"gorm.io/gorm"
)

type WafRuleRepository struct {
	db *gorm.DB
}

func NewWafRuleRepository(db *gorm.DB) *WafRuleRepository {
	return &WafRuleRepository{db: db}
}

func (r *WafRuleRepository) Create(rule *entity.WafRule) error {
	return r.db.Create(rule).Error
}

func (r *WafRuleRepository) Update(rule *entity.WafRule) error {
	return r.db.Save(rule).Error
}

func (r *WafRuleRepository) Delete(id uint) error {
	return r.db.Delete(&entity.WafRule{}, id).Error
}

func (r *WafRuleRepository) FindByID(id uint) (*entity.WafRule, error) {
	var rule entity.WafRule
	err := r.db.First(&rule, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &rule, err
}

func (r *WafRuleRepository) List(limit, offset int) ([]*entity.WafRule, int64, error) {
	var rules []*entity.WafRule
	var total int64

	if err := r.db.Model(&entity.WafRule{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("id ASC").Limit(limit).Offset(offset).Find(&rules).Error
	return rules, total, err
}

func (r *WafRuleRepository) FindAllEnabled() ([]*entity.WafRule, error) {
	var rules []*entity.WafRule
	err := r.db.Where("enabled = ?", true).Order("priority ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (r *WafRuleRepository) FindByName(name string) (*entity.WafRule, error) {
	var rule entity.WafRule
	err := r.db.Where("name = ?", name).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &rule, err
}

func (r *WafRuleRepository) UpdateGroupEnabled(groupName string, enabled bool) error {
	return r.db.Model(&entity.WafRule{}).Where("group_name = ?", groupName).Update("enabled", enabled).Error
}
