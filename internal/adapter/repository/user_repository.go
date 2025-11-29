package repository

import (
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

var _ _interface.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(user *entity.User) error {
	if user.Password != "" {
		user.Password = pkg.MD5Hash(user.Password)
	}
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByUsernameAndPassword(username string, password string) (*entity.User, error) {
	var user entity.User
	hashedPassword := pkg.MD5Hash(password)
	err := r.db.Where("username = ? AND password = ?", username, hashedPassword).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Update(user *entity.User) error {
	if user.Password != "" {
		var existingUser entity.User
		if err := r.db.First(&existingUser, user.ID).Error; err == nil {
			if existingUser.Password != user.Password {
				user.Password = pkg.MD5Hash(user.Password)
			}
		}
	}
	return r.db.Save(user).Error
}
func (r *UserRepository) Delete(id string) error {
	return r.db.Delete(&entity.User{}, "id = ?", id).Error
}
func (r *UserRepository) List(limit, offset int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	if err := r.db.Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}
