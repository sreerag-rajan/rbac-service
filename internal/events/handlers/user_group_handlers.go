package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"rbac-service/internal/app"
	"rbac-service/internal/events"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"
	"time"

	"github.com/google/uuid"
)

// UserGroupHandlers contains handlers for user-group events
type UserGroupHandlers struct {
	groupApp  *app.GroupAppService
	publisher *events.Publisher
}

// NewUserGroupHandlers creates new user-group handlers
func NewUserGroupHandlers(groupApp *app.GroupAppService, publisher *events.Publisher) *UserGroupHandlers {
	return &UserGroupHandlers{
		groupApp:  groupApp,
		publisher: publisher,
	}
}

// HandleAssignRequest handles user-group assignment requests
func (h *UserGroupHandlers) HandleAssignRequest(ctx context.Context, event model.Event) error {
	// Parse payload
	var payload model.UserGroupPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	logger.Info(ctx, "Processing user-group assign request", nil,
		"group_id", payload.GroupID,
		"user_count", fmt.Sprintf("%d", len(payload.UserIDs)),
	)

	// Call application service
	req := model.BulkUserGroupRequest{UserIDs: payload.UserIDs}
	err = h.groupApp.BulkAssignUsers(ctx, payload.GroupID, req)

	// Prepare completion event
	completionEvent := model.Event{
		ID:        uuid.New().String(),
		Type:      model.EventUserGroupAssignSuccess,
		Timestamp: time.Now(),
	}

	if err != nil {
		// Failed event
		completionEvent.Type = model.EventUserGroupAssignFailed
		completionEvent.Payload = model.ErrorPayload{
			UserIDs: payload.UserIDs,
			GroupID: payload.GroupID,
			Error:   err.Error(),
		}

		logger.Error(ctx, "Failed to assign users to group", err, "group_id", payload.GroupID)
	} else {
		// Success event
		completionEvent.Payload = payload
		logger.Info(ctx, "Successfully assigned users to group", nil, "group_id", payload.GroupID)
	}

	// Publish completion event
	publishErr := h.publisher.PublishWithRetry(ctx, completionEvent, 3)
	if publishErr != nil {
		logger.Error(ctx, "Failed to publish completion event", publishErr,
			"event_type", completionEvent.Type,
			"event_id", completionEvent.ID,
		)
		// Don't fail the handler if publishing fails
	}

	return err
}

// HandleRemoveRequest handles user-group removal requests
func (h *UserGroupHandlers) HandleRemoveRequest(ctx context.Context, event model.Event) error {
	// Parse payload
	var payload model.UserGroupPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	logger.Info(ctx, "Processing user-group remove request", nil,
		"group_id", payload.GroupID,
		"user_count", fmt.Sprintf("%d", len(payload.UserIDs)),
	)

	// Call application service
	req := model.BulkUserGroupRequest{UserIDs: payload.UserIDs}
	err = h.groupApp.BulkRemoveUsers(ctx, payload.GroupID, req)

	// Prepare completion event
	completionEvent := model.Event{
		ID:        uuid.New().String(),
		Type:      model.EventUserGroupRemoveSuccess,
		Timestamp: time.Now(),
	}

	if err != nil {
		// Failed event
		completionEvent.Type = model.EventUserGroupRemoveFailed
		completionEvent.Payload = model.ErrorPayload{
			UserIDs: payload.UserIDs,
			GroupID: payload.GroupID,
			Error:   err.Error(),
		}

		logger.Error(ctx, "Failed to remove users from group", err, "group_id", payload.GroupID)
	} else {
		// Success event
		completionEvent.Payload = payload
		logger.Info(ctx, "Successfully removed users from group", nil, "group_id", payload.GroupID)
	}

	// Publish completion event
	publishErr := h.publisher.PublishWithRetry(ctx, completionEvent, 3)
	if publishErr != nil {
		logger.Error(ctx, "Failed to publish completion event", publishErr,
			"event_type", completionEvent.Type,
			"event_id", completionEvent.ID,
		)
		// Don't fail the handler if publishing fails
	}

	return err
}
