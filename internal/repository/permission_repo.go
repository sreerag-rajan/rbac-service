package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/model"

	"github.com/jackc/pgx/v5"
)

type PermissionRepository struct{}

func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{}
}

// GetUserPermissions fetches all permissions for a user within a tenant context.
// It considers permissions from Roles and Groups assigned to the user.
// It also enforces that the permission must be valid for the tenant (present in resource_action_tenant).
func (r *PermissionRepository) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]model.Permission, error) {
	pool := GetPool()
	var query string
	var args []interface{}

	if tenantID == "" {
		// Global permissions only:
		// 1. Roles with tenant_id IS NULL
		// 2. Groups with tenant_id IS NULL
		// 3. NO check against resource_action_tenant (global admins bypass tenant feature flags)
		query = `
			SELECT DISTINCT resource_id, action_id FROM (
				SELECT rp.resource_id, rp.action_id
				FROM pmsn.user_role ur
				JOIN pmsn.role r ON ur.role_id = r.id
				JOIN pmsn.role_permission rp ON r.id = rp.role_id
				WHERE ur.user_id = $1 AND r.tenant_id IS NULL

				UNION

				SELECT gp.resource_id, gp.action_id
				FROM pmsn.user_group ug
				JOIN pmsn.group g ON ug.group_id = g.id
				JOIN pmsn.group_permission gp ON g.id = gp.group_id
				WHERE ug.user_id = $1 AND g.tenant_id IS NULL
			) p
		`
		args = []interface{}{userID}
	} else {
		// Tenant-scoped permissions:
		// 1. Roles with tenant_id = $2 OR NULL
		// 2. Groups with tenant_id = $2 OR NULL
		// 3. MUST exist in resource_action_tenant for $2
		query = `
			SELECT DISTINCT p.resource_id, p.action_id
			FROM (
				SELECT rp.resource_id, rp.action_id
				FROM pmsn.user_role ur
				JOIN pmsn.role r ON ur.role_id = r.id
				JOIN pmsn.role_permission rp ON r.id = rp.role_id
				WHERE ur.user_id = $1 AND (r.tenant_id = $2 OR r.tenant_id IS NULL)

				UNION

				SELECT gp.resource_id, gp.action_id
				FROM pmsn.user_group ug
				JOIN pmsn.group g ON ug.group_id = g.id
				JOIN pmsn.group_permission gp ON g.id = gp.group_id
				WHERE ug.user_id = $1 AND (g.tenant_id = $2 OR g.tenant_id IS NULL)
			) p
			JOIN pmsn.resource_action_tenant rat ON p.resource_id = rat.resource_id AND p.action_id = rat.action_id
			WHERE rat.tenant_id = $2
		`
		args = []interface{}{userID, tenantID}
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ResourceID, &p.ActionID); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// IsUserAssociatedWithTenant checks if a user has any role or group associated with the given tenant.
func (r *PermissionRepository) IsUserAssociatedWithTenant(ctx context.Context, userID, tenantID string) (bool, error) {
	pool := GetPool()
	query := `
		SELECT 1
		FROM (
			SELECT r.tenant_id FROM pmsn.user_role ur JOIN pmsn.role r ON ur.role_id = r.id WHERE ur.user_id = $1
			UNION
			SELECT g.tenant_id FROM pmsn.user_group ug JOIN pmsn.group g ON ug.group_id = g.id WHERE ug.user_id = $1
		) t
		WHERE t.tenant_id = $2
		LIMIT 1
	`
	var exists int
	err := pool.QueryRow(ctx, query, userID, tenantID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check user association: %w", err)
	}
	return true, nil
}

// CheckMiddlewarePermissions performs an optimized check for:
// 1. Global Permission (tenant_id IS NULL)
// 2. Tenant Permission (tenant_id = $2)
// 3. Associated Permission (tenant_id = $2 AND user associated with tenant)
func (r *PermissionRepository) CheckMiddlewarePermissions(ctx context.Context, userID string, tenantID *string, permRes, permAct, assocRes, assocAct string) (bool, error) {
	pool := GetPool()
	query := `
		WITH 
			target_perm AS (
				SELECT r.id as res_id, a.id as act_id
				FROM pmsn.resource r JOIN pmsn.action a ON r.id = a.resource_id
				WHERE r.code = $3 AND a.code = $4
			),
			assoc_perm AS (
				SELECT r.id as res_id, a.id as act_id
				FROM pmsn.resource r JOIN pmsn.action a ON r.id = a.resource_id
				WHERE r.code = $5 AND a.code = $6
			),
			user_roles AS (
				SELECT r.id, r.tenant_id
				FROM pmsn.user_role ur JOIN pmsn.role r ON ur.role_id = r.id
				WHERE ur.user_id = $1
			),
			user_groups AS (
				SELECT g.id, g.tenant_id
				FROM pmsn.user_group ug JOIN pmsn.group g ON ug.group_id = g.id
				WHERE ug.user_id = $1
			),
			raw_perms AS (
				SELECT rp.resource_id, rp.action_id, ur.tenant_id
				FROM user_roles ur JOIN pmsn.role_permission rp ON ur.id = rp.role_id
				UNION
				SELECT gp.resource_id, gp.action_id, ug.tenant_id
				FROM user_groups ug JOIN pmsn.group_permission gp ON ug.id = gp.group_id
			)
		SELECT EXISTS (
			-- 1. Global Permission Check
			SELECT 1 FROM raw_perms p, target_perm tp
			WHERE p.resource_id = tp.res_id AND p.action_id = tp.act_id AND p.tenant_id IS NULL
			
			UNION
			
			-- 2. Tenant Permission Check
			SELECT 1 FROM raw_perms p, target_perm tp, pmsn.resource_action_tenant rat
			WHERE $2 IS NOT NULL 
			AND p.resource_id = tp.res_id AND p.action_id = tp.act_id 
			AND (p.tenant_id = $2 OR p.tenant_id IS NULL)
			AND rat.resource_id = tp.res_id AND rat.action_id = tp.act_id AND rat.tenant_id = $2
			
			UNION
			
			-- 3. Associated Permission Check
			SELECT 1 
			WHERE $2 IS NOT NULL
			AND EXISTS (
				SELECT 1 FROM user_roles WHERE tenant_id = $2
				UNION
				SELECT 1 FROM user_groups WHERE tenant_id = $2
			)
			AND EXISTS (
				SELECT 1 FROM raw_perms p, assoc_perm ap, pmsn.resource_action_tenant rat
				WHERE p.resource_id = ap.res_id AND p.action_id = ap.act_id 
				AND (p.tenant_id = $2 OR p.tenant_id IS NULL)
				AND rat.resource_id = ap.res_id AND rat.action_id = ap.act_id AND rat.tenant_id = $2
			)
		)
	`

	var allowed bool
	err := pool.QueryRow(ctx, query, userID, tenantID, permRes, permAct, assocRes, assocAct).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}
	return allowed, nil
}
