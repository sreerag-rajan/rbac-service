package events

import (
	"context"
	"encoding/json"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

// Publisher handles publishing events with audit trail
type Publisher struct {
	provider  QueueProvider
	auditRepo *repository.EventAuditRepository
	exchange  string
}

// NewPublisher creates a new publisher
func NewPublisher(provider QueueProvider, auditRepo *repository.EventAuditRepository, exchange string) *Publisher {
	return &Publisher{
		provider:  provider,
		auditRepo: auditRepo,
		exchange:  exchange,
	}
}

// PublishRaw publishes an event with audit trail (internal use)
func (p *Publisher) PublishRaw(ctx context.Context, event model.Event) error {
	// Generate ID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize payload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// Create audit entry with pending status
	auditEvent := &model.PublishedEvent{
		ID:        event.ID,
		EventType: event.Type,
		Payload:   payloadBytes,
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = p.auditRepo.CreatePublishedEvent(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "Failed to create published event audit entry", err, "event_id", event.ID, "event_type", event.Type)
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	// Serialize entire event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		errMsg := err.Error()
		p.auditRepo.UpdatePublishedEvent(ctx, event.ID, model.StatusFailed, &errMsg)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to exchange
	err = p.provider.Publish(ctx, p.exchange, event.Type, eventBytes)
	if err != nil {
		errMsg := err.Error()
		p.auditRepo.UpdatePublishedEvent(ctx, event.ID, model.StatusFailed, &errMsg)
		logger.Error(ctx, "Failed to publish event", err, "event_id", event.ID, "event_type", event.Type)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Update audit entry to published
	err = p.auditRepo.UpdatePublishedEvent(ctx, event.ID, model.StatusPublished, nil)
	if err != nil {
		logger.Error(ctx, "Failed to update published event audit entry", err, "event_id", event.ID)
		// Don't return error as the event was published successfully
	}

	logger.Info(ctx, "Event published successfully", nil, "event_id", event.ID, "event_type", event.Type)
	return nil
}

// PublishWithRetry publishes an event with retry logic
func (p *Publisher) PublishWithRetry(ctx context.Context, event model.Event, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := p.PublishRaw(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err
		logger.Warn(ctx, "Failed to publish event, retrying", nil,
			"event_id", event.ID,
			"event_type", event.Type,
			"attempt", fmt.Sprintf("%d", attempt+1),
			"max_retries", fmt.Sprintf("%d", maxRetries),
			"error", err.Error(),
		)

		if attempt < maxRetries {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed to publish event after %d retries: %w", maxRetries, lastErr)
}

// Publish implements the app.EventPublisher interface
func (p *Publisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	event := model.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return p.PublishWithRetry(ctx, event, 3)
}
