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
	Enabled  bool
	Scopes   string
	Email    string
	Remark   string
}

type UpdateUserDTO struct {
	Username string
	Password string
	GitToken *string
	IsAdmin  *bool
	Enabled  *bool
	Scopes   *string
	Email    *string
	Remark   *string
}

type UserDTO struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	GitToken    string `json:"git_token,omitempty"`
	IsAdmin     bool   `json:"is_admin"`
	Enabled     bool   `json:"enabled"`
	Scopes      string `json:"scopes"`
	Email       string `json:"email"`
	Remark      string `json:"remark"`
	TotpEnabled bool   `json:"totp_enabled"`
}

type LoginInputDTO struct {
	Username           string
	Password           string
	TrustedDeviceToken string
}

type LoginOutputDTO struct {
	Token              string
	TrustedDeviceToken string
	User               *UserInfoDTO
}

type UserInfoDTO struct {
	ID          uint
	Username    string
	Scopes      string
	IsAdmin     bool
	Enabled     bool
	Email       string
	TotpEnabled bool
}

type AdminLoginOutputDTO struct {
	RequiresTwoFA  bool
	ChallengeToken string
	Token          string
	User           *UserInfoDTO
}

type GuardLoginOutputDTO struct {
	RequiresTwoFA  bool
	ChallengeToken string
	Token          string
	User           *UserInfoDTO
}

type TotpSetupOutputDTO struct {
	Secret        string
	QRCodeDataURI string
}

type TwoFALoginInputDTO struct {
	ChallengeToken string
	Code           string
	TrustDevice    bool
}
