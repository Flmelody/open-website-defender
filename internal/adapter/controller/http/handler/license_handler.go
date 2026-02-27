package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/license"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateLicense(c *gin.Context) {
	service := license.GetLicenseService()

	var req request.CreateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &license.CreateLicenseDTO{
		Name:   req.Name,
		Remark: req.Remark,
	}

	dto, err := service.Create(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create license: %v", err)
		response.InternalServerError(c, "Failed to create license")
		return
	}

	response.Created(c, dto)
}

func DeleteLicense(c *gin.Context) {
	service := license.GetLicenseService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete license: %v", err)
		response.InternalServerError(c, "Failed to delete license")
		return
	}

	response.NoContent(c)
}

func ListLicenses(c *gin.Context) {
	service := license.GetLicenseService()

	var req request.ListLicenseRequest
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
		logging.Sugar.Errorf("Failed to list licenses: %v", err)
		response.InternalServerError(c, "Failed to list licenses")
		return
	}

	response.PageSuccess(c, list, total, req.Page, req.Size)
}
