package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"net/http"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"open-website-defender/internal/usecase/iplist"
	"open-website-defender/internal/usecase/system"
	"open-website-defender/internal/usecase/threat"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

//go:embed templates/js_challenge.html
var challengeTemplateHTML string
var challengeTemplate = template.Must(template.New("challenge").Parse(challengeTemplateHTML))

// jsChallengeSkipRoutes holds route paths that should bypass JS Challenge.
// Register paths during route setup (before server starts) via JSChallengeSkipRoute.
var jsChallengeSkipRoutes = make(map[string]bool)

// JSChallengeSkipRoute marks route paths to bypass JS Challenge middleware.
func JSChallengeSkipRoute(paths ...string) {
	for _, p := range paths {
		jsChallengeSkipRoutes[p] = true
	}
}

func signChallenge(data string) string {
	return pkg.SignCookieData(data)
}

func verifySignature(data, signature string) bool {
	return pkg.VerifyCookieSignature(data, signature)
}

// JSChallenge returns a middleware that serves a JavaScript Proof-of-Work challenge
// to clients before allowing them through. This helps filter out bots that don't
// execute JavaScript.
func getJSChallengeSettings() (enabled bool, mode string, difficulty int) {
	settings, err := system.GetSystemService().GetSettings()
	if err == nil && settings != nil {
		return settings.JSChallengeEnabled, settings.JSChallengeMode, settings.JSChallengeDifficulty
	}
	// Fallback to config file
	return viper.GetBool("js-challenge.enabled"), viper.GetString("js-challenge.mode"), viper.GetInt("js-challenge.difficulty")
}

func JSChallenge() gin.HandlerFunc {
	return func(c *gin.Context) {
		enabled, mode, cfgDifficulty := getJSChallengeSettings()

		// Check if upstream middleware (WAF/BotManagement) flagged this request for challenge
		forcedChallenge := c.GetBool("waf_challenge")

		// If not forced by upstream, check own enabled/mode settings
		if !forcedChallenge {
			if !enabled || mode == "off" || mode == "" {
				c.Next()
				return
			}
		}

		// Skip non-GET requests (PoW only works for browser navigation)
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Only challenge browser navigation requests (Accept contains text/html)
		accept := c.GetHeader("Accept")
		if !strings.Contains(accept, "text/html") {
			c.Next()
			return
		}

		// Skip routes explicitly marked to bypass JS Challenge (e.g. nginx auth subrequest)
		if jsChallengeSkipRoutes[c.FullPath()] || challengeSkipRoutes[c.FullPath()] {
			c.Next()
			return
		}

		if settings, err := system.GetSystemService().GetSettings(); err == nil && settings != nil {
			if (settings.GitTokenHeader != "" && c.GetHeader(settings.GitTokenHeader) != "") ||
				(settings.LicenseHeader != "" && c.GetHeader(settings.LicenseHeader) != "") {
				c.Next()
				return
			}
		}

		// Skip whitelisted IPs
		if wl, _ := iplist.GetIpWhiteListService().FindByIP(c.ClientIP()); wl != nil {
			c.Next()
			return
		}

		// Check if client already has a valid pass cookie
		passCookie, err := c.Cookie("_defender_pow")
		if err == nil && passCookie != "" {
			parts := strings.SplitN(passCookie, "~", 2)
			if len(parts) == 2 {
				data := parts[0]
				sig := parts[1]
				if verifySignature(data, sig) {
					// Check IP encoded in data: ip|timestamp
					cookieParts := strings.SplitN(data, "|", 2)
					if len(cookieParts) == 2 && cookieParts[0] == c.ClientIP() {
						// Valid pass cookie
						c.Next()
						return
					}
					logging.Sugar.Debugf("JS Challenge: pass cookie IP mismatch, cookie=%s clientIP=%s", data, c.ClientIP())
				} else {
					logging.Sugar.Debugf("JS Challenge: pass cookie signature invalid")
				}
			}
		} else {
			logging.Sugar.Debugf("JS Challenge: no _defender_pow cookie for %s %s", c.Request.Method, c.Request.URL.Path)
		}

		// In "suspicious" mode, only challenge IPs with elevated threat score
		// Skip this check if forced by upstream (WAF/BotManagement already decided to challenge)
		if !forcedChallenge && mode == "suspicious" {
			td := threat.GetThreatDetector()
			threshold := 10
			score := td.GetThreatScore(c.ClientIP())
			if score < threshold {
				c.Next()
				return
			}
			logging.Sugar.Infof("JS Challenge triggered for IP %s (threat score: %d)", c.ClientIP(), score)
		}

		// Check if client is submitting a solution
		solutionCookie, err := c.Cookie("_defender_challenge")
		if err == nil && solutionCookie != "" {
			// Solution format: nonce:solution:signature
			parts := strings.SplitN(solutionCookie, ":", 3)
			if len(parts) == 3 {
				nonce := parts[0]
				solution := parts[1]
				sig := parts[2]

				if verifySignature(nonce, sig) {
					logging.Sugar.Debugf("JS Challenge: solution signature valid, verifying PoW")
					difficulty := cfgDifficulty
					if difficulty <= 0 {
						difficulty = 4
					}

					hash := sha256.Sum256([]byte(nonce + solution))
					hashHex := hex.EncodeToString(hash[:])
					prefix := strings.Repeat("0", difficulty)

					if strings.HasPrefix(hashHex, prefix) {
						logging.Sugar.Infof("JS Challenge: PoW verified for %s, setting pass cookie", c.ClientIP())
						cookieTTL := viper.GetInt("js-challenge.cookie-ttl")
						if cookieTTL <= 0 {
							cookieTTL = 86400
						}

						passData := fmt.Sprintf("%s|%d", c.ClientIP(), time.Now().Unix())
						passSig := signChallenge(passData)
						passValue := passData + "~" + passSig

						c.SetCookie("_defender_pow", passValue, cookieTTL, "/", "", false, true)
						// Clear challenge cookie
						c.SetCookie("_defender_challenge", "", -1, "/", "", false, false)

						// Redirect to same URL to proceed normally
						c.Redirect(http.StatusFound, c.Request.URL.String())
						c.Abort()
						return
					}
				}
			}
		}

		logging.Sugar.Debugf("JS Challenge: serving challenge page for %s %s", c.Request.Method, c.Request.URL.Path)
		// Serve the challenge page
		difficulty := cfgDifficulty
		if difficulty <= 0 {
			difficulty = 4
		}

		nonce := generateNonce()
		sig := signChallenge(nonce)

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.String(http.StatusOK, challengeHTML(nonce, sig, difficulty))
		c.Abort()
	}
}

func generateNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func challengeHTML(nonce, signature string, difficulty int) string {
	rootPath := viper.GetString("ROOT_PATH")
	if rootPath == "" {
		rootPath = "/wall"
	}
	adminPath := viper.GetString("ADMIN_PATH")
	if adminPath == "" {
		adminPath = "/admin"
	}

	prefix := strings.Repeat("0", difficulty)
	var buf bytes.Buffer
	_ = challengeTemplate.Execute(&buf, map[string]string{
		"Nonce":      nonce,
		"Signature":  signature,
		"Prefix":     prefix,
		"FaviconURL": fmt.Sprintf("%s%s/favicon.ico", rootPath, adminPath),
	})
	return buf.String()
}
