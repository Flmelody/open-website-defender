package _interface

import "open-website-defender/internal/domain/entity"

type UserRepository interface {
	Save(user *entity.User) error
	FindByID(id string) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	FindByUsernameAndPassword(username string, password string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id string) error
	List(limit, offset int) ([]*entity.User, int64, error)
}

type IpWhiteListRepository interface {
	Create(ip *entity.IpWhiteList) error
	Delete(id uint) error
	List(limit, offset int) ([]*entity.IpWhiteList, int64, error)
	FindByIP(ip string) (*entity.IpWhiteList, error)
}

type IpBlackListRepository interface {
	Create(ip *entity.IpBlackList) error
	Delete(id uint) error
	List(limit, offset int) ([]*entity.IpBlackList, int64, error)
	FindByIP(ip string) (*entity.IpBlackList, error)
}
