package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mojocn/base64Captcha"
)

type captchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes,omitempty"`
}

var builtinCaptchaStore = base64Captcha.DefaultMemStore

// GenerateBuiltinCaptcha creates a new image CAPTCHA and returns its ID and base64 PNG.
func GenerateBuiltinCaptcha() (string, string, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, builtinCaptchaStore)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate captcha: %w", err)
	}
	return id, b64s, nil
}

// VerifyBuiltinCaptcha checks the user's answer against the stored captcha.
func VerifyBuiltinCaptcha(id, answer string) bool {
	return builtinCaptchaStore.Verify(id, answer, true)
}

// VerifyTurnstile verifies a Cloudflare Turnstile response token.
func VerifyTurnstile(secret, token, remoteIP string) (bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
		"secret":   {secret},
		"response": {token},
		"remoteip": {remoteIP},
	})
	if err != nil {
		return false, fmt.Errorf("Turnstile verification request failed: %w", err)
	}
	defer resp.Body.Close()

	var result captchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("Turnstile response decode failed: %w", err)
	}

	return result.Success, nil
}
