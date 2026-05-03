package handler

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"castellum/internal/adapter/controller/http/request"
	"castellum/internal/adapter/controller/http/response"
	domainError "castellum/internal/domain/error"
	"castellum/internal/infrastructure/config"
	"castellum/internal/infrastructure/logging"
	"castellum/internal/pkg"
	"castellum/internal/usecase/iplist"
	"castellum/internal/usecase/license"
	"castellum/internal/usecase/system"
	"castellum/internal/usecase/threat"
	"castellum/internal/usecase/user"

	"github.com/gin-gonic/gin"
)

const authCookieName = "flmelody.token"

func isGitRequest(c *gin.Context) bool {
	if !strings.HasPrefix(c.GetHeader("User-Agent"), "git/") {
		return false
	}
	uri := c.GetHeader("X-Original-URI")
	if uri == "" {
		return false
	}
	// Strip query string
	if idx := strings.IndexByte(uri, '?'); idx != -1 {
		uri = uri[:idx]
	}
	return strings.HasSuffix(uri, "/info/refs") ||
		strings.HasSuffix(uri, "/git-upload-pack") ||
		strings.HasSuffix(uri, "/git-receive-pack") ||
		strings.HasSuffix(uri, "/HEAD") ||
		strings.Contains(uri, "/objects/")
}

func getRequestedDomain(c *gin.Context) string {
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return pkg.StripPort(host)
}

func checkUserScope(userInfo *user.UserInfoDTO, domain string) bool {
	if userInfo.IsAdmin {
		return true
	}
	return pkg.CheckDomainScope(userInfo.Scopes, domain)
}

func getAuthCookieDomain() string {
	if appCfg := config.GetAppConfig(); appCfg != nil {
		return appCfg.GuardDomain
	}
	return ""
}

func getAuthCookieMaxAge() int {
	expirationHrs := config.Get().Security.TokenExpirationHrs
	if expirationHrs <= 0 {
		expirationHrs = 24
	}
	return expirationHrs * 3600
}

func setAuthCookie(c *gin.Context, token string) {
	if token == "" {
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, token, getAuthCookieMaxAge(), "/", getAuthCookieDomain(), config.Get().Security.SecureCookies, true)
}

func clearAuthCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, "", -1, "/", getAuthCookieDomain(), config.Get().Security.SecureCookies, true)
}

func authTokenFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Castellum-Authorization")
	if authHeader != "" {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	if cookieToken, err := c.Cookie(authCookieName); err == nil {
		return cookieToken
	}
	return ""
}

func userInfoResponse(userInfo *user.UserInfoDTO) *UserInfoResponse {
	return &UserInfoResponse{
		ID:       userInfo.ID,
		Username: userInfo.Username,
	}
}

type LoginResponse struct {
	Token string            `json:"token,omitempty"`
	User  *UserInfoResponse `json:"user"`
}

type AdminLoginResponse struct {
	RequiresTwoFactor bool              `json:"requires_two_factor"`
	ChallengeToken    string            `json:"challenge_token,omitempty"`
	Token             string            `json:"token,omitempty"`
	User              *UserInfoResponse `json:"user"`
}

type GuardLoginResponse struct {
	RequiresTwoFA  bool              `json:"requires_two_fa"`
	ChallengeToken string            `json:"challenge_token,omitempty"`
	Token          string            `json:"token,omitempty"`
	User           *UserInfoResponse `json:"user"`
}

type UserInfoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// AuthMiddleware is used as a middleware to protect routes.
func AuthMiddleware(c *gin.Context) {
	// Check Token
	service := user.GetAuthService()
	tokenString := authTokenFromRequest(c)
	if tokenString == "" {
		response.Unauthorized(c, "No authentication token provided")
		c.Abort()
		return
	}

	userInfo, err := service.ValidateToken(tokenString)
	if err != nil {
		logging.Sugar.Warnf("Token validation failed: %v", err)
		response.Unauthorized(c, "Invalid or expired token")
		c.Abort()
		return
	}

	c.Set("user", userInfo)
	c.Next()
}

