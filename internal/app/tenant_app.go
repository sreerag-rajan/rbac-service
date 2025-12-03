package app

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/service"
)

type TenantAppService struct {
	tenantService *service.TenantService
}

func NewTenantAppService(tenantService *service.TenantService) *TenantAppService {
	return &TenantAppService{
		tenantService: tenantService,
	}
}

func (a *TenantAppService) BulkAssignPermissions(ctx context.Context, req model.BulkTenantPermissionRequest) error {
	return a.tenantService.AssignPermissions(ctx, req.TenantID, req.Permissions)
}

func (a *TenantAppService) BulkRemovePermissions(ctx context.Context, req model.BulkTenantPermissionRequest) error {
	return a.tenantService.RemovePermissions(ctx, req.TenantID, req.Permissions)
}

func (a *TenantAppService) BulkSyncPermissions(ctx context.Context, req model.BulkTenantPermissionRequest) error {
	return a.tenantService.SyncPermissions(ctx, req.TenantID, req.Permissions)
}
