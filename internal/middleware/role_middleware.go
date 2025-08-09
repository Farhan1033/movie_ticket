package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists || role != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin onlys"})
			return
		}
	}
}

func UserOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists || role != "user" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Foribidden: User onlys"})
		}
	}
}
