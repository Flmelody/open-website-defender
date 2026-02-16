package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/oauth"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateOAuthClient(c *gin.Context) {
	service := oauth.GetOAuthService()

	var req request.CreateOAuthClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &oauth.CreateOAuthClientDTO{
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		Trusted:      req.Trusted,
	}

	dto, err := service.CreateClient(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create OAuth client: %v", err)
		response.InternalServerError(c, "Failed to create OAuth client")
		return
	}

	logging.Sugar.Infof("OAuth client created: %s (client_id: %s)", dto.Name, dto.ClientID)
	response.Created(c, dto)
}

func UpdateOAuthClient(c *gin.Context) {
	service := oauth.GetOAuthService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid client ID")
		return
	}

	var req request.UpdateOAuthClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &oauth.UpdateOAuthClientDTO{
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		Trusted:      req.Trusted,
		Active:       req.Active,
	}

	dto, err := service.UpdateClient(uint(id), input)
	if err != nil {
		logging.Sugar.Errorf("Failed to update OAuth client: %v", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "OAuth client not found")
			return
		}
		response.InternalServerError(c, "Failed to update OAuth client")
		return
	}

	logging.Sugar.Infof("OAuth client updated: ID=%d", id)
	response.Success(c, dto)
}

func DeleteOAuthClient(c *gin.Context) {
	service := oauth.GetOAuthService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid client ID")
		return
	}

	if err := service.DeleteClient(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete OAuth client: %v", err)
		if err == oauth.ErrClientNotFound {
			response.NotFound(c, "OAuth client not found")
			return
		}
		response.InternalServerError(c, "Failed to delete OAuth client")
		return
	}

	logging.Sugar.Infof("OAuth client deleted: ID=%d", id)
	response.NoContent(c)
}

func GetOAuthClient(c *gin.Context) {
	service := oauth.GetOAuthService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid client ID")
		return
	}

	dto, err := service.GetClient(uint(id))
	if err != nil {
		logging.Sugar.Errorf("Failed to get OAuth client: %v", err)
		if err == oauth.ErrClientNotFound {
			response.NotFound(c, "OAuth client not found")
			return
		}
		response.InternalServerError(c, "Failed to get OAuth client")
		return
	}

	response.Success(c, dto)
}

func ListOAuthClients(c *gin.Context) {
	service := oauth.GetOAuthService()

	var req request.ListOAuthClientRequest
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

	list, total, err := service.ListClients(req.Page, req.Size)
	if err != nil {
		logging.Sugar.Errorf("Failed to list OAuth clients: %v", err)
		response.InternalServerError(c, "Failed to list OAuth clients")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}
