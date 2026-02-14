package handler

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/geoblock"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateGeoBlockRule(c *gin.Context) {
	service := geoblock.GetGeoBlockService()

	var req struct {
		CountryCode string `json:"country_code" binding:"required"`
		CountryName string `json:"country_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	dto, err := service.Create(req.CountryCode, req.CountryName)
	if err != nil {
		logging.Sugar.Errorf("Failed to create geo-block rule: %v", err)
		if strings.Contains(err.Error(), "already blocked") {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create geo-block rule")
		return
	}

	response.Created(c, dto)
}

func DeleteGeoBlockRule(c *gin.Context) {
	service := geoblock.GetGeoBlockService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete geo-block rule: %v", err)
		response.InternalServerError(c, "Failed to delete geo-block rule")
		return
	}

	response.NoContent(c)
}

func ListGeoBlockRules(c *gin.Context) {
	service := geoblock.GetGeoBlockService()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	list, total, err := service.List(page, size)
	if err != nil {
		logging.Sugar.Errorf("Failed to list geo-block rules: %v", err)
		response.InternalServerError(c, "Failed to list geo-block rules")
		return
	}

	response.PageSuccess(c, list, total, page, size)
}
