package app

import "context"

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
}
