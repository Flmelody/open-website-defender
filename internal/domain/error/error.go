package error

import "errors"

var (
	// User errors
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserAlreadyActive   = errors.New("user already active")
	ErrUserAlreadyInactive = errors.New("user already inactive")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAdminRequired       = errors.New("admin privileges required")
	ErrScopeDenied         = errors.New("domain not in user scope")
	ErrTotpRequired        = errors.New("2FA verification required")
	ErrTotpInvalidCode     = errors.New("invalid 2FA code")
	ErrTotpAlreadyEnabled  = errors.New("2FA is already enabled")
	ErrTotpNotEnabled      = errors.New("2FA is not enabled")
)
