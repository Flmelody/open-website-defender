package _interface

import "open-website-defender/internal/domain/entity"

type UserRepository interface {
	Save(user *entity.User) error
	FindByID(id string) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	FindByUsernameAndPassword(username string, password string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id string) error
	List(limit, offset int) ([]*entity.User, int64, error)
	RemoveScopeFromAll(scope string) error
}

type IpWhiteListRepository interface {
	Create(ip *entity.IpWhiteList) error
	Update(ip *entity.IpWhiteList) error
	Delete(id uint) error
	DeleteByDomain(domain string) error
	FindByID(id uint) (*entity.IpWhiteList, error)
	List(limit, offset int) ([]*entity.IpWhiteList, int64, error)
	FindByIP(ip string) (*entity.IpWhiteList, error)
}

type IpBlackListRepository interface {
	Create(ip *entity.IpBlackList) error
	Delete(id uint) error
	List(limit, offset int) ([]*entity.IpBlackList, int64, error)
	FindByIP(ip string) (*entity.IpBlackList, error)
}

type LicenseRepository interface {
	Create(license *entity.License) error
	Delete(id uint) error
	List(limit, offset int) ([]*entity.License, int64, error)
	FindByTokenHash(tokenHash string) (*entity.License, error)
}

type AuthorizedDomainRepository interface {
	Create(domain *entity.AuthorizedDomain) error
	Delete(id uint) error
	FindByID(id uint) (*entity.AuthorizedDomain, error)
	List(limit, offset int) ([]*entity.AuthorizedDomain, int64, error)
	ListAll() ([]*entity.AuthorizedDomain, error)
	FindByName(name string) (*entity.AuthorizedDomain, error)
}

type SystemRepository interface {
	Get() (*entity.System, error)
	Save(system *entity.System) error
}

type OAuthClientRepository interface {
	Create(client *entity.OAuthClient) error
	Update(client *entity.OAuthClient) error
	Delete(id uint) error
	FindByID(id uint) (*entity.OAuthClient, error)
	FindByClientID(clientID string) (*entity.OAuthClient, error)
	List(limit, offset int) ([]*entity.OAuthClient, int64, error)
}

type OAuthAuthorizationCodeRepository interface {
	Create(code *entity.OAuthAuthorizationCode) error
	FindByCode(code string) (*entity.OAuthAuthorizationCode, error)
	MarkUsed(id uint) error
	DeleteExpired() error
}

type OAuthRefreshTokenRepository interface {
	Create(token *entity.OAuthRefreshToken) error
	FindByToken(token string) (*entity.OAuthRefreshToken, error)
	Revoke(id uint) error
	RevokeByClientAndUser(clientID string, userID uint) error
	FindActiveByUserID(userID uint) ([]*entity.OAuthRefreshToken, error)
	DeleteExpired() error
}
