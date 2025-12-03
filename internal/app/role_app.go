package app

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/service"
)

type RoleAppService struct {
	roleService *service.RoleService
	publisher   EventPublisher
}

func NewRoleAppService(roleService *service.RoleService, publisher EventPublisher) *RoleAppService {
	return &RoleAppService{
		roleService: roleService,
		publisher:   publisher,
	}
}

func (a *RoleAppService) CreateRole(ctx context.Context, req model.CreateRoleRequest) (*model.Role, error) {
	return a.roleService.CreateRole(ctx, req.Name, req.TenantID)
}

func (a *RoleAppService) BulkAssignPermissions(ctx context.Context, roleID string, req model.BulkRolePermissionRequest) error {
	return a.roleService.AssignPermissions(ctx, roleID, req.Permissions)
}

func (a *RoleAppService) BulkRemovePermissions(ctx context.Context, roleID string, req model.BulkRolePermissionRequest) error {
	return a.roleService.RemovePermissions(ctx, roleID, req.Permissions)
}

func (a *RoleAppService) BulkSyncPermissions(ctx context.Context, roleID string, req model.BulkRolePermissionRequest) error {
	return a.roleService.SyncPermissions(ctx, roleID, req.Permissions)
}

func (a *RoleAppService) BulkAssignUsers(ctx context.Context, roleID string, req model.BulkUserRoleRequest) error {
	err := a.roleService.AssignUsers(ctx, roleID, req.UserIDs)
	if err != nil {
		return err
	}

	if a.publisher != nil {
		payload := map[string]interface{}{
			"role_id":  roleID,
			"user_ids": req.UserIDs,
		}
		_ = a.publisher.Publish(ctx, "rbac.user_role.assign.success", payload)
	}

	return nil
}

func (a *RoleAppService) BulkRemoveUsers(ctx context.Context, roleID string, req model.BulkUserRoleRequest) error {
	err := a.roleService.RemoveUsers(ctx, roleID, req.UserIDs)
	if err != nil {
		return err
	}

	if a.publisher != nil {
		payload := map[string]interface{}{
			"role_id":  roleID,
			"user_ids": req.UserIDs,
		}
		_ = a.publisher.Publish(ctx, "rbac.user_role.remove.success", payload)
	}

	return nil
}
