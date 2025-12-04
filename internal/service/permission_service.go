package service

import (
	"context"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"
)

type PermissionService struct {
	permRepo *repository.PermissionRepository
	resRepo  *repository.ResourceRepository
}

func NewPermissionService(permRepo *repository.PermissionRepository, resRepo *repository.ResourceRepository) *PermissionService {
	return &PermissionService{
		permRepo: permRepo,
		resRepo:  resRepo,
	}
}

func (s *PermissionService) CheckPermission(ctx context.Context, req model.CheckPermissionRequest) (bool, error) {
	// 1. Fetch all effective permissions for the user in the tenant
	userPerms, err := s.permRepo.GetUserPermissions(ctx, req.UserID, req.TenantID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Create a map for O(1) lookup
	userPermMap := make(map[string]bool)
	for _, p := range userPerms {
		key := fmt.Sprintf("%s:%s", p.ResourceID, p.ActionID)
		userPermMap[key] = true
	}

	// 2. Resolve requested codes to IDs and check
	matches := 0
	for _, pCode := range req.Permissions {
		// Resolve Resource Code -> ID
		res, err := s.resRepo.GetResourceByCode(ctx, pCode.ResourceCode)
		if err != nil {
			logger.Error(ctx, "Failed to resolve resource code", err, "code", pCode.ResourceCode)
			return false, fmt.Errorf("invalid resource code: %s", pCode.ResourceCode)
		}

		// Resolve Action Code -> ID
		act, err := s.resRepo.GetActionByCode(ctx, res.ID, pCode.ActionCode)
		if err != nil {
			logger.Error(ctx, "Failed to resolve action code", err, "resource_code", pCode.ResourceCode, "action_code", pCode.ActionCode)
			return false, fmt.Errorf("invalid action code: %s for resource: %s", pCode.ActionCode, pCode.ResourceCode)
		}

		key := fmt.Sprintf("%s:%s", res.ID, act.ID)
		if userPermMap[key] {
			matches++
		}
	}

	// 3. Apply Condition
	if req.Condition == "OR" {
		return matches > 0, nil
	} else {
		// Default to AND
		return matches == len(req.Permissions), nil
	}
}

func (s *PermissionService) IsUserAssociatedWithTenant(ctx context.Context, userID, tenantID string) (bool, error) {
	return s.permRepo.IsUserAssociatedWithTenant(ctx, userID, tenantID)
}

func (s *PermissionService) CheckMiddlewarePermissions(ctx context.Context, userID string, tenantID *string, permRes, permAct, assocRes, assocAct string) (bool, error) {
	return s.permRepo.CheckMiddlewarePermissions(ctx, userID, tenantID, permRes, permAct, assocRes, assocAct)
}
