# Architecture

The service follows a 4-layered architecture to separate concerns and ensure modularity.

## Layers

### 1. Controller Layer (`internal/controller`)
- **Responsibility**: Handles HTTP requests, parsing, validation, and response formatting.
- **Framework**: Gin.
- **Input**: HTTP Request.
- **Output**: HTTP Response (JSON).
- **Dependencies**: Application Service.

### 1b. Event Layer (`internal/events`)
- **Responsibility**: Handles asynchronous event-driven communication. Consumes request events from message queues and publishes completion events.
- **Framework**: RabbitMQ (configurable via provider interface).
- **Input**: Queue messages (events).
- **Output**: Completion events to exchange.
- **Dependencies**: Application Service (shared with Controller layer), Event Publisher, Event Audit Repository.
- **Note**: This is an alternative entry point to the system alongside the Controller layer. Both layers use the same Application Service layer for business logic.

### 2. Application Service Layer (`internal/app`)
- **Responsibility**: Orchestrates business use cases. It calls Domain Services and Repositories as needed. It handles transaction boundaries.
- **Input**: DTOs from Controller.
- **Output**: DTOs to Controller.
- **Dependencies**: Domain Service, Repository.

### 3. Domain Service Layer (`internal/service`)
- **Responsibility**: Contains pure business logic and rules. For example, the logic to resolve permissions from multiple roles/groups.
- **Input**: Domain Models.
- **Output**: Domain Models / Results.
- **Dependencies**: Repository (interfaces).

### 4. Repository Layer (`internal/repository`)
- **Responsibility**: Direct interaction with the database. Executes SQL queries.
- **Technology**: Postgres driver (pgx or similar).
- **Input**: Domain Models / Query Parameters.
- **Output**: Domain Models / Errors.

## Cross-Cutting Concerns

### Logging
- Structured logging with 5 levels (DEBUG, INFO, WARN, ERROR, FATAL).
- Fields: `service`, `file`, `func`, `line`, `msg`, `data`, `tags`.

### Database
- Postgres with schema `pmsn`.
