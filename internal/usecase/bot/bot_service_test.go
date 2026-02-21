package bot

import (
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/logging"
	"os"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	// Initialize a no-op logger to avoid nil pointer panics in tests
	logging.Logger = zap.NewNop()
	logging.Sugar = logging.Logger.Sugar()
	os.Exit(m.Run())
}

// setupTestDB creates an in-memory SQLite database with BotSignature table.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	err = db.AutoMigrate(&entity.BotSignature{})
	if err != nil {
		t.Fatalf("Failed to migrate BotSignature: %v", err)
	}

	return db
}

func boolPtr(b bool) *bool {
	return &b
}

// newTestBotService creates a BotService backed by the given test DB,
// bypassing the singleton.
func newTestBotService(db *gorm.DB) *BotService {
	return &BotService{
		repo: repository.NewBotSignatureRepository(db),
	}
}

// --- Bot detection tests ---

func TestCheckRequest_BotDetection_SQLMap(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	sig := &entity.BotSignature{
		Name:        "SQLMap Scanner",
		Pattern:     `(?i)sqlmap`,
		MatchTarget: "ua",
		Category:    "malicious",
		Action:      "block",
		Enabled:     boolPtr(true),
	}
	if err := db.Create(sig).Error; err != nil {
		t.Fatalf("Failed to seed signature: %v", err)
	}

	result := svc.CheckRequest("sqlmap/1.4.7#stable (http://sqlmap.org)", nil, "1.2.3.4")

	if result == nil {
		t.Fatal("Expected SQLMap UA to be detected, got nil result")
	}
	if !result.Matched {
		t.Error("Expected Matched=true, got false")
	}
	if result.SignatureName != "SQLMap Scanner" {
		t.Errorf("Expected SignatureName='SQLMap Scanner', got %q", result.SignatureName)
	}
	if result.Category != "malicious" {
		t.Errorf("Expected Category='malicious', got %q", result.Category)
	}
	if result.Action != "block" {
		t.Errorf("Expected Action='block', got %q", result.Action)
	}
}

func TestCheckRequest_SearchEngineDetection(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	sig := &entity.BotSignature{
		Name:        "Googlebot",
		Pattern:     `(?i)googlebot`,
		MatchTarget: "ua",
		Category:    "search_engine",
		Action:      "allow",
		Enabled:     boolPtr(true),
	}
	if err := db.Create(sig).Error; err != nil {
		t.Fatalf("Failed to seed signature: %v", err)
	}

	// Use a non-Google IP so DNS verification fails, which changes action to "block"
	// but the pattern still matches
	result := svc.CheckRequest(
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		nil,
		"1.2.3.4",
	)

	if result == nil {
		t.Fatal("Expected Googlebot UA to be detected, got nil result")
	}
	if !result.Matched {
		t.Error("Expected Matched=true, got false")
	}
	if result.SignatureName != "Googlebot" {
		t.Errorf("Expected SignatureName='Googlebot', got %q", result.SignatureName)
	}
	if result.Category != "search_engine" {
		t.Errorf("Expected Category='search_engine', got %q", result.Category)
	}
	// With a non-Google IP, DNS verification fails, so IsVerified=false and action becomes "block"
	if result.IsVerified {
		t.Error("Expected IsVerified=false for non-Google IP")
	}
	if result.Action != "block" {
		t.Errorf("Expected Action='block' (failed DNS verification), got %q", result.Action)
	}
}

