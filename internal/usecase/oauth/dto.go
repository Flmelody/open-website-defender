package oauth

import "time"

// Client DTOs

type CreateOAuthClientDTO struct {
	Name         string
	RedirectURIs []string
	Scopes       string
	Trusted      bool
}

type UpdateOAuthClientDTO struct {
	Name         string
	RedirectURIs []string
	Scopes       string
	Trusted      bool
	Active       bool
}

type OAuthClientDTO struct {
	ID           uint      `json:"id"`
	ClientID     string    `json:"client_id"`
	Name         string    `json:"name"`
	RedirectURIs []string  `json:"redirect_uris"`
	Scopes       string    `json:"scopes"`
	Trusted      bool      `json:"trusted"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type OAuthClientCreatedDTO struct {
	OAuthClientDTO
	ClientSecret string `json:"client_secret"`
}

// User OAuth Authorization DTOs

type UserOAuthAuthorizationDTO struct {
	ClientID     string `json:"client_id"`
	ClientName   string `json:"client_name"`
	Scope        string `json:"scope"`
	AuthorizedAt string `json:"authorized_at"`
}

// Authorization DTOs

type AuthorizeRequestDTO struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

type TokenRequestDTO struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	RefreshToken string
	CodeVerifier string
}

type TokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type UserInfoResponseDTO struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
}

// OIDC Discovery

type OIDCDiscoveryDTO struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}

type JWKSDTO struct {
	Keys []JWKDTO `json:"keys"`
}

type JWKDTO struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
