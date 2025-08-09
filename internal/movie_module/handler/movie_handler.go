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
	r.POST("/movie/create", h.Create)
	r.PUT("/movie/update/:id", h.Update)
	r.DELETE("/movie/delete/:id", h.Delete)
}

func NewMoviehandlerUser(r *gin.RouterGroup, svc services.MoviesService) {
	h := MovieHandler{svc: svc}
	r.GET("/movie", h.Get)
	r.GET("/movie/:id", h.GetById)
}

func (h *MovieHandler) Create(c *gin.Context) {
	var req dto.CreateMovieRequest
	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

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
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.MoviesResponse{Message: "successfully retrieved the data", Data: movies})

}

func (h *MovieHandler) GetById(c *gin.Context) {
	idParams := c.Param("id")

	movie, err := h.svc.GetMovieById(idParams)

	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidMovieId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrMovieNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.MoviesResponse{Message: "successfully retrieved the data", Data: movie})
}

func (h *MovieHandler) Update(c *gin.Context) {
	var req dto.UpdateMovieRequest
	idParam := c.Param("id")

	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updateMovie, err := h.svc.UpdateMovie(userRole, idParam, &req)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrMovieExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidPosterUrl):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidMovieId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.MoviesResponse{Message: "successfully updated data", Data: updateMovie})
}

func (h *MovieHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
	}

	deleteMovie := h.svc.DeleteMovie(userRole, idParam)
	if deleteMovie != nil {
		switch {
		case errors.Is(err, customerror.ErrMovieNotFound):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidMovieId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted data"})
}
