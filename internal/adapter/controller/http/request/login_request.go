package request

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TwoFactorVerifyRequest struct {
	ChallengeToken string `json:"challenge_token" binding:"required"`
	Code           string `json:"code" binding:"required,len=6"`
	TrustDevice    bool   `json:"trust_device"`
}

type AdminRecover2FARequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	RecoveryKey string `json:"recovery_key" binding:"required"`
}
