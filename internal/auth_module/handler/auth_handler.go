package handler

import (
	"movie-ticket/internal/auth_module/entities"
	"movie-ticket/internal/auth_module/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc services.AuthService
}

func NewAuthHandler(r *gin.RouterGroup, svc services.AuthService) {
	h := AuthHandler{svc: svc}
	r.Group("/register", h.Register)
	r.Group("/login", h.Login)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input entities.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.Register(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Account created successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {}
