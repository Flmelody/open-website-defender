package repository

import (
	"castellum/internal/domain/entity"
	_interface "castellum/internal/usecase/interface"
	"strings"
	"time"

	"gorm.io/gorm"
)

// escapeLikeKeyword escapes LIKE wildcards so user-supplied search terms are
// matched literally. Pair with `ESCAPE '\'` in the SQL clause.
func escapeLikeKeyword(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

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

func (r *IpBlackListRepository) Update(ip *entity.IpBlackList) error {
	return r.db.Save(ip).Error
}

func (r *IpBlackListRepository) FindByID(id uint) (*entity.IpBlackList, error) {
	var item entity.IpBlackList
	err := r.db.First(&item, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}

func (r *IpBlackListRepository) Delete(id uint) error {
	return r.db.Delete(&entity.IpBlackList{}, id).Error
}

func (r *IpBlackListRepository) List(limit, offset int, keyword string) ([]*entity.IpBlackList, int64, error) {
	var list []*entity.IpBlackList
	var total int64
	query := r.db.Model(&entity.IpBlackList{})
	if keyword != "" {
		like := "%" + escapeLikeKeyword(keyword) + "%"
		query = query.Where(`LOWER(ip) LIKE LOWER(?) ESCAPE '\' OR LOWER(remark) LIKE LOWER(?) ESCAPE '\'`, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
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
