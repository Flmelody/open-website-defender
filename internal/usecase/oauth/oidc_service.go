package oauth

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/config"
	"open-website-defender/internal/pkg"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type OIDCService struct {
	keyID string
}

var (
	oidcService *OIDCService
	oidcOnce    sync.Once
)

func GetOIDCService() *OIDCService {
	oidcOnce.Do(func() {
		// Derive a stable key ID from the public key
		pubKey := pkg.GetRSAPublicKey()
		kid := deriveKeyID(pubKey)
		oidcService = &OIDCService{
			keyID: kid,
		}
	})
	return oidcService
}

func (s *OIDCService) GenerateAccessToken(user *entity.User, clientID string, scope string, lifetime int) (string, error) {
	privateKey := pkg.GetRSAPrivateKey()
	if privateKey == nil {
		return "", fmt.Errorf("RSA private key not initialized")
	}

	issuer := getIssuer()
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":        issuer,
		"sub":        fmt.Sprintf("%d", user.ID),
		"client_id":  clientID,
		"scope":      scope,
		"token_type": "access_token",
		"exp":        now.Add(time.Duration(lifetime) * time.Second).Unix(),
		"iat":        now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID

	return token.SignedString(privateKey)
}

func (s *OIDCService) GenerateIDToken(user *entity.User, clientID string, nonce string, scope string) (string, error) {
	privateKey := pkg.GetRSAPrivateKey()
	if privateKey == nil {
		return "", fmt.Errorf("RSA private key not initialized")
	}

	issuer := getIssuer()
	idTokenLifetime := config.Get().OAuth.IDTokenLifetime
	if idTokenLifetime <= 0 {
		idTokenLifetime = 3600
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":       issuer,
		"sub":       fmt.Sprintf("%d", user.ID),
		"aud":       clientID,
		"exp":       now.Add(time.Duration(idTokenLifetime) * time.Second).Unix(),
		"iat":       now.Unix(),
		"auth_time": now.Unix(),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	// Add profile claims if scope includes "profile"
	if containsScope(scope, "profile") {
		claims["preferred_username"] = user.Username
	}

	// Add email claims if scope includes "email"
	if containsScope(scope, "email") && user.Email != "" {
		claims["email"] = user.Email
		claims["email_verified"] = true
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID

	return token.SignedString(privateKey)
}

func (s *OIDCService) GetDiscoveryDocument() *OIDCDiscoveryDTO {
	issuer := getIssuer()

	return &OIDCDiscoveryDTO{
		Issuer:                            issuer,
		AuthorizationEndpoint:             issuer + "/oauth/authorize",
		TokenEndpoint:                     issuer + "/oauth/token",
		UserInfoEndpoint:                  issuer + "/oauth/userinfo",
		JwksURI:                           issuer + "/.well-known/jwks.json",
		RevocationEndpoint:                issuer + "/oauth/revoke",
		ResponseTypesSupported:            []string{"code"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ScopesSupported:                   []string{"openid", "profile", "email"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "client_secret_basic"},
		ClaimsSupported:                   []string{"sub", "iss", "aud", "exp", "iat", "nonce", "auth_time", "preferred_username", "email", "email_verified"},
		CodeChallengeMethodsSupported:     []string{"S256", "plain"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
	}
}

func (s *OIDCService) GetJWKS() *JWKSDTO {
	pubKey := pkg.GetRSAPublicKey()
	if pubKey == nil {
		return &JWKSDTO{Keys: []JWKDTO{}}
	}

	return &JWKSDTO{
		Keys: []JWKDTO{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: s.keyID,
				Alg: "RS256",
				N:   base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes()),
			},
		},
	}
}

func (s *OIDCService) GetUserInfo(userID uint) (*UserInfoResponseDTO, error) {
	oauthSvc := GetOAuthService()
	user, err := oauthSvc.userRepo.FindByID(fmt.Sprintf("%d", userID))
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &UserInfoResponseDTO{
		Sub:               fmt.Sprintf("%d", user.ID),
		PreferredUsername: user.Username,
		Email:             user.Email,
		EmailVerified:     user.Email != "",
	}, nil
}

func getIssuer() string {
	issuer := viper.GetString("oauth.issuer")
	if issuer == "" {
		// Fallback: construct from BACKEND_HOST
		backendHost := viper.GetString("BACKEND_HOST")
		if backendHost == "" {
			backendHost = "http://localhost:9999/wall"
		}
		issuer = backendHost
	}
	return issuer
}

func deriveKeyID(pubKey *rsa.PublicKey) string {
	if pubKey == nil {
		return "default"
	}
	// Create a stable key ID from the public key modulus
	h := sha256.Sum256(pubKey.N.Bytes())
	return base64.RawURLEncoding.EncodeToString(h[:8])
}
