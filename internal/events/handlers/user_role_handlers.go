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

// UserRoleHandlers contains handlers for user-role events
type UserRoleHandlers struct {
	roleApp   *app.RoleAppService
	publisher *events.Publisher
}

// NewUserRoleHandlers creates new user-role handlers
func NewUserRoleHandlers(roleApp *app.RoleAppService, publisher *events.Publisher) *UserRoleHandlers {
	return &UserRoleHandlers{
		roleApp:   roleApp,
		publisher: publisher,
	}
}

// HandleAssignRequest handles user-role assignment requests
func (h *UserRoleHandlers) HandleAssignRequest(ctx context.Context, event model.Event) error {
	// Parse payload
	var payload model.UserRolePayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	logger.Info(ctx, "Processing user-role assign request", nil,
		"role_id", payload.RoleID,
		"user_count", fmt.Sprintf("%d", len(payload.UserIDs)),
	)

	// Call application service
	req := model.BulkUserRoleRequest{UserIDs: payload.UserIDs}
	err = h.roleApp.BulkAssignUsers(ctx, payload.RoleID, req)

	// Prepare completion event
	completionEvent := model.Event{
		ID:        uuid.New().String(),
		Type:      model.EventUserRoleAssignSuccess,
		Timestamp: time.Now(),
	}

	if err != nil {
		// Failed event
		completionEvent.Type = model.EventUserRoleAssignFailed
		completionEvent.Payload = model.ErrorPayload{
			UserIDs: payload.UserIDs,
			RoleID:  payload.RoleID,
			Error:   err.Error(),
		}

		logger.Error(ctx, "Failed to assign users to role", err, "role_id", payload.RoleID)
	} else {
		// Success event
		completionEvent.Payload = payload
		logger.Info(ctx, "Successfully assigned users to role", nil, "role_id", payload.RoleID)
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

// HandleRemoveRequest handles user-role removal requests
func (h *UserRoleHandlers) HandleRemoveRequest(ctx context.Context, event model.Event) error {
	// Parse payload
	var payload model.UserRolePayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	logger.Info(ctx, "Processing user-role remove request", nil,
		"role_id", payload.RoleID,
		"user_count", fmt.Sprintf("%d", len(payload.UserIDs)),
	)

	// Call application service
	req := model.BulkUserRoleRequest{UserIDs: payload.UserIDs}
	err = h.roleApp.BulkRemoveUsers(ctx, payload.RoleID, req)

	// Prepare completion event
	completionEvent := model.Event{
		ID:        uuid.New().String(),
		Type:      model.EventUserRoleRemoveSuccess,
		Timestamp: time.Now(),
	}

	if err != nil {
		// Failed event
		completionEvent.Type = model.EventUserRoleRemoveFailed
		completionEvent.Payload = model.ErrorPayload{
			UserIDs: payload.UserIDs,
			RoleID:  payload.RoleID,
			Error:   err.Error(),
		}

		logger.Error(ctx, "Failed to remove users from role", err, "role_id", payload.RoleID)
	} else {
		// Success event
		completionEvent.Payload = payload
		logger.Info(ctx, "Successfully removed users from role", nil, "role_id", payload.RoleID)
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
