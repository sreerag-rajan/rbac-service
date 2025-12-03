# Event-Driven Architecture

## Overview

The RBAC service supports event-driven integration with other services through a message queue system. This enables asynchronous communication for both consuming requests from other services and publishing completion events.

## Supported Queue Providers

- **RabbitMQ**: Primary implementation with connection pooling and auto-reconnection

## Event Naming Convention

Events follow a standardized naming pattern:

```
<service>.<noun>.<verb>.<type>
```

**Components:**
- `service`: Service name (e.g., `rbac`)
- `noun`: Entity being operated on (e.g., `user_role`, `user_group`)
- `verb`: Action being performed (e.g., `assign`, `remove`)
- `type`: Event type (`request`, `success`, `failed`)

**Examples:**
- `rbac.user_role.assign.request`
- `rbac.user_role.assign.success`
- `rbac.user_role.assign.failed`

## Queue and Exchange Architecture

### Consumer Queue
- **Queue Name**: `permissions`
- **Purpose**: Consumes request events from other services
- **Binding**: Binds to routing keys matching `rbac.*.*.request`

### Publisher Exchange
- **Exchange Name**: `rbac_permissions`
- **Exchange Type**: Topic
- **Purpose**: Publishes completion events (success/failed)
- **Routing Keys**: Event type (e.g., `rbac.user_role.assign.success`)

## Supported Event Types

### User-Role Events

#### Request Events (Consumed)
- **`rbac.user_role.assign.request`**
  - Assigns users to a role
  - Payload: `{"user_ids": ["uuid1", "uuid2"], "role_id": "role-uuid"}`
  
- **`rbac.user_role.remove.request`**
  - Removes users from a role
  - Payload: `{"user_ids": ["uuid1", "uuid2"], "role_id": "role-uuid"}`

#### Completion Events (Published)
- **`rbac.user_role.assign.success`**
  - Published after successful user-role assignment
  - Payload: `{"user_ids": ["uuid1"], "role_id": "role-uuid"}`

- **`rbac.user_role.assign.failed`**
  - Published when user-role assignment fails
  - Payload: `{"user_ids": ["uuid1"], "role_id": "role-uuid", "error": "error message"}`

- **`rbac.user_role.remove.success`**
  - Published after successful user-role removal
  - Payload: `{"user_ids": ["uuid1"], "role_id": "role-uuid"}`

- **`rbac.user_role.remove.failed`**
  - Published when user-role removal fails
  - Payload: `{"user_ids": ["uuid1"], "role_id": "role-uuid", "error": "error message"}`

### User-Group Events

#### Request Events (Consumed)
- **`rbac.user_group.assign.request`**
  - Assigns users to a group
  - Payload: `{"user_ids": ["uuid1", "uuid2"], "group_id": "group-uuid"}`
  
- **`rbac.user_group.remove.request`**
  - Removes users from a group
  - Payload: `{"user_ids": ["uuid1", "uuid2"], "group_id": "group-uuid"}`

#### Completion Events (Published)
- **`rbac.user_group.assign.success`**
  - Published after successful user-group assignment
  - Payload: `{"user_ids": ["uuid1"], "group_id": "group-uuid"}`

- **`rbac.user_group.assign.failed`**
  - Published when user-group assignment fails
  - Payload: `{"user_ids": ["uuid1"], "group_id": "group-uuid", "error": "error message"}`

- **`rbac.user_group.remove.success`**
  - Published after successful user-group removal
  - Payload: `{"user_ids": ["uuid1"], "group_id": "group-uuid"}`

- **`rbac.user_group.remove.failed`**
  - Published when user-group removal fails
  - Payload: `{"user_ids": ["uuid1"], "group_id": "group-uuid", "error": "error message"}`

## Configuration

### Environment Variables

```env
# Queue Provider Configuration
QUEUE_PROVIDER=RABBITMQ          # Provider type (RABBITMQ or empty to disable)

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_MAX_CONNECTIONS=5       # Maximum connection pool size
RABBITMQ_MAX_CHANNELS_PER_CONN=10 # Maximum channels per connection
```

### Disabling Event System

To disable the event system entirely, leave `QUEUE_PROVIDER` empty or unset:

```env
QUEUE_PROVIDER=
```

## Event Processing Flow

### Consuming Request Events

