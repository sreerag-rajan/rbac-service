package events

import (
	"context"
	"fmt"
	"rbac-service/internal/model"
)

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, event model.Event) error

// EventRouter routes events to appropriate handlers
type EventRouter struct {
	handlers map[string]EventHandler
}

// NewEventRouter creates a new event router
func NewEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make(map[string]EventHandler),
	}
}

// Register registers a handler for an event type
func (r *EventRouter) Register(eventType string, handler EventHandler) {
	r.handlers[eventType] = handler
}

// Dispatch routes an event to its handler
func (r *EventRouter) Dispatch(ctx context.Context, event model.Event) error {
	handler, exists := r.handlers[event.Type]
	if !exists {
		return fmt.Errorf("no handler registered for event type: %s", event.Type)
	}

	return handler(ctx, event)
}
