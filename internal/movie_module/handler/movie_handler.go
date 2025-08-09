package handler

import (
	"errors"
	"movie-ticket/internal/middleware"
	customerror "movie-ticket/internal/movie_module/custom_error"
	"movie-ticket/internal/movie_module/dto"
	"movie-ticket/internal/movie_module/services"
	"net/http"
	"strconv"

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
	// GET from role redis
	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
	}

	// Binding JSON ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Panggil service untuk membuat movie
	movie, err := h.svc.CreateMovie(userRole, &req)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrMovieExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidPosterUrl):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": movie})
}

func (h *MovieHandler) Get(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 1
	}

	movies, err := h.svc.GetMovies(page, limit)

	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrMovieNotFound):
			c.JSON(409, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(200, gin.H{"data": movies})

}
