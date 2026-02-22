package waf

import (
	"errors"
	"fmt"
	"open-website-defender/internal/adapter/repository"
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/system"
	"open-website-defender/internal/usecase/waf/semantic"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

type compiledRule struct {
	ID          uint
	Name        string
	Pattern     *regexp.Regexp // nil for non-regex operators
	RawPattern  string
	Operator    string
	Target      string
	Action      string
	Category    string
	Priority    int
	RedirectURL string
	RateLimit   int
}

type compiledExclusion struct {
	RuleID   uint
	Pattern  *regexp.Regexp // nil for non-regex operators
	Path     string
	Operator string
}

type WafService struct {
	repo               *repository.WafRuleRepository
	exclusionRepo      *repository.WafExclusionRepository
	compiledRules      []compiledRule
	compiledExclusions []compiledExclusion
	mu                 sync.RWMutex
}

var (
	wafService *WafService
	wafOnce    sync.Once
)

func GetWafService() *WafService {
	wafOnce.Do(func() {
		wafService = &WafService{
			repo:          repository.NewWafRuleRepository(database.DB),
			exclusionRepo: repository.NewWafExclusionRepository(database.DB),
		}
		event.Bus().Subscribe(event.WafRulesChanged, func(_ event.Event, _ any) {
			wafService.invalidateCache()
		})
	})
	return wafService
}

func (s *WafService) getCompiledRules() ([]compiledRule, error) {
	s.mu.RLock()
	if s.compiledRules != nil {
		rules := s.compiledRules
		s.mu.RUnlock()
		return rules, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.compiledRules != nil {
		return s.compiledRules, nil
	}

	dbRules, err := s.repo.FindAllEnabled()
	if err != nil {
		return nil, err
	}

	rules := make([]compiledRule, 0, len(dbRules))
	for _, r := range dbRules {
		cr := compiledRule{
			ID:          r.ID,
			Name:        r.Name,
			RawPattern:  r.Pattern,
			Operator:    r.Operator,
			Target:      r.Target,
			Action:      r.Action,
			Category:    r.Category,
			Priority:    r.Priority,
			RedirectURL: r.RedirectURL,
			RateLimit:   r.RateLimit,
		}
		if cr.Operator == "" {
			cr.Operator = "regex"
		}
		if cr.Target == "" {
			cr.Target = "all"
		}
		if cr.Operator == "regex" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				logging.Sugar.Warnf("Invalid WAF rule pattern '%s' (rule: %s): %v", r.Pattern, r.Name, err)
				continue
			}
			cr.Pattern = re
		}
		rules = append(rules, cr)
	}

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority < rules[j].Priority
		}
		return rules[i].ID < rules[j].ID
	})

	s.compiledRules = rules

	// Also load exclusions
	if s.exclusionRepo != nil {
		dbExclusions, err := s.exclusionRepo.FindAllEnabled()
		if err != nil {
			logging.Sugar.Warnf("Failed to load WAF exclusions: %v", err)
		} else {
			exclusions := make([]compiledExclusion, 0, len(dbExclusions))
			for _, e := range dbExclusions {
				ce := compiledExclusion{
					RuleID:   e.RuleID,
					Path:     e.Path,
					Operator: e.Operator,
				}
				if ce.Operator == "" {
					ce.Operator = "prefix"
				}
				if ce.Operator == "regex" {
					re, err := regexp.Compile(e.Path)
					if err != nil {
						logging.Sugar.Warnf("Invalid WAF exclusion pattern '%s': %v", e.Path, err)
						continue
					}
					ce.Pattern = re
				}
				exclusions = append(exclusions, ce)
			}
			s.compiledExclusions = exclusions
		}
	}

	return rules, nil
}

func (s *WafService) getExclusions() []compiledExclusion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.compiledExclusions
}

