package model

import "time"

// Event statuses
const (
	StatusPending    = "pending"
	StatusPublished  = "published"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// Event types
const (
	EventUserRoleAssignRequest  = "rbac.user_role.assign.request"
	EventUserRoleAssignSuccess  = "rbac.user_role.assign.success"
	EventUserRoleAssignFailed   = "rbac.user_role.assign.failed"
	EventUserRoleRemoveRequest  = "rbac.user_role.remove.request"
	EventUserRoleRemoveSuccess  = "rbac.user_role.remove.success"
	EventUserRoleRemoveFailed   = "rbac.user_role.remove.failed"
	EventUserGroupAssignRequest = "rbac.user_group.assign.request"
	EventUserGroupAssignSuccess = "rbac.user_group.assign.success"
	EventUserGroupAssignFailed  = "rbac.user_group.assign.failed"
	EventUserGroupRemoveRequest = "rbac.user_group.remove.request"
	EventUserGroupRemoveSuccess = "rbac.user_group.remove.success"
	EventUserGroupRemoveFailed  = "rbac.user_group.remove.failed"
)

// Event represents a message in the event system
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// UserRolePayload represents the payload for user-role events
type UserRolePayload struct {
	UserIDs []string `json:"user_ids"`
	RoleID  string   `json:"role_id"`
}

// UserGroupPayload represents the payload for user-group events
type UserGroupPayload struct {
	UserIDs []string `json:"user_ids"`
	GroupID string   `json:"group_id"`
}

// ErrorPayload represents the payload for failed events
type ErrorPayload struct {
	UserIDs []string `json:"user_ids,omitempty"`
	RoleID  string   `json:"role_id,omitempty"`
	GroupID string   `json:"group_id,omitempty"`
	Error   string   `json:"error"`
}

// PublishedEvent represents an event in the published_events table
type PublishedEvent struct {
	ID           string
	EventType    string
	Payload      []byte
	Status       string
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ConsumedEvent represents an event in the consumed_events table
type ConsumedEvent struct {
	ID           string
	EventType    string
	Payload      []byte
	Status       string
	ErrorMessage *string
	RetryCount   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
