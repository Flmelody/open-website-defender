package user

type AuthLoginDto struct {
	Name string `form:"name" binding:"required"`
	Pwd  string `form:"pwd" binding:"required"`
}

type CreateUserDto struct {
	Username string
	Password string
	GitToken string
	IsAdmin  bool
}

type UpdateUserDto struct {
	Username string
	Password string
	GitToken string
	IsAdmin  bool
}

type UserDto struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	GitToken string `json:"git_token,omitempty"`
	IsAdmin  bool   `json:"is_admin"`
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
}
