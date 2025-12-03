package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/jackc/pgx/v5"
)

type GroupRepository struct{}

func NewGroupRepository() *GroupRepository {
	return &GroupRepository{}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, group *model.Group) error {
	pool := GetPool()
	_, err := pool.Exec(ctx, "INSERT INTO pmsn.group (id, name, tenant_id) VALUES ($1, $2, $3)", group.ID, group.Name, group.TenantID)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (r *GroupRepository) BulkAssignPermissions(ctx context.Context, groupID string, permissions []model.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, p := range permissions {
		batch.Queue("INSERT INTO pmsn.group_permission (group_id, resource_id, action_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", groupID, p.ResourceID, p.ActionID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(permissions); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to assign permission to group", err, "group_id", groupID)
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *GroupRepository) BulkRemovePermissions(ctx context.Context, groupID string, permissions []model.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, p := range permissions {
		batch.Queue("DELETE FROM pmsn.group_permission WHERE group_id = $1 AND resource_id = $2 AND action_id = $3", groupID, p.ResourceID, p.ActionID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(permissions); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to remove permission from group", err, "group_id", groupID)
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *GroupRepository) BulkSyncPermissions(ctx context.Context, groupID string, permissions []model.Permission) error {
	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delete all existing permissions for the group
	if _, err := tx.Exec(ctx, "DELETE FROM pmsn.group_permission WHERE group_id = $1", groupID); err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// 2. Insert new permissions
	if len(permissions) > 0 {
		batch := &pgx.Batch{}
		for _, p := range permissions {
			batch.Queue("INSERT INTO pmsn.group_permission (group_id, resource_id, action_id) VALUES ($1, $2, $3)", groupID, p.ResourceID, p.ActionID)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for i := 0; i < len(permissions); i++ {
			if _, err := br.Exec(); err != nil {
				logger.Error(ctx, "Failed to insert permission for group sync", err, "group_id", groupID)
				return fmt.Errorf("failed to execute batch: %w", err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("failed to close batch results: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *GroupRepository) BulkAssignUsers(ctx context.Context, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, uid := range userIDs {
		batch.Queue("INSERT INTO pmsn.user_group (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", uid, groupID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(userIDs); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to assign user to group", err, "group_id", groupID)
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *GroupRepository) BulkRemoveUsers(ctx context.Context, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, uid := range userIDs {
		batch.Queue("DELETE FROM pmsn.user_group WHERE user_id = $1 AND group_id = $2", uid, groupID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(userIDs); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to remove user from group", err, "group_id", groupID)
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
