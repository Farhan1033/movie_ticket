package router

import (
	"movie-ticket/internal/middleware"
	handlers "movie-ticket/internal/studio_module/handler"
	"movie-ticket/internal/studio_module/repositories"
	"movie-ticket/internal/studio_module/services"

	"github.com/gin-gonic/gin"
)

func InitStudioRouter(r *gin.Engine) {
	studioRepo := repositories.NewStudioRepo()
	studioSvc := services.NewStudioService(studioRepo)

	api := r.Group("/api/v1")
	api.Use(middleware.JwtMiddleware(), middleware.RequireRole("admin", "user"))
	{
		handlers.NewStudioHandlerUser(api, &studioSvc)
	}

	apiAdmin := r.Group("/api/v1/admin")
	api.Use(middleware.JwtMiddleware(), middleware.RequireRole("admin"))
	{
		handlers.NewStudioHandlerAdmin(apiAdmin, &studioSvc)
	}
}
