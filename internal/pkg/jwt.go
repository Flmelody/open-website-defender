package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"open-website-defender/internal/infrastructure/logging"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	JWTSecret          = []byte("your-secret-key-change-in-production")
	TokenExpirationHrs = 24
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("expired token")
)

type Claims struct {
	Username string `json:"username"`
	UserID   uint   `json:"user_id"`
	Purpose  string `json:"purpose,omitempty"`
	jwt.RegisteredClaims
}

// InitJWTSecret initializes the JWT secret from configuration.
// If the secret is empty, a random 32-byte secret is generated.
func InitJWTSecret(secret string, expirationHrs int) {
	if secret == "" {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			logging.Sugar.Fatal("Failed to generate random JWT secret: ", err)
		}
		JWTSecret = randomBytes
		logging.Sugar.Warn("No JWT secret configured, using randomly generated secret. Tokens will be invalidated on restart.")
	} else {
		JWTSecret = []byte(secret)
	}
	if expirationHrs > 0 {
		TokenExpirationHrs = expirationHrs
	}
}

func GenerateToken(username string, userID uint) (string, error) {
	claims := &Claims{
		Username: username,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(TokenExpirationHrs) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Reject challenge tokens from being used as regular auth tokens
	if claims.Purpose != "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func Generate2FAToken(username string, userID uint) (string, error) {
	claims := &Claims{
		Username: username,
		UserID:   userID,
		Purpose:  "2fa_challenge",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func Parse2FAToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Purpose != "2fa_challenge" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func ValidateToken(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return err == nil
}

// generateRandomHex returns a hex-encoded random string (exported for testing).
func generateRandomHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
