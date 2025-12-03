package app

import (
	"context"
	"rbac-service/internal/model"
	"rbac-service/internal/service"
)

type GroupAppService struct {
	groupService *service.GroupService
	publisher    EventPublisher
}

func NewGroupAppService(groupService *service.GroupService, publisher EventPublisher) *GroupAppService {
	return &GroupAppService{
		groupService: groupService,
		publisher:    publisher,
	}
}

func (a *GroupAppService) CreateGroup(ctx context.Context, req model.CreateGroupRequest) (*model.Group, error) {
	return a.groupService.CreateGroup(ctx, req.Name, req.TenantID)
}

func (a *GroupAppService) BulkAssignPermissions(ctx context.Context, groupID string, req model.BulkGroupPermissionRequest) error {
	return a.groupService.AssignPermissions(ctx, groupID, req.Permissions)
}

func (a *GroupAppService) BulkRemovePermissions(ctx context.Context, groupID string, req model.BulkGroupPermissionRequest) error {
	return a.groupService.RemovePermissions(ctx, groupID, req.Permissions)
}

func (a *GroupAppService) BulkSyncPermissions(ctx context.Context, groupID string, req model.BulkGroupPermissionRequest) error {
	return a.groupService.SyncPermissions(ctx, groupID, req.Permissions)
}

func (a *GroupAppService) BulkAssignUsers(ctx context.Context, groupID string, req model.BulkUserGroupRequest) error {
	err := a.groupService.AssignUsers(ctx, groupID, req.UserIDs)
	if err != nil {
		return err
	}

	if a.publisher != nil {
		payload := map[string]interface{}{
			"group_id": groupID,
			"user_ids": req.UserIDs,
		}
		_ = a.publisher.Publish(ctx, "rbac.user_group.assign.success", payload)
	}

	return nil
}

func (a *GroupAppService) BulkRemoveUsers(ctx context.Context, groupID string, req model.BulkUserGroupRequest) error {
	err := a.groupService.RemoveUsers(ctx, groupID, req.UserIDs)
	if err != nil {
		return err
	}

	if a.publisher != nil {
		payload := map[string]interface{}{
			"group_id": groupID,
			"user_ids": req.UserIDs,
		}
		_ = a.publisher.Publish(ctx, "rbac.user_group.remove.success", payload)
	}

	return nil
}
