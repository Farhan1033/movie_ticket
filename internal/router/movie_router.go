package router

import (
	"movie-ticket/internal/middleware"
	"movie-ticket/internal/movie_module/handler"
	"movie-ticket/internal/movie_module/repositories"
	"movie-ticket/internal/movie_module/services"

	"github.com/gin-gonic/gin"
)

func InitMovieRoute(r *gin.Engine) {
	movies := repositories.NewMovieRepo()
	moviesSvc := services.NewMoviesService(movies)

	api := r.Group("/api/v1/")
	api.Use(middleware.JwtMiddleware(), middleware.GinRoleChecker("admin", "user"))
	{
		handler.NewMoviehandlerUser(api, moviesSvc)
	}

	apiAdmin := r.Group("/api/v1/admin")
	api.Use(middleware.JwtMiddleware(), middleware.GinRoleChecker("admin"))
	{
		handler.NewMovieHandlerAdmin(apiAdmin, moviesSvc)
	}
}
