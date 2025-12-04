BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS pmsn;

CREATE TABLE pmsn.resource (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE pmsn.action (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_id UUID NOT NULL REFERENCES pmsn.resource(id),
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    UNIQUE(resource_id, code)
);

CREATE TABLE pmsn.resource_action_tenant (
    resource_id UUID NOT NULL REFERENCES pmsn.resource(id),
    action_id UUID NOT NULL REFERENCES pmsn.action(id),
    tenant_id UUID NOT NULL,
    PRIMARY KEY (resource_id, action_id, tenant_id)
);

CREATE TABLE pmsn.role (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    tenant_id UUID
);

CREATE TABLE pmsn.group (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    tenant_id UUID
);

CREATE TABLE pmsn.role_permission (
    role_id UUID NOT NULL REFERENCES pmsn.role(id),
    resource_id UUID NOT NULL REFERENCES pmsn.resource(id),
    action_id UUID NOT NULL REFERENCES pmsn.action(id),
    PRIMARY KEY (role_id, resource_id, action_id)
);

CREATE TABLE pmsn.group_permission (
    group_id UUID NOT NULL REFERENCES pmsn.group(id),
    resource_id UUID NOT NULL REFERENCES pmsn.resource(id),
    action_id UUID NOT NULL REFERENCES pmsn.action(id),
    PRIMARY KEY (group_id, resource_id, action_id)
);

CREATE TABLE pmsn.user_role (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES pmsn.role(id),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE pmsn.user_group (
    user_id UUID NOT NULL,
    group_id UUID NOT NULL REFERENCES pmsn.group(id),
    PRIMARY KEY (user_id, group_id)
);

COMMIT;
