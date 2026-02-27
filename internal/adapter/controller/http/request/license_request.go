package request

type CreateLicenseRequest struct {
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

type ListLicenseRequest struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
