package error

import "errors"

var (
	// User errors
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserAlreadyActive   = errors.New("user already active")
	ErrUserAlreadyInactive = errors.New("user already inactive")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrScopeDenied         = errors.New("domain not in user scope")
)
