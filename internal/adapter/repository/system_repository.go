package repository

import (
	_interface "castellum/internal/usecase/interface"

	"castellum/internal/domain/entity"

	"gorm.io/gorm"
)

var _ _interface.SystemRepository = (*SystemRepository)(nil)

type SystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

func (r *SystemRepository) Get() (*entity.System, error) {
	var system entity.System
	result := r.db.First(&system)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &system, nil
}

func (r *SystemRepository) Save(system *entity.System) error {
	return r.db.Save(system).Error
}
