package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/config"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	_interface "open-website-defender/internal/usecase/interface"
	"strings"
	"sync"
	"time"
)

var (
	ErrClientNotFound      = errors.New("oauth client not found")
	ErrClientInactive      = errors.New("oauth client is inactive")
	ErrInvalidRedirectURI  = errors.New("invalid redirect_uri")
	ErrInvalidScope        = errors.New("invalid scope")
	ErrInvalidGrant        = errors.New("invalid grant")
	ErrCodeExpired         = errors.New("authorization code expired")
	ErrCodeUsed            = errors.New("authorization code already used")
	ErrInvalidClientSecret = errors.New("invalid client secret")
	ErrInvalidCodeVerifier = errors.New("invalid code verifier")
	ErrTokenRevoked        = errors.New("token has been revoked")
	ErrTokenExpired        = errors.New("token has expired")
)

type OAuthService struct {
	clientRepo  _interface.OAuthClientRepository
	codeRepo    _interface.OAuthAuthorizationCodeRepository
	refreshRepo _interface.OAuthRefreshTokenRepository
	userRepo    _interface.UserRepository
}

var (
	oauthService *OAuthService
	oauthOnce    sync.Once
)

func GetOAuthService() *OAuthService {
	oauthOnce.Do(func() {
		oauthService = &OAuthService{
			clientRepo:  repository.NewOAuthClientRepository(database.DB),
			codeRepo:    repository.NewOAuthAuthorizationCodeRepository(database.DB),
			refreshRepo: repository.NewOAuthRefreshTokenRepository(database.DB),
			userRepo:    repository.NewUserRepository(database.DB),
		}
		// Start periodic cleanup
		go oauthService.periodicCleanup()
	})
	return oauthService
}

func (s *OAuthService) periodicCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if err := s.codeRepo.DeleteExpired(); err != nil {
			logging.Sugar.Warnf("Failed to cleanup expired auth codes: %v", err)
		}
		if err := s.refreshRepo.DeleteExpired(); err != nil {
			logging.Sugar.Warnf("Failed to cleanup expired refresh tokens: %v", err)
		}
	}
}

// --- Client CRUD ---

func (s *OAuthService) CreateClient(input *CreateOAuthClientDTO) (*OAuthClientCreatedDTO, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if len(input.RedirectURIs) == 0 {
		return nil, errors.New("at least one redirect_uri is required")
	}

	clientID := pkg.GenerateRandomToken(32)
	rawSecret := pkg.GenerateRandomToken(32)
	hashedSecret, err := pkg.HashPassword(rawSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash client secret: %w", err)
	}

	redirectURIsJSON, err := json.Marshal(input.RedirectURIs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal redirect URIs: %w", err)
	}

	scopes := input.Scopes
	if scopes == "" {
		scopes = "openid profile email"
	}

	client := &entity.OAuthClient{
		ClientID:     clientID,
		ClientSecret: hashedSecret,
		Name:         input.Name,
		RedirectURIs: string(redirectURIsJSON),
		Scopes:       scopes,
		Trusted:      input.Trusted,
		Active:       true,
	}

	if err := s.clientRepo.Create(client); err != nil {
		return nil, fmt.Errorf("failed to create oauth client: %w", err)
	}

	return &OAuthClientCreatedDTO{
		OAuthClientDTO: OAuthClientDTO{
			ID:           client.ID,
			ClientID:     client.ClientID,
			Name:         client.Name,
			RedirectURIs: input.RedirectURIs,
			Scopes:       client.Scopes,
			Trusted:      client.Trusted,
			Active:       client.Active,
			CreatedAt:    client.CreatedAt,
		},
		ClientSecret: rawSecret,
	}, nil
}

