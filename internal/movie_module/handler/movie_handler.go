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
	r.PATCH("/movie/:id/status", h.PatchStatus)
}

func NewMoviehandlerUser(r *gin.RouterGroup, svc services.MoviesService) {
	h := MovieHandler{svc: svc}
	r.GET("/movie", h.Get)
	r.GET("/movie/:id", h.GetById)
}

// Create godoc
// @Summary Membuat movie baru (Admin only)
// @Description Membuat movie baru dengan informasi lengkap. Hanya admin yang dapat mengakses endpoint ini
// @Tags Movies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.CreateMovieRequest true "Movie creation data"
// @Success 201 {object} map[string]interface{} "Movie created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input, session, atau poster URL"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 409 {object} map[string]interface{} "Conflict - Movie sudah ada"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/movie/create [post]
// @Security BearerAuth
func (h *MovieHandler) Create(c *gin.Context) {
	var req dto.CreateMovieRequest
	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
		return
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

// Get godoc
// @Summary Mendapatkan daftar movie dengan pagination
// @Description Mengambil daftar semua movie yang tersedia dengan dukungan pagination
// @Tags Movies
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman" default(1) minimum(1)
// @Param limit query int false "Jumlah data per halaman" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.MoviesResponse "Data movie berhasil diambil"
// @Failure 404 {object} map[string]interface{} "Movie tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /movie [get]
func (h *MovieHandler) Get(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
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

// GetById godoc
// @Summary Mendapatkan detail movie berdasarkan ID
// @Description Mengambil informasi lengkap movie berdasarkan ID yang diberikan
// @Tags Movies
// @Accept json
// @Produce json
// @Param id path string true "Movie ID" format(uuid)
// @Success 200 {object} dto.MoviesResponse "Detail movie berhasil diambil"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid movie ID atau movie tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /movie/{id} [get]
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

// Update godoc
// @Summary Update movie (Admin only)
// @Description Mengupdate informasi movie yang sudah ada. Hanya admin yang dapat mengakses endpoint ini
// @Tags Movies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Movie ID" format(uuid)
// @Param request body dto.UpdateMovieRequest true "Movie update data"
// @Success 201 {object} dto.MoviesResponse "Movie updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input, session, poster URL, atau movie ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 409 {object} map[string]interface{} "Conflict - Movie sudah ada"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/movie/update/{id} [put]
// @Security BearerAuth
func (h *MovieHandler) Update(c *gin.Context) {
	var req dto.UpdateMovieRequest
	idParam := c.Param("id")

	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
		return
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

// Delete godoc
// @Summary Hapus movie (Admin only)
// @Description Menghapus movie berdasarkan ID. Hanya admin yang dapat mengakses endpoint ini
// @Tags Movies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Movie ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Movie deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid movie ID atau session failed"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 409 {object} map[string]interface{} "Conflict - Movie tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/movie/delete/{id} [delete]
// @Security BearerAuth
func (h *MovieHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed Get session from redis"})
		return
	}

	deleteMovie := h.svc.DeleteMovie(userRole, idParam)
	if deleteMovie != nil {
		switch {
		case errors.Is(deleteMovie, customerror.ErrMovieNotFound):
			c.JSON(http.StatusConflict, gin.H{"error": deleteMovie.Error()})
		case errors.Is(deleteMovie, customerror.ErrInvalidMovieId):
			c.JSON(http.StatusBadRequest, gin.H{"error": deleteMovie.Error()})
		case errors.Is(deleteMovie, customerror.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": deleteMovie.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": deleteMovie.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted data"})
}

// PatchStatus godoc
// @Summary Update status movie (Admin only)
// @Description Mengubah status movie (aktif/non-aktif) berdasarkan ID. Hanya admin yang dapat mengakses endpoint ini
// @Tags Movies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Movie ID" format(uuid)
// @Param request body dto.StatusMovieRequest true "Movie status data"
// @Success 200 {object} map[string]interface{} "Status movie berhasil diupdate"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid request body, movie ID, input, atau session failed"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User bukan admin"
// @Failure 404 {object} map[string]interface{} "Not Found - Movie tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/movie/{id}/status [patch]
// @Security BearerAuth
func (h *MovieHandler) PatchStatus(c *gin.Context) {
	var reqStatus dto.StatusMovieRequest
	idParam := c.Param("id")

	if err := c.ShouldBindJSON(&reqStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userRole, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get session from redis"})
		return
	}

	if err := h.svc.PatchStatus(userRole, idParam, &reqStatus); err != nil {
		switch {
		case errors.Is(err, customerror.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidMovieId):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrUnauthorizedUser):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// pastikan status tidak nil sebelum akses
	if reqStatus.Status != nil && !*reqStatus.Status {
		c.JSON(http.StatusOK, gin.H{"message": "Successfully disabled the film"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully enabled the film"})
}
