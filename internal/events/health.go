package events

import (
	"context"
	"rbac-service/internal/logger"
	"time"
)

// HealthChecker manages periodic health checks for the queue provider
type HealthChecker struct {
	provider      QueueProvider
	interval      time.Duration
	stopChan      chan struct{}
	reconnectFunc func(ctx context.Context) error
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(provider QueueProvider, interval time.Duration, reconnectFunc func(ctx context.Context) error) *HealthChecker {
	if interval == 0 {
		interval = 30 * time.Second
	}

	return &HealthChecker{
		provider:      provider,
		interval:      interval,
		stopChan:      make(chan struct{}),
		reconnectFunc: reconnectFunc,
	}
}

// Start begins periodic health checks
func (h *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	logger.Info(ctx, "Health checker started", nil, "interval", h.interval.String())

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "Health checker stopped", nil)
			return

		case <-h.stopChan:
			logger.Info(ctx, "Health checker stopped", nil)
			return

		case <-ticker.C:
			err := h.provider.HealthCheck(ctx)
			if err != nil {
				logger.Error(ctx, "Health check failed, attempting reconnection", err)

				if h.reconnectFunc != nil {
					if reconnectErr := h.reconnectFunc(ctx); reconnectErr != nil {
						logger.Error(ctx, "Reconnection failed", reconnectErr)
					} else {
						logger.Info(ctx, "Reconnection successful", nil)
					}
				}
			}
		}
	}
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	close(h.stopChan)
}
