package pkg

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/spf13/viper"
)

var (
	cookieSecret     []byte
	cookieSecretOnce sync.Once
)

// GetCookieSecret returns the shared cookie signing secret.
// It reads from config "js-challenge.cookie-secret" or generates a random 32-byte key.
func GetCookieSecret() []byte {
	cookieSecretOnce.Do(func() {
		secret := viper.GetString("js-challenge.cookie-secret")
		if secret != "" {
			cookieSecret = []byte(secret)
			return
		}
		cookieSecret = make([]byte, 32)
		_, _ = rand.Read(cookieSecret)
	})
	return cookieSecret
}

// SignCookieData signs data with HMAC-SHA256 using the shared cookie secret.
func SignCookieData(data string) string {
	mac := hmac.New(sha256.New, GetCookieSecret())
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyCookieSignature verifies that the signature matches the data.
func VerifyCookieSignature(data, signature string) bool {
	expected := SignCookieData(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}
