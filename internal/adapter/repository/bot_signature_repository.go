package repository

import (
	"open-website-defender/internal/domain/entity"

	"gorm.io/gorm"
)

type BotSignatureRepository struct {
	db *gorm.DB
}

func NewBotSignatureRepository(db *gorm.DB) *BotSignatureRepository {
	return &BotSignatureRepository{db: db}
}

func (r *BotSignatureRepository) Create(sig *entity.BotSignature) error {
	return r.db.Create(sig).Error
}

func (r *BotSignatureRepository) Update(sig *entity.BotSignature) error {
	return r.db.Save(sig).Error
}

func (r *BotSignatureRepository) Delete(id uint) error {
	return r.db.Delete(&entity.BotSignature{}, id).Error
}

func (r *BotSignatureRepository) FindByID(id uint) (*entity.BotSignature, error) {
	var sig entity.BotSignature
	err := r.db.First(&sig, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sig, err
}

func (r *BotSignatureRepository) List(limit, offset int) ([]*entity.BotSignature, int64, error) {
	var sigs []*entity.BotSignature
	var total int64

	if err := r.db.Model(&entity.BotSignature{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("id ASC").Limit(limit).Offset(offset).Find(&sigs).Error
	return sigs, total, err
}

func (r *BotSignatureRepository) FindAllEnabled() ([]*entity.BotSignature, error) {
	var sigs []*entity.BotSignature
	err := r.db.Where("enabled = ?", true).Find(&sigs).Error
	return sigs, err
}
