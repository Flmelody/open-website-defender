package handler

import (
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/bot"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateBotSignature(c *gin.Context) {
	service := bot.GetBotService()

	var input bot.CreateBotSignatureDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	dto, err := service.Create(&input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create bot signature: %v", err)
		if strings.Contains(err.Error(), "invalid regex") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create bot signature")
		return
	}

	response.Created(c, dto)
}

func UpdateBotSignature(c *gin.Context) {
	service := bot.GetBotService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var input bot.UpdateBotSignatureDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	dto, err := service.Update(uint(id), &input)
	if err != nil {
		logging.Sugar.Errorf("Failed to update bot signature: %v", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid regex") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update bot signature")
		return
	}

	response.Success(c, dto)
}

func DeleteBotSignature(c *gin.Context) {
	service := bot.GetBotService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := service.Delete(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete bot signature: %v", err)
		response.InternalServerError(c, "Failed to delete bot signature")
		return
	}

	response.NoContent(c)
}

func ListBotSignatures(c *gin.Context) {
	service := bot.GetBotService()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	list, total, err := service.List(page, size)
	if err != nil {
		logging.Sugar.Errorf("Failed to list bot signatures: %v", err)
		response.InternalServerError(c, "Failed to list bot signatures")
		return
	}

	response.PageSuccess(c, list, total, page, size)
}
