package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"

	"open-website-defender/internal/infrastructure/logging"
)

var (
	rsaPrivateKey *rsa.PrivateKey
	rsaOnce       sync.Once
)

// InitRSAKey initializes the RSA key pair for OIDC token signing.
// If keyPath is provided and exists, the key is loaded from disk.
// Otherwise, a new 2048-bit key is generated in memory.
func InitRSAKey(keyPath string) {
	rsaOnce.Do(func() {
		if keyPath != "" {
			key, err := loadRSAKeyFromFile(keyPath)
			if err == nil {
				rsaPrivateKey = key
				logging.Sugar.Info("RSA private key loaded from file")
				return
			}
			logging.Sugar.Warnf("Failed to load RSA key from %s: %v, generating new key", keyPath, err)
		}

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			logging.Sugar.Fatalf("Failed to generate RSA key: %v", err)
		}
		rsaPrivateKey = key
		logging.Sugar.Warn("Using auto-generated RSA key. ID tokens will be invalidated on restart. Set oauth.rsa-private-key-path for persistence.")
	})
}

// GetRSAPrivateKey returns the RSA private key used for signing.
func GetRSAPrivateKey() *rsa.PrivateKey {
	return rsaPrivateKey
}

// GetRSAPublicKey returns the RSA public key used for verification.
func GetRSAPublicKey() *rsa.PublicKey {
	if rsaPrivateKey == nil {
		return nil
	}
	return &rsaPrivateKey.PublicKey
}

func loadRSAKeyFromFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, os.ErrInvalid
	}

	// Try PKCS8 first, then PKCS1
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
