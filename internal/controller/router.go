package controller

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	tenantHandler *TenantHandler,
	roleHandler *RoleHandler,
	groupHandler *GroupHandler,
	validationHandler *ValidationHandler,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		// Tenant
		v1.POST("/tenant/permissions/bulk", tenantHandler.BulkAssignPermissions) // Deprecated: Use /add
		v1.POST("/tenant/permissions/add", tenantHandler.BulkAssignPermissions)
		v1.POST("/tenant/permissions/remove", tenantHandler.BulkRemovePermissions)
		v1.PUT("/tenant/permissions", tenantHandler.BulkSyncPermissions)

		// Roles
		v1.POST("/roles", roleHandler.CreateRole)
		v1.POST("/roles/:role_id/permissions/bulk", roleHandler.BulkAssignPermissions) // Deprecated: Use /add
		v1.POST("/roles/:role_id/permissions/add", roleHandler.BulkAssignPermissions)
		v1.POST("/roles/:role_id/permissions/remove", roleHandler.BulkRemovePermissions)
		v1.PUT("/roles/:role_id/permissions", roleHandler.BulkSyncPermissions)
		v1.POST("/roles/:role_id/users/bulk", roleHandler.BulkAssignUsers)
		v1.DELETE("/roles/:role_id/users/bulk", roleHandler.BulkRemoveUsers)

		// Groups
		v1.POST("/groups", groupHandler.CreateGroup)
		v1.POST("/groups/:group_id/permissions/bulk", groupHandler.BulkAssignPermissions) // Deprecated: Use /add
		v1.POST("/groups/:group_id/permissions/add", groupHandler.BulkAssignPermissions)
		v1.POST("/groups/:group_id/permissions/remove", groupHandler.BulkRemovePermissions)
		v1.PUT("/groups/:group_id/permissions", groupHandler.BulkSyncPermissions)
		v1.POST("/groups/:group_id/users/bulk", groupHandler.BulkAssignUsers)
		v1.DELETE("/groups/:group_id/users/bulk", groupHandler.BulkRemoveUsers)

		// Validation
		v1.POST("/check-permission", validationHandler.CheckPermission)
	}

	return r
}
