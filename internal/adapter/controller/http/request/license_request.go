package request

type CreateLicenseRequest struct {
	Name string `json:"name" binding:"required"`
}

type ListLicenseRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
