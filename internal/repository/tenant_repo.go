package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"

	"github.com/jackc/pgx/v5"
)

type TenantRepository struct{}

func NewTenantRepository() *TenantRepository {
	return &TenantRepository{}
}

func (r *TenantRepository) BulkAssignPermissions(ctx context.Context, tenantID string, permissions []model.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare batch insert
	batch := &pgx.Batch{}
	for _, p := range permissions {
		batch.Queue("INSERT INTO pmsn.resource_action_tenant (resource_id, action_id, tenant_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", p.ResourceID, p.ActionID, tenantID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(permissions); i++ {
		_, err := br.Exec()
		if err != nil {
			logger.Error(ctx, "Failed to execute batch insert for tenant permission", err, "tenant_id", tenantID)
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

func (r *TenantRepository) BulkRemovePermissions(ctx context.Context, tenantID string, permissions []model.Permission) error {
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
		batch.Queue("DELETE FROM pmsn.resource_action_tenant WHERE resource_id = $1 AND action_id = $2 AND tenant_id = $3", p.ResourceID, p.ActionID, tenantID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(permissions); i++ {
		if _, err := br.Exec(); err != nil {
			logger.Error(ctx, "Failed to remove permission from tenant", err, "tenant_id", tenantID)
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

func (r *TenantRepository) BulkSyncPermissions(ctx context.Context, tenantID string, permissions []model.Permission) error {
	pool := GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delete all existing permissions for the tenant
	if _, err := tx.Exec(ctx, "DELETE FROM pmsn.resource_action_tenant WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// 2. Insert new permissions
	if len(permissions) > 0 {
		batch := &pgx.Batch{}
		for _, p := range permissions {
			batch.Queue("INSERT INTO pmsn.resource_action_tenant (resource_id, action_id, tenant_id) VALUES ($1, $2, $3)", p.ResourceID, p.ActionID, tenantID)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for i := 0; i < len(permissions); i++ {
			if _, err := br.Exec(); err != nil {
				logger.Error(ctx, "Failed to insert permission for tenant sync", err, "tenant_id", tenantID)
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

func (r *TenantRepository) GetTenantPermissions(ctx context.Context, tenantID string) ([]model.Permission, error) {
	pool := GetPool()
	rows, err := pool.Query(ctx, "SELECT resource_id, action_id FROM pmsn.resource_action_tenant WHERE tenant_id = $1", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant permissions: %w", err)
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ResourceID, &p.ActionID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}
