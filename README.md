# RBAC Service

A robust, multi-tenant Role-Based Access Control (RBAC) service built with Go, featuring event-driven architecture with RabbitMQ integration.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Documentation](#documentation)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Event System](#event-system)
- [Development](#development)
- [License](#license)

## 🎯 Overview

This service provides a comprehensive RBAC solution for multi-tenant applications, enabling fine-grained access control through roles, groups, and permissions. It supports both synchronous REST API operations and asynchronous event-driven workflows via RabbitMQ.

## ✨ Features

- **Multi-Tenancy**: Complete tenant isolation with dedicated schemas
- **Role Management**: Create and manage roles with hierarchical permissions
- **Group Management**: Organize users into groups with inherited permissions
- **Permission Resolution**: Efficient permission validation and resolution
- **Event-Driven Architecture**: Asynchronous processing with RabbitMQ
- **Audit Trail**: Complete event tracking for published and consumed events
- **Health Checks**: Built-in health monitoring and auto-reconnection
- **Docker Support**: Fully containerized with Docker Compose
- **Structured Logging**: JSON-formatted logs with context tracking

## 🏗️ Architecture

The service follows a clean 4-layered architecture:

```
┌─────────────────────────────────────────┐
│         Controller Layer (HTTP)         │
│         Event Layer (RabbitMQ)          │
├─────────────────────────────────────────┤
│       Application Service Layer         │
├─────────────────────────────────────────┤
│         Domain Service Layer            │
├─────────────────────────────────────────┤
│         Repository Layer (DB)           │
└─────────────────────────────────────────┘
```

### Key Components

- **Controllers**: HTTP request handlers (Gin framework)
- **Event Handlers**: RabbitMQ message processors
- **Application Services**: Business orchestration and event publishing
- **Domain Services**: Core business logic
- **Repositories**: Data access layer (PostgreSQL)

## 🛠️ Tech Stack

- **Language**: Go 1.23+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15
- **Message Queue**: RabbitMQ 3
- **Containerization**: Docker & Docker Compose
- **Database Driver**: pgx/v5

## 📚 Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Architecture](docs/@architecture/architecture.md)**: System design and component interaction
- **[API Specification](docs/@apis/api_spec.md)**: Complete REST API documentation
- **[Database Schema](docs/@tables/schema.md)**: Table structures and relationships
- **[RBAC Concepts](docs/@concept/concept.md)**: Core RBAC principles and terminology
- **[Event System](docs/@events/events.md)**: Event-driven architecture details

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)
- Git

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/sreerag-rajan/rbac-service.git
   cd rbac-service
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Verify the service is running**
   ```bash
   curl http://localhost:9980/health
   ```

The service will be available at `http://localhost:9980`

RabbitMQ Management UI: `http://localhost:15672` (guest/guest)

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `rbac` |
| `PORT` | Service HTTP port | `9980` |
| `QUEUE_PROVIDER` | Queue provider (`RABBITMQ` or empty) | - |
| `RABBITMQ_URL` | RabbitMQ connection URL | `amqp://guest:guest@localhost:5672/` |
| `RABBITMQ_MAX_CONNECTIONS` | Max connection pool size | `1` |
| `RABBITMQ_MAX_CHANNELS_PER_CONN` | Max channels per connection | `10` |

### Disabling Event System

To run without RabbitMQ, simply omit `QUEUE_PROVIDER` or set it to an empty string in your `.env` file.

## 🔌 API Endpoints

### Tenants
- `POST /tenants` - Create a new tenant
- `GET /tenants/:id` - Get tenant details

### Roles
- `POST /tenants/:tenant_id/roles` - Create a role
- `POST /roles/:role_id/permissions` - Assign permissions to role
- `POST /roles/:role_id/users` - Assign users to role
- `DELETE /roles/:role_id/users` - Remove users from role

### Groups
- `POST /tenants/:tenant_id/groups` - Create a group
- `POST /groups/:group_id/permissions` - Assign permissions to group
- `POST /groups/:group_id/users` - Assign users to group
- `DELETE /groups/:group_id/users` - Remove users from group

### Validation
- `POST /validate` - Validate user permissions

For complete API documentation, see [API Specification](docs/@apis/api_spec.md).

## 📡 Event System

The service supports asynchronous event processing via RabbitMQ:

### Event Types

**Request Events** (consumed):
- `rbac.user_role.assign.request`
- `rbac.user_role.remove.request`
- `rbac.user_group.assign.request`
- `rbac.user_group.remove.request`

**Completion Events** (published):
- `rbac.user_role.assign.success/failed`
- `rbac.user_role.remove.success/failed`
- `rbac.user_group.assign.success/failed`
- `rbac.user_group.remove.success/failed`

### Event Architecture

- **Exchange**: `rbac_permissions` (topic)
- **Queue**: `permissions`
- **Routing Pattern**: `rbac.*.*.request`
- **Retry Strategy**: Exponential backoff (max 3 retries)
- **Audit Tables**: `published_events`, `consumed_events`

For detailed event documentation, see [Event System](docs/@events/events.md).

## 🔧 Development

### Local Development Setup

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Run database migrations**
   ```bash
   # Migrations run automatically on startup
   # Or manually: psql -U postgres -d rbac -f migrations/001_initial_schema.sql
   ```

3. **Run the service**
   ```bash
   go run cmd/server/main.go
   ```

### Project Structure

```
rbac-service/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── app/             # Application services
│   ├── controller/      # HTTP handlers
│   ├── events/          # Event infrastructure
│   │   ├── handlers/    # Event handlers
│   │   └── rabbitmq/    # RabbitMQ implementation
│   ├── logger/          # Structured logging
│   ├── model/           # Data models and DTOs
│   ├── repository/      # Database layer
│   └── service/         # Domain services
├── migrations/          # Database migrations
├── docs/                # Documentation
├── docker-compose.yml   # Docker orchestration
└── Dockerfile          # Container definition
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o rbac-service cmd/server/main.go
```

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Contact

For questions or support, please open an issue on GitHub.

---

**Built with ❤️ using Go and RabbitMQ**
