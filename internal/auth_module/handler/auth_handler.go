package handler

import (
	"errors"
	customerror "movie-ticket/internal/auth_module/custom_error"
	"movie-ticket/internal/auth_module/dto"
	"movie-ticket/internal/auth_module/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc services.AuthService
}

func NewAuthHandler(r *gin.RouterGroup, svc services.AuthService) {
	h := AuthHandler{svc: svc}
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
	r.POST("/refresh", h.RefreshToken)
}

// Register godoc
// @Summary Register user baru
// @Description Membuat akun baru dengan email & password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Data"
// @Success 201 {object} map[string]interface{} "Account created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input"
// @Failure 302 {object} map[string]interface{} "Found - Email sudah ada"
// @Failure 404 {object} map[string]interface{} "Not Found - Email not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.Register(&input)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrEmailExist):
			c.JSON(http.StatusFound, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrEmailNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"data":    user,
	})
}

// Login godoc
// @Summary Login user
// @Description Masuk akun dengan email & password untuk mendapatkan access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Data"
// @Success 200 {object} map[string]interface{} "Login berhasil dengan token"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid input atau session gagal"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Password salah"
// @Failure 404 {object} map[string]interface{} "Not Found - Email tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := h.svc.Login(&req)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrEmailNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrWrongPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, customerror.ErrFailedSession):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": users})
}

// Logout godoc
// @Summary Logout user
// @Description Keluar dari akun dan invalidate token yang aktif
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Success 200 {object} map[string]interface{} "Logged out successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid authorization header"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /logout [post]
// @Security BearerAuth
func (h *AuthHandler) Logout(ctx *gin.Context) {
	// Ambil token dari header
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := h.svc.Logout(token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Mendapatkan access token baru menggunakan token yang masih valid
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Success 200 {object} map[string]interface{} "Token refreshed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid authorization header"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Token expired atau invalid"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /refresh [post]
// @Security BearerAuth
func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header"})
		return
	}

	oldToken := strings.TrimPrefix(authHeader, "Bearer ")

	response, err := h.svc.RefreshToken(oldToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
