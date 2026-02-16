package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"
	"time"

	"gorm.io/gorm"
)

type OAuthRefreshTokenRepository struct {
	db *gorm.DB
}

var _ _interface.OAuthRefreshTokenRepository = (*OAuthRefreshTokenRepository)(nil)

func NewOAuthRefreshTokenRepository(db *gorm.DB) *OAuthRefreshTokenRepository {
	return &OAuthRefreshTokenRepository{db: db}
}

func (r *OAuthRefreshTokenRepository) Create(token *entity.OAuthRefreshToken) error {
	return r.db.Create(token).Error
}

func (r *OAuthRefreshTokenRepository) FindByToken(token string) (*entity.OAuthRefreshToken, error) {
	var item entity.OAuthRefreshToken
	err := r.db.Where("token = ?", token).First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}

func (r *OAuthRefreshTokenRepository) Revoke(id uint) error {
	return r.db.Model(&entity.OAuthRefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

func (r *OAuthRefreshTokenRepository) RevokeByClientAndUser(clientID string, userID uint) error {
	return r.db.Model(&entity.OAuthRefreshToken{}).
		Where("client_id = ? AND user_id = ? AND revoked = ?", clientID, userID, false).
		Update("revoked", true).Error
}

func (r *OAuthRefreshTokenRepository) FindActiveByUserID(userID uint) ([]*entity.OAuthRefreshToken, error) {
	var tokens []*entity.OAuthRefreshToken
	err := r.db.Where("user_id = ? AND revoked = ? AND expires_at > ?", userID, false, time.Now().UTC()).
		Order("created_at DESC").Find(&tokens).Error
	return tokens, err
}

func (r *OAuthRefreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? OR revoked = ?", time.Now().UTC(), true).
		Delete(&entity.OAuthRefreshToken{}).Error
}
