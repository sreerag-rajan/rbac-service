package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/jackc/pgx/v5"
)

type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

func (r *RoleRepository) CreateRole(ctx context.Context, role *model.Role) error {
	pool := GetPool()
	_, err := pool.Exec(ctx, "INSERT INTO pmsn.role (id, name, tenant_id) VALUES ($1, $2, $3)", role.ID, role.Name, role.TenantID)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *RoleRepository) BulkAssignPermissions(ctx context.Context, roleID string, permissions []model.Permission) error {
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
		batch.Queue("INSERT INTO pmsn.role_permission (role_id, resource_id, action_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", roleID, p.ResourceID, p.ActionID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(permissions); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to assign permission to role", err, "role_id", roleID)
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

func (r *RoleRepository) BulkRemovePermissions(ctx context.Context, roleID string, permissions []model.Permission) error {
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
		batch.Queue("DELETE FROM pmsn.role_permission WHERE role_id = $1 AND resource_id = $2 AND action_id = $3", roleID, p.ResourceID, p.ActionID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(permissions); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to remove permission from role", err, "role_id", roleID)
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

func (r *RoleRepository) BulkSyncPermissions(ctx context.Context, roleID string, permissions []model.Permission) error {
	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delete all existing permissions for the role
	if _, err := tx.Exec(ctx, "DELETE FROM pmsn.role_permission WHERE role_id = $1", roleID); err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// 2. Insert new permissions
	if len(permissions) > 0 {
		batch := &pgx.Batch{}
		for _, p := range permissions {
			batch.Queue("INSERT INTO pmsn.role_permission (role_id, resource_id, action_id) VALUES ($1, $2, $3)", roleID, p.ResourceID, p.ActionID)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for i := 0; i < len(permissions); i++ {
			if _, err := br.Exec(); err != nil {
				logger.Error(ctx, "Failed to insert permission for role sync", err, "role_id", roleID)
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

func (r *RoleRepository) BulkAssignUsers(ctx context.Context, roleID string, userIDs []string) error {
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
		batch.Queue("INSERT INTO pmsn.user_role (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", uid, roleID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(userIDs); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to assign user to role", err, "role_id", roleID)
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

func (r *RoleRepository) BulkRemoveUsers(ctx context.Context, roleID string, userIDs []string) error {
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
		batch.Queue("DELETE FROM pmsn.user_role WHERE user_id = $1 AND role_id = $2", uid, roleID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(userIDs); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to remove user from role", err, "role_id", roleID)
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