func TestCheckRequest_NoMatchNormalUA(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	// Seed bot signatures
	sigs := []entity.BotSignature{
		{Name: "SQLMap", Pattern: `(?i)sqlmap`, MatchTarget: "ua", Category: "malicious", Action: "block", Enabled: boolPtr(true)},
		{Name: "Nikto", Pattern: `(?i)nikto`, MatchTarget: "ua", Category: "malicious", Action: "block", Enabled: boolPtr(true)},
		{Name: "DirBuster", Pattern: `(?i)dirbuster`, MatchTarget: "ua", Category: "malicious", Action: "block", Enabled: boolPtr(true)},
	}
	for i := range sigs {
		if err := db.Create(&sigs[i]).Error; err != nil {
			t.Fatalf("Failed to seed signature %q: %v", sigs[i].Name, err)
		}
	}

	normalUAs := []struct {
		name string
		ua   string
	}{
		{"Chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		{"Firefox", "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0"},
		{"Safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15"},
		{"curl", "curl/8.4.0"},
		{"axios", "axios/1.6.0"},
		{"Postman", "PostmanRuntime/7.36.0"},
	}

	for _, tt := range normalUAs {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CheckRequest(tt.ua, nil, "10.0.0.1")
			if result != nil {
				t.Errorf("Normal UA %q was flagged by signature %q", tt.ua, result.SignatureName)
			}
		})
	}
}

func TestCheckRequest_ChallengeAction(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	sig := &entity.BotSignature{
		Name:        "Suspicious Crawler",
		Pattern:     `(?i)crawler\s*bot`,
		MatchTarget: "ua",
		Category:    "malicious",
		Action:      "challenge",
		Enabled:     boolPtr(true),
	}
	if err := db.Create(sig).Error; err != nil {
		t.Fatalf("Failed to seed signature: %v", err)
	}

	result := svc.CheckRequest("CrawlerBot/1.0", nil, "1.2.3.4")

	if result == nil {
		t.Fatal("Expected challenge action result, got nil")
	}
	if result.Action != "challenge" {
		t.Errorf("Expected Action='challenge', got %q", result.Action)
	}
	if !result.Matched {
		t.Error("Expected Matched=true")
	}
}

func TestCheckRequest_MonitorAction(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	sig := &entity.BotSignature{
		Name:        "Monitor Bot",
		Pattern:     `(?i)monitorbot`,
		MatchTarget: "ua",
		Category:    "good_bot",
		Action:      "monitor",
		Enabled:     boolPtr(true),
	}
	if err := db.Create(sig).Error; err != nil {
		t.Fatalf("Failed to seed signature: %v", err)
	}

	result := svc.CheckRequest("MonitorBot/2.0", nil, "10.0.0.1")

	if result == nil {
		t.Fatal("Expected monitor action result, got nil")
	}
	if result.Action != "monitor" {
		t.Errorf("Expected Action='monitor', got %q", result.Action)
	}
}

func TestCheckRequest_HeaderMatching(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	sig := &entity.BotSignature{
		Name:        "Suspicious Header",
		Pattern:     `(?i)evil-bot-framework`,
		MatchTarget: "header",
		Category:    "malicious",
		Action:      "block",
		Enabled:     boolPtr(true),
	}
	if err := db.Create(sig).Error; err != nil {
		t.Fatalf("Failed to seed signature: %v", err)
	}

	headers := map[string]string{
		"X-Custom-Header": "evil-bot-framework/1.0",
	}

	result := svc.CheckRequest("Mozilla/5.0", headers, "1.2.3.4")

	if result == nil {
		t.Fatal("Expected header-based detection, got nil")
	}
	if !result.Matched {
		t.Error("Expected Matched=true")
	}
	if result.SignatureName != "Suspicious Header" {
		t.Errorf("Expected SignatureName='Suspicious Header', got %q", result.SignatureName)
	}
}

func TestCheckRequest_DisabledSignature(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	sig := &entity.BotSignature{
		Name:        "Disabled SQLMap",
		Pattern:     `(?i)sqlmap`,
		MatchTarget: "ua",
		Category:    "malicious",
		Action:      "block",
		Enabled:     boolPtr(false),
	}
	if err := db.Create(sig).Error; err != nil {
		t.Fatalf("Failed to seed signature: %v", err)
	}

	result := svc.CheckRequest("sqlmap/1.4.7", nil, "1.2.3.4")

	if result != nil {
		t.Errorf("Disabled signature should not match, but got SignatureName=%q", result.SignatureName)
	}
}

