package user

import (
	"errors"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	domainError "castellum/internal/domain/error"
	"castellum/internal/infrastructure/cache"
)

func TestValidateTotpCodeOnceRejectsReplay(t *testing.T) {
	if err := cache.Store().Clear(); err != nil {
		t.Fatalf("clear cache: %v", err)
	}

	secret := "JBSWY3DPEHPK3PXP"
	now := time.Unix(1700000000, 0).UTC()
	code := mustGenerateTotpCode(t, secret, now)

	if err := validateTotpCodeOnce(1, code, secret, now); err != nil {
		t.Fatalf("first TOTP validation failed: %v", err)
	}

	if err := validateTotpCodeOnce(1, code, secret, now); !errors.Is(err, domainError.ErrTotpInvalidCode) {
		t.Fatalf("expected replay to be rejected as invalid TOTP code, got %v", err)
	}
}

func TestValidateTotpCodeOnceScopesReplayByUser(t *testing.T) {
	if err := cache.Store().Clear(); err != nil {
		t.Fatalf("clear cache: %v", err)
	}

	secret := "JBSWY3DPEHPK3PXP"
	now := time.Unix(1700000000, 0).UTC()
	code := mustGenerateTotpCode(t, secret, now)

	if err := validateTotpCodeOnce(1, code, secret, now); err != nil {
		t.Fatalf("first user TOTP validation failed: %v", err)
	}

	if err := validateTotpCodeOnce(2, code, secret, now); err != nil {
		t.Fatalf("same code for a different user should not be treated as replay: %v", err)
	}
}

func TestValidateTotpCodeOnceInvalidCodeDoesNotPoisonReplayCache(t *testing.T) {
	if err := cache.Store().Clear(); err != nil {
		t.Fatalf("clear cache: %v", err)
	}

	secret := "JBSWY3DPEHPK3PXP"
	now := time.Unix(1700000000, 0).UTC()
	code := mustGenerateTotpCode(t, secret, now)

	if err := validateTotpCodeOnce(1, "000000", secret, now); !errors.Is(err, domainError.ErrTotpInvalidCode) {
		t.Fatalf("expected invalid TOTP code, got %v", err)
	}

	if err := validateTotpCodeOnce(1, code, secret, now); err != nil {
		t.Fatalf("valid code should still pass after invalid attempt: %v", err)
	}
}

func mustGenerateTotpCode(t *testing.T, secret string, at time.Time) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secret, at, totp.ValidateOpts{
		Period:    totpPeriodSeconds,
		Skew:      totpAllowedSkew,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate TOTP code: %v", err)
	}
	return code
}
