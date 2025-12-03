package controller

import (
	"net/http"
	"rbac-service/internal/app"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupApp *app.GroupAppService
}

func NewGroupHandler(groupApp *app.GroupAppService) *GroupHandler {
	return &GroupHandler{
		groupApp: groupApp,
	}
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req model.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.groupApp.CreateGroup(c.Request.Context(), req)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to create group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

func (h *GroupHandler) BulkAssignPermissions(c *gin.Context) {
	groupID := c.Param("group_id")
	var req model.BulkGroupPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupApp.BulkAssignPermissions(c.Request.Context(), groupID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to assign permissions to group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions assigned successfully"})
}

func (h *GroupHandler) BulkRemovePermissions(c *gin.Context) {
	groupID := c.Param("group_id")
	var req model.BulkGroupPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupApp.BulkRemovePermissions(c.Request.Context(), groupID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to remove permissions from group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions removed successfully"})
}

func (h *GroupHandler) BulkSyncPermissions(c *gin.Context) {
	groupID := c.Param("group_id")
	var req model.BulkGroupPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupApp.BulkSyncPermissions(c.Request.Context(), groupID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to sync permissions for group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions synced successfully"})
}

func (h *GroupHandler) BulkAssignUsers(c *gin.Context) {
	groupID := c.Param("group_id")
	var req model.BulkUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupApp.BulkAssignUsers(c.Request.Context(), groupID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to assign users to group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users assigned successfully"})
}

func (h *GroupHandler) BulkRemoveUsers(c *gin.Context) {
	groupID := c.Param("group_id")
	var req model.BulkUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupApp.BulkRemoveUsers(c.Request.Context(), groupID, req); err != nil {
		logger.Error(c.Request.Context(), "Failed to remove users from group", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users removed successfully"})
}
