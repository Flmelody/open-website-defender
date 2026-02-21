package bot

import (
	"context"
	"errors"
	"fmt"
	"net"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
	"regexp"
	"strings"
	"sync"
	"time"
)

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

type compiledSignature struct {
	Name        string
	Pattern     *regexp.Regexp
	MatchTarget string
	Category    string
	Action      string
}

type BotService struct {
	repo     *repository.BotSignatureRepository
	compiled []compiledSignature
	mu       sync.RWMutex
}

var (
	botService *BotService
	botOnce    sync.Once
)

func GetBotService() *BotService {
	botOnce.Do(func() {
		botService = &BotService{
			repo: repository.NewBotSignatureRepository(database.DB),
		}
		event.Bus().Subscribe(event.BotSignaturesChanged, func(_ event.Event, _ any) {
			botService.invalidateCache()
		})
	})
	return botService
}

func (s *BotService) getCompiledSignatures() ([]compiledSignature, error) {
	s.mu.RLock()
	if s.compiled != nil {
		sigs := s.compiled
		s.mu.RUnlock()
		return sigs, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.compiled != nil {
		return s.compiled, nil
	}

	dbSigs, err := s.repo.FindAllEnabled()
	if err != nil {
		return nil, err
	}

	sigs := make([]compiledSignature, 0, len(dbSigs))
	for _, sig := range dbSigs {
		re, err := regexp.Compile(sig.Pattern)
		if err != nil {
			logging.Sugar.Warnf("Invalid bot signature pattern '%s' (sig: %s): %v", sig.Pattern, sig.Name, err)
			continue
		}
		sigs = append(sigs, compiledSignature{
			Name:        sig.Name,
			Pattern:     re,
			MatchTarget: sig.MatchTarget,
			Category:    sig.Category,
			Action:      sig.Action,
		})
	}

	s.compiled = sigs
	return sigs, nil
}

// CheckRequest checks a request against bot signatures.
func (s *BotService) CheckRequest(ua string, headers map[string]string, clientIP string) *BotCheckResult {
	sigs, err := s.getCompiledSignatures()
	if err != nil {
		logging.Sugar.Errorf("Failed to get bot signatures: %v", err)
		return nil
	}

	for _, sig := range sigs {
		var matched bool
		switch sig.MatchTarget {
		case "ua":
			matched = sig.Pattern.MatchString(ua)
		case "header":
			for _, v := range headers {
				if sig.Pattern.MatchString(v) {
					matched = true
					break
				}
			}
		default:
			matched = sig.Pattern.MatchString(ua)
		}

		if matched {
			result := &BotCheckResult{
				Matched:       true,
				SignatureName: sig.Name,
				Category:      sig.Category,
				Action:        sig.Action,
			}

			// Verify search engine bots via reverse DNS
			if sig.Category == "search_engine" && sig.Action == "allow" {
				result.IsVerified = VerifySearchEngine(ua, clientIP)
				if !result.IsVerified {
					result.Action = "block"
				}
			}

			return result
		}
	}

	return nil
}

// VerifySearchEngine performs reverse DNS verification for search engine bots.
func VerifySearchEngine(ua, clientIP string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := net.DefaultResolver
	names, err := resolver.LookupAddr(ctx, clientIP)
	if err != nil || len(names) == 0 {
		return false
	}

	hostname := strings.TrimSuffix(names[0], ".")

	// Google: *.googlebot.com or *.google.com
	if strings.Contains(strings.ToLower(ua), "googlebot") {
		return strings.HasSuffix(hostname, ".googlebot.com") || strings.HasSuffix(hostname, ".google.com")
	}

	// Bing: *.search.msn.com
	if strings.Contains(strings.ToLower(ua), "bingbot") {
		return strings.HasSuffix(hostname, ".search.msn.com")
	}

	// Yahoo/Slurp: *.crawl.yahoo.net
	if strings.Contains(strings.ToLower(ua), "slurp") {
		return strings.HasSuffix(hostname, ".crawl.yahoo.net")
	}

	// Baidu: *.crawl.baidu.com
	if strings.Contains(strings.ToLower(ua), "baiduspider") {
		return strings.HasSuffix(hostname, ".crawl.baidu.com") || strings.HasSuffix(hostname, ".baidu.jp")
	}

	// Forward DNS verification
	addrs, err := resolver.LookupHost(ctx, hostname)
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if addr == clientIP {
			return true
		}
	}

	return false
}

// DetermineChallenge returns the appropriate challenge type based on threat score.
func DetermineChallenge(threatScore int) string {
	switch {
	case threatScore >= 90:
		return "block"
	case threatScore >= 60:
		return "captcha"
	case threatScore >= 30:
		return "js_challenge"
	case threatScore >= 10:
		return "rate_limit"
	default:
		return "allow"
	}
}

func (s *BotService) invalidateCache() {
	s.mu.Lock()
	s.compiled = nil
	s.mu.Unlock()
}

func (s *BotService) Create(input *CreateBotSignatureDto) (*BotSignatureDto, error) {
	if input.Name == "" || input.Pattern == "" || input.Category == "" {
		return nil, errors.New("name, pattern, and category are required")
	}

	if _, err := regexp.Compile(input.Pattern); err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	matchTarget := input.MatchTarget
	if matchTarget == "" {
		matchTarget = "ua"
	}
	action := input.Action
	if action == "" {
		action = "block"
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	sig := &entity.BotSignature{
		Name:        input.Name,
		Pattern:     input.Pattern,
		MatchTarget: matchTarget,
		Category:    input.Category,
		Action:      action,
		Enabled:     &enabled,
	}

	if err := s.repo.Create(sig); err != nil {
		return nil, fmt.Errorf("failed to create bot signature: %w", err)
	}

	event.Bus().Publish(event.BotSignaturesChanged)

	return sigToDto(sig), nil
}

func (s *BotService) Update(id uint, input *UpdateBotSignatureDto) (*BotSignatureDto, error) {
	sig, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if sig == nil {
		return nil, errors.New("bot signature not found")
	}

	if input.Name != "" {
		sig.Name = input.Name
	}
	if input.Pattern != "" {
		if _, err := regexp.Compile(input.Pattern); err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		sig.Pattern = input.Pattern
	}
	if input.MatchTarget != "" {
		sig.MatchTarget = input.MatchTarget
	}
	if input.Category != "" {
		sig.Category = input.Category
	}
	if input.Action != "" {
		sig.Action = input.Action
	}
	if input.Enabled != nil {
		sig.Enabled = input.Enabled
	}

	if err := s.repo.Update(sig); err != nil {
		return nil, err
	}

	event.Bus().Publish(event.BotSignaturesChanged)

	return sigToDto(sig), nil
}

func (s *BotService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	event.Bus().Publish(event.BotSignaturesChanged)
	return nil
}

func (s *BotService) List(page, size int) ([]*BotSignatureDto, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	offset := (page - 1) * size
	list, total, err := s.repo.List(size, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*BotSignatureDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, sigToDto(item))
	}
	return dtos, total, nil
}

func sigToDto(s *entity.BotSignature) *BotSignatureDto {
	return &BotSignatureDto{
		ID:          s.ID,
		Name:        s.Name,
		Pattern:     s.Pattern,
		MatchTarget: s.MatchTarget,
		Category:    s.Category,
		Action:      s.Action,
		Enabled:     derefBool(s.Enabled),
		CreatedAt:   s.CreatedAt,
	}
}
