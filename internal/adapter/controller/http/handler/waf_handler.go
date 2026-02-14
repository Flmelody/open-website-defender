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
		Name:     req.Name,
		Pattern:  req.Pattern,
		Category: req.Category,
		Action:   req.Action,
		Enabled:  req.Enabled,
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
		Name:     req.Name,
		Pattern:  req.Pattern,
		Category: req.Category,
		Action:   req.Action,
		Enabled:  req.Enabled,
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
