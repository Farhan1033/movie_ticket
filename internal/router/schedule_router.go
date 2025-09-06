package router

import (
	"movie-ticket/internal/middleware"
	"movie-ticket/internal/schedule_module/handler"
	"movie-ticket/internal/schedule_module/repositories"
	"movie-ticket/internal/schedule_module/services"

	"github.com/gin-gonic/gin"
)

func InitialScheduleRouter(c *gin.Engine) {
	r := repositories.NewScheduleRepo()
	svc := services.NewShceduleSvc(r)

	apiAdmin := c.Group("/api/v1/admin")
	apiAdmin.Use(middleware.JwtMiddleware(), middleware.RequireRole("admin"))
	{
		handler.NewScheduleHandlerAdmin(apiAdmin, &svc)
	}

	api := c.Group("/api/v1")
	api.Use(middleware.JwtMiddleware(), middleware.RequireRole("user", "admin"))
	{
		handler.NewScheduleHandlerUser(api, &svc)
	}
}
