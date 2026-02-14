package request

type UpdateSystemSettingsRequest struct {
	GitTokenHeader string `json:"git_token_header" binding:"required"`
	LicenseHeader  string `json:"license_header" binding:"required"`
}