func TestCheckRequest_NoSignatures(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	// No signatures seeded
	result := svc.CheckRequest("sqlmap/1.4.7", nil, "1.2.3.4")

	if result != nil {
		t.Errorf("Expected nil result when no signatures exist, got SignatureName=%q", result.SignatureName)
	}
}

// --- DetermineChallenge tests ---

func TestDetermineChallenge(t *testing.T) {
	tests := []struct {
		name        string
		threatScore int
		expected    string
	}{
		{"score 100 - block", 100, "block"},
		{"score 95 - block", 95, "block"},
		{"score 90 - block", 90, "block"},
		{"score 89 - captcha", 89, "captcha"},
		{"score 75 - captcha", 75, "captcha"},
		{"score 60 - captcha", 60, "captcha"},
		{"score 59 - js_challenge", 59, "js_challenge"},
		{"score 45 - js_challenge", 45, "js_challenge"},
		{"score 30 - js_challenge", 30, "js_challenge"},
		{"score 29 - rate_limit", 29, "rate_limit"},
		{"score 15 - rate_limit", 15, "rate_limit"},
		{"score 10 - rate_limit", 10, "rate_limit"},
		{"score 9 - allow", 9, "allow"},
		{"score 5 - allow", 5, "allow"},
		{"score 0 - allow", 0, "allow"},
		{"score negative - allow", -1, "allow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineChallenge(tt.threatScore)
			if result != tt.expected {
				t.Errorf("DetermineChallenge(%d) = %q, want %q", tt.threatScore, result, tt.expected)
			}
		})
	}
}

// --- CRUD tests ---

