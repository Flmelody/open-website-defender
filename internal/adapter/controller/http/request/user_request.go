package request

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	GitToken string `json:"git_token" binding:"omitempty,max=300"`
	IsAdmin  bool   `json:"is_admin" binding:"omitempty"`
	Scopes   string `json:"scopes" binding:"omitempty,max=1000"`
	Email    string `json:"email" binding:"omitempty,max=255"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Password string `json:"password" binding:"omitempty,min=6,max=50"`
	GitToken string `json:"git_token" binding:"omitempty,max=300"`
	IsAdmin  bool   `json:"is_admin" binding:"omitempty"`
	Scopes   string `json:"scopes" binding:"omitempty,max=1000"`
	Email    string `json:"email" binding:"omitempty,max=255"`
}

type ListUserRequest struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"size" binding:"omitempty,min=1,max=100"`
}
