# Database Schema

The service uses a Postgres database with a schema named `pmsn`.

## Tables

### `pmsn.resource`
| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique ID |
| `code` | VARCHAR | Unique code (e.g., "order") |
| `name` | VARCHAR | Display name |
| `description` | TEXT | |

### `pmsn.action`
| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique ID |
| `resource_id` | VARCHAR | FK to `pmsn.resource.id` |
| `code` | VARCHAR | Unique code per resource (e.g., "read") |
| `name` | VARCHAR | Display name |
| `description` | TEXT | |

### `pmsn.resource_action_tenant`
| Column | Type | Description |
|---|---|---|
| `resource_id` | VARCHAR | FK to `pmsn.resource.id` |
| `action_id` | VARCHAR | FK to `pmsn.action.id` |
| `tenant_id` | VARCHAR | Tenant ID |
| **PK** | | `(resource_id, action_id, tenant_id)` |

### `pmsn.role`
| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique ID |
| `name` | VARCHAR | Role name |
| `tenant_id` | VARCHAR | Optional Tenant ID |

### `pmsn.group`
| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique ID |
| `name` | VARCHAR | Group name |
| `tenant_id` | VARCHAR | Optional Tenant ID |

### `pmsn.role_permission`
| Column | Type | Description |
|---|---|---|
| `role_id` | VARCHAR | FK to `pmsn.role.id` |
| `resource_id` | VARCHAR | FK to `pmsn.resource.id` |
| `action_id` | VARCHAR | FK to `pmsn.action.id` |
| **PK** | | `(role_id, resource_id, action_id)` |

### `pmsn.group_permission`
| Column | Type | Description |
|---|---|---|
| `group_id` | VARCHAR | FK to `pmsn.group.id` |
| `resource_id` | VARCHAR | FK to `pmsn.resource.id` |
| `action_id` | VARCHAR | FK to `pmsn.action.id` |
| **PK** | | `(group_id, resource_id, action_id)` |

### `pmsn.user_role`
| Column | Type | Description |
|---|---|---|
| `user_id` | VARCHAR | User UUID |
| `role_id` | VARCHAR | FK to `pmsn.role.id` |
| **PK** | | `(user_id, role_id)` |

### `pmsn.user_group`
| Column | Type | Description |
|---|---|---|
| `user_id` | VARCHAR | User UUID |
| `group_id` | VARCHAR | FK to `pmsn.group.id` |
| **PK** | | `(user_id, group_id)` |

### `pmsn.published_events`
| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique event ID |
| `event_type` | VARCHAR | Event type (e.g., `rbac.user_role.assign.success`) |
| `payload` | JSONB | Event payload |
| `status` | VARCHAR | `pending`, `published`, `failed` |
| `error_message` | TEXT | Error details if failed |
| `created_at` | TIMESTAMP | Event creation time |
| `updated_at` | TIMESTAMP | Last update time |

### `pmsn.consumed_events`
| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique event ID |
| `event_type` | VARCHAR | Event type (e.g., `rbac.user_role.assign.request`) |
| `payload` | JSONB | Event payload |
| `status` | VARCHAR | `processing`, `completed`, `failed` |
| `error_message` | TEXT | Error details if failed |
| `retry_count` | INT | Number of retry attempts |
| `created_at` | TIMESTAMP | Event creation time |
| `updated_at` | TIMESTAMP | Last update time |
