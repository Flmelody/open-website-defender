package handler

import (
	"errors"
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	domainError "open-website-defender/internal/domain/error"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"open-website-defender/internal/usecase/iplist"
	"open-website-defender/internal/usecase/license"
	"open-website-defender/internal/usecase/system"
	"open-website-defender/internal/usecase/user"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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

type LoginResponse struct {
	Token string            `json:"token"`
	User  *UserInfoResponse `json:"user"`
}

type UserInfoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// AuthMiddleware is used as a middleware to protect routes.
func AuthMiddleware(c *gin.Context) {
	// Check Token
	service := user.GetAuthService()
	authHeader := c.GetHeader("Defender-Authorization")
	if len(authHeader) == 0 {
		response.Unauthorized(c, "No authentication token provided")
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
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
	clientIP := pkg.GetClientIP(c)

	// 2. Check Blacklist
	blackListService := iplist.GetIpBlackListService()
	if blackListItem, _ := blackListService.FindByIP(clientIP); blackListItem != nil {
		logging.Sugar.Warnf("Access denied for blacklisted IP: %s", clientIP)
		response.Forbidden(c, "Access denied")
		return
	}

	// 3. Check Whitelist
	whiteListService := iplist.GetIpWhiteListService()
	if whiteListItem, _ := whiteListService.FindByIP(clientIP); whiteListItem != nil {
		logging.Sugar.Infof("Access granted for whitelisted IP: %s", clientIP)
		response.Success(c, gin.H{
			"message": "Access granted via IP whitelist",
			"ip":      clientIP,
		})
		return
	}

	// 4. Check Token
	service := user.GetAuthService()
	clientToken, err := c.Cookie("flmelody.token")
	if err != nil {
		clientToken = c.GetHeader("Defender-Authorization")
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

	requestedDomain := getRequestedDomain(c)

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

	// 5. Check Git Token header (configurable)
	systemService := system.GetSystemService()
	gitHeaderName, licenseHeaderName := systemService.GetHeaderNames()

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
				logging.Sugar.Infof("Access granted via git token for user: %s", parts[0])
				response.Success(c, userInfo)
				return
			}
			logging.Sugar.Warnf("Git token validation failed for user '%s': %v", parts[0], err)
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
	authHeader := c.GetHeader("Defender-Authorization")
	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userInfo, err := service.ValidateToken(tokenString)
		if err == nil && userInfo != nil {
			logging.Sugar.Infof("User '%s' already logged in", userInfo.Username)
			response.SuccessWithMessage(c, "Already logged in", LoginResponse{
				Token: tokenString,
				User: &UserInfoResponse{
					ID:       userInfo.ID,
					Username: userInfo.Username,
				},
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

	input := &user.LoginInputDTO{
		Username: req.Username,
		Password: req.Password,
	}

	output, err := service.Login(input)
	if err != nil {
		if errors.Is(err, domainError.ErrInvalidCredentials) {
			logging.Sugar.Warnf("Login failed for user '%s': invalid credentials", req.Username)
			response.Unauthorized(c, "Invalid username or password")
			return
		}

		logging.Sugar.Errorf("Login failed for user '%s': %v", req.Username, err)
		response.InternalServerError(c, "Login failed, please try again later")
		return
	}

	loginResponse := LoginResponse{
		Token: output.Token,
		User: &UserInfoResponse{
			ID:       output.User.ID,
			Username: output.User.Username,
		},
	}

	logging.Sugar.Infof("User '%s' logged in successfully", req.Username)
	response.SuccessWithMessage(c, "Login successful", loginResponse)
}
