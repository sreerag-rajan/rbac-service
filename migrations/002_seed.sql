BEGIN;

-- Clean up existing data (optional, but good for re-seeding)
TRUNCATE TABLE pmsn.role_permission CASCADE;
TRUNCATE TABLE pmsn.group_permission CASCADE;
TRUNCATE TABLE pmsn.user_role CASCADE;
TRUNCATE TABLE pmsn.user_group CASCADE;
TRUNCATE TABLE pmsn.role CASCADE;
TRUNCATE TABLE pmsn.group CASCADE;
TRUNCATE TABLE pmsn.action CASCADE;
TRUNCATE TABLE pmsn.resource CASCADE;

-- Insert Resources
INSERT INTO pmsn.resource (code, name, description) VALUES
('role', 'Role', 'Role management'),
('group', 'Group', 'Group management'),
('tenant_permission', 'Tenant Permission', 'Tenant permission management')
ON CONFLICT (code) DO NOTHING;

-- Insert Actions
WITH res AS (SELECT id, code FROM pmsn.resource)
INSERT INTO pmsn.action (resource_id, code, name, description) VALUES
-- Role Actions
((SELECT id FROM res WHERE code = 'role'), 'manage', 'Manage Roles', 'Create, delete, update roles'),
((SELECT id FROM res WHERE code = 'role'), 'manage_permissions', 'Manage Role Permissions', 'Assign or remove permissions from a role'),
((SELECT id FROM res WHERE code = 'role'), 'manage_permissions_tenant_associated', 'Manage Associated Role Permissions', 'Assign/remove role permissions for associated tenant only'),
((SELECT id FROM res WHERE code = 'role'), 'manage_tenant_associated', 'Manage Associated Roles', 'Create/delete roles within associated tenant only'),

-- Group Actions
((SELECT id FROM res WHERE code = 'group'), 'manage', 'Manage Groups', 'Create, delete, update groups'),
((SELECT id FROM res WHERE code = 'group'), 'manage_permissions', 'Manage Group Permissions', 'Assign or remove permissions from a group'),
((SELECT id FROM res WHERE code = 'group'), 'manage_permissions_tenant_associated', 'Manage Associated Group Permissions', 'Assign/remove group permissions for associated tenant only'),
((SELECT id FROM res WHERE code = 'group'), 'manage_tenant_associated', 'Manage Associated Groups', 'Create/delete groups within associated tenant only'),

-- Tenant Permission Actions
((SELECT id FROM res WHERE code = 'tenant_permission'), 'manage', 'Manage Tenant Permissions', 'Manage permissions in a tenant')
ON CONFLICT (resource_id, code) DO NOTHING;

-- Create Superadmin Role
INSERT INTO pmsn.role (name, tenant_id) VALUES
('superadmin', NULL) -- Global role
ON CONFLICT DO NOTHING;

-- Assign All Permissions to Superadmin
WITH sa_role AS (
    SELECT id FROM pmsn.role WHERE name = 'superadmin' LIMIT 1
),
all_actions AS (
    SELECT r.id as resource_id, a.id as action_id
    FROM pmsn.resource r
    JOIN pmsn.action a ON r.id = a.resource_id
)
INSERT INTO pmsn.role_permission (role_id, resource_id, action_id)
SELECT sa_role.id, all_actions.resource_id, all_actions.action_id
FROM sa_role, all_actions
ON CONFLICT DO NOTHING;

COMMIT;