func (s *OAuthService) UpdateClient(id uint, input *UpdateOAuthClientDTO) (*OAuthClientDTO, error) {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrClientNotFound
	}

	if input.Name != "" {
		client.Name = input.Name
	}
	if len(input.RedirectURIs) > 0 {
		redirectURIsJSON, err := json.Marshal(input.RedirectURIs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal redirect URIs: %w", err)
		}
		client.RedirectURIs = string(redirectURIsJSON)
	}
	if input.Scopes != "" {
		client.Scopes = input.Scopes
	}
	client.Trusted = input.Trusted
	client.Active = input.Active

	if err := s.clientRepo.Update(client); err != nil {
		return nil, fmt.Errorf("failed to update oauth client: %w", err)
	}

	return clientToDTO(client), nil
}

func (s *OAuthService) DeleteClient(id uint) error {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrClientNotFound
	}
	return s.clientRepo.Delete(id)
}

func (s *OAuthService) GetClient(id uint) (*OAuthClientDTO, error) {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrClientNotFound
	}
	return clientToDTO(client), nil
}

func (s *OAuthService) ListClients(page, size int) ([]*OAuthClientDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	clients, total, err := s.clientRepo.List(size, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*OAuthClientDTO, 0, len(clients))
	for _, c := range clients {
		dtos = append(dtos, clientToDTO(c))
	}
	return dtos, total, nil
}

// --- Authorization Code Flow ---

func (s *OAuthService) Authorize(req *AuthorizeRequestDTO, userID uint) (string, error) {
	client, err := s.clientRepo.FindByClientID(req.ClientID)
	if err != nil {
		return "", err
	}
	if client == nil {
		return "", ErrClientNotFound
	}
	if !client.Active {
		return "", ErrClientInactive
	}

	// Validate redirect URI
	if !s.isValidRedirectURI(client, req.RedirectURI) {
		return "", ErrInvalidRedirectURI
	}

	// Validate scope
	if !s.isValidScope(client, req.Scope) {
		return "", ErrInvalidScope
	}

	codeLifetime := config.Get().OAuth.AuthorizationCodeLifetime
	if codeLifetime <= 0 {
		codeLifetime = 300
	}

	code := pkg.GenerateRandomToken(64)
	authCode := &entity.OAuthAuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		UserID:              userID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().UTC().Add(time.Duration(codeLifetime) * time.Second),
	}

	if err := s.codeRepo.Create(authCode); err != nil {
		return "", fmt.Errorf("failed to create authorization code: %w", err)
	}

	return code, nil
}

func (s *OAuthService) ExchangeCode(req *TokenRequestDTO) (*TokenResponseDTO, error) {
	// Find the authorization code
	authCode, err := s.codeRepo.FindByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if authCode == nil {
		return nil, ErrInvalidGrant
	}
	if authCode.Used {
		return nil, ErrCodeUsed
	}
	if time.Now().UTC().After(authCode.ExpiresAt) {
		return nil, ErrCodeExpired
	}

	// Mark code as used immediately
	if err := s.codeRepo.MarkUsed(authCode.ID); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Validate client
	client, err := s.clientRepo.FindByClientID(authCode.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrClientNotFound
	}

	// Validate client credentials
	if req.ClientID != authCode.ClientID {
		return nil, ErrInvalidGrant
	}
	if !pkg.CheckPassword(client.ClientSecret, req.ClientSecret) {
		return nil, ErrInvalidClientSecret
	}

	// Validate redirect URI matches
	if req.RedirectURI != authCode.RedirectURI {
		return nil, ErrInvalidRedirectURI
	}

	// Validate PKCE
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			return nil, ErrInvalidCodeVerifier
		}
		if !verifyCodeChallenge(authCode.CodeChallenge, authCode.CodeChallengeMethod, req.CodeVerifier) {
			return nil, ErrInvalidCodeVerifier
		}
	}

	// Get user
	user, err := s.userRepo.FindByID(fmt.Sprintf("%d", authCode.UserID))
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Generate tokens
	accessTokenLifetime := config.Get().OAuth.AccessTokenLifetime
	if accessTokenLifetime <= 0 {
		accessTokenLifetime = 3600
	}

	oidcService := GetOIDCService()
	accessToken, err := oidcService.GenerateAccessToken(user, authCode.ClientID, authCode.Scope, accessTokenLifetime)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshTokenLifetime := config.Get().OAuth.RefreshTokenLifetime
	if refreshTokenLifetime <= 0 {
		refreshTokenLifetime = 2592000
	}

	refreshTokenStr := pkg.GenerateRandomToken(32)
	refreshToken := &entity.OAuthRefreshToken{
		Token:     refreshTokenStr,
		ClientID:  authCode.ClientID,
		UserID:    authCode.UserID,
		Scope:     authCode.Scope,
		ExpiresAt: time.Now().UTC().Add(time.Duration(refreshTokenLifetime) * time.Second),
	}
	if err := s.refreshRepo.Create(refreshToken); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Generate ID token if openid scope requested
	var idToken string
	if containsScope(authCode.Scope, "openid") {
		idToken, err = oidcService.GenerateIDToken(user, authCode.ClientID, authCode.Nonce, authCode.Scope)
		if err != nil {
			logging.Sugar.Warnf("Failed to generate ID token: %v", err)
		}
	}

	return &TokenResponseDTO{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenLifetime,
		RefreshToken: refreshTokenStr,
		IDToken:      idToken,
		Scope:        authCode.Scope,
	}, nil
}

