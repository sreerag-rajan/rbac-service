package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/model"
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

// CheckMiddlewarePermissionsWithMV performs an optimized check using the materialized view.
// This is significantly faster than CheckMiddlewarePermissions as it uses pre-computed permissions.
func (r *PermissionRepository) CheckMiddlewarePermissionsWithMV(ctx context.Context, userID string, tenantID *string, permRes, permAct, assocRes, assocAct string) (bool, error) {
	pool := GetPool()

	query := `
		SELECT EXISTS (
			-- 1. Global Permission Check
			SELECT 1 FROM pmsn.mv_user_permissions mvp
			WHERE mvp.user_id = $1 
			AND mvp.resource_code = $2 
			AND mvp.action_code = $3
			AND mvp.tenant_id IS NULL
			
			UNION
			
			-- 2. Tenant Permission Check (if tenant provided)
			SELECT 1 FROM pmsn.mv_user_permissions mvp
			JOIN pmsn.resource_action_tenant rat 
				ON mvp.resource_id = rat.resource_id 
				AND mvp.action_id = rat.action_id
			WHERE $4 IS NOT NULL
			AND mvp.user_id = $1 
			AND mvp.resource_code = $2 
			AND mvp.action_code = $3
			AND (mvp.tenant_id = $4 OR mvp.tenant_id IS NULL)
			AND rat.tenant_id = $4
			
			UNION
			
			-- 3. Associated Permission Check
			SELECT 1 
			WHERE $4 IS NOT NULL
			AND EXISTS (
				-- Check user is associated with tenant
				SELECT 1 FROM pmsn.mv_user_permissions
				WHERE user_id = $1 AND tenant_id = $4
				LIMIT 1
			)
			AND EXISTS (
				-- Check user has associated permission
				SELECT 1 FROM pmsn.mv_user_permissions mvp
				JOIN pmsn.resource_action_tenant rat 
					ON mvp.resource_id = rat.resource_id 
					AND mvp.action_id = rat.action_id
				WHERE mvp.user_id = $1 
				AND mvp.resource_code = $5 
				AND mvp.action_code = $6
				AND (mvp.tenant_id = $4 OR mvp.tenant_id IS NULL)
				AND rat.tenant_id = $4
			)
		)
	`

	var allowed bool
	err := pool.QueryRow(ctx, query, userID, permRes, permAct, tenantID, assocRes, assocAct).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions with MV: %w", err)
	}
	return allowed, nil
}
