package rabbitmq

import (
	"context"
	"fmt"
	"rbac-service/internal/events"
	"rbac-service/internal/logger"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQProvider implements the QueueProvider interface for RabbitMQ
type RabbitMQProvider struct {
	url                 string
	maxConnections      int
	maxChannelsPerConn  int
	connections         []*amqp.Connection
	channelPools        [][]*amqp.Channel
	currentConnIndex    int
	currentChannelIndex []int
	mu                  sync.RWMutex
	closed              bool
	consumerCancelFuncs []context.CancelFunc
}

// NewRabbitMQProvider creates a new RabbitMQ provider
func NewRabbitMQProvider(url string, maxConnections, maxChannelsPerConn int) (*RabbitMQProvider, error) {
	if url == "" {
		return nil, fmt.Errorf("RabbitMQ URL is required")
	}

	return &RabbitMQProvider{
		url:                 url,
		maxConnections:      maxConnections,
		maxChannelsPerConn:  maxChannelsPerConn,
		connections:         make([]*amqp.Connection, 0, maxConnections),
		channelPools:        make([][]*amqp.Channel, 0, maxConnections),
		currentChannelIndex: make([]int, maxConnections),
	}, nil
}

// Connect establishes connections to RabbitMQ
func (r *RabbitMQProvider) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("provider is closed")
	}

	// Close existing connections if any
	r.closeConnectionsUnsafe()

	// Create connection pool
	for i := 0; i < r.maxConnections; i++ {
		conn, err := amqp.Dial(r.url)
		if err != nil {
			r.closeConnectionsUnsafe()
			return fmt.Errorf("failed to connect to RabbitMQ (connection %d): %w", i, err)
		}

		r.connections = append(r.connections, conn)
		r.channelPools = append(r.channelPools, make([]*amqp.Channel, 0, r.maxChannelsPerConn))

		logger.Info(ctx, "RabbitMQ connection established", nil, "connection_index", fmt.Sprintf("%d", i))
	}

	logger.Info(ctx, "RabbitMQ provider connected", nil, "connections", fmt.Sprintf("%d", r.maxConnections))
	return nil
}

// Close closes all connections and channels
func (r *RabbitMQProvider) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true

	// Cancel all consumers
	for _, cancel := range r.consumerCancelFuncs {
		cancel()
	}
	r.consumerCancelFuncs = nil

	r.closeConnectionsUnsafe()
	return nil
}

// closeConnectionsUnsafe closes all connections without locking (must be called with lock held)
func (r *RabbitMQProvider) closeConnectionsUnsafe() {
	// Close all channels
	for _, channelPool := range r.channelPools {
		for _, ch := range channelPool {
			if ch != nil && !ch.IsClosed() {
				ch.Close()
			}
		}
	}

	// Close all connections
	for _, conn := range r.connections {
		if conn != nil && !conn.IsClosed() {
			conn.Close()
		}
	}

	r.connections = make([]*amqp.Connection, 0, r.maxConnections)
	r.channelPools = make([][]*amqp.Channel, 0, r.maxConnections)
	r.currentChannelIndex = make([]int, r.maxConnections)
	r.currentConnIndex = 0
}

// getChannel gets or creates a channel from the pool (round-robin)
func (r *RabbitMQProvider) getChannel(ctx context.Context) (*amqp.Channel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("provider is closed")
	}

	if len(r.connections) == 0 {
		return nil, fmt.Errorf("no connections available")
	}

	// Round-robin connection selection
	connIndex := r.currentConnIndex
	r.currentConnIndex = (r.currentConnIndex + 1) % len(r.connections)

	conn := r.connections[connIndex]
	if conn.IsClosed() {
		return nil, fmt.Errorf("connection %d is closed", connIndex)
	}

	// Try to reuse existing channel
	channelPool := r.channelPools[connIndex]
	if len(channelPool) < r.maxChannelsPerConn {
		// Create new channel
		ch, err := conn.Channel()
		if err != nil {
			return nil, fmt.Errorf("failed to create channel: %w", err)
		}

		r.channelPools[connIndex] = append(r.channelPools[connIndex], ch)
		return ch, nil
	}

	// Reuse existing channel (round-robin)
	channelIndex := r.currentChannelIndex[connIndex]
	r.currentChannelIndex[connIndex] = (r.currentChannelIndex[connIndex] + 1) % len(channelPool)

	ch := channelPool[channelIndex]
	if ch.IsClosed() {
		// Recreate channel
		newCh, err := conn.Channel()
		if err != nil {
			return nil, fmt.Errorf("failed to recreate channel: %w", err)
		}
		r.channelPools[connIndex][channelIndex] = newCh
		return newCh, nil
	}

	return ch, nil
}

// Publish publishes a message to an exchange
func (r *RabbitMQProvider) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	ch, err := r.getChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume starts consuming messages from a queue
func (r *RabbitMQProvider) Consume(ctx context.Context, queue string, handler events.MessageHandler) error {
	ch, err := r.getChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel for consumer: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		queue, // queue
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	// Create cancellable context for this consumer
	consumerCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.consumerCancelFuncs = append(r.consumerCancelFuncs, cancel)
	r.mu.Unlock()

	// Start consumer goroutine
	go func() {
		logger.Info(consumerCtx, "Started consuming from queue", nil, "queue", queue)

		for {
			select {
			case <-consumerCtx.Done():
				logger.Info(consumerCtx, "Consumer stopped", nil, "queue", queue)
				return

			case msg, ok := <-msgs:
				if !ok {
					logger.Warn(consumerCtx, "Message channel closed", nil, "queue", queue)
					return
				}

				// Process message
				err := handler(consumerCtx, msg.Body)
				if err != nil {
					logger.Error(consumerCtx, "Failed to process message", err, "queue", queue)
					msg.Nack(false, true) // Requeue on error
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

// HealthCheck verifies the connection is healthy
func (r *RabbitMQProvider) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return fmt.Errorf("provider is closed")
	}

	if len(r.connections) == 0 {
		return fmt.Errorf("no connections available")
	}

	// Check if at least one connection is alive
	for i, conn := range r.connections {
		if conn != nil && !conn.IsClosed() {
			return nil
		}
		logger.Warn(ctx, "Connection is closed", nil, "connection_index", fmt.Sprintf("%d", i))
	}

	return fmt.Errorf("all connections are closed")
}

// DeclareExchange declares an exchange
func (r *RabbitMQProvider) DeclareExchange(ctx context.Context, exchange, exchangeType string) error {
	ch, err := r.getChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	logger.Info(ctx, "Exchange declared", nil, "exchange", exchange, "type", exchangeType)
	return nil
}

// DeclareQueue declares a queue
func (r *RabbitMQProvider) DeclareQueue(ctx context.Context, queue string) (string, error) {
	ch, err := r.getChannel(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return "", fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Info(ctx, "Queue declared", nil, "queue", q.Name)
	return q.Name, nil
}

// BindQueue binds a queue to an exchange
func (r *RabbitMQProvider) BindQueue(ctx context.Context, queue, exchange, routingKey string) error {
	ch, err := r.getChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	err = ch.QueueBind(
		queue,      // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	logger.Info(ctx, "Queue bound to exchange", nil, "queue", queue, "exchange", exchange, "routing_key", routingKey)
	return nil
}
