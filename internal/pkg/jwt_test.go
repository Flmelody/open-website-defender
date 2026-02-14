package pkg

import (
	"open-website-defender/internal/infrastructure/logging"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize a no-op logger so that InitJWTSecret (which calls logging.Sugar.Warn)
	// does not panic on nil Sugar. We use zap.NewNop() to avoid file I/O issues in tests.
	logging.Logger = zap.NewNop()
	logging.Sugar = logging.Logger.Sugar()
	os.Exit(m.Run())
}

func TestInitJWTSecret_WithCustomSecret(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	customSecret := "my-custom-jwt-secret-key"
	InitJWTSecret(customSecret, 48)

	if string(JWTSecret) != customSecret {
		t.Errorf("JWTSecret = %q, want %q", string(JWTSecret), customSecret)
	}
	if TokenExpirationHrs != 48 {
		t.Errorf("TokenExpirationHrs = %d, want 48", TokenExpirationHrs)
	}
}

func TestInitJWTSecret_EmptySecret_GeneratesRandom(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	InitJWTSecret("", 12)

	if len(JWTSecret) != 32 {
		t.Errorf("Random JWTSecret length = %d, want 32", len(JWTSecret))
	}
	if TokenExpirationHrs != 12 {
		t.Errorf("TokenExpirationHrs = %d, want 12", TokenExpirationHrs)
	}
}

func TestInitJWTSecret_EmptySecret_DifferentEachTime(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	InitJWTSecret("", 1)
	secret1 := make([]byte, len(JWTSecret))
	copy(secret1, JWTSecret)

	InitJWTSecret("", 1)
	secret2 := make([]byte, len(JWTSecret))
	copy(secret2, JWTSecret)

	match := true
	for i := range secret1 {
		if secret1[i] != secret2[i] {
			match = false
			break
		}
	}
	if match {
		t.Error("Two random secrets should not be identical")
	}
}

func TestInitJWTSecret_ZeroExpiration_KeepsDefault(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	TokenExpirationHrs = 24
	InitJWTSecret("test-secret", 0)

	if TokenExpirationHrs != 24 {
		t.Errorf("TokenExpirationHrs should remain 24 when 0 is passed, got %d", TokenExpirationHrs)
	}
}

func TestInitJWTSecret_NegativeExpiration_KeepsDefault(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	TokenExpirationHrs = 24
	InitJWTSecret("test-secret", -5)

	if TokenExpirationHrs != 24 {
		t.Errorf("TokenExpirationHrs should remain 24 when negative is passed, got %d", TokenExpirationHrs)
	}
}