1. **Receive Event**: Consumer receives event from `permissions` queue
2. **Create Audit Entry**: Insert record in `pmsn.consumed_events` with status `processing`
3. **Route to Handler**: Event router dispatches to appropriate handler based on event type
4. **Execute Business Logic**: Handler calls application service layer
5. **Publish Completion Event**: On success/failure, publish corresponding event to `rbac_permissions` exchange
6. **Update Audit Entry**: Update status to `completed` or `failed` with error details
7. **Retry on Failure**: If handler fails, retry with exponential backoff (max 3 retries)

### Publishing Completion Events

1. **Create Audit Entry**: Insert record in `pmsn.published_events` with status `pending`
2. **Publish to Exchange**: Send event to `rbac_permissions` exchange
3. **Update Audit Entry**: Update status to `published` or `failed` with error details

## Retry Strategy

### Consumer Retry
- **Max Retries**: 3
- **Backoff**: Exponential (1s, 2s, 4s)
- **After Exhaustion**: Mark as `failed` in audit table

### Publisher Retry
- **Max Retries**: 3
- **Backoff**: Exponential (1s, 2s, 4s)
- **After Exhaustion**: Mark as `failed` in audit table

## Health Check and Reconnection

### Health Check
- **Interval**: 30 seconds
- **Action**: Verifies connection to queue provider
- **On Failure**: Triggers reconnection

### Auto-Reconnection
- **Trigger**: Connection loss or health check failure
- **Strategy**: Exponential backoff with max 10 retries
- **Consumer Restart**: Automatically restarts consumer after successful reconnection

## Audit Tables

### Published Events (`pmsn.published_events`)

Tracks all events published by this service.

| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | PK, Unique event ID |
| `event_type` | VARCHAR | Event type (e.g., `rbac.user_role.assign.success`) |
| `payload` | JSONB | Event payload |
| `status` | VARCHAR | `pending`, `published`, `failed` |
| `error_message` | TEXT | Error details if failed |
| `created_at` | TIMESTAMP | Event creation time |
| `updated_at` | TIMESTAMP | Last update time |

### Consumed Events (`pmsn.consumed_events`)

Tracks all events consumed by this service.

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

## Integration with Application Services

The event system integrates with the existing 4-layer architecture:

1. **Event Handlers** (`internal/events/handlers`): Entry point for consumed events
2. **Application Services** (`internal/app`): Business logic orchestration (shared with API layer)
3. **Domain Services** (`internal/service`): Core business rules
4. **Repository Layer** (`internal/repository`): Database operations

Event handlers call the same application service methods used by the API layer, ensuring consistency.

## Error Handling

### Consumer Errors
- **Validation Errors**: Mark as `failed` immediately (no retry)
- **Business Logic Errors**: Retry with backoff
- **Database Errors**: Retry with backoff
- **Unknown Errors**: Retry with backoff

### Publisher Errors
- **Connection Errors**: Retry with backoff
- **Exchange/Queue Errors**: Mark as `failed` immediately
- **Serialization Errors**: Mark as `failed` immediately

## Monitoring and Observability

### Logs
All event operations are logged with structured logging:
- Event type
- Event ID
- Status transitions
- Error details
- Retry attempts

### Audit Trail
Both `published_events` and `consumed_events` tables provide complete audit trail for:
- Debugging failed events
- Tracking event processing times
- Analyzing retry patterns
- Compliance and auditing

## Example Usage

### Publishing an Event (from Application Service)

```go
// After successful user-role assignment
event := events.Event{
    ID:        uuid.New().String(),
    Type:      events.EventUserRoleAssignSuccess,
    Payload:   events.UserRolePayload{UserIDs: userIDs, RoleID: roleID},
    Timestamp: time.Now(),
}

err := publisher.Publish(ctx, "rbac_permissions", event.Type, event)
if err != nil {
    logger.Error(ctx, "Failed to publish event", err)
}
```

### Consuming an Event (Handler)

```go
func HandleUserRoleAssignRequest(ctx context.Context, event events.Event) error {
    var payload events.UserRolePayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return err
    }

    // Call application service
    err := roleAppService.BulkAssignUsers(ctx, payload.RoleID, payload.UserIDs)
    
    // Publish completion event
    completionEvent := events.Event{
        ID:        uuid.New().String(),
        Type:      events.EventUserRoleAssignSuccess,
        Payload:   payload,
        Timestamp: time.Now(),
    }
    
    if err != nil {
        completionEvent.Type = events.EventUserRoleAssignFailed
        completionEvent.Payload = map[string]interface{}{
            "user_ids": payload.UserIDs,
            "role_id":  payload.RoleID,
            "error":    err.Error(),
        }
    }
    
    publisher.Publish(ctx, "rbac_permissions", completionEvent.Type, completionEvent)
    return err
}
```
