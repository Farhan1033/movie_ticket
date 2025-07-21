package router

import (
	"movie-ticket/internal/auth_module/handler"
	"movie-ticket/internal/auth_module/repositories"
	"movie-ticket/internal/auth_module/services"

	"github.com/gin-gonic/gin"
)

func InitAuthRoutes(r *gin.Engine) {
	auth := repositories.NewAuthRepo()
	authSvc := services.NewAuthSvc(auth)

	api := r.Group("api/v1")
	{
		handler.NewAuthHandler(api, authSvc)
	}
}
