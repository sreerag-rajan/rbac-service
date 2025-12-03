package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/model"
	"time"
)

// EventAuditRepository handles database operations for event audit tables
type EventAuditRepository struct{}

// NewEventAuditRepository creates a new event audit repository
func NewEventAuditRepository() *EventAuditRepository {
	return &EventAuditRepository{}
}

// CreatePublishedEvent creates a new published event record
func (r *EventAuditRepository) CreatePublishedEvent(ctx context.Context, event *model.PublishedEvent) error {
	query := `
		INSERT INTO pmsn.published_events (id, event_type, payload, status, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := GetPool().Exec(ctx, query,
		event.ID,
		event.EventType,
		event.Payload,
		event.Status,
		event.ErrorMessage,
		event.CreatedAt,
		event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create published event: %w", err)
	}

	return nil
}

// UpdatePublishedEvent updates a published event record
func (r *EventAuditRepository) UpdatePublishedEvent(ctx context.Context, id, status string, errorMessage *string) error {
	query := `
		UPDATE pmsn.published_events
		SET status = $1, error_message = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := GetPool().Exec(ctx, query, status, errorMessage, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update published event: %w", err)
	}

	return nil
}

// CreateConsumedEvent creates a new consumed event record
func (r *EventAuditRepository) CreateConsumedEvent(ctx context.Context, event *model.ConsumedEvent) error {
	query := `
		INSERT INTO pmsn.consumed_events (id, event_type, payload, status, error_message, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := GetPool().Exec(ctx, query,
		event.ID,
		event.EventType,
		event.Payload,
		event.Status,
		event.ErrorMessage,
		event.RetryCount,
		event.CreatedAt,
		event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create consumed event: %w", err)
	}

	return nil
}

// UpdateConsumedEvent updates a consumed event record
func (r *EventAuditRepository) UpdateConsumedEvent(ctx context.Context, id, status string, errorMessage *string, retryCount int) error {
	query := `
		UPDATE pmsn.consumed_events
		SET status = $1, error_message = $2, retry_count = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := GetPool().Exec(ctx, query, status, errorMessage, retryCount, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update consumed event: %w", err)
	}

	return nil
}

// GetConsumedEvent retrieves a consumed event by ID
func (r *EventAuditRepository) GetConsumedEvent(ctx context.Context, id string) (*model.ConsumedEvent, error) {
	query := `
		SELECT id, event_type, payload, status, error_message, retry_count, created_at, updated_at
		FROM pmsn.consumed_events
		WHERE id = $1
	`

	var event model.ConsumedEvent
	err := GetPool().QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.EventType,
		&event.Payload,
		&event.Status,
		&event.ErrorMessage,
		&event.RetryCount,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get consumed event: %w", err)
	}

	return &event, nil
}
