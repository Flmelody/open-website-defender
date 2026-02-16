package request

type CreateOAuthClientRequest struct {
	Name         string   `json:"name" binding:"required,max=255"`
	RedirectURIs []string `json:"redirect_uris" binding:"required,min=1"`
	Scopes       string   `json:"scopes" binding:"omitempty,max=1000"`
	Trusted      bool     `json:"trusted"`
}

type UpdateOAuthClientRequest struct {
	Name         string   `json:"name" binding:"omitempty,max=255"`
	RedirectURIs []string `json:"redirect_uris" binding:"omitempty"`
	Scopes       string   `json:"scopes" binding:"omitempty,max=1000"`
	Trusted      bool     `json:"trusted"`
	Active       bool     `json:"active"`
}

type ListOAuthClientRequest struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"size" binding:"omitempty,min=1,max=100"`
}

type OAuthTokenRequest struct {
	GrantType    string `form:"grant_type" json:"grant_type"`
	Code         string `form:"code" json:"code"`
	RedirectURI  string `form:"redirect_uri" json:"redirect_uri"`
	ClientID     string `form:"client_id" json:"client_id"`
	ClientSecret string `form:"client_secret" json:"client_secret"`
	RefreshToken string `form:"refresh_token" json:"refresh_token"`
	CodeVerifier string `form:"code_verifier" json:"code_verifier"`
}

type OAuthRevokeRequest struct {
	Token        string `form:"token" json:"token" binding:"required"`
	ClientID     string `form:"client_id" json:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" json:"client_secret" binding:"required"`
}
