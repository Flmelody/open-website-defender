package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/iplist"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// BlackList Handlers

func CreateIpBlackList(c *gin.Context) {
	service := iplist.GetIpBlackListService()

	var req request.CreateIpBlackListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &iplist.CreateIpBlackListDto{
		Ip: req.Ip,
	}

	dto, err := service.Create(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create blacklist item: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create blacklist item")
		return
	}

	response.Created(c, dto)
}

func DeleteIpBlackList(c *gin.Context) {
	service := iplist.GetIpBlackListService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete blacklist item: %v", err)
		response.InternalServerError(c, "Failed to delete blacklist item")
		return
	}

	response.NoContent(c)
}

func ListIpBlackList(c *gin.Context) {
	service := iplist.GetIpBlackListService()

	var req request.ListIpBlackListRequest
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
		logging.Sugar.Errorf("Failed to list blacklist items: %v", err)
		response.InternalServerError(c, "Failed to list blacklist items")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}

// WhiteList Handlers

func CreateIpWhiteList(c *gin.Context) {
	service := iplist.GetIpWhiteListService()

	var req request.CreateIpWhiteListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &iplist.CreateIpWhiteListDto{
		Ip:     req.Ip,
		Domain: req.Domain,
	}

	dto, err := service.Create(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create whitelist item: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create whitelist item")
		return
	}

	response.Created(c, dto)
}

func DeleteIpWhiteList(c *gin.Context) {
	service := iplist.GetIpWhiteListService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete whitelist item: %v", err)
		response.InternalServerError(c, "Failed to delete whitelist item")
		return
	}

	response.NoContent(c)
}

func ListIpWhiteList(c *gin.Context) {
	service := iplist.GetIpWhiteListService()

	var req request.ListIpWhiteListRequest
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
		logging.Sugar.Errorf("Failed to list whitelist items: %v", err)
		response.InternalServerError(c, "Failed to list whitelist items")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}
