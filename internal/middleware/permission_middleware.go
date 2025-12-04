package middleware

import (
	"net/http"
	"rbac-service/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type PermissionMiddleware struct {
	permService *service.PermissionService
}

func NewPermissionMiddleware(permService *service.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permService: permService,
	}
}

func (m *PermissionMiddleware) RequirePermission(permCode string, associatedPermCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header is required"})
			return
		}

		targetTenantID := c.GetHeader("X-Tenant-ID")
		if targetTenantID == "" {
			targetTenantID = c.Query("tenant_id")
		}

		// Parse permission codes
		permParts := strings.Split(permCode, ".")
		if len(permParts) != 2 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid permission code format"})
			return
		}

		assocPermParts := strings.Split(associatedPermCode, ".")
		if len(assocPermParts) != 2 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid associated permission code format"})
			return
		}

		var tenantIDPtr *string
		if targetTenantID != "" {
			tenantIDPtr = &targetTenantID
		}

		// Optimized single query check
		allowed, err := m.permService.CheckMiddlewarePermissions(
			c.Request.Context(),
			userID,
			tenantIDPtr,
			permParts[0], permParts[1],
			assocPermParts[0], assocPermParts[1],
		)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			return
		}

		if allowed {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
	}
}
