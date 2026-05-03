package repository

import (
	"testing"
	"time"

	"castellum/internal/domain/entity"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestOAuthAuthorizationCodeRepositoryMarkUsedIsSingleUse(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&entity.OAuthAuthorizationCode{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	code := &entity.OAuthAuthorizationCode{
		Code:        "code-1",
		ClientID:    "client-1",
		UserID:      1,
		RedirectURI: "https://client.example/callback",
		ExpiresAt:   time.Now().Add(time.Minute),
	}
	if err := db.Create(code).Error; err != nil {
		t.Fatalf("create code: %v", err)
	}

	repo := NewOAuthAuthorizationCodeRepository(db)
	claimed, err := repo.MarkUsed(code.ID)
	if err != nil {
		t.Fatalf("first MarkUsed returned error: %v", err)
	}
	if !claimed {
		t.Fatal("first MarkUsed should claim the code")
	}

	claimed, err = repo.MarkUsed(code.ID)
	if err != nil {
		t.Fatalf("second MarkUsed returned error: %v", err)
	}
	if claimed {
		t.Fatal("second MarkUsed should not claim an already used code")
	}
}
