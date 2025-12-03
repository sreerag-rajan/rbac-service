package events

import (
	"context"
	"encoding/json"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"
	"time"
)

// Consumer handles consuming events with audit trail
type Consumer struct {
	provider   QueueProvider
	auditRepo  *repository.EventAuditRepository
	router     *EventRouter
	queue      string
	maxRetries int
}

// NewConsumer creates a new consumer
func NewConsumer(
	provider QueueProvider,
	auditRepo *repository.EventAuditRepository,
	router *EventRouter,
	queue string,
	maxRetries int,
) *Consumer {
	return &Consumer{
		provider:   provider,
		auditRepo:  auditRepo,
		router:     router,
		queue:      queue,
		maxRetries: maxRetries,
	}
}

// Start starts the consumer
func (c *Consumer) Start(ctx context.Context) error {
	logger.Info(ctx, "Starting consumer", nil, "queue", c.queue)

	return c.provider.Consume(ctx, c.queue, c.handleMessage)
}

// handleMessage processes a single message
func (c *Consumer) handleMessage(ctx context.Context, body []byte) error {
	// Parse event
	var event model.Event
	err := json.Unmarshal(body, &event)
	if err != nil {
		logger.Error(ctx, "Failed to unmarshal event", err)
		return nil // Don't retry malformed events
	}

	logger.Info(ctx, "Received event", nil, "event_id", event.ID, "event_type", event.Type)

	// Create audit entry with processing status
	auditEvent := &model.ConsumedEvent{
		ID:        event.ID,
		EventType: event.Type,
		Payload:   body,
		Status:    model.StatusProcessing,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = c.auditRepo.CreateConsumedEvent(ctx, auditEvent)
	if err != nil {
		logger.Error(ctx, "Failed to create consumed event audit entry", err, "event_id", event.ID)
		// Continue processing even if audit fails? Maybe better to retry.
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	// Process event with retry
	err = c.processWithRetry(ctx, event)

	if err != nil {
		// Update audit entry to failed
		errMsg := err.Error()
		updateErr := c.auditRepo.UpdateConsumedEvent(ctx, event.ID, model.StatusFailed, &errMsg, c.maxRetries)
		if updateErr != nil {
			logger.Error(ctx, "Failed to update consumed event audit entry", updateErr, "event_id", event.ID)
		}
		return err
	}

	// Update audit entry to completed
	updateErr := c.auditRepo.UpdateConsumedEvent(ctx, event.ID, model.StatusCompleted, nil, 0)
	if updateErr != nil {
		logger.Error(ctx, "Failed to update consumed event audit entry", updateErr, "event_id", event.ID)
	}

	return nil
}

// processWithRetry processes an event with retry logic
func (c *Consumer) processWithRetry(ctx context.Context, event model.Event) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := c.router.Dispatch(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err
		logger.Warn(ctx, "Failed to process event, retrying", nil,
			"event_id", event.ID,
			"event_type", event.Type,
			"attempt", fmt.Sprintf("%d", attempt+1),
			"max_retries", fmt.Sprintf("%d", c.maxRetries),
			"error", err.Error(),
		)

		if attempt < c.maxRetries {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			logger.Info(ctx, "Waiting before retry", nil, "backoff", backoff.String())
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed to process event after %d retries: %w", c.maxRetries, lastErr)
}
