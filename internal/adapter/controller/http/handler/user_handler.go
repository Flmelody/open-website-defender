package handler

import (
	"open-website-defender/internal/adapter/controller/http/request"
	"open-website-defender/internal/adapter/controller/http/response"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/user"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	service := user.GetUserService()

	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &user.CreateUserDTO{
		Username: req.Username,
		Password: req.Password,
		GitToken: req.GitToken,
		IsAdmin:  req.IsAdmin,
	}

	userDto, err := service.CreateUser(input)
	if err != nil {
		logging.Sugar.Errorf("Failed to create user: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			response.Conflict(c, "Username already exists")
			return
		}
		response.InternalServerError(c, "Failed to create user")
		return
	}

	logging.Sugar.Infof("User created successfully: %s", userDto.Username)
	response.Created(c, userDto)
}

func UpdateUser(c *gin.Context) {
	service := user.GetUserService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Sugar.Errorf("Invalid request format: %v", err)
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	input := &user.UpdateUserDTO{
		Username: req.Username,
		Password: req.Password,
		GitToken: req.GitToken,
		IsAdmin:  req.IsAdmin,
	}

	userDto, err := service.UpdateUser(uint(id), input)
	if err != nil {
		logging.Sugar.Errorf("Failed to update user: %v", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User not found")
			return
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			response.Conflict(c, "Username already exists")
			return
		}
		response.InternalServerError(c, "Failed to update user")
		return
	}

	logging.Sugar.Infof("User updated successfully: ID=%d", id)
	response.Success(c, userDto)
}

func DeleteUser(c *gin.Context) {
	service := user.GetUserService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := service.DeleteUser(uint(id)); err != nil {
		logging.Sugar.Errorf("Failed to delete user: %v", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalServerError(c, "Failed to delete user")
		return
	}

	logging.Sugar.Infof("User deleted successfully: ID=%d", id)
	response.NoContent(c)
}

func GetUser(c *gin.Context) {
	service := user.GetUserService()

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	userDto, err := service.GetUser(uint(id))
	if err != nil {
		logging.Sugar.Errorf("Failed to get user: %v", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalServerError(c, "Failed to get user")
		return
	}

	response.Success(c, userDto)
}

func ListUser(c *gin.Context) {
	service := user.GetUserService()

	var req request.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logging.Sugar.Errorf("Invalid query parameters: %v", err)
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}

	users, total, err := service.ListUsers(req.Page, req.Size)
	if err != nil {
		logging.Sugar.Errorf("Failed to list users: %v", err)
		response.InternalServerError(c, "Failed to list users")
		return
	}

	response.PageSuccess(c, users, total, req.Page, req.Size)
}