// matchOperator checks if the value matches the pattern using the given operator.
func matchOperator(operator, pattern, value string) bool {
	if value == "" {
		return false
	}
	switch operator {
	case "regex":
		// Should not be called for regex; handled separately with compiled pattern
		return false
	case "contains":
		return strings.Contains(strings.ToLower(value), strings.ToLower(pattern))
	case "prefix":
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(pattern))
	case "suffix":
		return strings.HasSuffix(strings.ToLower(value), strings.ToLower(pattern))
	case "equals":
		return strings.EqualFold(value, pattern)
	case "gt":
		v, err1 := strconv.ParseFloat(value, 64)
		p, err2 := strconv.ParseFloat(pattern, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return v > p
	case "lt":
		v, err1 := strconv.ParseFloat(value, 64)
		p, err2 := strconv.ParseFloat(pattern, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return v < p
	default:
		return false
	}
}

// matchValue checks a single value against a compiled rule.
func matchValue(rule *compiledRule, value string) bool {
	if value == "" {
		return false
	}
	if rule.Operator == "regex" {
		return rule.Pattern != nil && rule.Pattern.MatchString(value)
	}
	return matchOperator(rule.Operator, rule.RawPattern, value)
}

// getTargetValues returns the values to check based on the rule's target.
func getTargetValues(ctx *RequestContext, target string) []string {
	switch target {
	case "url":
		return []string{ctx.Path}
	case "query":
		return []string{ctx.Query}
	case "headers":
		vals := make([]string, 0, len(ctx.Headers))
		for _, v := range ctx.Headers {
			vals = append(vals, v)
		}
		return vals
	case "body":
		return []string{ctx.Body}
	case "cookies":
		vals := make([]string, 0, len(ctx.Cookies))
		for _, v := range ctx.Cookies {
			vals = append(vals, v)
		}
		return vals
	case "all":
		targets := []string{ctx.Path, ctx.Query, ctx.UA, ctx.Body}
		for _, v := range ctx.Headers {
			targets = append(targets, v)
		}
		for _, v := range ctx.Cookies {
			targets = append(targets, v)
		}
		return targets
	default:
		return []string{ctx.Path, ctx.Query, ctx.UA, ctx.Body}
	}
}

// isExcluded checks if the request path matches any exclusion for the given rule.
func isExcluded(exclusions []compiledExclusion, ruleID uint, path string) bool {
	for _, e := range exclusions {
		if e.RuleID != 0 && e.RuleID != ruleID {
			continue
		}
		switch e.Operator {
		case "exact":
			if path == e.Path {
				return true
			}
		case "prefix":
			if strings.HasPrefix(path, e.Path) {
				return true
			}
		case "regex":
			if e.Pattern != nil && e.Pattern.MatchString(path) {
				return true
			}
		}
	}
	return false
}

// CheckRequest inspects a request for malicious patterns.
// This is the new API using RequestContext.
func (s *WafService) CheckRequest(method, path, queryString, userAgent, body string) *WafCheckResult {
	ctx := &RequestContext{
		Method: method,
		Path:   path,
		Query:  queryString,
		UA:     userAgent,
		Body:   body,
	}
	return s.CheckRequestContext(ctx)
}

// CheckRequestContext inspects a request using the full RequestContext.
func (s *WafService) CheckRequestContext(ctx *RequestContext) *WafCheckResult {
	rules, err := s.getCompiledRules()
	if err != nil {
		logging.Sugar.Errorf("Failed to get WAF rules: %v", err)
		return nil
	}

	exclusions := s.getExclusions()

	semanticEnabled := isSemanticAnalysisEnabled()

	// Track values already checked by semantic analysis during regex confirmation,
	// so the independent fallback doesn't re-check them.
	var sqliChecked, xssChecked map[string]struct{}
	if semanticEnabled {
		sqliChecked = make(map[string]struct{})
		xssChecked = make(map[string]struct{})
	}

	for _, rule := range rules {
		// Skip response-only targets in request checking
		if rule.Target == "response_body" || rule.Target == "response_headers" {
			continue
		}

		// Check exclusions
		if isExcluded(exclusions, rule.ID, ctx.Path) {
			continue
		}

		targets := getTargetValues(ctx, rule.Target)
		for _, target := range targets {
			if matchValue(&rule, target) {
				result := &WafCheckResult{
					Blocked:     rule.Action == "block",
					RuleName:    rule.Name,
					Action:      rule.Action,
					RedirectURL: rule.RedirectURL,
					RateLimit:   rule.RateLimit,
				}

				// Semantic analysis confirmation for sqli/xss categories
				if semanticEnabled && (rule.Category == "sqli" || rule.Category == "xss") {
					confirmed := false
					if rule.Category == "sqli" {
						sqliChecked[target] = struct{}{}
						isSQLi, fp := semantic.IsSQLi(target)
						confirmed = isSQLi
						result.SemanticFingerprint = fp
					} else if rule.Category == "xss" {
						xssChecked[target] = struct{}{}
						confirmed = semantic.IsXSS(target)
					}
					result.SemanticConfirmed = confirmed

					// If semantic analysis doesn't confirm,
					// treat as false positive and skip this rule match
					if !confirmed {
						continue
					}
				}

				return result
			}
		}
	}

	// Independent semantic detection fallback:
	// When semantic analysis is enabled and no regex rule matched,
	// scan all request fields for SQLi/XSS using semantic analysis alone.
	// Skip values already checked during regex confirmation phase.
	if semanticEnabled {
		allTargets := []string{ctx.Path, ctx.Query, ctx.UA, ctx.Body}
		for _, v := range ctx.Headers {
			allTargets = append(allTargets, v)
		}
		for _, v := range ctx.Cookies {
			allTargets = append(allTargets, v)
		}
		for _, target := range allTargets {
			if target == "" {
				continue
			}
			// Check SQLi
			if _, checked := sqliChecked[target]; !checked {
				if isSQLi, fp := semantic.IsSQLi(target); isSQLi {
					return &WafCheckResult{
						Blocked:             true,
						RuleName:            "Semantic SQLi Detection",
						Action:              "block",
						SemanticConfirmed:   true,
						SemanticFingerprint: fp,
					}
				}
			}
			// Check XSS
			if _, checked := xssChecked[target]; !checked {
				if semantic.IsXSS(target) {
					return &WafCheckResult{
						Blocked:           true,
						RuleName:          "Semantic XSS Detection",
						Action:            "block",
						SemanticConfirmed: true,
					}
				}
			}
		}
	}

	return nil
}

// isSemanticAnalysisEnabled checks if semantic analysis is enabled via system settings.
func isSemanticAnalysisEnabled() (enabled bool) {
	defer func() {
		if r := recover(); r != nil {
			enabled = false
		}
	}()
	settings, err := system.GetSystemService().GetSettings()
	if err != nil || settings == nil {
		return false
	}
	return settings.SemanticAnalysisEnabled
}

func (s *WafService) invalidateCache() {
	s.mu.Lock()
	s.compiledRules = nil
	s.compiledExclusions = nil
	s.mu.Unlock()
}

func (s *WafService) Create(input *CreateWafRuleDto) (*WafRuleDto, error) {
	if input.Name == "" || input.Pattern == "" || input.Category == "" {
		return nil, errors.New("name, pattern, and category are required")
	}

	operator := input.Operator
	if operator == "" {
		operator = "regex"
	}

	// Validate regex if operator is regex
	if operator == "regex" {
		if _, err := regexp.Compile(input.Pattern); err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	action := input.Action
	if action == "" {
		action = "block"
	}

	target := input.Target
	if target == "" {
		target = "all"
	}

	priority := input.Priority
	if priority == 0 {
		priority = 100
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	rule := &entity.WafRule{
		Name:        input.Name,
		Pattern:     input.Pattern,
		Category:    input.Category,
		Action:      action,
		Operator:    operator,
		Target:      target,
		Priority:    priority,
		GroupName:   input.GroupName,
		RedirectURL: input.RedirectURL,
		RateLimit:   input.RateLimit,
		Description: input.Description,
		Enabled:     &enabled,
	}

	if err := s.repo.Create(rule); err != nil {
		return nil, fmt.Errorf("failed to create WAF rule: %w", err)
	}

	event.Bus().Publish(event.WafRulesChanged)

	return ruleToDto(rule), nil
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
		operator := rule.Operator
		if input.Operator != "" {
			operator = input.Operator
		}
		if operator == "" || operator == "regex" {
			if _, err := regexp.Compile(input.Pattern); err != nil {
				return nil, fmt.Errorf("invalid regex pattern: %w", err)
			}
		}
		rule.Pattern = input.Pattern
	}
	if input.Category != "" {
		rule.Category = input.Category
	}
	if input.Action != "" {
		rule.Action = input.Action
	}
	if input.Operator != "" {
		rule.Operator = input.Operator
	}
	if input.Target != "" {
		rule.Target = input.Target
	}
	if input.Priority != nil {
		rule.Priority = *input.Priority
	}
	if input.GroupName != nil {
		rule.GroupName = *input.GroupName
	}
	if input.RedirectURL != nil {
		rule.RedirectURL = *input.RedirectURL
	}
	if input.RateLimit != nil {
		rule.RateLimit = *input.RateLimit
	}
	if input.Description != nil {
		rule.Description = *input.Description
	}
	if input.Enabled != nil {
		rule.Enabled = input.Enabled
	}

	if err := s.repo.Update(rule); err != nil {
		return nil, err
	}

	event.Bus().Publish(event.WafRulesChanged)

	return ruleToDto(rule), nil
}

func (s *WafService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	event.Bus().Publish(event.WafRulesChanged)
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
		dtos = append(dtos, ruleToDto(item))
	}
	return dtos, total, nil
}

func (s *WafService) BatchEnableGroup(groupName string, enabled bool) error {
	if err := s.repo.UpdateGroupEnabled(groupName, enabled); err != nil {
		return err
	}
	event.Bus().Publish(event.WafRulesChanged)
	return nil
}

// --- Exclusion CRUD ---

func (s *WafService) CreateExclusion(input *CreateWafExclusionDto) (*WafExclusionDto, error) {
	if input.Path == "" {
		return nil, errors.New("path is required")
	}

	operator := input.Operator
	if operator == "" {
		operator = "prefix"
	}

	if operator == "regex" {
		if _, err := regexp.Compile(input.Path); err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	exclusion := &entity.WafExclusion{
		RuleID:   input.RuleID,
		Path:     input.Path,
		Operator: operator,
		Enabled:  &enabled,
	}

	if err := s.exclusionRepo.Create(exclusion); err != nil {
		return nil, fmt.Errorf("failed to create WAF exclusion: %w", err)
	}

	event.Bus().Publish(event.WafRulesChanged)

	return exclusionToDto(exclusion), nil
}

func (s *WafService) DeleteExclusion(id uint) error {
	if err := s.exclusionRepo.Delete(id); err != nil {
		return err
	}
	event.Bus().Publish(event.WafRulesChanged)
	return nil
}

func (s *WafService) ListExclusions(page, size int) ([]*WafExclusionDto, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	offset := (page - 1) * size
	list, total, err := s.exclusionRepo.List(size, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*WafExclusionDto, 0, len(list))
	for _, item := range list {
		dtos = append(dtos, exclusionToDto(item))
	}
	return dtos, total, nil
}

func ruleToDto(r *entity.WafRule) *WafRuleDto {
	operator := r.Operator
	if operator == "" {
		operator = "regex"
	}
	target := r.Target
	if target == "" {
		target = "all"
	}
	return &WafRuleDto{
		ID:          r.ID,
		Name:        r.Name,
		Pattern:     r.Pattern,
		Category:    r.Category,
		Action:      r.Action,
		Operator:    operator,
		Target:      target,
		Priority:    r.Priority,
		GroupName:   r.GroupName,
		RedirectURL: r.RedirectURL,
		RateLimit:   r.RateLimit,
		Description: r.Description,
		Enabled:     derefBool(r.Enabled),
		CreatedAt:   r.CreatedAt,
	}
}

func exclusionToDto(e *entity.WafExclusion) *WafExclusionDto {
	return &WafExclusionDto{
		ID:        e.ID,
		RuleID:    e.RuleID,
		Path:      e.Path,
		Operator:  e.Operator,
		Enabled:   derefBool(e.Enabled),
		CreatedAt: e.CreatedAt,
	}
}
