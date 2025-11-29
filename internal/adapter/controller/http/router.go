package http

import (
	"open-website-defender/internal/adapter/controller/http/handler"
	"open-website-defender/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func Setup(appConfig *config.AppConfig) *gin.Engine {
	router := gin.Default()

	api := router.Group(appConfig.RootPath)
	{
		// Standalone auth check endpoint
		api.GET("/auth", handler.Auth)
		api.POST("/login", handler.Login)

		authorized := api.Group("")
		// Middleware for route protection
		authorized.Use(handler.AuthMiddleware)
		{
			authorized.POST("/users", handler.CreateUser)
			authorized.GET("/users", handler.ListUser)
			authorized.GET("/users/:id", handler.GetUser)
			authorized.PUT("/users/:id", handler.UpdateUser)
			authorized.DELETE("/users/:id", handler.DeleteUser)

			// IP Blacklist
			authorized.POST("/ip-black-list", handler.CreateIpBlackList)
			authorized.GET("/ip-black-list", handler.ListIpBlackList)
			authorized.DELETE("/ip-black-list/:id", handler.DeleteIpBlackList)

			// IP Whitelist
			authorized.POST("/ip-white-list", handler.CreateIpWhiteList)
			authorized.GET("/ip-white-list", handler.ListIpWhiteList)
			authorized.DELETE("/ip-white-list/:id", handler.DeleteIpWhiteList)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
