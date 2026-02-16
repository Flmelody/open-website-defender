package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	authorized_domain "open-website-defender/internal/usecase/authorized_domain"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateAuthorizedDomain(c *gin.Context) {
	service := authorized_domain.GetAuthorizedDomainService()

	var req request.CreateAuthorizedDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	if !pkg.ValidateDomainPattern(req.Name) {
		response.BadRequest(c, "Invalid domain format")
		return
	}

	input := &authorized_domain.CreateAuthorizedDomainDTO{
		Name: req.Name,
	}

	dto, err := service.Create(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create authorized domain: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create authorized domain")
		return
	}

	response.Created(c, dto)
}

func DeleteAuthorizedDomain(c *gin.Context) {
	service := authorized_domain.GetAuthorizedDomainService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete authorized domain: %v", err)
		response.InternalServerError(c, "Failed to delete authorized domain")
		return
	}

	response.NoContent(c)
}

func ListAuthorizedDomains(c *gin.Context) {
	service := authorized_domain.GetAuthorizedDomainService()

	var req request.ListAuthorizedDomainRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logging.Sugar.Errorf("Invalid query parameters: %v", err)
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if req.All == "true" {
		list, err := service.ListAll()
		if err != nil {
			logging.Sugar.Errorf("Failed to list authorized domains: %v", err)
			response.InternalServerError(c, "Failed to list authorized domains")
			return
		}
		response.Success(c, list)
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
		logging.Sugar.Errorf("Failed to list authorized domains: %v", err)
		response.InternalServerError(c, "Failed to list authorized domains")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}
