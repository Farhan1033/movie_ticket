package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GinRoleChecker(roles ...string) gin.HandlerFunc {
	// Bikin map untuk lookup lebih cepat
	roleMap := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleMap[strings.ToLower(r)] = struct{}{}
	}

	return func(c *gin.Context) {
		userRole, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Role not found in context"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid role type in context"})
			return
		}

		if _, allowed := roleMap[strings.ToLower(roleStr)]; allowed {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access forbidden for role: " + roleStr})
	}
}
