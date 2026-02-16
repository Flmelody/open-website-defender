package user

type AuthLoginDTO struct {
	Name string `form:"name" binding:"required"`
	Pwd  string `form:"pwd" binding:"required"`
}

type CreateUserDTO struct {
	Username string
	Password string
	GitToken string
	IsAdmin  bool
	Scopes   string
	Email    string
}

type UpdateUserDTO struct {
	Username string
	Password string
	GitToken string
	IsAdmin  bool
	Scopes   string
	Email    string
}

type UserDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	GitToken string `json:"git_token,omitempty"`
	IsAdmin  bool   `json:"is_admin"`
	Scopes   string `json:"scopes"`
	Email    string `json:"email"`
}

type LoginInputDTO struct {
	Username string
	Password string
}

type LoginOutputDTO struct {
	Token string
	User  *UserInfoDTO
}

type UserInfoDTO struct {
	ID       uint
	Username string
	Scopes   string
	IsAdmin  bool
	Email    string
}
