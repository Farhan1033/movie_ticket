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
	r.POST("/schedule/create", h.CreateSchedule)
	r.PUT("/schedule/update/:id", h.UpdateSchedule)
	r.DELETE("/schedule/delete/:id", h.DeleteSchedule)
}

func NewScheduleHandlerUser(r *gin.RouterGroup, svc *services.ScheduleServices) {
	h := ScheduleHandler{svc: *svc}
	r.GET("/schedule", h.Get)
	r.GET("/schedule/:id", h.GetById)
}

// CreateSchedule godoc
// @Summary Membuat jadwal tayang baru (Admin only)
// @Description Membuat jadwal tayang baru untuk movie tertentu dengan studio, waktu, dan harga yang ditentukan. Hanya admin yang dapat mengakses endpoint ini
// @Tags Schedules
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.ScheduleCreateRequest true "Schedule creation data"
// @Success 201 {object} map[string]interface{} "Schedule created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input, waktu mulai, movie tidak aktif, atau harga invalid"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 404 {object} map[string]interface{} "Not Found - Movie tidak ditemukan"
// @Failure 409 {object} map[string]interface{} "Conflict - Jadwal bertabrakan dengan jadwal lain"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/schedule/create [post]
// @Security BearerAuth
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var reqSchedule dto.ScheduleCreateRequest
	role, err := middleware.GetUserRoleFromRedis(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorize user"})
		return
	}

	if err := c.ShouldBindJSON(&reqSchedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	schedule, err := h.svc.Create(role, &reqSchedule)
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
		case errors.Is(err, customerrors.ErrScheduleConflict):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrPriceInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": schedule})
}

// Get godoc
// @Summary Mendapatkan daftar semua jadwal tayang
// @Description Mengambil semua jadwal tayang yang tersedia untuk semua movie
// @Tags Schedules
// @Accept json
// @Produce json
// @Success 200 {object} dto.MessageResponse "Data jadwal berhasil diambil"
// @Failure 404 {object} map[string]interface{} "Not Found - Jadwal tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /schedule [get]
func (h *ScheduleHandler) Get(c *gin.Context) {
	schedules, err := h.svc.Get()
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Successfully displaying data",
		Data:    schedules})
}

// GetById godoc
// @Summary Mendapatkan detail jadwal berdasarkan ID
// @Description Mengambil informasi lengkap jadwal tayang berdasarkan ID yang diberikan
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID" format(uuid)
// @Success 200 {object} dto.MessageResponse "Detail jadwal berhasil diambil"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid schedule ID"
// @Failure 404 {object} map[string]interface{} "Not Found - Jadwal tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /schedule/{id} [get]
func (h *ScheduleHandler) GetById(c *gin.Context) {
	id := c.Param("id")

	schedule, err := h.svc.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Successfully displaying data",
		Data:    schedule})
}

// UpdateSchedule godoc
// @Summary Update jadwal tayang (Admin only)
// @Description Mengupdate informasi jadwal tayang yang sudah ada. Hanya admin yang dapat mengakses endpoint ini
// @Tags Schedules
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Schedule ID" format(uuid)
// @Param request body dto.ScheduleUpdateRequest true "Schedule update data"
// @Success 200 {object} dto.MessageResponse "Schedule updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input, schedule ID, waktu mulai, atau harga"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 404 {object} map[string]interface{} "Not Found - Jadwal tidak ditemukan"
// @Failure 409 {object} map[string]interface{} "Conflict - Jadwal bertabrakan dengan jadwal lain"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/schedule/update/{id} [put]
// @Security BearerAuth
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")
	role, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}

	var req dto.ScheduleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	schedule, err := h.svc.Update(role, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrInvalidScheduleId),
			errors.Is(err, customerrors.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrTimeStart):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrScheduleConflict):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrPriceInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Successfully updated schedule",
		Data:    schedule,
	})
}

// DeleteSchedule godoc
// @Summary Hapus jadwal tayang (Admin only)
// @Description Menghapus jadwal tayang berdasarkan ID. Hanya admin yang dapat mengakses endpoint ini
// @Tags Schedules
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Schedule ID" format(uuid)
// @Success 200 {object} dto.MessageResponse "Schedule deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid schedule ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 404 {object} map[string]interface{} "Not Found - Jadwal tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/schedule/delete/{id} [delete]
// @Security BearerAuth
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")
	role, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}

	if err := h.svc.Delete(role, id); err != nil {
		switch {
		case errors.Is(err, customerrors.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrInvalidScheduleId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerrors.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Successfully deleted schedule",
	})
}
