package handlers

import (
	"errors"
	"movie-ticket/internal/middleware"
	customerror "movie-ticket/internal/studio_module/custom_error"
	"movie-ticket/internal/studio_module/dto"
	"movie-ticket/internal/studio_module/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StudioHandler struct {
	service services.StudioService
}

func NewStudioHandlerAdmin(r *gin.RouterGroup, svc *services.StudioService) {
	h := StudioHandler{service: *svc}
	r.POST("/studio/create", h.Create)
	r.PUT("/studio/update/:id", h.Update)
	r.DELETE("/studio/delete/:id", h.Delete)
}

func NewStudioHandlerUser(r *gin.RouterGroup, svc *services.StudioService) {
	h := StudioHandler{service: *svc}
	r.GET("/studios", h.Get)
	r.GET("/studio", h.GetByName)
	r.GET("/studio/:id", h.GetById)
}

// Create godoc
// @Summary Membuat studio baru (Admin only)
// @Description Membuat studio baru dengan informasi nama, kapasitas, dan tipe. Hanya admin yang dapat mengakses endpoint ini
// @Tags Studios
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.CreateStudioRequest true "Studio creation data"
// @Success 201 {object} map[string]interface{} "Studio created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 406 {object} map[string]interface{} "Not Acceptable - Studio sudah ada"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/studio/create [post]
// @Security BearerAuth
func (h *StudioHandler) Create(c *gin.Context) {
	role, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized User!"})
		return
	}

	var req dto.CreateStudioRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	studio, err := h.service.Create(role, &req)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrStudioExists):
			c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, studio)
}

// Get godoc
// @Summary Mendapatkan daftar semua studio
// @Description Mengambil daftar semua studio yang tersedia di bioskop
// @Tags Studios
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{} "Data studio berhasil diambil"
// @Failure 404 {object} map[string]interface{} "Not Found - Studio tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error - Database error"
// @Router /studios [get]
func (h *StudioHandler) Get(c *gin.Context) {
	studios, err := h.service.Get()
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrStudioNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, customerror.ErrDatabaseError)
		}
		return
	}
	c.JSON(http.StatusOK, studios)
}

// GetByName godoc
// @Summary Mencari studio berdasarkan nama
// @Description Mengambil informasi studio berdasarkan nama yang diberikan sebagai query parameter
// @Tags Studios
// @Accept json
// @Produce json
// @Param name query string true "Nama studio yang dicari"
// @Success 200 {object} map[string]interface{} "Data studio berhasil diambil"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input atau nama kosong"
// @Failure 404 {object} map[string]interface{} "Not Found - Studio tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error - Database error"
// @Router /studio [get]
func (h *StudioHandler) GetByName(c *gin.Context) {
	name := c.Query("name")
	studio, err := h.service.GetByName(name)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrStudioNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, customerror.ErrDatabaseError)
		}
		return
	}
	c.JSON(http.StatusOK, studio)
}

// GetById godoc
// @Summary Mendapatkan detail studio berdasarkan ID
// @Description Mengambil informasi lengkap studio berdasarkan ID yang diberikan
// @Tags Studios
// @Accept json
// @Produce json
// @Param id path string true "Studio ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Detail studio berhasil diambil"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid studio ID"
// @Failure 404 {object} map[string]interface{} "Not Found - Studio tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error - Database error"
// @Router /studio/{id} [get]
func (h *StudioHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	studio, err := h.service.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrStudioNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, customerror.ErrDatabaseError)
		}
		return
	}
	c.JSON(http.StatusOK, studio)
}

// Update godoc
// @Summary Update studio (Admin only)
// @Description Mengupdate informasi studio yang sudah ada. Hanya admin yang dapat mengakses endpoint ini
// @Tags Studios
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Studio ID" format(uuid)
// @Param request body dto.UpdateStudioRequest true "Studio update data"
// @Success 200 {object} map[string]interface{} "Studio updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input atau studio ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 404 {object} map[string]interface{} "Not Found - Studio tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error - Database error"
// @Router /admin/studio/update/{id} [put]
// @Security BearerAuth
func (h *StudioHandler) Update(c *gin.Context) {
	role, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized User!"})
		return
	}
	id := c.Param("id")

	var req dto.UpdateStudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	studio, err := h.service.Update(role, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidStudioId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrStudioNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, customerror.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, studio)
}

// Delete godoc
// @Summary Hapus studio (Admin only)
// @Description Menghapus studio berdasarkan ID. Hanya admin yang dapat mengakses endpoint ini. Studio yang masih memiliki jadwal aktif tidak dapat dihapus
// @Tags Studios
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Studio ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Studio deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input atau studio ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 404 {object} map[string]interface{} "Not Found - Studio tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error - Database error"
// @Router /admin/studio/delete/{id} [delete]
// @Security BearerAuth
func (h *StudioHandler) Delete(c *gin.Context) {
	role, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized User!"})
		return
	}
	id := c.Param("id")

	if err := h.service.Delete(role, id); err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidStudioId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrStudioNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, customerror.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Studio deleted successfully"})
}
