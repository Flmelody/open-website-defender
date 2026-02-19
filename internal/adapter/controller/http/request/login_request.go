package request

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TwoFactorVerifyRequest struct {
	ChallengeToken string `json:"challenge_token" binding:"required"`
	Code           string `json:"code" binding:"required,len=6"`
}
