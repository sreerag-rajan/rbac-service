package controller

import (
	"net/http"
	"rbac-service/internal/app"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/gin-gonic/gin"
)

type ValidationHandler struct {
	validationApp *app.ValidationAppService
}

func NewValidationHandler(validationApp *app.ValidationAppService) *ValidationHandler {
	return &ValidationHandler{
		validationApp: validationApp,
	}
}

func (h *ValidationHandler) CheckPermission(c *gin.Context) {
	var req model.CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowed, err := h.validationApp.CheckPermission(c.Request.Context(), req)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to check permission", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"allowed": allowed})
}
