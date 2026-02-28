package handler

import (
	"fmt"
	"net/http"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/config"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"time"
	"open-website-defender/internal/usecase/bot"
	"open-website-defender/internal/usecase/system"

	"github.com/gin-gonic/gin"
)

// GenerateCaptcha generates a new built-in image CAPTCHA.
func GenerateCaptcha(c *gin.Context) {
	id, image, err := bot.GenerateBuiltinCaptcha()
	if err != nil {
		logging.Sugar.Errorf("Failed to generate captcha: %v", err)
		response.InternalServerError(c, "failed to generate captcha")
		return
	}
	response.Success(c, gin.H{"id": id, "image": image})
}

// VerifyCaptcha handles CAPTCHA verification and sets a signed pass cookie.
func VerifyCaptcha(c *gin.Context) {
	settings, err := system.GetSystemService().GetSettings()
	if err != nil || settings == nil {
		response.InternalServerError(c, "failed to load settings")
		return
	}

	provider := settings.CaptchaProvider

	var verified bool

	switch provider {
	case "builtin":
		captchaID := c.PostForm("captcha_id")
		captchaAnswer := c.PostForm("captcha_answer")
		if captchaID == "" || captchaAnswer == "" {
			response.BadRequest(c, "captcha_id and captcha_answer required")
			return
		}
		verified = bot.VerifyBuiltinCaptcha(captchaID, captchaAnswer)

	case "turnstile":
		secret := settings.CaptchaSecretKey
		if secret == "" {
			response.InternalServerError(c, "CAPTCHA not configured")
			return
		}

		token := c.PostForm("cf-turnstile-response")
		if token == "" {
			var body struct {
				Token string `json:"token"`
			}
			if err := c.ShouldBindJSON(&body); err == nil {
				token = body.Token
			}
		}
		if token == "" {
			response.BadRequest(c, "CAPTCHA token required")
			return
		}

		verified, err = bot.VerifyTurnstile(secret, token, c.ClientIP())
		if err != nil {
			logging.Sugar.Errorf("CAPTCHA verification error: %v", err)
			response.InternalServerError(c, "CAPTCHA verification failed")
			return
		}

	default:
		response.BadRequest(c, "unsupported CAPTCHA provider")
		return
	}

	if !verified {
		response.Forbidden(c, "CAPTCHA verification failed")
		return
	}

	// Set signed pass cookie: IP|timestamp~HMAC
	cookieTTL := settings.CaptchaCookieTTL
	if cookieTTL <= 0 {
		cookieTTL = 86400
	}

	passData := fmt.Sprintf("%s|%d", c.ClientIP(), time.Now().Unix())
	passSig := pkg.SignCookieData(passData)
	passValue := passData + "~" + passSig

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("owd_captcha_pass", passValue, cookieTTL, "/", "", config.Get().Security.SecureCookies, true)

	// Redirect back to original URL or return success
	redirectURL := c.Query("redirect")
	if redirectURL != "" {
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	response.Success(c, gin.H{"verified": true})
}
