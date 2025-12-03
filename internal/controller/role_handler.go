package controller

import (
	"net/http"
	"rbac-service/internal/app"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleApp *app.RoleAppService
}

func NewRoleHandler(roleApp *app.RoleAppService) *RoleHandler {
	return &RoleHandler{
		roleApp: roleApp,
	}
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.roleApp.CreateRole(c.Request.Context(), req)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to create role", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) BulkAssignPermissions(c *gin.Context) {
	roleID := c.Param("role_id")
	var req model.BulkRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleApp.BulkAssignPermissions(c.Request.Context(), roleID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to assign permissions to role", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions assigned successfully"})
}

func (h *RoleHandler) BulkRemovePermissions(c *gin.Context) {
	roleID := c.Param("role_id")
	var req model.BulkRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleApp.BulkRemovePermissions(c.Request.Context(), roleID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to remove permissions from role", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions removed successfully"})
}

func (h *RoleHandler) BulkSyncPermissions(c *gin.Context) {
	roleID := c.Param("role_id")
	var req model.BulkRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleApp.BulkSyncPermissions(c.Request.Context(), roleID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to sync permissions for role", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions synced successfully"})
}

func (h *RoleHandler) BulkAssignUsers(c *gin.Context) {
	roleID := c.Param("role_id")
	var req model.BulkUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleApp.BulkAssignUsers(c.Request.Context(), roleID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to assign users to role", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users assigned successfully"})
}

func (h *RoleHandler) BulkRemoveUsers(c *gin.Context) {
	roleID := c.Param("role_id")
	var req model.BulkUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleApp.BulkRemoveUsers(c.Request.Context(), roleID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to remove users from role", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users removed successfully"})
}
