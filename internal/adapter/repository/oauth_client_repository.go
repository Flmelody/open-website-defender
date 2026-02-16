package repository

import (
	"open-website-defender/internal/domain/entity"
	_interface "open-website-defender/internal/usecase/interface"

	"gorm.io/gorm"
)

type OAuthClientRepository struct {
	db *gorm.DB
}

var _ _interface.OAuthClientRepository = (*OAuthClientRepository)(nil)

func NewOAuthClientRepository(db *gorm.DB) *OAuthClientRepository {
	return &OAuthClientRepository{db: db}
}

func (r *OAuthClientRepository) Create(client *entity.OAuthClient) error {
	return r.db.Create(client).Error
}

func (r *OAuthClientRepository) Update(client *entity.OAuthClient) error {
	return r.db.Save(client).Error
}

func (r *OAuthClientRepository) Delete(id uint) error {
	return r.db.Delete(&entity.OAuthClient{}, id).Error
}

func (r *OAuthClientRepository) FindByID(id uint) (*entity.OAuthClient, error) {
	var client entity.OAuthClient
	err := r.db.First(&client, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &client, err
}

func (r *OAuthClientRepository) FindByClientID(clientID string) (*entity.OAuthClient, error) {
	var client entity.OAuthClient
	err := r.db.Where("client_id = ?", clientID).First(&client).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &client, err
}

func (r *OAuthClientRepository) List(limit, offset int) ([]*entity.OAuthClient, int64, error) {
	var list []*entity.OAuthClient
	var total int64
	if err := r.db.Model(&entity.OAuthClient{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}
