package model

type Resource struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Action struct {
	ID          string `json:"id"`
	ResourceID  string `json:"resource_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TenantPermission struct {
	ResourceID string `json:"resource_id"`
	ActionID   string `json:"action_id"`
	TenantID   string `json:"tenant_id"`
}

type Role struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id,omitempty"`
}

type Group struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id,omitempty"`
}

type RolePermission struct {
	RoleID     string `json:"role_id"`
	ResourceID string `json:"resource_id"`
	ActionID   string `json:"action_id"`
}

type GroupPermission struct {
	GroupID    string `json:"group_id"`
	ResourceID string `json:"resource_id"`
	ActionID   string `json:"action_id"`
}

type UserRole struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

type UserGroup struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

// DTOs for API requests

type BulkTenantPermissionRequest struct {
	TenantID    string       `json:"tenant_id"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	ResourceID string `json:"resource_id"`
	ActionID   string `json:"action_id"`
}

type CreateRoleRequest struct {
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
}

type BulkRolePermissionRequest struct {
	Permissions []Permission `json:"permissions"`
}

type BulkUserRoleRequest struct {
	UserIDs []string `json:"user_ids"`
}

type CreateGroupRequest struct {
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
}

type BulkGroupPermissionRequest struct {
	Permissions []Permission `json:"permissions"`
}

type BulkUserGroupRequest struct {
	UserIDs []string `json:"user_ids"`
}

type CheckPermissionRequest struct {
	UserID      string           `json:"user_id"`
	TenantID    string           `json:"tenant_id"`
	Permissions []PermissionCode `json:"permissions"`
	Condition   string           `json:"condition"` // AND / OR
}

type PermissionCode struct {
	ResourceCode string `json:"resource_code"`
	ActionCode   string `json:"action_code"`
}
