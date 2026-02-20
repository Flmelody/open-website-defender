package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"
	"time"

	"gorm.io/gorm"
)

type IpBlackListRepository struct {
	db *gorm.DB
}

var _ _interface.IpBlackListRepository = (*IpBlackListRepository)(nil)

func NewIpBlackListRepository(db *gorm.DB) *IpBlackListRepository {
	return &IpBlackListRepository{db: db}
}

func (r *IpBlackListRepository) Create(ip *entity.IpBlackList) error {
	return r.db.Create(ip).Error
}

func (r *IpBlackListRepository) Delete(id uint) error {
	return r.db.Delete(&entity.IpBlackList{}, id).Error
}

func (r *IpBlackListRepository) List(limit, offset int) ([]*entity.IpBlackList, int64, error) {
	var list []*entity.IpBlackList
	var total int64
	if err := r.db.Model(&entity.IpBlackList{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *IpBlackListRepository) FindByIP(ip string) (*entity.IpBlackList, error) {
	var item entity.IpBlackList
	// Exact match for management/duplicate check
	err := r.db.Where("ip = ?", ip).First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}

// DeleteExpired removes all blacklist entries whose ExpiresAt is in the past.
func (r *IpBlackListRepository) DeleteExpired() (int64, error) {
	result := r.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now().UTC()).Delete(&entity.IpBlackList{})
	return result.RowsAffected, result.Error
}
