package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"open-website-defender/internal/usecase/oauth"
	"open-website-defender/internal/usecase/user"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OIDCDiscovery handles GET /.well-known/openid-configuration
func OIDCDiscovery(c *gin.Context) {
	service := oauth.GetOIDCService()
	c.JSON(http.StatusOK, service.GetDiscoveryDocument())
}

// JWKS handles GET /.well-known/jwks.json
func JWKS(c *gin.Context) {
	service := oauth.GetOIDCService()
	c.JSON(http.StatusOK, service.GetJWKS())
}

// OAuthAuthorize handles GET /oauth/authorize
func OAuthAuthorize(c *gin.Context) {
	// Parse query parameters
	responseType := c.Query("response_type")
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")
	state := c.Query("state")
	nonce := c.Query("nonce")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	if responseType != "code" {
		redirectWithError(c, redirectURI, "unsupported_response_type", "Only 'code' response type is supported", state)
		return
	}

	if clientID == "" {
		response.BadRequest(c, "client_id is required")
		return
	}

	// Authenticate user via OWD session cookie
	userInfo := authenticateOWDUser(c)
	if userInfo == nil {
		// Redirect to guard login with return URL
		guardURL := buildGuardLoginURL(c)
		c.Redirect(http.StatusFound, guardURL)
		return
	}

	oauthService := oauth.GetOAuthService()

	// Check if client is trusted (skip consent)
	if oauthService.IsTrustedClient(clientID) {
		// Auto-approve for trusted clients
		code, err := oauthService.Authorize(&oauth.AuthorizeRequestDTO{
			ResponseType:        responseType,
			ClientID:            clientID,
			RedirectURI:         redirectURI,
			Scope:               scope,
			State:               state,
			Nonce:               nonce,
			CodeChallenge:       codeChallenge,
			CodeChallengeMethod: codeChallengeMethod,
		}, userInfo.ID)
		if err != nil {
			handleAuthorizeError(c, redirectURI, err, state)
			return
		}

		redirectWithCode(c, redirectURI, code, state)
		return
	}

	// Non-trusted client: show consent page
	// Build consent URL with all params forwarded
	consentParams := url.Values{
		"response_type":         {responseType},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {scope},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {codeChallengeMethod},
	}

	// Get client name for display
	client, _ := oauthService.FindClientByID(clientID)
	clientName := clientID
	if client != nil {
		clientName = client.Name
	}
	consentParams.Set("client_name", clientName)

	// Redirect to the guard's consent page
	guardConsentURL := buildGuardConsentURL(c, consentParams)
	c.Redirect(http.StatusFound, guardConsentURL)
}

// OAuthConsent handles POST /oauth/consent (user approves/denies)
func OAuthConsent(c *gin.Context) {
	action := c.PostForm("action")
	if action != "approve" {
		redirectURI := c.PostForm("redirect_uri")
		state := c.PostForm("state")
		redirectWithError(c, redirectURI, "access_denied", "User denied the authorization request", state)
		return
	}

	// Authenticate user
	userInfo := authenticateOWDUser(c)
	if userInfo == nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	oauthService := oauth.GetOAuthService()

	code, err := oauthService.Authorize(&oauth.AuthorizeRequestDTO{
		ResponseType:        c.PostForm("response_type"),
		ClientID:            c.PostForm("client_id"),
		RedirectURI:         c.PostForm("redirect_uri"),
		Scope:               c.PostForm("scope"),
		State:               c.PostForm("state"),
		Nonce:               c.PostForm("nonce"),
		CodeChallenge:       c.PostForm("code_challenge"),
		CodeChallengeMethod: c.PostForm("code_challenge_method"),
	}, userInfo.ID)

	if err != nil {
		handleAuthorizeError(c, c.PostForm("redirect_uri"), err, c.PostForm("state"))
		return
	}

	redirectWithCode(c, c.PostForm("redirect_uri"), code, c.PostForm("state"))
}

