package controller

import (
	"rbac-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	tenantHandler *TenantHandler,
	roleHandler *RoleHandler,
	groupHandler *GroupHandler,
	validationHandler *ValidationHandler,
	permMiddleware *middleware.PermissionMiddleware,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		// Tenant
		tenant := v1.Group("/tenant")
		tenant.Use(permMiddleware.RequirePermission("tenant_permission.manage", "tenant_permission.manage")) // No associated variant for tenant permissions explicitly defined as different logic, but using same code for now or maybe just one. User said "tenant_permission.manage".
		{
			tenant.POST("/permissions/bulk", tenantHandler.BulkAssignPermissions) // Deprecated
			tenant.POST("/permissions/add", tenantHandler.BulkAssignPermissions)
			tenant.POST("/permissions/remove", tenantHandler.BulkRemovePermissions)
			tenant.PUT("/permissions", tenantHandler.BulkSyncPermissions)
		}

		// Roles
		roles := v1.Group("/roles")
		{
			roles.POST("", permMiddleware.RequirePermission("role.manage", "role.manage_tenant_associated"), roleHandler.CreateRole)

			rolePerms := roles.Group("/:role_id")
			rolePerms.Use(permMiddleware.RequirePermission("role.manage_permissions", "role.manage_permissions_tenant_associated"))
			{
				rolePerms.POST("/permissions/bulk", roleHandler.BulkAssignPermissions) // Deprecated
				rolePerms.POST("/permissions/add", roleHandler.BulkAssignPermissions)
				rolePerms.POST("/permissions/remove", roleHandler.BulkRemovePermissions)
				rolePerms.PUT("/permissions", roleHandler.BulkSyncPermissions)
			}

			roleUsers := roles.Group("/:role_id/users")
			roleUsers.Use(permMiddleware.RequirePermission("role.manage", "role.manage_tenant_associated")) // Assuming role management covers user assignment
			{
				roleUsers.POST("/bulk", roleHandler.BulkAssignUsers)
				roleUsers.DELETE("/bulk", roleHandler.BulkRemoveUsers)
			}
		}

		// Groups
		groups := v1.Group("/groups")
		{
			groups.POST("", permMiddleware.RequirePermission("group.manage", "group.manage_tenant_associated"), groupHandler.CreateGroup)

			groupPerms := groups.Group("/:group_id")
			groupPerms.Use(permMiddleware.RequirePermission("group.manage_permissions", "group.manage_permissions_tenant_associated"))
			{
				groupPerms.POST("/permissions/bulk", groupHandler.BulkAssignPermissions) // Deprecated
				groupPerms.POST("/permissions/add", groupHandler.BulkAssignPermissions)
				groupPerms.POST("/permissions/remove", groupHandler.BulkRemovePermissions)
				groupPerms.PUT("/permissions", groupHandler.BulkSyncPermissions)
			}

			groupUsers := groups.Group("/:group_id/users")
			groupUsers.Use(permMiddleware.RequirePermission("group.manage", "group.manage_tenant_associated")) // Assuming group management covers user assignment
			{
				groupUsers.POST("/bulk", groupHandler.BulkAssignUsers)
				groupUsers.DELETE("/bulk", groupHandler.BulkRemoveUsers)
			}
		}

		// Validation
		v1.POST("/check-permission", validationHandler.CheckPermission)
	}

	return r
}
