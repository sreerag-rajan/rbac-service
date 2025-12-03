package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"rbac-service/internal/app"
	"rbac-service/internal/controller"
	"rbac-service/internal/events"
	"rbac-service/internal/events/handlers"
	"rbac-service/internal/events/rabbitmq"
	"rbac-service/internal/logger"
	"rbac-service/internal/model"
	"rbac-service/internal/repository"
	"rbac-service/internal/service"
	"strconv"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	// 1. Init DB
	if err := repository.InitDB(ctx); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer repository.CloseDB()

	// Run Migrations (controlled by RUN_MIGRATIONS env var, defaults to false)
	runMigrations := os.Getenv("RUN_MIGRATIONS")
	if runMigrations == "true" {
		logger.Info(ctx, "Running database migrations", nil)
		if err := repository.RunMigrations(ctx, "./migrations"); err != nil {
			logger.Fatal(ctx, "Failed to run migrations", err)
		}
	} else {
		logger.Info(ctx, "Skipping database migrations (RUN_MIGRATIONS not set to 'true')", nil)
	}

	// 2. Init Repositories
	tenantRepo := repository.NewTenantRepository()
	roleRepo := repository.NewRoleRepository()
	groupRepo := repository.NewGroupRepository()
	resRepo := repository.NewResourceRepository()
	permRepo := repository.NewPermissionRepository()
	eventAuditRepo := repository.NewEventAuditRepository()

	// 3. Init Domain Services
	tenantService := service.NewTenantService(tenantRepo)
	roleService := service.NewRoleService(roleRepo)
	groupService := service.NewGroupService(groupRepo)
	permService := service.NewPermissionService(permRepo, resRepo)

	// 4. Init Event System
	queueProvider, err := createQueueProvider()
	if err != nil {
		logger.Fatal(ctx, "Failed to create queue provider", err)
	}

	// Check if external queue manager is used (defaults to false)
	hasExternalQueueManager := os.Getenv("HAS_EXTERNAL_QUEUE_MANAGER") == "true"
	if hasExternalQueueManager {
		logger.Info(ctx, "External queue manager enabled - will skip infrastructure setup", nil)
	}

	eventManager, err := events.NewEventManager(queueProvider, eventAuditRepo, hasExternalQueueManager)
	if err != nil {
		logger.Fatal(ctx, "Failed to create event manager", err)
	}

	var publisher app.EventPublisher
	if eventManager != nil {
		publisher = eventManager.GetPublisher()
	}

	// 5. Init App Services
	tenantApp := app.NewTenantAppService(tenantService)
	roleApp := app.NewRoleAppService(roleService, publisher)
	groupApp := app.NewGroupAppService(groupService, publisher)
	validationApp := app.NewValidationAppService(permService)

	// 6. Register Event Handlers
	if eventManager != nil {
		router := eventManager.GetRouter()
		publisher := eventManager.GetPublisher()

		// Create handler instances
		userRoleHandlers := handlers.NewUserRoleHandlers(roleApp, publisher)
		userGroupHandlers := handlers.NewUserGroupHandlers(groupApp, publisher)

		// Register user-role handlers
		router.Register(model.EventUserRoleAssignRequest, userRoleHandlers.HandleAssignRequest)
		router.Register(model.EventUserRoleRemoveRequest, userRoleHandlers.HandleRemoveRequest)

		// Register user-group handlers
		router.Register(model.EventUserGroupAssignRequest, userGroupHandlers.HandleAssignRequest)
		router.Register(model.EventUserGroupRemoveRequest, userGroupHandlers.HandleRemoveRequest)

		if err := eventManager.Start(ctx); err != nil {
			logger.Fatal(ctx, "Failed to start event system", err)
		}
		defer eventManager.Stop()
	}

	// 6. Init Handlers
	tenantHandler := controller.NewTenantHandler(tenantApp)
	roleHandler := controller.NewRoleHandler(roleApp)
	groupHandler := controller.NewGroupHandler(groupApp)
	validationHandler := controller.NewValidationHandler(validationApp)

	// 7. Setup Router
	r := controller.SetupRouter(tenantHandler, roleHandler, groupHandler, validationHandler)

	// 8. Start Server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "9980"
	}

	// Setup graceful shutdown
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Info(ctx, "Starting server", nil, "port", port)
		if err := r.Run(":" + port); err != nil {
			logger.Fatal(ctx, "Failed to start server", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	logger.Info(serverCtx, "Shutting down server gracefully...", nil)

	// Give outstanding requests time to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop event system
	if eventManager != nil {
		logger.Info(shutdownCtx, "Stopping event system...", nil)
		if err := eventManager.Stop(); err != nil {
			logger.Error(shutdownCtx, "Error stopping event system", err)
		}
	}

	logger.Info(shutdownCtx, "Server stopped", nil)
}

func createQueueProvider() (events.QueueProvider, error) {
	providerType := os.Getenv("QUEUE_PROVIDER")
	if providerType == "" {
		return nil, nil
	}

	if providerType == "RABBITMQ" {
		url := os.Getenv("RABBITMQ_URL")
		maxConnsStr := os.Getenv("RABBITMQ_MAX_CONNECTIONS")
		maxChannelsStr := os.Getenv("RABBITMQ_MAX_CHANNELS_PER_CONN")

		maxConns := 1
		if maxConnsStr != "" {
			var err error
			maxConns, err = strconv.Atoi(maxConnsStr)
			if err != nil {
				return nil, fmt.Errorf("invalid RABBITMQ_MAX_CONNECTIONS: %w", err)
			}
		}

		maxChannels := 10
		if maxChannelsStr != "" {
			var err error
			maxChannels, err = strconv.Atoi(maxChannelsStr)
			if err != nil {
				return nil, fmt.Errorf("invalid RABBITMQ_MAX_CHANNELS_PER_CONN: %w", err)
			}
		}

		return rabbitmq.NewRabbitMQProvider(url, maxConns, maxChannels)
	}

	return nil, fmt.Errorf("unsupported queue provider: %s", providerType)
}
