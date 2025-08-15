package handler

import (
	"errors"
	"movie-ticket/internal/middleware"
	movieError "movie-ticket/internal/movie_module/custom_error"
	customerrors "movie-ticket/internal/schedule_module/custom_errors"
	"movie-ticket/internal/schedule_module/dto"
	"movie-ticket/internal/schedule_module/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	svc services.ScheduleServices
}

func NewScheduleHandlerAdmin(r *gin.RouterGroup, svc *services.ScheduleServices) {
	h := ScheduleHandler{svc: *svc}
	r.POST("/schedule/create", h.CreateStudio)
}

func (h *ScheduleHandler) CreateStudio(c *gin.Context) {
	var reqSchedule *dto.ScheduleCreateRequest
	role, err := middleware.GetUserRoleFromRedis(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorize user"})
		return
	}

	if err := c.ShouldBindJSON(&reqSchedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	schedule, err := h.svc.Create(role, reqSchedule)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrTimeStart):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, movieError.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrInactiveMovie):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": schedule})
}