func TestCreate_ValidSignature(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	enabled := true
	dto, err := svc.Create(&CreateBotSignatureDto{
		Name:        "Test Bot",
		Pattern:     `(?i)testbot`,
		MatchTarget: "ua",
		Category:    "malicious",
		Action:      "block",
		Enabled:     &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dto.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if dto.Name != "Test Bot" {
		t.Errorf("Expected Name='Test Bot', got %q", dto.Name)
	}
	if dto.Pattern != `(?i)testbot` {
		t.Errorf("Unexpected Pattern: %q", dto.Pattern)
	}
	if dto.MatchTarget != "ua" {
		t.Errorf("Expected MatchTarget='ua', got %q", dto.MatchTarget)
	}
	if dto.Action != "block" {
		t.Errorf("Expected Action='block', got %q", dto.Action)
	}
}

func TestCreate_InvalidRegex(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	enabled := true
	_, err := svc.Create(&CreateBotSignatureDto{
		Name:     "Bad Regex",
		Pattern:  `(?i)(unclosed group`,
		Category: "malicious",
		Enabled:  &enabled,
	})
	if err == nil {
		t.Error("Expected error for invalid regex pattern, got nil")
	}
}

func TestCreate_MissingRequiredFields(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	_, err := svc.Create(&CreateBotSignatureDto{
		Name:     "",
		Pattern:  `test`,
		Category: "malicious",
	})
	if err == nil {
		t.Error("Expected error for missing name, got nil")
	}

	_, err = svc.Create(&CreateBotSignatureDto{
		Name:     "Test",
		Pattern:  "",
		Category: "malicious",
	})
	if err == nil {
		t.Error("Expected error for missing pattern, got nil")
	}

	_, err = svc.Create(&CreateBotSignatureDto{
		Name:     "Test",
		Pattern:  `test`,
		Category: "",
	})
	if err == nil {
		t.Error("Expected error for missing category, got nil")
	}
}

func TestCreate_DefaultAction(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	dto, err := svc.Create(&CreateBotSignatureDto{
		Name:     "Default Action Bot",
		Pattern:  `test`,
		Category: "malicious",
		Action:   "", // should default to "block"
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dto.Action != "block" {
		t.Errorf("Expected default Action='block', got %q", dto.Action)
	}
}

func TestCreate_DefaultMatchTarget(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	dto, err := svc.Create(&CreateBotSignatureDto{
		Name:        "Default Target Bot",
		Pattern:     `test`,
		Category:    "malicious",
		MatchTarget: "", // should default to "ua"
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dto.MatchTarget != "ua" {
		t.Errorf("Expected default MatchTarget='ua', got %q", dto.MatchTarget)
	}
}

func TestUpdate_Integration(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	enabled := true
	created, err := svc.Create(&CreateBotSignatureDto{
		Name:        "Original Bot",
		Pattern:     `original`,
		MatchTarget: "ua",
		Category:    "malicious",
		Action:      "block",
		Enabled:     &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	disabled := false
	updated, err := svc.Update(created.ID, &UpdateBotSignatureDto{
		Name:    "Updated Bot",
		Pattern: `updated`,
		Action:  "challenge",
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated Bot" {
		t.Errorf("Expected Name='Updated Bot', got %q", updated.Name)
	}
	if updated.Pattern != "updated" {
		t.Errorf("Expected Pattern='updated', got %q", updated.Pattern)
	}
	if updated.Action != "challenge" {
		t.Errorf("Expected Action='challenge', got %q", updated.Action)
	}
	if updated.Enabled != false {
		t.Error("Expected Enabled=false")
	}
}

func TestUpdate_InvalidRegex(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	enabled := true
	created, err := svc.Create(&CreateBotSignatureDto{
		Name:     "Signature",
		Pattern:  `test`,
		Category: "malicious",
		Enabled:  &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = svc.Update(created.ID, &UpdateBotSignatureDto{
		Pattern: `(?i)(unclosed`,
	})
	if err == nil {
		t.Error("Expected error for invalid regex in update, got nil")
	}
}

func TestUpdate_NonExistentSignature(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	_, err := svc.Update(99999, &UpdateBotSignatureDto{
		Name: "Does Not Exist",
	})
	if err == nil {
		t.Error("Expected error updating non-existent signature, got nil")
	}
}

func TestDelete_Integration(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	enabled := true
	created, err := svc.Create(&CreateBotSignatureDto{
		Name:     "To Delete",
		Pattern:  `delete_me`,
		Category: "malicious",
		Enabled:  &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = svc.Delete(created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, total, err := svc.List(1, 100)
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected 0 signatures after deletion, got %d", total)
	}
}

func TestList_Pagination(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	// Seed 5 signatures
	for i := 0; i < 5; i++ {
		enabled := true
		_, err := svc.Create(&CreateBotSignatureDto{
			Name:     "Bot " + string(rune('A'+i)),
			Pattern:  `bot` + string(rune('0'+i)),
			Category: "malicious",
			Enabled:  &enabled,
		})
		if err != nil {
			t.Fatalf("Failed to create signature %d: %v", i, err)
		}
	}

	// List first page
	dtos, total, err := svc.List(1, 3)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected total=5, got %d", total)
	}
	if len(dtos) != 3 {
		t.Errorf("Expected 3 items on page 1, got %d", len(dtos))
	}

	// List second page
	dtos2, total2, err := svc.List(2, 3)
	if err != nil {
		t.Fatalf("List page 2 failed: %v", err)
	}
	if total2 != 5 {
		t.Errorf("Expected total=5, got %d", total2)
	}
	if len(dtos2) != 2 {
		t.Errorf("Expected 2 items on page 2, got %d", len(dtos2))
	}
}

func TestList_InvalidPagination(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestBotService(db)

	// Seed 3 signatures
	for i := 0; i < 3; i++ {
		enabled := true
		_, err := svc.Create(&CreateBotSignatureDto{
			Name:     "Bot " + string(rune('A'+i)),
			Pattern:  `bot` + string(rune('0'+i)),
			Category: "malicious",
			Enabled:  &enabled,
		})
		if err != nil {
			t.Fatalf("Failed to create signature %d: %v", i, err)
		}
	}

	// Negative page and size should be corrected
	dtos, _, err := svc.List(-1, -1)
	if err != nil {
		t.Fatalf("List with invalid pagination failed: %v", err)
	}
	// page defaults to 1, size defaults to 10
	if len(dtos) != 3 {
		t.Errorf("Expected 3 items (all signatures fit in size=10), got %d", len(dtos))
	}
}
