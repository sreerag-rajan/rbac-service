CREATE SCHEMA IF NOT EXISTS pmsn;

CREATE TABLE pmsn.resource (
    id VARCHAR(255) PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE pmsn.action (
    id VARCHAR(255) PRIMARY KEY,
    resource_id VARCHAR(255) NOT NULL REFERENCES pmsn.resource(id),
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    UNIQUE(resource_id, code)
);

CREATE TABLE pmsn.resource_action_tenant (
    resource_id VARCHAR(255) NOT NULL REFERENCES pmsn.resource(id),
    action_id VARCHAR(255) NOT NULL REFERENCES pmsn.action(id),
    tenant_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (resource_id, action_id, tenant_id)
);

CREATE TABLE pmsn.role (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255)
);

CREATE TABLE pmsn.group (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255)
);

CREATE TABLE pmsn.role_permission (
    role_id VARCHAR(255) NOT NULL REFERENCES pmsn.role(id),
    resource_id VARCHAR(255) NOT NULL REFERENCES pmsn.resource(id),
    action_id VARCHAR(255) NOT NULL REFERENCES pmsn.action(id),
    PRIMARY KEY (role_id, resource_id, action_id)
);

CREATE TABLE pmsn.group_permission (
    group_id VARCHAR(255) NOT NULL REFERENCES pmsn.group(id),
    resource_id VARCHAR(255) NOT NULL REFERENCES pmsn.resource(id),
    action_id VARCHAR(255) NOT NULL REFERENCES pmsn.action(id),
    PRIMARY KEY (group_id, resource_id, action_id)
);

CREATE TABLE pmsn.user_role (
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL REFERENCES pmsn.role(id),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE pmsn.user_group (
    user_id VARCHAR(255) NOT NULL,
    group_id VARCHAR(255) NOT NULL REFERENCES pmsn.group(id),
    PRIMARY KEY (user_id, group_id)
);
