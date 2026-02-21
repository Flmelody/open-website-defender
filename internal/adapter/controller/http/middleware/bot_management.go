package middleware

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"open-website-defender/internal/usecase/bot"
	"open-website-defender/internal/usecase/system"
	"open-website-defender/internal/usecase/threat"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// RenderCaptchaPage renders the CAPTCHA challenge HTML page using the embedded template.
// Exported so that handler or test routes can reuse the same template.
func RenderCaptchaPage(c *gin.Context, provider, siteKey, redirectURL string, statusCode int) {
	rootPath := viper.GetString("ROOT_PATH")
	if rootPath == "" {
		rootPath = "/wall"
	}

	verifyAction := fmt.Sprintf("%s/captcha/verify?redirect=%s", rootPath, url.QueryEscape(redirectURL))
	generateURL := fmt.Sprintf("%s/captcha/generate", rootPath)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")

	var buf bytes.Buffer
	_ = captchaTemplate.Execute(&buf, map[string]string{
		"Provider":     provider,
		"SiteKey":      siteKey,
		"VerifyAction": verifyAction,
		"GenerateURL":  generateURL,
	})
	c.String(statusCode, buf.String())
}

//go:embed templates/captcha.html
var captchaTemplateHTML string
var captchaTemplate = template.Must(template.New("captcha").Parse(captchaTemplateHTML))

// captchaConfigured returns true if a captcha provider is properly configured.
func captchaConfigured(settings *system.SystemSettingsDTO) bool {
	switch settings.CaptchaProvider {
	case "builtin":
		return true
	case "turnstile":
		return settings.CaptchaSiteKey != "" && settings.CaptchaSecretKey != ""
	default:
		return false
	}
}

// applyChallenge sets the appropriate challenge flag based on escalation settings.
// With escalation: threat score determines JS PoW / captcha / block.
// Without escalation: use configured captcha provider, fallback to JS PoW if not configured.
func applyChallenge(c *gin.Context, settings *system.SystemSettingsDTO) {
	if settings.ChallengeEscalation {
		score := threat.GetThreatDetector().GetThreatScore(c.ClientIP())
		decision := bot.DetermineChallenge(score)
		logging.Sugar.Infof("Challenge escalation: ip=%s, score=%d, decision=%s", c.ClientIP(), score, decision)
		switch decision {
		case "block":
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		case "captcha":
			c.Set("bot_captcha", true)
			c.Next()
			return
		default:
			c.Set("waf_challenge", true)
			c.Next()
			return
		}
	}
	// Challenge escalation disabled — use captcha if configured, otherwise JS PoW
	if captchaConfigured(settings) {
		c.Set("bot_captcha", true)
	} else {
		c.Set("waf_challenge", true)
	}
	c.Next()
}

// BotManagement returns a middleware that checks requests against bot signatures.
// When enabled, ALL requests are challenged by default; signatures can override
// the action (allow known bots, block malicious ones, etc.).
func BotManagement() gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := system.GetSystemService().GetSettings()
		if err != nil || settings == nil || !settings.BotManagementEnabled {
			c.Next()
			return
		}

		service := bot.GetBotService()

		ua := c.GetHeader("User-Agent")

		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}

		result := service.CheckRequest(ua, headers, c.ClientIP())
		if result == nil {
			// No signature matched — apply default challenge for all unknown traffic
			applyChallenge(c, settings)
			return
		}

		logging.Sugar.Infof("Bot detected: %s (category=%s, action=%s, ip=%s, verified=%v)",
			result.SignatureName, result.Category, result.Action, c.ClientIP(), result.IsVerified)

		c.Set("bot_detected", true)
		c.Set("bot_category", result.Category)
		c.Set("bot_signature", result.SignatureName)

		switch result.Action {
		case "allow":
			c.Next()
			return
		case "block":
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		case "challenge":
			applyChallenge(c, settings)
			return
		case "monitor":
			// Log only, let it through
			c.Next()
			return
		default:
			c.Next()
			return
		}
	}
}

// CaptchaPage serves a CAPTCHA challenge page when bot_captcha flag is set.
// It reads CAPTCHA configuration dynamically from DB settings.
// If CAPTCHA is not configured, it degrades to JS PoW challenge.
func CaptchaPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("bot_captcha") {
			c.Next()
			return
		}

		// Skip non-GET requests (POST to /captcha/verify must not be intercepted)
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Check for valid signed captcha pass cookie
		passCookie, err := c.Cookie("owd_captcha_pass")
		if err == nil && passCookie != "" {
			parts := strings.SplitN(passCookie, "~", 2)
			if len(parts) == 2 {
				data := parts[0]
				sig := parts[1]
				if pkg.VerifyCookieSignature(data, sig) {
					cookieParts := strings.SplitN(data, "|", 2)
					if len(cookieParts) == 2 && cookieParts[0] == c.ClientIP() {
						c.Next()
						return
					}
				}
			}
		}

		settings, sErr := system.GetSystemService().GetSettings()
		if sErr != nil || settings == nil {
			logging.Sugar.Info("CaptchaPage: settings unavailable, degrading to JS challenge")
			c.Set("waf_challenge", true)
			c.Next()
			return
		}

		provider := settings.CaptchaProvider

		// For builtin provider, no site key needed. For turnstile, require site key.
		if provider == "turnstile" && settings.CaptchaSiteKey == "" {
			logging.Sugar.Info("CaptchaPage: Turnstile not configured (no site key), degrading to JS challenge")
			c.Set("waf_challenge", true)
			c.Next()
			return
		}

		if provider != "builtin" && provider != "turnstile" {
			logging.Sugar.Info("CaptchaPage: unknown provider, degrading to JS challenge")
			c.Set("waf_challenge", true)
			c.Next()
			return
		}

		currentURL := c.Request.URL.String()
		RenderCaptchaPage(c, provider, settings.CaptchaSiteKey, currentURL, http.StatusForbidden)
		c.Abort()
	}
}
