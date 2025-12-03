# API Specification

## Tenant Management

### POST /api/v1/tenant/permissions/add
Add permissions to a tenant.
**Body**:
```json
{
  "tenant_id": "string",
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### POST /api/v1/tenant/permissions/remove
Remove permissions from a tenant.
**Body**:
```json
{
  "tenant_id": "string",
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### PUT /api/v1/tenant/permissions
Sync permissions for a tenant (Replace all).
**Body**:
```json
{
  "tenant_id": "string",
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

## Role Management

### POST /api/v1/roles
Create a new role.
**Body**:
```json
{
  "name": "string",
  "tenant_id": "string" // optional
}
```

### POST /api/v1/roles/:role_id/permissions/add
Add permissions to a role.
**Body**:
```json
{
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### POST /api/v1/roles/:role_id/permissions/remove
Remove permissions from a role.
**Body**:
```json
{
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### PUT /api/v1/roles/:role_id/permissions
Sync permissions for a role (Replace all).
**Body**:
```json
{
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### POST /api/v1/roles/:role_id/users/bulk
Assign users to a role.
**Body**:
```json
{
  "user_ids": ["string"]
}
```
### DELETE /api/v1/roles/:role_id/users/bulk
Remove users from a role.

## Group Management

### POST /api/v1/groups
Create a new group.
**Body**:
```json
{
  "name": "string",
  "tenant_id": "string" // optional
}
```

### POST /api/v1/groups/:group_id/permissions/add
Add permissions to a group.
**Body**:
```json
{
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### POST /api/v1/groups/:group_id/permissions/remove
Remove permissions from a group.
**Body**:
```json
{
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### PUT /api/v1/groups/:group_id/permissions
Sync permissions for a group (Replace all).
**Body**:
```json
{
  "permissions": [
    { "resource_id": "string", "action_id": "string" }
  ]
}
```

### POST /api/v1/groups/:group_id/users/bulk
Assign users to a group.
**Body**:
```json
{
  "user_ids": ["string"]
}
```
### DELETE /api/v1/groups/:group_id/users/bulk
Remove users from a group.

## Validation

### POST /api/v1/check-permission
Check if a user has specific permissions.
**Body**:
```json
{
  "user_id": "string",
  "tenant_id": "string",
  "permissions": [
    { "resource_code": "string", "action_code": "string" }
  ],
  "condition": "AND" // or "OR" - default AND
}
```
**Response**:
```json
{
  "allowed": true
}
```
