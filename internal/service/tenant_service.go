package service

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"
)

type TenantService struct {
	tenantRepo *repository.TenantRepository
}

func NewTenantService(tenantRepo *repository.TenantRepository) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
	}
}

func (s *TenantService) AssignPermissions(ctx context.Context, tenantID string, permissions []model.Permission) error {
	return s.tenantRepo.BulkAssignPermissions(ctx, tenantID, permissions)
}

func (s *TenantService) RemovePermissions(ctx context.Context, tenantID string, permissions []model.Permission) error {
	return s.tenantRepo.BulkRemovePermissions(ctx, tenantID, permissions)
}

func (s *TenantService) SyncPermissions(ctx context.Context, tenantID string, permissions []model.Permission) error {
	return s.tenantRepo.BulkSyncPermissions(ctx, tenantID, permissions)
}
