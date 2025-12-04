BEGIN;

-- Migration 004: Materialized View for User Permissions
-- This view pre-computes all user permissions for fast lookup

-- Create the materialized view
CREATE MATERIALIZED VIEW pmsn.mv_user_permissions AS
SELECT 
    ur.user_id,
    r.tenant_id,
    rp.resource_id,
    rp.action_id,
    res.code as resource_code,
    act.code as action_code
FROM pmsn.user_role ur
JOIN pmsn.role r ON ur.role_id = r.id
JOIN pmsn.role_permission rp ON r.id = rp.role_id
JOIN pmsn.resource res ON rp.resource_id = res.id
JOIN pmsn.action act ON rp.action_id = act.id

UNION

SELECT 
    ug.user_id,
    g.tenant_id,
    gp.resource_id,
    gp.action_id,
    res.code as resource_code,
    act.code as action_code
FROM pmsn.user_group ug
JOIN pmsn.group g ON ug.group_id = g.id
JOIN pmsn.group_permission gp ON g.id = gp.group_id
JOIN pmsn.resource res ON gp.resource_id = res.id
JOIN pmsn.action act ON gp.action_id = act.id;

-- Create indexes for fast lookups
CREATE UNIQUE INDEX idx_mv_user_perms_unique 
    ON pmsn.mv_user_permissions(user_id, resource_id, action_id, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::UUID));

CREATE INDEX idx_mv_user_perms_lookup 
    ON pmsn.mv_user_permissions(user_id, resource_code, action_code, tenant_id);

CREATE INDEX idx_mv_user_perms_user 
    ON pmsn.mv_user_permissions(user_id);

CREATE INDEX idx_mv_user_perms_tenant 
    ON pmsn.mv_user_permissions(tenant_id) WHERE tenant_id IS NOT NULL;

-- Create function to refresh the materialized view
CREATE OR REPLACE FUNCTION pmsn.refresh_user_permissions()
RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY pmsn.mv_user_permissions;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers on all tables that affect user permissions
CREATE TRIGGER trg_refresh_perms_user_role
AFTER INSERT OR UPDATE OR DELETE ON pmsn.user_role
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_user_group
AFTER INSERT OR UPDATE OR DELETE ON pmsn.user_group
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_role_permission
AFTER INSERT OR UPDATE OR DELETE ON pmsn.role_permission
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_group_permission
AFTER INSERT OR UPDATE OR DELETE ON pmsn.group_permission
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_role
AFTER INSERT OR UPDATE OR DELETE ON pmsn.role
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_group
AFTER INSERT OR UPDATE OR DELETE ON pmsn.group
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_resource
AFTER INSERT OR UPDATE OR DELETE ON pmsn.resource
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

CREATE TRIGGER trg_refresh_perms_action
AFTER INSERT OR UPDATE OR DELETE ON pmsn.action
FOR EACH STATEMENT EXECUTE FUNCTION pmsn.refresh_user_permissions();

COMMIT;
