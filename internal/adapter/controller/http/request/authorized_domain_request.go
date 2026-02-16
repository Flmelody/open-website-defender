package request

type CreateAuthorizedDomainRequest struct {
	Name string `json:"name" binding:"required"`
}

type ListAuthorizedDomainRequest struct {
	Page int    `form:"page"`
	Size int    `form:"size"`
	All  string `form:"all"`
}