func (s *OAuthService) RefreshAccessToken(req *TokenRequestDTO) (*TokenResponseDTO, error) {
	// Find refresh token
	rt, err := s.refreshRepo.FindByToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, ErrInvalidGrant
	}
	if rt.Revoked {
		return nil, ErrTokenRevoked
	}
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Validate client
	client, err := s.clientRepo.FindByClientID(rt.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrClientNotFound
	}
	if !pkg.CheckPassword(client.ClientSecret, req.ClientSecret) {
		return nil, ErrInvalidClientSecret
	}

	// Get user
	user, err := s.userRepo.FindByID(fmt.Sprintf("%d", rt.UserID))
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Revoke old refresh token
	if err := s.refreshRepo.Revoke(rt.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Generate new tokens
	accessTokenLifetime := config.Get().OAuth.AccessTokenLifetime
	if accessTokenLifetime <= 0 {
		accessTokenLifetime = 3600
	}
	refreshTokenLifetime := config.Get().OAuth.RefreshTokenLifetime
	if refreshTokenLifetime <= 0 {
		refreshTokenLifetime = 2592000
	}

	oidcService := GetOIDCService()
	newAccessToken, err := oidcService.GenerateAccessToken(user, rt.ClientID, rt.Scope, accessTokenLifetime)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	newRefreshTokenStr := pkg.GenerateRandomToken(32)
	newRefreshToken := &entity.OAuthRefreshToken{
		Token:     newRefreshTokenStr,
		ClientID:  rt.ClientID,
		UserID:    rt.UserID,
		Scope:     rt.Scope,
		ExpiresAt: time.Now().UTC().Add(time.Duration(refreshTokenLifetime) * time.Second),
	}
	if err := s.refreshRepo.Create(newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	var idToken string
	if containsScope(rt.Scope, "openid") {
		idToken, err = oidcService.GenerateIDToken(user, rt.ClientID, "", rt.Scope)
		if err != nil {
			logging.Sugar.Warnf("Failed to generate ID token on refresh: %v", err)
		}
	}

	return &TokenResponseDTO{
		AccessToken:  newAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenLifetime,
		RefreshToken: newRefreshTokenStr,
		IDToken:      idToken,
		Scope:        rt.Scope,
	}, nil
}

func (s *OAuthService) RevokeToken(token string, clientID string, clientSecret string) error {
	// Validate client
	client, err := s.clientRepo.FindByClientID(clientID)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrClientNotFound
	}
	if !pkg.CheckPassword(client.ClientSecret, clientSecret) {
		return ErrInvalidClientSecret
	}

	// Try as refresh token
	rt, err := s.refreshRepo.FindByToken(token)
	if err != nil {
		return err
	}
	if rt != nil && rt.ClientID == clientID {
		return s.refreshRepo.Revoke(rt.ID)
	}

	// Token not found is not an error per RFC 7009
	return nil
}

