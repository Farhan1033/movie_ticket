package handler

import (
	"errors"
	customerror "movie-ticket/internal/movie_module/custom_error"
	"movie-ticket/internal/movie_module/dto"
	"movie-ticket/internal/movie_module/services"

	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	svc services.MoviesService
}

func NewMovieHandlerAdmin(r *gin.RouterGroup, svc services.MoviesService) {
	h := MovieHandler{svc: svc}
	r.POST("/create-movie", h.Create)
}

func NewMoviehandlerUser(r *gin.RouterGroup, svc services.MoviesService) {
	h := MovieHandler{svc: svc}
	r.GET("/get-movie", h.Get)
}

func (h *MovieHandler) Create(c *gin.Context) {
	var req dto.CreateMovieRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
	}

	movie, err := h.svc.CreateMovie(&req)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrMovieExists):
			c.JSON(409, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(400, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidPosterUrl):
			c.JSON(400, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(201, gin.H{"data": movie})
}

func (h *MovieHandler) Get(c *gin.Context) {}
