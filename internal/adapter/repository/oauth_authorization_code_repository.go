package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"
	"time"

	"gorm.io/gorm"
)

type OAuthAuthorizationCodeRepository struct {
	db *gorm.DB
}

var _ _interface.OAuthAuthorizationCodeRepository = (*OAuthAuthorizationCodeRepository)(nil)

func NewOAuthAuthorizationCodeRepository(db *gorm.DB) *OAuthAuthorizationCodeRepository {
	return &OAuthAuthorizationCodeRepository{db: db}
}

func (r *OAuthAuthorizationCodeRepository) Create(code *entity.OAuthAuthorizationCode) error {
	return r.db.Create(code).Error
}

func (r *OAuthAuthorizationCodeRepository) FindByCode(code string) (*entity.OAuthAuthorizationCode, error) {
	var item entity.OAuthAuthorizationCode
	err := r.db.Where("code = ?", code).First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}

func (r *OAuthAuthorizationCodeRepository) MarkUsed(id uint) error {
	return r.db.Model(&entity.OAuthAuthorizationCode{}).Where("id = ?", id).Update("used", true).Error
}

func (r *OAuthAuthorizationCodeRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? OR used = ?", time.Now().UTC(), true).
		Delete(&entity.OAuthAuthorizationCode{}).Error
}
