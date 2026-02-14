package system

type SystemSettingsDTO struct {
	GitTokenHeader string `json:"git_token_header"`
	LicenseHeader  string `json:"license_header"`
}

type UpdateSystemSettingsDTO struct {
	GitTokenHeader string `json:"git_token_header" binding:"required"`
	LicenseHeader  string `json:"license_header" binding:"required"`
}
