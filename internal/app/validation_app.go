package app

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/service"
)

type ValidationAppService struct {
	permService *service.PermissionService
}

func NewValidationAppService(permService *service.PermissionService) *ValidationAppService {
	return &ValidationAppService{
		permService: permService,
	}
}

func (a *ValidationAppService) CheckPermission(ctx context.Context, req model.CheckPermissionRequest) (bool, error) {
	return a.permService.CheckPermission(ctx, req)
}
