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
	api.Use(middleware.JWTAuth())
	{
		handler.NewMoviehandlerUser(api, moviesSvc)
	}

	apiAdmin := r.Group("/api/v1/admin")
	{
		handler.NewMovieHandlerAdmin(apiAdmin, moviesSvc)
	}
}
