package service

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"

	"github.com/google/uuid"
)

type GroupService struct {
	groupRepo *repository.GroupRepository
}

func NewGroupService(groupRepo *repository.GroupRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, name, tenantID string) (*model.Group, error) {
	group := &model.Group{
		ID:       uuid.New().String(),
		Name:     name,
		TenantID: tenantID,
	}
	if err := s.groupRepo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) AssignPermissions(ctx context.Context, groupID string, permissions []model.Permission) error {
	return s.groupRepo.BulkAssignPermissions(ctx, groupID, permissions)
}

func (s *GroupService) RemovePermissions(ctx context.Context, groupID string, permissions []model.Permission) error {
	return s.groupRepo.BulkRemovePermissions(ctx, groupID, permissions)
}

func (s *GroupService) SyncPermissions(ctx context.Context, groupID string, permissions []model.Permission) error {
	return s.groupRepo.BulkSyncPermissions(ctx, groupID, permissions)
}

func (s *GroupService) AssignUsers(ctx context.Context, groupID string, userIDs []string) error {
	return s.groupRepo.BulkAssignUsers(ctx, groupID, userIDs)
}

func (s *GroupService) RemoveUsers(ctx context.Context, groupID string, userIDs []string) error {
	return s.groupRepo.BulkRemoveUsers(ctx, groupID, userIDs)
}
