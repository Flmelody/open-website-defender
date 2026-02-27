package license

import "time"

type CreateLicenseDTO struct {
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

type LicenseDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Remark    string    `json:"remark"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// LicenseCreatedDTO is returned only on creation; Token is the plaintext shown once.
type LicenseCreatedDTO struct {
	LicenseDTO
	Token string `json:"token"`
}
