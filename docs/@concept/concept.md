# RBAC Service Concepts

## Overview
This service provides Role-Based Access Control (RBAC) with multi-tenancy support. It manages resources, actions, roles, groups, and their associations to determine user permissions.

## Core Entities

### Resource & Action
- **Resource**: Represents an entity in the system (e.g., "Order", "User").
- **Action**: Represents an operation on a resource (e.g., "Create", "Read").
- **Resource-Action**: A specific action on a specific resource constitutes a base permission unit.

### Multi-Tenancy
- **Tenant**: An isolation unit.
- **Resource-Action-Tenant**: Defines which permissions are available/relevant for a specific tenant. A tenant cannot have permissions that are not mapped here.

### Roles & Groups
- **Role**: A collection of permissions (Resource-Action pairs). Can be tenant-specific.
- **Group**: A collection of users and permissions. Can be tenant-specific.
- **User**: An external entity (UUID) assigned to Roles and Groups.

## Permission Resolution
- **Additive Model**: Permissions are additive. If a user has a permission via *any* assigned Role or Group, they have that permission.
- **Default Deny**: If no Role or Group grants the permission, the user does not have it.
- **Conflict Resolution**: Since permissions are additive, "True" overrides "False" (absence).

## User Validation
To validate if a user `U` has permission `P` (Resource `R` + Action `A`) in Tenant `T`:
1.  Find all Roles and Groups assigned to `U`.
2.  Filter Roles/Groups relevant to Tenant `T` (or global).
3.  Check if `P` exists in the permission set of any of these Roles or Groups.
4.  Return `True` if found, `False` otherwise.
