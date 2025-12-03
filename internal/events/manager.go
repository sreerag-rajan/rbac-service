package events

import (
	"context"
	"fmt"
	"rbac-service/internal/logger"
	"rbac-service/internal/repository"
	"time"
)

const (
	ExchangeName = "rbac_permissions"
	QueueName    = "permissions"
)

// RoleAppService interface to avoid circular dependency
type RoleAppService interface {
	BulkAssignUsers(ctx context.Context, roleID string, req interface{}) error
	BulkRemoveUsers(ctx context.Context, roleID string, req interface{}) error
}

// GroupAppService interface to avoid circular dependency
type GroupAppService interface {
	BulkAssignUsers(ctx context.Context, groupID string, req interface{}) error
	BulkRemoveUsers(ctx context.Context, groupID string, req interface{}) error
}

// EventManager manages the event system lifecycle
type EventManager struct {
	provider      QueueProvider
	publisher     *Publisher
	consumer      *Consumer
	healthChecker *HealthChecker
	router        *EventRouter
}

// NewEventManager creates a new event manager
func NewEventManager(
	provider QueueProvider,
	auditRepo *repository.EventAuditRepository,
) (*EventManager, error) {
	// If no provider configured, return nil manager
	if provider == nil {
		return nil, nil
	}

	// Create publisher
	publisher := NewPublisher(provider, auditRepo, ExchangeName)

	// Create router
	router := NewEventRouter()

	// Create consumer
	consumer := NewConsumer(provider, auditRepo, router, QueueName, 3)

	// Create health checker with reconnect function
	reconnectFunc := func(ctx context.Context) error {
		return provider.Connect(ctx)
	}
	healthChecker := NewHealthChecker(provider, 30*time.Second, reconnectFunc)

	return &EventManager{
		provider:      provider,
		publisher:     publisher,
		consumer:      consumer,
		healthChecker: healthChecker,
		router:        router,
	}, nil
}

// GetRouter returns the event router
func (m *EventManager) GetRouter() *EventRouter {
	if m == nil {
		return nil
	}
	return m.router
}

// Start initializes and starts the event system
func (m *EventManager) Start(ctx context.Context) error {
	if m == nil {
		logger.Info(ctx, "Event system disabled (no provider configured)", nil)
		return nil
	}

	logger.Info(ctx, "Starting event system", nil)

	// Connect to provider
	err := m.provider.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to queue provider: %w", err)
	}

	// Declare exchange
	err = m.provider.DeclareExchange(ctx, ExchangeName, "topic")
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = m.provider.DeclareQueue(ctx, QueueName)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with routing key pattern
	err = m.provider.BindQueue(ctx, QueueName, ExchangeName, "rbac.*.*.request")
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consumer
	err = m.consumer.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	// Start health checker
	go m.healthChecker.Start(ctx)

	logger.Info(ctx, "Event system started successfully", nil)
	return nil
}

// Stop gracefully stops the event system
func (m *EventManager) Stop() error {
	if m == nil {
		return nil
	}

	logger.Info(context.Background(), "Stopping event system", nil)

	// Stop health checker
	m.healthChecker.Stop()

	// Close provider (this will also cancel consumers)
	err := m.provider.Close()
	if err != nil {
		return fmt.Errorf("failed to close provider: %w", err)
	}

	logger.Info(context.Background(), "Event system stopped", nil)
	return nil
}

// GetPublisher returns the publisher for use by application services
func (m *EventManager) GetPublisher() *Publisher {
	if m == nil {
		return nil
	}
	return m.publisher
}