func TestGenerateToken(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	JWTSecret = []byte("test-secret-for-generate")
	TokenExpirationHrs = 24

	tokenString, err := GenerateToken("admin", 1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if tokenString == "" {
		t.Error("GenerateToken returned empty token")
	}
}

func TestGenerateToken_DifferentUsersProduceDifferentTokens(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	JWTSecret = []byte("test-secret")
	TokenExpirationHrs = 24

	token1, err := GenerateToken("user1", 1)
	if err != nil {
		t.Fatalf("GenerateToken for user1 failed: %v", err)
	}

	token2, err := GenerateToken("user2", 2)
	if err != nil {
		t.Fatalf("GenerateToken for user2 failed: %v", err)
	}

	if token1 == token2 {
		t.Error("Different users should produce different tokens")
	}
}

func TestParseToken_ValidToken(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	JWTSecret = []byte("test-secret-for-parse")
	TokenExpirationHrs = 24

	username := "testuser"
	var userID uint = 42

	tokenString, err := GenerateToken(username, userID)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Username != username {
		t.Errorf("claims.Username = %q, want %q", claims.Username, username)
	}
	if claims.UserID != userID {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, userID)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	origSecret := JWTSecret
	defer func() {
		JWTSecret = origSecret
	}()

	JWTSecret = []byte("test-secret")

	_, err := ParseToken("this.is.not.a.valid.token")
	if err == nil {
		t.Error("ParseToken should return error for invalid token")
	}
	if err != ErrInvalidToken {
		t.Errorf("ParseToken error = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_EmptyToken(t *testing.T) {
	_, err := ParseToken("")
	if err == nil {
		t.Error("ParseToken should return error for empty token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	// Generate token with one secret
	JWTSecret = []byte("secret-one")
	TokenExpirationHrs = 24
	tokenString, err := GenerateToken("user", 1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Try to parse with different secret
	JWTSecret = []byte("secret-two")
	_, err = ParseToken(tokenString)
	if err == nil {
		t.Error("ParseToken should fail with wrong secret")
	}
	if err != ErrInvalidToken {
		t.Errorf("ParseToken error = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	JWTSecret = []byte("test-secret-expired")

	// Create a token that is already expired by building claims manually
	claims := &Claims{
		Username: "expired-user",
		UserID:   1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ParseToken(tokenString)
	if err == nil {
		t.Error("ParseToken should return error for expired token")
	}
	if err != ErrExpiredToken {
		t.Errorf("ParseToken error = %v, want ErrExpiredToken", err)
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	JWTSecret = []byte("test-secret-validate")
	TokenExpirationHrs = 24

	tokenString, err := GenerateToken("user", 1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if !ValidateToken(tokenString) {
		t.Error("ValidateToken should return true for valid token")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	if ValidateToken("invalid-token-string") {
		t.Error("ValidateToken should return false for invalid token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	if ValidateToken("") {
		t.Error("ValidateToken should return false for empty token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	origSecret := JWTSecret
	defer func() {
		JWTSecret = origSecret
	}()

	JWTSecret = []byte("test-secret-validate-expired")

	claims := &Claims{
		Username: "expired-user",
		UserID:   1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	if ValidateToken(tokenString) {
		t.Error("ValidateToken should return false for expired token")
	}
}

func TestGenerateAndParseToken_RoundTrip(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	JWTSecret = []byte("roundtrip-secret")
	TokenExpirationHrs = 24

	testCases := []struct {
		username string
		userID   uint
	}{
		{"admin", 1},
		{"user", 100},
		{"test-user@example.com", 999},
		{"", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			tokenString, err := GenerateToken(tc.username, tc.userID)
			if err != nil {
				t.Fatalf("GenerateToken(%q, %d) failed: %v", tc.username, tc.userID, err)
			}

			claims, err := ParseToken(tokenString)
			if err != nil {
				t.Fatalf("ParseToken failed: %v", err)
			}

			if claims.Username != tc.username {
				t.Errorf("Username = %q, want %q", claims.Username, tc.username)
			}
			if claims.UserID != tc.userID {
				t.Errorf("UserID = %d, want %d", claims.UserID, tc.userID)
			}

			// Verify expiration is set correctly (within a small margin)
			expectedExpiry := time.Now().Add(time.Duration(TokenExpirationHrs) * time.Hour)
			actualExpiry := claims.ExpiresAt.Time
			diff := expectedExpiry.Sub(actualExpiry)
			if diff < -5*time.Second || diff > 5*time.Second {
				t.Errorf("Token expiry off by %v", diff)
			}
		})
	}
}

func TestInitJWTSecret_ThenGenerateAndParse(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	// Test the full flow: init with custom secret, generate, parse
	InitJWTSecret("integration-test-secret", 8)

	tokenString, err := GenerateToken("integration-user", 5)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Username != "integration-user" {
		t.Errorf("Username = %q, want %q", claims.Username, "integration-user")
	}
	if claims.UserID != 5 {
		t.Errorf("UserID = %d, want 5", claims.UserID)
	}

	if !ValidateToken(tokenString) {
		t.Error("ValidateToken should return true")
	}
}

func TestInitJWTSecret_RandomSecret_TokensStillWork(t *testing.T) {
	origSecret := JWTSecret
	origExpiration := TokenExpirationHrs
	defer func() {
		JWTSecret = origSecret
		TokenExpirationHrs = origExpiration
	}()

	// Init with empty secret (random generation)
	InitJWTSecret("", 24)

	tokenString, err := GenerateToken("random-secret-user", 10)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Username != "random-secret-user" {
		t.Errorf("Username = %q, want %q", claims.Username, "random-secret-user")
	}

	if !ValidateToken(tokenString) {
		t.Error("ValidateToken should return true for token generated with same secret")
	}
}