// OAuthToken handles POST /oauth/token
func OAuthToken(c *gin.Context) {
	var req request.OAuthTokenRequest

	// Support both form-encoded and JSON
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			oauthError(c, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			oauthError(c, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
	}

	// Support HTTP Basic auth for client credentials
	if req.ClientID == "" || req.ClientSecret == "" {
		clientID, clientSecret, ok := parseBasicAuth(c)
		if ok {
			if req.ClientID == "" {
				req.ClientID = clientID
			}
			if req.ClientSecret == "" {
				req.ClientSecret = clientSecret
			}
		}
	}

	oauthService := oauth.GetOAuthService()

	switch req.GrantType {
	case "authorization_code":
		tokenResp, err := oauthService.ExchangeCode(&oauth.TokenRequestDTO{
			GrantType:    req.GrantType,
			Code:         req.Code,
			RedirectURI:  req.RedirectURI,
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
			CodeVerifier: req.CodeVerifier,
		})
		if err != nil {
			handleTokenError(c, err)
			return
		}
		c.JSON(http.StatusOK, tokenResp)

	case "refresh_token":
		tokenResp, err := oauthService.RefreshAccessToken(&oauth.TokenRequestDTO{
			GrantType:    req.GrantType,
			RefreshToken: req.RefreshToken,
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
		})
		if err != nil {
			handleTokenError(c, err)
			return
		}
		c.JSON(http.StatusOK, tokenResp)

	default:
		oauthError(c, http.StatusBadRequest, "unsupported_grant_type", "Supported grant types: authorization_code, refresh_token")
	}
}

// OAuthUserInfo handles GET/POST /oauth/userinfo
func OAuthUserInfo(c *gin.Context) {
	// Extract bearer token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.Header("WWW-Authenticate", "Bearer")
		oauthError(c, http.StatusUnauthorized, "invalid_token", "Bearer token required")
		return
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse the JWT access token to extract user identity
	claims, err := parseIDToken(accessToken)
	if err != nil {
		c.Header("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		oauthError(c, http.StatusUnauthorized, "invalid_token", "Invalid or expired access token")
		return
	}

	// Verify this is an access token (not an ID token used by mistake)
	if tokenType, _ := claims["token_type"].(string); tokenType != "access_token" {
		c.Header("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		oauthError(c, http.StatusUnauthorized, "invalid_token", "Expected an access token")
		return
	}

	sub, _ := claims["sub"].(string)
	var userID uint
	fmt.Sscanf(sub, "%d", &userID)
	if userID == 0 {
		c.Header("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		oauthError(c, http.StatusUnauthorized, "invalid_token", "Invalid subject in token")
		return
	}

	oidcService := oauth.GetOIDCService()
	info, err := oidcService.GetUserInfo(userID)
	if err != nil {
		oauthError(c, http.StatusInternalServerError, "server_error", "Failed to get user info")
		return
	}
	c.JSON(http.StatusOK, info)
}

// OAuthRevoke handles POST /oauth/revoke
func OAuthRevoke(c *gin.Context) {
	var req request.OAuthRevokeRequest
	if err := c.ShouldBind(&req); err != nil {
		oauthError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Support HTTP Basic auth
	if req.ClientID == "" || req.ClientSecret == "" {
		clientID, clientSecret, ok := parseBasicAuth(c)
		if ok {
			if req.ClientID == "" {
				req.ClientID = clientID
			}
			if req.ClientSecret == "" {
				req.ClientSecret = clientSecret
			}
		}
	}

	oauthService := oauth.GetOAuthService()
	if err := oauthService.RevokeToken(req.Token, req.ClientID, req.ClientSecret); err != nil {
		if err == oauth.ErrInvalidClientSecret || err == oauth.ErrClientNotFound {
			oauthError(c, http.StatusUnauthorized, "invalid_client", err.Error())
			return
		}
		oauthError(c, http.StatusInternalServerError, "server_error", "Failed to revoke token")
		return
	}

	c.Status(http.StatusOK)
}

// --- Helpers ---

func authenticateOWDUser(c *gin.Context) *user.UserInfoDTO {
	authService := user.GetAuthService()

	// Check cookie
	clientToken, err := c.Cookie("flmelody.token")
	if err != nil || clientToken == "" {
		// Fallback: check Cookie header manually
		cookieHeader := c.GetHeader("Cookie")
		if cookieHeader != "" {
			cookies := strings.Split(cookieHeader, ";")
			for _, cookie := range cookies {
				cookie = strings.TrimSpace(cookie)
				if strings.HasPrefix(cookie, "flmelody.token=") {
					clientToken = strings.TrimPrefix(cookie, "flmelody.token=")
					break
				}
			}
		}
	}

	if clientToken == "" {
		// Also check Authorization header
		authHeader := c.GetHeader("Defender-Authorization")
		if authHeader != "" {
			clientToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if clientToken == "" {
		return nil
	}

	userInfo, err := authService.ValidateToken(clientToken)
	if err != nil {
		return nil
	}
	return userInfo
}

func buildGuardLoginURL(c *gin.Context) string {
	// Build the full current URL as the redirect target
	scheme := "https"
	if c.GetHeader("X-Forwarded-Proto") != "" {
		scheme = c.GetHeader("X-Forwarded-Proto")
	} else if c.Request.TLS == nil {
		scheme = "http"
	}
	currentURL := fmt.Sprintf("%s://%s%s?%s", scheme, c.Request.Host, c.Request.URL.Path, c.Request.URL.RawQuery)

	// Get guard path from config
	guardPath := "/wall/guard"
	return fmt.Sprintf("%s/login?redirect=%s", guardPath, url.QueryEscape(currentURL))
}

func buildGuardConsentURL(c *gin.Context, params url.Values) string {
	guardPath := "/wall/guard"
	return fmt.Sprintf("%s/consent?%s", guardPath, params.Encode())
}

func redirectWithCode(c *gin.Context, redirectURI string, code string, state string) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		response.BadRequest(c, "Invalid redirect_uri")
		return
	}
	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

func redirectWithError(c *gin.Context, redirectURI string, errCode string, description string, state string) {
	if redirectURI == "" {
		response.BadRequest(c, description)
		return
	}
	u, err := url.Parse(redirectURI)
	if err != nil {
		response.BadRequest(c, "Invalid redirect_uri")
		return
	}
	q := u.Query()
	q.Set("error", errCode)
	q.Set("error_description", description)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

func handleAuthorizeError(c *gin.Context, redirectURI string, err error, state string) {
	switch err {
	case oauth.ErrClientNotFound:
		response.BadRequest(c, "Unknown client_id")
	case oauth.ErrClientInactive:
		response.BadRequest(c, "Client is inactive")
	case oauth.ErrInvalidRedirectURI:
		response.BadRequest(c, "Invalid redirect_uri")
	case oauth.ErrInvalidScope:
		redirectWithError(c, redirectURI, "invalid_scope", "Requested scope is not allowed", state)
	default:
		logging.Sugar.Errorf("OAuth authorize error: %v", err)
		redirectWithError(c, redirectURI, "server_error", "Internal server error", state)
	}
}

func handleTokenError(c *gin.Context, err error) {
	switch err {
	case oauth.ErrInvalidGrant, oauth.ErrCodeExpired, oauth.ErrCodeUsed:
		oauthError(c, http.StatusBadRequest, "invalid_grant", err.Error())
	case oauth.ErrInvalidClientSecret, oauth.ErrClientNotFound:
		oauthError(c, http.StatusUnauthorized, "invalid_client", err.Error())
	case oauth.ErrInvalidRedirectURI:
		oauthError(c, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
	case oauth.ErrInvalidCodeVerifier:
		oauthError(c, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
	case oauth.ErrTokenRevoked:
		oauthError(c, http.StatusBadRequest, "invalid_grant", "token has been revoked")
	case oauth.ErrTokenExpired:
		oauthError(c, http.StatusBadRequest, "invalid_grant", "token has expired")
	default:
		logging.Sugar.Errorf("OAuth token error: %v", err)
		oauthError(c, http.StatusInternalServerError, "server_error", "Internal server error")
	}
}

func oauthError(c *gin.Context, status int, errCode string, description string) {
	c.JSON(status, gin.H{
		"error":             errCode,
		"error_description": description,
	})
}

func parseBasicAuth(c *gin.Context) (string, string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
	if err != nil {
		return "", "", false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	// URL-decode client_id and client_secret per RFC 6749
	clientID, _ := url.QueryUnescape(parts[0])
	clientSecret, _ := url.QueryUnescape(parts[1])
	return clientID, clientSecret, true
}

func parseIDToken(tokenStr string) (map[string]interface{}, error) {
	pubKey := pkg.GetRSAPublicKey()
	if pubKey == nil {
		return nil, fmt.Errorf("no RSA public key")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
