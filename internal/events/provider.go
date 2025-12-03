package events

import "context"

// QueueProvider defines the interface for message queue providers
type QueueProvider interface {
	// Connect establishes connection to the queue provider
	Connect(ctx context.Context) error

	// Close closes the connection to the queue provider
	Close() error

	// Publish publishes an event to an exchange with a routing key
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error

	// Consume starts consuming messages from a queue
	Consume(ctx context.Context, queue string, handler MessageHandler) error

	// HealthCheck verifies the connection is healthy
	HealthCheck(ctx context.Context) error

	// DeclareExchange declares an exchange
	DeclareExchange(ctx context.Context, exchange, exchangeType string) error

	// DeclareQueue declares a queue and returns the queue name
	DeclareQueue(ctx context.Context, queue string) (string, error)

	// BindQueue binds a queue to an exchange with a routing key
	BindQueue(ctx context.Context, queue, exchange, routingKey string) error
}

// MessageHandler is a function that processes incoming messages
type MessageHandler func(ctx context.Context, body []byte) error
