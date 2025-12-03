package service

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"

	"github.com/google/uuid"
)

type RoleService struct {
	roleRepo *repository.RoleRepository
}

func NewRoleService(roleRepo *repository.RoleRepository) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, name, tenantID string) (*model.Role, error) {
	role := &model.Role{
		ID:       uuid.New().String(),
		Name:     name,
		TenantID: tenantID,
	}
	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, permissions []model.Permission) error {
	return s.roleRepo.BulkAssignPermissions(ctx, roleID, permissions)
}

func (s *RoleService) RemovePermissions(ctx context.Context, roleID string, permissions []model.Permission) error {
	return s.roleRepo.BulkRemovePermissions(ctx, roleID, permissions)
}

func (s *RoleService) SyncPermissions(ctx context.Context, roleID string, permissions []model.Permission) error {
	return s.roleRepo.BulkSyncPermissions(ctx, roleID, permissions)
}

func (s *RoleService) AssignUsers(ctx context.Context, roleID string, userIDs []string) error {
	return s.roleRepo.BulkAssignUsers(ctx, roleID, userIDs)
}

func (s *RoleService) RemoveUsers(ctx context.Context, roleID string, userIDs []string) error {
	return s.roleRepo.BulkRemoveUsers(ctx, roleID, userIDs)
}
