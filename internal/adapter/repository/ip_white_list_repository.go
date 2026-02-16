package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"

	"gorm.io/gorm"
)

type IpWhiteListRepository struct {
	db *gorm.DB
}

var _ _interface.IpWhiteListRepository = (*IpWhiteListRepository)(nil)

func NewIpWhiteListRepository(db *gorm.DB) *IpWhiteListRepository {
	return &IpWhiteListRepository{db: db}
}

func (r *IpWhiteListRepository) Create(ip *entity.IpWhiteList) error {
	return r.db.Create(ip).Error
}

func (r *IpWhiteListRepository) Delete(id uint) error {
	return r.db.Delete(&entity.IpWhiteList{}, id).Error
}

func (r *IpWhiteListRepository) List(limit, offset int) ([]*entity.IpWhiteList, int64, error) {
	var list []*entity.IpWhiteList
	var total int64
	if err := r.db.Model(&entity.IpWhiteList{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *IpWhiteListRepository) DeleteByDomain(domain string) error {
	return r.db.Where("domain = ?", domain).Delete(&entity.IpWhiteList{}).Error
}

func (r *IpWhiteListRepository) FindByIP(ip string) (*entity.IpWhiteList, error) {
	var item entity.IpWhiteList
	// Exact match for management/duplicate check
	err := r.db.Where("ip = ?", ip).First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}
