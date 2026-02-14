package pkg

import (
	"strings"
	"testing"
)

func TestMD5Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "simple string",
			input:    "hello",
			expected: "5d41402abc4b2a76b9719d911017c592",
		},
		{
			name:     "password string",
			input:    "defender",
			expected: "975ecb719692fa2bc7255b0c2dd2f3a4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MD5Hash(tt.input)
			if result != tt.expected {
				t.Errorf("MD5Hash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMD5Hash_Deterministic(t *testing.T) {
	input := "test-password-123"
	hash1 := MD5Hash(input)
	hash2 := MD5Hash(input)
	if hash1 != hash2 {
		t.Errorf("MD5Hash is not deterministic: %q != %q", hash1, hash2)
	}
}

func TestMD5Hash_DifferentInputs(t *testing.T) {
	hash1 := MD5Hash("password1")
	hash2 := MD5Hash("password2")
	if hash1 == hash2 {
		t.Error("MD5Hash produced same hash for different inputs")
	}
}

func TestHashPassword(t *testing.T) {
	password := "my-secure-password"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword(%q) returned error: %v", password, err)
	}

	if hashed == "" {
		t.Error("HashPassword returned empty string")
	}

	if hashed == password {
		t.Error("HashPassword returned the plaintext password")
	}

	// bcrypt hashes start with "$2a$" or "$2b$"
	if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") {
		t.Errorf("HashPassword result does not look like a bcrypt hash: %q", hashed)
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	password := "same-password"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword call failed: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword call failed: %v", err)
	}

	// bcrypt uses random salt, so same password should produce different hashes
	if hash1 == hash2 {
		t.Error("HashPassword produced identical hashes for the same password (bcrypt should use random salt)")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hashed, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword with empty string returned error: %v", err)
	}
	if hashed == "" {
		t.Error("HashPassword with empty string returned empty hash")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "test-password"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		expected       bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashed,
			password:       password,
			expected:       true,
		},
		{
			name:           "wrong password",
			hashedPassword: hashed,
			password:       "wrong-password",
			expected:       false,
		},
		{
			name:           "empty password against hash",
			hashedPassword: hashed,
			password:       "",
			expected:       false,
		},
		{
			name:           "invalid hash format",
			hashedPassword: "not-a-bcrypt-hash",
			password:       password,
			expected:       false,
		},
		{
			name:           "empty hash",
			hashedPassword: "",
			password:       password,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.hashedPassword, tt.password)
			if result != tt.expected {
				t.Errorf("CheckPassword(%q, %q) = %v, want %v",
					tt.hashedPassword, tt.password, result, tt.expected)
			}
		})
	}
}

func TestCheckPassword_MultipleHashes(t *testing.T) {
	password := "shared-password"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Both different hashes should validate the same password
	if !CheckPassword(hash1, password) {
		t.Error("CheckPassword failed for hash1 with correct password")
	}
	if !CheckPassword(hash2, password) {
		t.Error("CheckPassword failed for hash2 with correct password")
	}
}

func TestIsMD5Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid MD5 lowercase",
			input:    "d41d8cd98f00b204e9800998ecf8427e",
			expected: true,
		},
		{
			name:     "valid MD5 uppercase",
			input:    "D41D8CD98F00B204E9800998ECF8427E",
			expected: true,
		},
		{
			name:     "valid MD5 mixed case",
			input:    "d41D8cd98F00b204E9800998ecF8427e",
			expected: true,
		},
		{
			name:     "actual MD5 output",
			input:    MD5Hash("hello"),
			expected: true,
		},
		{
			name:     "too short",
			input:    "d41d8cd98f00b204",
			expected: false,
		},
		{
			name:     "too long",
			input:    "d41d8cd98f00b204e9800998ecf8427e00",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "non-hex characters",
			input:    "g41d8cd98f00b204e9800998ecf8427e",
			expected: false,
		},
		{
			name:     "bcrypt hash",
			input:    "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234",
			expected: false,
		},
		{
			name:     "32 chars with special chars",
			input:    "d41d8cd98f00b204e9800998ecf8427!",
			expected: false,
		},
		{
			name:     "plain text 32 chars",
			input:    "this is exactly 32 characters!!!", // 32 chars but not hex
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMD5Hash(tt.input)
			if result != tt.expected {
				t.Errorf("IsMD5Hash(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsMD5Hash_BcryptHashNotMD5(t *testing.T) {
	password := "test"
	bcryptHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if IsMD5Hash(bcryptHash) {
		t.Error("IsMD5Hash should return false for bcrypt hashes")
	}
}

func TestMigrationPath_MD5ToBcrypt(t *testing.T) {
	// Simulate the migration scenario:
	// 1. Old password stored as MD5
	// 2. User provides plaintext password
	// 3. Verify with MD5, then upgrade to bcrypt

	plainPassword := "defender"
	md5Hash := MD5Hash(plainPassword)

	// Step 1: Verify old hash is recognized as MD5
	if !IsMD5Hash(md5Hash) {
		t.Fatal("MD5Hash output should be recognized as MD5")
	}

	// Step 2: Verify MD5 match
	if md5Hash != MD5Hash(plainPassword) {
		t.Fatal("MD5 verification failed")
	}

	// Step 3: Create bcrypt hash (the upgrade)
	bcryptHash, err := HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("Failed to create bcrypt hash: %v", err)
	}

	// Step 4: New hash should not be detected as MD5
	if IsMD5Hash(bcryptHash) {
		t.Error("bcrypt hash should not be detected as MD5")
	}

	// Step 5: New hash should validate correctly
	if !CheckPassword(bcryptHash, plainPassword) {
		t.Error("bcrypt hash should validate with correct password")
	}
}
