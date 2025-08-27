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
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 302 {object} map[string]interface{} // Email sudah ada
// @Failure 500 {object} map[string]interface{}
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func (c *AuthHandler) Logout(ctx *gin.Context) {
	// Ambil token dari header
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := c.svc.Logout(token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (c *AuthHandler) RefreshToken(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header"})
		return
	}

	oldToken := strings.TrimPrefix(authHeader, "Bearer ")

	response, err := c.svc.RefreshToken(oldToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