// --- User OAuth Authorizations ---

func (s *OAuthService) ListUserAuthorizations(userID uint) ([]*UserOAuthAuthorizationDTO, error) {
	tokens, err := s.refreshRepo.FindActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Group by client_id, keep the earliest created_at
	type clientAuth struct {
		clientID     string
		scope        string
		authorizedAt time.Time
	}
	seen := make(map[string]*clientAuth)
	for _, t := range tokens {
		if existing, ok := seen[t.ClientID]; ok {
			if t.CreatedAt.Before(existing.authorizedAt) {
				existing.authorizedAt = t.CreatedAt
			}
		} else {
			seen[t.ClientID] = &clientAuth{
				clientID:     t.ClientID,
				scope:        t.Scope,
				authorizedAt: t.CreatedAt,
			}
		}
	}

	result := make([]*UserOAuthAuthorizationDTO, 0, len(seen))
	for _, auth := range seen {
		clientName := auth.clientID
		client, err := s.clientRepo.FindByClientID(auth.clientID)
		if err == nil && client != nil {
			clientName = client.Name
		}
		result = append(result, &UserOAuthAuthorizationDTO{
			ClientID:     auth.clientID,
			ClientName:   clientName,
			Scope:        auth.scope,
			AuthorizedAt: auth.authorizedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *OAuthService) RevokeUserAuthorization(userID uint, clientID string) error {
	return s.refreshRepo.RevokeByClientAndUser(clientID, userID)
}

func (s *OAuthService) FindClientByID(clientID string) (*entity.OAuthClient, error) {
	return s.clientRepo.FindByClientID(clientID)
}

func (s *OAuthService) IsTrustedClient(clientID string) bool {
	client, err := s.clientRepo.FindByClientID(clientID)
	if err != nil || client == nil {
		return false
	}
	return client.Trusted && client.Active
}

// --- Helpers ---

func (s *OAuthService) isValidRedirectURI(client *entity.OAuthClient, uri string) bool {
	var uris []string
	if err := json.Unmarshal([]byte(client.RedirectURIs), &uris); err != nil {
		return false
	}
	for _, allowed := range uris {
		if allowed == uri {
			return true
		}
	}
	return false
}

func (s *OAuthService) isValidScope(client *entity.OAuthClient, requestedScope string) bool {
	if requestedScope == "" {
		return true
	}
	allowedScopes := strings.Fields(client.Scopes)
	allowedSet := make(map[string]bool, len(allowedScopes))
	for _, s := range allowedScopes {
		allowedSet[s] = true
	}
	for _, s := range strings.Fields(requestedScope) {
		if !allowedSet[s] {
			return false
		}
	}
	return true
}

func clientToDTO(client *entity.OAuthClient) *OAuthClientDTO {
	var uris []string
	_ = json.Unmarshal([]byte(client.RedirectURIs), &uris)

	return &OAuthClientDTO{
		ID:           client.ID,
		ClientID:     client.ClientID,
		Name:         client.Name,
		RedirectURIs: uris,
		Scopes:       client.Scopes,
		Trusted:      client.Trusted,
		Active:       client.Active,
		CreatedAt:    client.CreatedAt,
	}
}

func containsScope(scopeStr string, target string) bool {
	for _, s := range strings.Fields(scopeStr) {
		if s == target {
			return true
		}
	}
	return false
}

func verifyCodeChallenge(challenge, method, verifier string) bool {
	switch method {
	case "S256":
		h := sha256.Sum256([]byte(verifier))
		computed := base64.RawURLEncoding.EncodeToString(h[:])
		return computed == challenge
	case "plain", "":
		return verifier == challenge
	default:
		return false
	}
}