// Auth is used as a standalone endpoint (GET /auth) to verify credentials.
// It checks Blacklist -> Whitelist -> Token and returns status.
func Auth(c *gin.Context) {
	// 1. Get Client IP
	clientIP := c.ClientIP()

	// 2. Check Blacklist
	blackListService := iplist.GetIpBlackListService()
	if blackListItem, _ := blackListService.FindByIP(clientIP); blackListItem != nil {
		logging.Sugar.Warnf("Access denied for blacklisted IP: %s", clientIP)
		response.Forbidden(c, "Access denied")
		return
	}

	// 3. Check Whitelist
	requestedDomain := getRequestedDomain(c)
	whiteListService := iplist.GetIpWhiteListService()
	if whiteListItem, _ := whiteListService.FindByIP(clientIP); whiteListItem != nil {
		if whiteListItem.Domain == "" || pkg.MatchDomain(whiteListItem.Domain, requestedDomain) {
			logging.Sugar.Infof("Access granted for whitelisted IP: %s (domain: %s)", clientIP, whiteListItem.Domain)
			response.Success(c, gin.H{
				"message": "Access granted via IP whitelist",
				"ip":      clientIP,
			})
			return
		}
		logging.Sugar.Infof("Whitelist IP %s matched but domain '%s' not in bound domain '%s', falling through to token auth", clientIP, requestedDomain, whiteListItem.Domain)
	}

	// 4. Check Token
	service := user.GetAuthService()
	clientToken, err := c.Cookie("flmelody.token")
	if err != nil {
		clientToken = c.GetHeader("Castellum-Authorization")
	}
	if clientToken == "" {
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

	if clientToken != "" {
		tokenString := strings.TrimPrefix(clientToken, "Bearer ")
		userInfo, err := service.ValidateToken(tokenString)
		if err == nil {
			if !checkUserScope(userInfo, requestedDomain) {
				logging.Sugar.Warnf("Scope denied for user '%s': domain '%s' not in scopes '%s'", userInfo.Username, requestedDomain, userInfo.Scopes)
				response.Forbidden(c, "Domain not in user scope")
				return
			}
			response.Success(c, userInfo)
			return
		}
		logging.Sugar.Warnf("Token validation failed: %v", err)
	}

	// 5. Check Git Token header (configurable, only for git requests)
	systemService := system.GetSystemService()
	gitHeaderName, licenseHeaderName := systemService.GetHeaderNames()

	if isGitRequest(c) {
		gitTokenHeader := c.GetHeader(gitHeaderName)
		if gitTokenHeader != "" {
			parts := strings.SplitN(gitTokenHeader, ":", 2)
			if len(parts) == 2 {
				userInfo, err := service.ValidateGitToken(parts[0], parts[1])
				if err == nil {
					if !checkUserScope(userInfo, requestedDomain) {
						logging.Sugar.Warnf("Scope denied for user '%s': domain '%s' not in scopes '%s'", userInfo.Username, requestedDomain, userInfo.Scopes)
						response.Forbidden(c, "Domain not in user scope")
						return
					}
					// Auto-trust: create temporary whitelist for GCM's subsequent OAuth requests
					if err := whiteListService.GrantTemporaryAccess(clientIP, requestedDomain, 300, "git-token-auto-trust"); err != nil {
						logging.Sugar.Warnf("Failed to grant temporary whitelist for git token auto-trust: %v", err)
					}
					logging.Sugar.Infof("Access granted via git token for user: %s", parts[0])
					response.Success(c, userInfo)
					return
				}
				logging.Sugar.Warnf("Git token validation failed for user '%s': %v", parts[0], err)
			}
		}
	}

	// 6. Check License header (configurable)
	licenseToken := c.GetHeader(licenseHeaderName)
	if licenseToken != "" {
		licenseService := license.GetLicenseService()
		valid, err := licenseService.ValidateToken(licenseToken)
		if err == nil && valid {
			logging.Sugar.Infof("Access granted via license token from IP: %s", clientIP)
			response.Success(c, gin.H{
				"message": "Access granted via license",
				"ip":      clientIP,
			})
			return
		}
		if err != nil {
			logging.Sugar.Warnf("License validation error: %v", err)
		}
	}

	response.Unauthorized(c, "No valid authentication provided")
}

func Login(c *gin.Context) {
	service := user.GetAuthService()

	startTime := time.Now()
	defer func() {
		logging.Sugar.Infof("Login request processed in %v", time.Since(startTime))
	}()

	// 检查是否已经登录
	tokenString := authTokenFromRequest(c)
	if tokenString != "" {
		userInfo, err := service.ValidateToken(tokenString)
		if err == nil && userInfo != nil {
			logging.Sugar.Infof("User '%s' already logged in", userInfo.Username)
			setAuthCookie(c, tokenString)
			response.SuccessWithMessage(c, "Already logged in", GuardLoginResponse{
				Token: tokenString,
				User: userInfoResponse(userInfo),
			})
			return
		}
		if err != nil {
			logging.Sugar.Warnf("Invalid token: %v", err)
		}
	}

	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	trustedDeviceCookie, _ := c.Cookie("flmelody.trusted_device")

	input := &user.LoginInputDTO{
		Username:           req.Username,
		Password:           req.Password,
		TrustedDeviceToken: trustedDeviceCookie,
	}

	output, err := service.GuardLogin(input)
	if err != nil {
		if errors.Is(err, domainError.ErrInvalidCredentials) {
			logging.Sugar.Warnf("Login failed for user '%s': invalid credentials", req.Username)
			threat.GetThreatDetector().RecordFailedLogin(c.ClientIP())
			response.Unauthorized(c, "Invalid username or password")
			return
		}
		if errors.Is(err, domainError.ErrUserDisabled) {
			logging.Sugar.Warnf("Login failed for user '%s': account disabled", req.Username)
			response.Forbidden(c, "Account is disabled")
			return
		}
		logging.Sugar.Errorf("Login failed for user '%s': %v", req.Username, err)
		response.InternalServerError(c, "Login failed, please try again later")
		return
	}

	guardResponse := GuardLoginResponse{
		RequiresTwoFA:  output.RequiresTwoFA,
		ChallengeToken: output.ChallengeToken,
		Token:          output.Token,
		User: &UserInfoResponse{
			ID:       output.User.ID,
			Username: output.User.Username,
		},
	}

	if output.RequiresTwoFA {
		logging.Sugar.Infof("User '%s' requires 2FA verification", req.Username)
		response.SuccessWithMessage(c, "2FA verification required", guardResponse)
	} else {
		setAuthCookie(c, output.Token)
		logging.Sugar.Infof("User '%s' logged in successfully", req.Username)
		response.SuccessWithMessage(c, "Login successful", guardResponse)
	}
}

// AdminLogin wraps Login with an additional admin privilege check.
func AdminLogin(c *gin.Context) {
	service := user.GetAuthService()

	startTime := time.Now()
	defer func() {
		logging.Sugar.Infof("Admin login request processed in %v", time.Since(startTime))
	}()

	// Check if already logged in
	tokenString := authTokenFromRequest(c)
	if tokenString != "" {
		userInfo, err := service.ValidateToken(tokenString)
		if err == nil && userInfo != nil {
			if !userInfo.IsAdmin {
				logging.Sugar.Warnf("Admin login denied for user '%s': not an admin", userInfo.Username)
				response.Forbidden(c, "Admin privileges required")
				return
			}
			logging.Sugar.Infof("Admin user '%s' already logged in", userInfo.Username)
			setAuthCookie(c, tokenString)
			response.SuccessWithMessage(c, "Already logged in", AdminLoginResponse{
				Token: tokenString,
				User: userInfoResponse(userInfo),
			})
			return
		}
	}

	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	trustedDeviceCookie, _ := c.Cookie("flmelody.trusted_device")

	input := &user.LoginInputDTO{
		Username:           req.Username,
		Password:           req.Password,
		TrustedDeviceToken: trustedDeviceCookie,
	}

	output, err := service.AdminLogin(input)
	if err != nil {
		if errors.Is(err, domainError.ErrInvalidCredentials) {
			logging.Sugar.Warnf("Admin login failed for user '%s': invalid credentials", req.Username)
			threat.GetThreatDetector().RecordFailedLogin(c.ClientIP())
			response.Unauthorized(c, "Invalid username or password")
			return
		}
		if errors.Is(err, domainError.ErrUserDisabled) {
			logging.Sugar.Warnf("Admin login failed for user '%s': account disabled", req.Username)
			response.Forbidden(c, "Account is disabled")
			return
		}
		if errors.Is(err, domainError.ErrAdminRequired) {
			logging.Sugar.Warnf("Admin login denied for user '%s': not an admin", req.Username)
			response.Forbidden(c, "Admin privileges required")
			return
		}
		logging.Sugar.Errorf("Admin login failed for user '%s': %v", req.Username, err)
		response.InternalServerError(c, "Login failed, please try again later")
		return
	}

	adminLoginResponse := AdminLoginResponse{
		RequiresTwoFactor: output.RequiresTwoFA,
		ChallengeToken:    output.ChallengeToken,
		Token:             output.Token,
		User: &UserInfoResponse{
			ID:       output.User.ID,
			Username: output.User.Username,
		},
	}

	if output.RequiresTwoFA {
		logging.Sugar.Infof("Admin user '%s' requires 2FA verification", req.Username)
		response.SuccessWithMessage(c, "2FA verification required", adminLoginResponse)
	} else {
		setAuthCookie(c, output.Token)
		logging.Sugar.Infof("Admin user '%s' logged in successfully", req.Username)
		response.SuccessWithMessage(c, "Login successful", adminLoginResponse)
	}
}

// Verify2FA handles the second step of 2FA login
func Verify2FA(c *gin.Context) {
	service := user.GetAuthService()

	var req request.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	output, err := service.Verify2FALogin(&user.TwoFALoginInputDTO{
		ChallengeToken: req.ChallengeToken,
		Code:           req.Code,
		TrustDevice:    req.TrustDevice,
	})
	if err != nil {
		if errors.Is(err, domainError.ErrInvalidCredentials) || errors.Is(err, domainError.ErrTotpInvalidCode) {
			response.Unauthorized(c, "Invalid 2FA code")
			return
		}
		if errors.Is(err, domainError.ErrUserDisabled) {
			response.Forbidden(c, "Account is disabled")
			return
		}
		if errors.Is(err, domainError.ErrTotpNotEnabled) {
			response.BadRequest(c, "2FA is not enabled for this account")
			return
		}
		logging.Sugar.Errorf("2FA verification failed: %v", err)
		response.InternalServerError(c, "Verification failed, please try again later")
		return
	}

	// Set trusted device cookie if token was generated
	if output.TrustedDeviceToken != "" {
		days := config.Get().Security.TrustedDeviceDays
		maxAge := days * 86400
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("flmelody.trusted_device", output.TrustedDeviceToken, maxAge, "/", getAuthCookieDomain(), config.Get().Security.SecureCookies, true)
	}

	loginResponse := LoginResponse{
		Token: output.Token,
		User: &UserInfoResponse{
			ID:       output.User.ID,
			Username: output.User.Username,
		},
	}

	logging.Sugar.Infof("User '%s' completed 2FA verification", output.User.Username)
	setAuthCookie(c, output.Token)
	response.SuccessWithMessage(c, "Login successful", loginResponse)
}

func AdminSession(c *gin.Context) {
	userInfo, ok := currentUserInfo(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}
	response.Success(c, gin.H{
		"user": userInfoResponse(userInfo),
	})
}

func Logout(c *gin.Context) {
	clearAuthCookie(c)
	response.SuccessWithMessage(c, "Logged out", nil)
}

// AdminRecover2FA resets 2FA for an admin user using a config-based recovery key.
func AdminRecover2FA(c *gin.Context) {
	// Check local-only restriction (default: true when not explicitly set).
	// Must satisfy both conditions to be considered a genuine local request:
	//   1. TCP peer (RemoteAddr) is loopback
	//   2. No forwarding headers present (rules out reverse-proxied external traffic)
	if config.Get().Security.AdminRecoveryLocalOnly {
		host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
		ip := net.ParseIP(host)
		proxied := c.GetHeader("X-Forwarded-For") != "" || c.GetHeader("X-Real-IP") != ""
		if ip == nil || !ip.IsLoopback() || proxied {
			response.Forbidden(c, "Recovery is only allowed from localhost")
			return
		}
	}

	var req request.AdminRecover2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	service := user.GetAuthService()
	err := service.RecoverAdmin2FA(req.Username, req.Password, req.RecoveryKey)
	if err != nil {
		if errors.Is(err, domainError.ErrRecoveryDisabled) || errors.Is(err, domainError.ErrRecoveryKeyInvalid) {
			response.Forbidden(c, "Recovery failed")
			return
		}
		if errors.Is(err, domainError.ErrInvalidCredentials) {
			threat.GetThreatDetector().RecordFailedLogin(c.ClientIP())
			response.Unauthorized(c, "Invalid username or password")
			return
		}
		if errors.Is(err, domainError.ErrAdminRequired) {
			response.Forbidden(c, "Admin privileges required")
			return
		}
		if errors.Is(err, domainError.ErrTotpNotEnabled) {
			response.BadRequest(c, "2FA is not enabled for this account")
			return
		}
		logging.Sugar.Errorf("Admin 2FA recovery failed for user '%s': %v", req.Username, err)
		response.InternalServerError(c, "Recovery failed, please try again later")
		return
	}

	logging.Sugar.Infof("Admin 2FA recovered for user '%s'", req.Username)
	response.SuccessWithMessage(c, "2FA has been reset successfully", nil)
}
