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

	// Query to get union of permissions from Roles and Groups
	// Joined with resource_action_tenant to ensure tenant validity
	query := `
		SELECT DISTINCT p.resource_id, p.action_id
		FROM (
			-- Permissions from Roles
			SELECT rp.resource_id, rp.action_id
			FROM pmsn.user_role ur
			JOIN pmsn.role r ON ur.role_id = r.id
			JOIN pmsn.role_permission rp ON r.id = rp.role_id
			WHERE ur.user_id = $1 AND (r.tenant_id = $2 OR r.tenant_id IS NULL)

			UNION

			-- Permissions from Groups
			SELECT gp.resource_id, gp.action_id
			FROM pmsn.user_group ug
			JOIN pmsn.group g ON ug.group_id = g.id
			JOIN pmsn.group_permission gp ON g.id = gp.group_id
			WHERE ug.user_id = $1 AND (g.tenant_id = $2 OR g.tenant_id IS NULL)
		) p
		JOIN pmsn.resource_action_tenant rat ON p.resource_id = rat.resource_id AND p.action_id = rat.action_id
		WHERE rat.tenant_id = $2
	`

	rows, err := pool.Query(ctx, query, userID, tenantID)
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
