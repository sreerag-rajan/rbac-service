package controller

import (
	"net/http"
	"rbac-service/internal/app"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantApp *app.TenantAppService
}

func NewTenantHandler(tenantApp *app.TenantAppService) *TenantHandler {
	return &TenantHandler{
		tenantApp: tenantApp,
	}
}

func (h *TenantHandler) BulkAssignPermissions(c *gin.Context) {
	var req model.BulkTenantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c.Request.Context(), "Invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenantApp.BulkAssignPermissions(c.Request.Context(), req); err != nil {
		logger.Error(c.Request.Context(), "Failed to assign permissions to tenant", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions assigned successfully"})
}

func (h *TenantHandler) BulkRemovePermissions(c *gin.Context) {
	var req model.BulkTenantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c.Request.Context(), "Invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenantApp.BulkRemovePermissions(c.Request.Context(), req); err != nil {
		logger.Error(c.Request.Context(), "Failed to remove permissions from tenant", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions removed successfully"})
}

func (h *TenantHandler) BulkSyncPermissions(c *gin.Context) {
	var req model.BulkTenantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c.Request.Context(), "Invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenantApp.BulkSyncPermissions(c.Request.Context(), req); err != nil {
		logger.Error(c.Request.Context(), "Failed to sync permissions for tenant", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions synced successfully"})
}
