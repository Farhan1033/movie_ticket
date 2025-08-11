package handlers

import (
	"movie-ticket/internal/middleware"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, studio)
}

func (h *StudioHandler) Get(c *gin.Context) {
	studios, err := h.service.Get()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, studios)
}

func (h *StudioHandler) GetByName(c *gin.Context) {
	name := c.Query("name")
	studio, err := h.service.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, studio)
}

func (h *StudioHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	studio, err := h.service.GetById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, studio)
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, studio)
}

func (h *StudioHandler) Delete(c *gin.Context) {
	role, err := middleware.GetUserRoleFromRedis(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized User!"})
		return
	}
	id := c.Param("id")

	if err := h.service.Delete(role, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Studio deleted successfully"})
}
