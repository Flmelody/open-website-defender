package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/waf"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateWafRule(c *gin.Context) {
	service := waf.GetWafService()

	var req request.CreateWafRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &waf.CreateWafRuleDto{
		Name:        req.Name,
		Pattern:     req.Pattern,
		Category:    req.Category,
		Action:      req.Action,
		Operator:    req.Operator,
		Target:      req.Target,
		Priority:    req.Priority,
		GroupName:   req.GroupName,
		RedirectURL: req.RedirectURL,
		RateLimit:   req.RateLimit,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	dto, err := service.Create(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create WAF rule: %v", err)
		if strings.Contains(err.Error(), "invalid regex") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create WAF rule")
		return
	}

	response.Created(c, dto)
}

func UpdateWafRule(c *gin.Context) {
	service := waf.GetWafService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req request.UpdateWafRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &waf.UpdateWafRuleDto{
		Name:        req.Name,
		Pattern:     req.Pattern,
		Category:    req.Category,
		Action:      req.Action,
		Operator:    req.Operator,
		Target:      req.Target,
		Priority:    req.Priority,
		GroupName:   req.GroupName,
		RedirectURL: req.RedirectURL,
		RateLimit:   req.RateLimit,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	dto, err := service.Update(uint(id), input)
	if err != nil {
		logging.Sugar.Errorf("Failed to update WAF rule: %v", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid regex") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update WAF rule")
		return
	}

	response.Success(c, dto)
}

func DeleteWafRule(c *gin.Context) {
	service := waf.GetWafService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete WAF rule: %v", err)
		response.InternalServerError(c, "Failed to delete WAF rule")
		return
	}

	response.NoContent(c)
}

func ListWafRules(c *gin.Context) {
	service := waf.GetWafService()

	var req request.ListWafRuleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logging.Sugar.Errorf("Invalid query parameters: %v", err)
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}

	list, total, err := service.List(req.Page, req.Size)
	if err != nil {
		logging.Sugar.Errorf("Failed to list WAF rules: %v", err)
		response.InternalServerError(c, "Failed to list WAF rules")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}

func BatchEnableWafGroup(c *gin.Context) {
	service := waf.GetWafService()
	groupName := c.Param("name")

	if err := service.BatchEnableGroup(groupName, true); err != nil {
		logging.Sugar.Errorf("Failed to enable WAF group: %v", err)
		response.InternalServerError(c, "Failed to enable WAF group")
		return
	}

	response.Success(c, gin.H{"message": "group enabled"})
}

func BatchDisableWafGroup(c *gin.Context) {
	service := waf.GetWafService()
	groupName := c.Param("name")

	if err := service.BatchEnableGroup(groupName, false); err != nil {
		logging.Sugar.Errorf("Failed to disable WAF group: %v", err)
		response.InternalServerError(c, "Failed to disable WAF group")
		return
	}

	response.Success(c, gin.H{"message": "group disabled"})
}

func CreateWafExclusion(c *gin.Context) {
	service := waf.GetWafService()

	var req request.CreateWafExclusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &waf.CreateWafExclusionDto{
		RuleID:   req.RuleID,
		Path:     req.Path,
		Operator: req.Operator,
		Enabled:  req.Enabled,
	}

	dto, err := service.CreateExclusion(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create WAF exclusion: %v", err)
		response.InternalServerError(c, "Failed to create WAF exclusion")
		return
	}

	response.Created(c, dto)
}

func DeleteWafExclusion(c *gin.Context) {
	service := waf.GetWafService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.DeleteExclusion(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete WAF exclusion: %v", err)
		response.InternalServerError(c, "Failed to delete WAF exclusion")
		return
	}

	response.NoContent(c)
}

func ListWafExclusions(c *gin.Context) {
	service := waf.GetWafService()

	var req request.ListWafExclusionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}

	list, total, err := service.ListExclusions(req.Page, req.Size)
	if err != nil {
		logging.Sugar.Errorf("Failed to list WAF exclusions: %v", err)
		response.InternalServerError(c, "Failed to list WAF exclusions")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}
