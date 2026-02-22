package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"
	"time"

	"gorm.io/gorm"
)

type TrustedDeviceRepository struct {
	db *gorm.DB
}

var _ _interface.TrustedDeviceRepository = (*TrustedDeviceRepository)(nil)

func NewTrustedDeviceRepository(db *gorm.DB) *TrustedDeviceRepository {
	return &TrustedDeviceRepository{db: db}
}

func (r *TrustedDeviceRepository) Create(device *entity.TrustedDevice) error {
	return r.db.Create(device).Error
}

func (r *TrustedDeviceRepository) FindValidByToken(token string) (*entity.TrustedDevice, error) {
	var device entity.TrustedDevice
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now().UTC()).First(&device).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &device, err
}

func (r *TrustedDeviceRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&entity.TrustedDevice{}).Error
}

func (r *TrustedDeviceRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now().UTC()).Delete(&entity.TrustedDevice{}).Error
}
