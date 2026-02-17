package waf

import (
	"errors"
	"fmt"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"regexp"
	"sync"
)

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

type compiledRule struct {
	Name    string
	Pattern *regexp.Regexp
	Action  string
}

type WafService struct {
	repo          *repository.WafRuleRepository
	compiledRules []compiledRule
	mu            sync.RWMutex
}

var (
	wafService *WafService
	wafOnce    sync.Once
)

func GetWafService() *WafService {
	wafOnce.Do(func() {
		wafService = &WafService{
			repo: repository.NewWafRuleRepository(database.DB),
		}
	})
	return wafService
}

func (s *WafService) getCompiledRules() ([]compiledRule, error) {
	// Fast path: return cached compiled rules
	s.mu.RLock()
	if s.compiledRules != nil {
		rules := s.compiledRules
		s.mu.RUnlock()
		return rules, nil
	}
	s.mu.RUnlock()

	// Slow path: load from DB and compile
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.compiledRules != nil {
		return s.compiledRules, nil
	}

	dbRules, err := s.repo.FindAllEnabled()
	if err != nil {
		return nil, err
	}

	rules := make([]compiledRule, 0, len(dbRules))
	for _, r := range dbRules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			logging.Sugar.Warnf("Invalid WAF rule pattern '%s' (rule: %s): %v", r.Pattern, r.Name, err)
			continue
		}
		rules = append(rules, compiledRule{Name: r.Name, Pattern: re, Action: r.Action})
	}

	s.compiledRules = rules
	return rules, nil
}

// CheckRequest inspects a request for malicious patterns.
func (s *WafService) CheckRequest(method, path, queryString, userAgent, body string) *WafCheckResult {
	rules, err := s.getCompiledRules()
	if err != nil {
		logging.Sugar.Errorf("Failed to get WAF rules: %v", err)
		return nil
	}

	// Combine all parts to check
	targets := []string{path, queryString, userAgent, body}

	for _, rule := range rules {
		for _, target := range targets {
			if target != "" && rule.Pattern.MatchString(target) {
				return &WafCheckResult{
					Blocked:  rule.Action == "block",
					RuleName: rule.Name,
					Action:   rule.Action,
				}
			}
		}
	}

	return nil
}

func (s *WafService) invalidateCache() {
	s.mu.Lock()
	s.compiledRules = nil
	s.mu.Unlock()
}

func (s *WafService) Create(input *CreateWafRuleDto) (*WafRuleDto, error) {
	if input.Name == "" || input.Pattern == "" || input.Category == "" {
		return nil, errors.New("name, pattern, and category are required")
	}

	// Validate regex
	if _, err := regexp.Compile(input.Pattern); err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	action := input.Action
	if action == "" {
		action = "block"
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	rule := &entity.WafRule{
		Name:     input.Name,
		Pattern:  input.Pattern,
		Category: input.Category,
		Action:   action,
		Enabled:  &enabled,
	}

	if err := s.repo.Create(rule); err != nil {
		return nil, fmt.Errorf("failed to create WAF rule: %w", err)
	}

	s.invalidateCache()

	return &WafRuleDto{
		ID:        rule.ID,
		Name:      rule.Name,
		Pattern:   rule.Pattern,
		Category:  rule.Category,
		Action:    rule.Action,
		Enabled:   derefBool(rule.Enabled),
		CreatedAt: rule.CreatedAt,
	}, nil
}

func (s *WafService) Update(id uint, input *UpdateWafRuleDto) (*WafRuleDto, error) {
	rule, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, errors.New("WAF rule not found")
	}

	if input.Name != "" {
		rule.Name = input.Name
	}
	if input.Pattern != "" {
		if _, err := regexp.Compile(input.Pattern); err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		rule.Pattern = input.Pattern
	}
	if input.Category != "" {
		rule.Category = input.Category
	}
	if input.Action != "" {
		rule.Action = input.Action
	}
	if input.Enabled != nil {
		rule.Enabled = input.Enabled
	}

	if err := s.repo.Update(rule); err != nil {
		return nil, err
	}

	s.invalidateCache()

	return &WafRuleDto{
		ID:        rule.ID,
		Name:      rule.Name,
		Pattern:   rule.Pattern,
		Category:  rule.Category,
		Action:    rule.Action,
		Enabled:   derefBool(rule.Enabled),
		CreatedAt: rule.CreatedAt,
	}, nil
}

func (s *WafService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.invalidateCache()
	return nil
}

func (s *WafService) List(page, size int) ([]*WafRuleDto, int64, error) {
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

	dtos := make([]*WafRuleDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, &WafRuleDto{
			ID:        item.ID,
			Name:      item.Name,
			Pattern:   item.Pattern,
			Category:  item.Category,
			Action:    item.Action,
			Enabled:   derefBool(item.Enabled),
			CreatedAt: item.CreatedAt,
		})
	}
	return dtos, total, nil
}
