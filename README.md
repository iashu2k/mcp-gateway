# MCP Gateway

A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations.

The project is inspired by the idea of an internal “USB-C for AI agents”: a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, understand their available tools, invoke those tools through centralized controls, and obtain auditable execution history.

> **Current status:** Phase 1 complete — Go service foundation and MCP Server Registry CRUD API are implemented and verified locally.

---

## Table of Contents

- [Project Vision](#project-vision)
- [Why This Project](#why-this-project)
- [System Architecture](#system-architecture)
- [Technology Stack](#technology-stack)
- [Current Features](#current-features)
- [Development Roadmap](#development-roadmap)
- [Repository Structure](#repository-structure)
- [Prerequisites](#prerequisites)
- [Local Setup](#local-setup)
- [Environment Variables](#environment-variables)
- [Running the Project](#running-the-project)
- [Database Migrations](#database-migrations)
- [API Reference](#api-reference)
- [Phase 0: Foundation](#phase-0-foundation)
- [Phase 1: MCP Server Registry](#phase-1-mcp-server-registry)
- [Validation and Testing](#validation-and-testing)
- [Design Decisions](#design-decisions)
- [Known Limitations](#known-limitations)
- [Planned Phases](#planned-phases)
- [Troubleshooting](#troubleshooting)
- [Contributing Workflow](#contributing-workflow)

---

## Project Vision

Internal organizations increasingly expose capabilities through APIs, automations, and MCP servers: GitHub repository operations, Jira issue management, Confluence search, Slack messaging, deployment workflows, analytics tools, and more.

Without a centralized gateway, teams often face several problems:

- Developers and agents cannot easily discover which tools exist.
- API credentials may be distributed across scripts, applications, and local environments.
- Access control is inconsistent across tools.
- Tool invocations are hard to audit.
- Teams cannot easily understand tool reliability, latency, errors, or usage.
- Mutating operations can be invoked without sufficiently clear policy controls.

MCP Gateway addresses this by serving as a control plane for internal MCP servers.

```text
                    ┌─────────────────────────────┐
                    │       React Web UI           │
                    │ Catalog -  Sandbox -  History  │
                    └──────────────┬──────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go MCP Gateway                            │
│                                                                 │
│  Server Registry -  Tool Discovery -  Authentication -  RBAC      │
│  Policy Enforcement -  Invocation Proxy -  Audit Logs -  Metrics  │
└─────────────┬──────────────────┬──────────────────┬─────────────┘
              │                  │                  │
              ▼                  ▼                  ▼
      ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
      │ GitHub MCP    │  │ Jira MCP      │  │ Slack MCP     │
      │ Server/API    │  │ Server/API    │  │ Server/API    │
      └───────────────┘  └───────────────┘  └───────────────┘
              │
              ▼
      ┌─────────────────────────────────────────────────────────┐
      │ PostgreSQL                                               │
      │ Server Metadata -  Tool Catalog -  Permissions -  Audits    │
      └─────────────────────────────────────────────────────────┘
```

---

## Why This Project

This project is designed to demonstrate practical backend and AI-platform engineering skills rather than only building another standalone LLM application.

It focuses on:

- Go backend development and service design
- PostgreSQL schema design and migrations
- API gateway patterns
- MCP ecosystem integration
- Authentication and authorization
- Secure handling of external-service credentials
- Tool invocation policy enforcement
- Auditability and observability
- React-based developer tooling
- Containerized local development
- Production-style testing and deployment practices

The long-term architecture intentionally separates deterministic infrastructure concerns from agentic reasoning:

```text
Go Gateway
├── Authentication
├── Authorization
├── Validation
├── Policy checks
├── Tool routing
├── Audit logging
└── Observability

Python Agent Service (future, optional)
├── LangGraph workflows
├── LLM orchestration
├── Retrieval-augmented generation
├── Evaluation pipelines
└── Agent planning and reasoning
```

The gateway should remain reliable, predictable, and policy-driven even when future AI agents consume it.

---

## System Architecture

### Current architecture

```text
Developer / API Client
         │
         │ HTTP / JSON
         ▼
┌───────────────────────────────┐
│ Go HTTP Service               │
│                               │
│ chi Router                    │
│ Request ID middleware         │
│ Timeout middleware            │
│ Structured logging            │
│ Health endpoint               │
│ MCP Server Registry API       │
└───────────────┬───────────────┘
                │
                │ pgx connection pool
                ▼
┌───────────────────────────────┐
│ PostgreSQL 16                 │
│                               │
│ mcp_servers                   │
│ schema_migrations             │
└───────────────────────────────┘
```

### Target architecture

```text
React Discovery UI
        │
        ▼
Go MCP Gateway API
├── Server Registry
├── Tool Catalog
├── Authentication / JWT
├── Role-Based Access Control
├── Server and Tool Permissions
├── Policy Engine
├── Tool Invocation Proxy
├── Audit Log Service
├── Metrics and Tracing
└── MCP Client / Adapter Layer
        │
        ├── GitHub MCP Server or GitHub API adapter
        ├── Jira MCP Server
        ├── Slack MCP Server
        ├── Confluence MCP Server
        └── Internal engineering tools
```

---

## Technology Stack

| Area | Technology | Purpose |
|---|---|---|
| Backend language | Go | Concurrent, strongly typed backend and gateway implementation |
| HTTP routing | `go-chi/chi` | Lightweight, idiomatic REST routing and middleware |
| Database | PostgreSQL 16 | Persistent registry, permissions, audit records, and metadata |
| PostgreSQL driver | `pgx/v5` | Native PostgreSQL client and connection pool for Go |
| Migrations | `golang-migrate` | Version-controlled database schema changes |
| Configuration | Environment variables + `godotenv` | Local configuration and secret loading |
| Logging | Go `log/slog` | Structured JSON logs |
| API testing | `curl`, `jq`, Go `testing` | Endpoint validation and automated unit tests |
| Containerization | Docker Compose | Local PostgreSQL environment |
| Frontend | React + TypeScript | Future discovery catalog and testing sandbox |
| Observability | Prometheus + OpenTelemetry | Future metrics, traces, and service health visibility |
| MCP integration | Official Go MCP SDK | Future MCP server discovery and tool invocation |

---

## Current Features

### Completed

- Go module initialized at `github.com/iashu2k/mcp-gateway/backend`
- Local PostgreSQL 16 container managed through Docker Compose
- Environment-based configuration
- PostgreSQL connection pool using `pgxpool`
- Health check endpoint with database connectivity validation
- Structured JSON application logs
- HTTP request ID middleware
- HTTP timeout middleware
- Graceful service shutdown
- Version-controlled PostgreSQL migrations
- `mcp_servers` registry table
- MCP server CRUD API
- Request validation with field-level error details
- Duplicate server-name prevention
- UUID validation for server routes
- Unit tests for service-layer create validation
- Successful `go test ./...` verification

### Not yet implemented

- MCP tool catalog
- Tool discovery from MCP servers
- Authentication
- JWT validation
- Role-based access control
- Server-level and tool-level permissions
- Real MCP invocation proxying
- GitHub/Jira/Slack/Confluence integrations
- Credential reference storage
- Audit logs
- Metrics and distributed tracing
- React UI
- CI/CD pipeline
- Production deployment

---

## Development Roadmap

| Phase | Status | Focus | Main Outcome |
|---|---|---|---|
| Phase 0 | Complete | Foundation | Go API, PostgreSQL, Docker Compose, health check |
| Phase 1 | Complete | MCP Server Registry | Persistent CRUD API for registered MCP servers |
| Phase 2 | Next | Tool Catalog | Register, list, and govern tools available from each server |
| Phase 3 | Planned | Authentication and RBAC | JWT authentication and role-based access policies |
| Phase 4 | Planned | Invocation Gateway | Validate, authorize, proxy, and log tool calls |
| Phase 5 | Planned | First Live Integration | GitHub or Jira tool invocation through the gateway |
| Phase 6 | Planned | Observability | Audit logs, Prometheus metrics, OpenTelemetry traces |
| Phase 7 | Planned | React UI | Discovery catalog, JSON sandbox, invocation history |
| Phase 8 | Planned | Delivery and Polish | CI, tests, Docker images, deployment, documentation |

---

## Repository Structure

```text
mcp-gateway/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   │
│   │   ├── domain/
│   │   │   └── server.go
│   │   │
│   │   ├── httpapi/
│   │   │   ├── errors.go
│   │   │   ├── handlers.go
│   │   │   ├── router.go
│   │   │   ├── router_test.go
│   │   │   └── server_handler.go
│   │   │
│   │   ├── platform/
│   │   │   └── database/
│   │   │       └── postgres.go
│   │   │
│   │   ├── repository/
│   │   │   └── server_repository.go
│   │   │
│   │   └── service/
│   │       ├── server_service.go
│   │       └── server_service_test.go
│   │
│   ├── migrations/
│   │   ├── 000001_create_mcp_servers.down.sql
│   │   └── 000001_create_mcp_servers.up.sql
│   │
│   ├── go.mod
│   └── go.sum
│
├── docs/
│
├── frontend/
│
├── infra/
│
├── .env.example
├── .gitignore
├── docker-compose.yml
└── README.md
```

### Package responsibilities

| Package | Responsibility |
|---|---|
| `cmd/api` | Application composition, server startup, dependency initialization, graceful shutdown |
| `internal/config` | Environment variable loading and configuration validation |
| `internal/domain` | Core application types, request models, constants, and API response domain models |
| `internal/httpapi` | HTTP routes, handlers, request decoding, JSON responses, and HTTP error mapping |
| `internal/platform/database` | PostgreSQL connection pool creation and connection verification |
| `internal/repository` | SQL queries and PostgreSQL persistence logic |
| `internal/service` | Business rules, validation, and application use cases |
| `migrations` | Ordered and versioned PostgreSQL schema changes |

The project follows this dependency direction:

```text
HTTP Handler
    ↓
Service
    ↓
Repository
    ↓
PostgreSQL
```

HTTP handlers do not contain SQL. This keeps transport logic, business rules, and persistence concerns separate and testable.

---

## Prerequisites

Install the following before running the project:

- Go 1.23 or newer
- Docker Desktop or Docker Engine with Docker Compose
- Git
- `curl`
- `jq` for readable JSON output
- `golang-migrate` CLI for database schema migrations

Verify the local environment:

```bash
go version
docker --version
docker compose version
git --version
curl --version
jq --version
```

Install the migration CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Ensure Go-installed binaries are accessible:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"

migrate -version
```

For a persistent shell configuration, add the `PATH` export to your shell profile, such as `~/.zshrc` or `~/.bashrc`.

---

## Local Setup

### 1. Clone the repository

```bash
git clone https://github.com/iashu2k/mcp-gateway.git
cd mcp-gateway
```

### 2. Create local environment configuration

```bash
cp .env.example .env
```

The local `.env` file is ignored by Git and must never contain production secrets.

### 3. Start PostgreSQL

```bash
docker compose up -d postgres
```

Check the service:

```bash
docker compose ps
```

Expected result:

```text
mcp-gateway-postgres ... Up (healthy)
```

View database logs if needed:

```bash
docker compose logs -f postgres
```

Stop log streaming with `Ctrl+C`; this does not stop the running container.

### 4. Load environment variables

Run this from the repository root:

```bash
set -a
source .env
set +a
```

### 5. Apply database migrations

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  up
```

Verify the current schema migration version:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

### 6. Install backend dependencies

```bash
cd backend
go mod download
go mod tidy
cd ..
```

### 7. Start the API

```bash
cd backend
go run ./cmd/api
```

The service runs at:

```text
http://localhost:8080
```

---

## Environment Variables

Create `.env` from `.env.example`.

```dotenv
APP_ENV=development
HTTP_PORT=8080

POSTGRES_DB=mcp_gateway
POSTGRES_USER=mcp_gateway
POSTGRES_PASSWORD=mcp_gateway_dev_password
DATABASE_URL=postgres://mcp_gateway:mcp_gateway_dev_password@localhost:5432/mcp_gateway?sslmode=disable
```

| Variable | Required | Description |
|---|---:|---|
| `APP_ENV` | Yes | Application environment, currently `development` |
| `HTTP_PORT` | Yes | Port for the Go HTTP server |
| `POSTGRES_DB` | Yes | PostgreSQL database name used by Docker Compose |
| `POSTGRES_USER` | Yes | PostgreSQL database user |
| `POSTGRES_PASSWORD` | Yes | PostgreSQL database password for local development |
| `DATABASE_URL` | Yes | Go application and migration connection string |

### Credential safety

The `.env` file is local-only and must remain in `.gitignore`.

Future external API credentials, such as `GITHUB_TOKEN`, `JIRA_TOKEN`, and `SLACK_BOT_TOKEN`, will also be loaded through environment variables during local development. In a deployed environment, they should be supplied by a secrets manager rather than committed configuration.

---

## Running the Project

### Start PostgreSQL

```bash
docker compose up -d postgres
```

### Start the backend

```bash
cd backend
go run ./cmd/api
```

### Check API health

```bash
curl -s http://localhost:8080/health | jq
```

Expected response:

```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2026-08-01T00:00:00Z"
}
```

### Check API root

```bash
curl -s http://localhost:8080/api/v1/ | jq
```

Expected response:

```json
{
  "message": "MCP Gateway API"
}
```

### Stop local services

Stop the backend with:

```text
Ctrl+C
```

Stop PostgreSQL while preserving data:

```bash
docker compose down
```

Stop PostgreSQL and remove all local database data:

```bash
docker compose down -v
```

> Warning: `docker compose down -v` deletes the persistent PostgreSQL volume and all locally stored registry data.

---

## Database Migrations

Database schemas are versioned under:

```text
backend/migrations/
```

Migration files use paired `up` and `down` files:

```text
000001_create_mcp_servers.up.sql
000001_create_mcp_servers.down.sql
```

### Apply pending migrations

```bash
set -a
source .env
set +a

migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  up
```

### Check migration version

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

### Roll back the latest migration

Use this only in local development:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  down 1
```

### Inspect tables manually

```bash
docker exec -it mcp-gateway-postgres \
  psql -U mcp_gateway -d mcp_gateway
```

Then in `psql`:

```sql
\dt
\d mcp_servers
SELECT * FROM mcp_servers;
```

---

## API Reference

Base URL:

```text
http://localhost:8080/api/v1
```

### Health check

```http
GET /health
```

Checks whether the HTTP service is running and PostgreSQL is reachable.

Example:

```bash
curl -s http://localhost:8080/health | jq
```

---

### Create MCP server

```http
POST /api/v1/servers
Content-Type: application/json
```

Request body:

```json
{
  "name": "github-mcp",
  "description": "Read-only GitHub repository, issue, and pull-request tools",
  "baseUrl": "http://localhost:3001",
  "transportType": "streamable_http",
  "ownerTeam": "developer-platform"
}
```

Example request:

```bash
curl -s -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "github-mcp",
    "description": "Read-only GitHub repository, issue, and pull-request tools",
    "baseUrl": "http://localhost:3001",
    "transportType": "streamable_http",
    "ownerTeam": "developer-platform"
  }' | jq
```

Successful response: `201 Created`

```json
{
  "id": "7db41f97-2356-4492-97c7-0d0843b56c3e",
  "name": "github-mcp",
  "description": "Read-only GitHub repository, issue, and pull-request tools",
  "baseUrl": "http://localhost:3001",
  "transportType": "streamable_http",
  "status": "active",
  "ownerTeam": "developer-platform",
  "createdAt": "2026-08-01T00:00:00Z",
  "updatedAt": "2026-08-01T00:00:00Z"
}
```

#### Supported transport types

| Value | Meaning |
|---|---|
| `streamable_http` | Default MCP transport for HTTP-based server communication |
| `sse` | Server-Sent Events transport |
| `stdio` | Standard input/output transport for locally managed MCP processes |

#### Default behavior

- `transportType` defaults to `streamable_http` if omitted.
- New servers are created with status `active`.
- `name` must be unique across registered servers.

---

### List MCP servers

```http
GET /api/v1/servers
```

Example:

```bash
curl -s http://localhost:8080/api/v1/servers | jq
```

Successful response: `200 OK`

```json
{
  "data": [],
  "count": 0
}
```

When records exist:

```json
{
  "data": [
    {
      "id": "7db41f97-2356-4492-97c7-0d0843b56c3e",
      "name": "github-mcp",
      "description": "Read-only GitHub repository, issue, and pull-request tools",
      "baseUrl": "http://localhost:3001",
      "transportType": "streamable_http",
      "status": "active",
      "ownerTeam": "developer-platform",
      "createdAt": "2026-08-01T00:00:00Z",
      "updatedAt": "2026-08-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

---

### Get one MCP server

```http
GET /api/v1/servers/{serverID}
```

Example:

```bash
export SERVER_ID="7db41f97-2356-4492-97c7-0d0843b56c3e"

curl -s "http://localhost:8080/api/v1/servers/$SERVER_ID" | jq
```

Successful response: `200 OK`

---

### Update an MCP server

```http
PATCH /api/v1/servers/{serverID}
Content-Type: application/json
```

All fields are optional. Only submitted fields are changed.

Example:

```bash
curl -s -X PATCH "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "GitHub tools for repository search, issue lookup, and pull request inspection",
    "status": "inactive"
  }' | jq
```

Successful response: `200 OK`

Supported `status` values:

| Value | Meaning |
|---|---|
| `active` | Server is enabled and available for use |
| `inactive` | Server is intentionally disabled |
| `unhealthy` | Server has failed health checks or is unavailable |

---

### Delete an MCP server

```http
DELETE /api/v1/servers/{serverID}
```

Example:

```bash
curl -i -X DELETE "http://localhost:8080/api/v1/servers/$SERVER_ID"
```

Successful response:

```text
HTTP/1.1 204 No Content
```

---

### API error format

Errors follow this shape:

```json
{
  "error": "validation_error",
  "message": "request validation failed",
  "details": [
    {
      "field": "baseUrl",
      "message": "must be a valid absolute HTTP URL"
    }
  ]
}
```

Current error codes:

| HTTP status | Error code | Meaning |
|---:|---|---|
| `400` | `invalid_json` | Invalid JSON or unknown request field |
| `400` | `validation_error` | One or more request fields are invalid |
| `400` | `invalid_server_id` | Route parameter is not a valid UUID |
| `404` | `server_not_found` | No registered MCP server has the requested ID |
| `409` | `duplicate_server_name` | Another registered server already uses the name |
| `500` | `internal_error` | Unexpected application or persistence failure |

---

## Phase 0: Foundation

**Status:** Complete

### Goal

Create a reliable local development baseline before adding product functionality.

### Implemented components

- Git repository initialization
- Go module initialization
- Go service entry point
- Environment configuration
- Docker Compose PostgreSQL service
- PostgreSQL connection pool
- Health endpoint
- Request ID middleware
- Request timeout middleware
- JSON structured logs
- Graceful shutdown
- Baseline tests

### Key implementation decisions

#### Go module

The backend module path is:

```text
github.com/iashu2k/mcp-gateway/backend
```

All internal imports use this module path.

#### PostgreSQL with Docker Compose

PostgreSQL runs independently from the Go process:

```text
Go API on localhost:8080
PostgreSQL on localhost:5432
```

This lets backend code be run directly with `go run` while the database remains reproducible through Docker.

#### Connection pooling

The application creates one `pgxpool.Pool` at startup and shares it across repositories.

```text
Application starts
    ↓
Create PostgreSQL pool
    ↓
Ping database
    ↓
Start HTTP server
    ↓
Close pool during graceful shutdown
```

This is preferable to opening a new database connection for each HTTP request.

#### Graceful shutdown

The application listens for `SIGINT` and `SIGTERM`.

On shutdown:

1. The server stops accepting new requests.
2. Existing requests are allowed up to 10 seconds to complete.
3. The PostgreSQL connection pool is closed.
4. The process exits.

### Phase 0 validation commands

```bash
docker compose up -d postgres

cd backend
go test ./...
go vet ./...
go run ./cmd/api
```

In a separate terminal:

```bash
curl -s http://localhost:8080/health | jq
curl -s http://localhost:8080/api/v1/ | jq
```

---

## Phase 1: MCP Server Registry

**Status:** Complete

### Goal

Create the persistent source of truth for MCP servers known to the gateway.

An MCP server record describes an external or internal system that can expose tools through a supported transport. Examples include:

- GitHub tool server
- Jira tool server
- Slack tool server
- Confluence tool server
- Internal deployment automation server
- Internal documentation search server
- Security remediation tool server

### Database schema

The initial table is `mcp_servers`.

```sql
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    base_url TEXT NOT NULL,
    transport_type TEXT NOT NULL DEFAULT 'streamable_http',
    status TEXT NOT NULL DEFAULT 'active',
    owner_team TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Data model

| Field | Description |
|---|---|
| `id` | UUID primary key generated by PostgreSQL |
| `name` | Unique, human-readable server identifier |
| `description` | Explanation of the server’s purpose and capabilities |
| `baseUrl` | Location where the future gateway will connect to the MCP server |
| `transportType` | `streamable_http`, `sse`, or `stdio` |
| `status` | `active`, `inactive`, or `unhealthy` |
| `ownerTeam` | Team responsible for server operations and maintenance |
| `createdAt` | UTC creation timestamp |
| `updatedAt` | UTC timestamp of the latest update |

### Layered implementation

```text
Request
   ↓
chi route
   ↓
ServerHandler
   ↓
ServerService
   ↓
ServerRepository
   ↓
pgxpool
   ↓
PostgreSQL
```

#### HTTP handler layer

Responsibilities:

- Decode JSON request bodies
- Read URL parameters
- Invoke service methods
- Return JSON success responses
- Translate known application errors into HTTP status codes

#### Service layer

Responsibilities:

- Trim whitespace
- Apply defaults
- Validate required fields
- Validate URLs
- Validate transport and status enums
- Support partial update semantics
- Convert repository failures into business-level errors where appropriate

#### Repository layer

Responsibilities:

- Execute parameterized SQL queries
- Map PostgreSQL records to Go domain structs
- Return a typed `ErrServerNotFound` when no record exists
- Keep SQL out of handlers and services

### Validation rules

| Field | Rules |
|---|---|
| `name` | Required, non-empty after trimming, maximum 100 characters, unique |
| `description` | Required, non-empty after trimming, maximum 1000 characters |
| `baseUrl` | Required for HTTP/SSE transports, must be an absolute `http` or `https` URL |
| `transportType` | Must be `streamable_http`, `sse`, or `stdio` |
| `status` | Must be `active`, `inactive`, or `unhealthy` |
| `ownerTeam` | Required, non-empty after trimming, maximum 100 characters |

### Unit tests implemented

The service test suite currently validates:

- New server records default to `streamable_http`.
- New server records default to `active`.
- Invalid URLs are rejected.
- Missing required server names are rejected.

Run all tests:

```bash
cd backend
go test ./...
```

### Verified Phase 1 output

The current test and registry output confirms the project builds and the registry is reachable:

```text
?       github.com/iashu2k/mcp-gateway/backend/cmd/api  [no test files]
?       github.com/iashu2k/mcp-gateway/backend/internal/config  [no test files]
?       github.com/iashu2k/mcp-gateway/backend/internal/domain  [no test files]
ok      github.com/iashu2k/mcp-gateway/backend/internal/httpapi (cached)
?       github.com/iashu2k/mcp-gateway/backend/internal/platform/database       [no test files]
?       github.com/iashu2k/mcp-gateway/backend/internal/repository      [no test files]
ok      github.com/iashu2k/mcp-gateway/backend/internal/service (cached)
```

```json
{
  "count": 0,
  "data": []
}
```

An empty registry is the correct result when no MCP servers have been created yet.

### Phase 1 acceptance criteria

- [x] PostgreSQL migration creates `mcp_servers`
- [x] Create endpoint persists a server and returns `201 Created`
- [x] List endpoint returns registered servers and a count
- [x] Get-by-ID endpoint returns a single server
- [x] Patch endpoint updates only provided fields
- [x] Delete endpoint removes a server and returns `204 No Content`
- [x] Duplicate names return `409 Conflict`
- [x] Invalid data returns `400 Bad Request`
- [x] Invalid UUIDs return `400 Bad Request`
- [x] Missing servers return `404 Not Found`
- [x] Tests pass with `go test ./...`

---

## Validation and Testing

### Run unit tests

```bash
cd backend
go test ./...
```

### Run static analysis

```bash
cd backend
go vet ./...
```

### Format all Go code

```bash
cd backend
gofmt -w .
```

### Run the complete local quality check

```bash
cd backend

gofmt -w .
go mod tidy
go test ./...
go vet ./...
```

### Manual API smoke test

```bash
curl -s http://localhost:8080/health | jq

curl -s -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "github-mcp",
    "description": "Read-only GitHub repository, issue, and pull-request tools",
    "baseUrl": "http://localhost:3001",
    "transportType": "streamable_http",
    "ownerTeam": "developer-platform"
  }' | jq

curl -s http://localhost:8080/api/v1/servers | jq
```

---

## Design Decisions

### Why Go

The gateway is primarily an I/O-heavy control-plane service. It will perform authentication checks, query PostgreSQL, contact external MCP servers and APIs, enforce policy, log audit events, and expose metrics.

Go is a good fit because it provides:

- Static typing
- Fast startup
- Straightforward container deployment
- Lightweight concurrency through goroutines
- Strong standard-library HTTP support
- Explicit error handling
- Good fit for API gateways, platform services, and developer tooling

Python remains a strong choice for future AI reasoning services, LangGraph workflows, RAG pipelines, evaluations, and model experimentation. This architecture uses Go where deterministic infrastructure matters most and leaves room for Python where AI tooling has the strongest ecosystem.

### Why PostgreSQL

PostgreSQL is used because it supports the project’s future needs:

- Relational server, user, permission, and audit models
- JSON/JSONB tool input schemas
- Indexing for usage history and server status
- Transactions for policy-controlled tool execution
- Strong data constraints
- Mature Docker and cloud deployment support

### Why a layered backend

Separating handlers, services, repositories, and domain models provides clear responsibilities:

```text
Handlers: HTTP concerns
Services: business rules
Repositories: persistence
Domain: shared application concepts
```

This structure makes later additions such as JWT authentication, authorization checks, audit records, and MCP clients easier to test and evolve.

### Why read-only first

The first real integrations will focus on low-risk, read-only actions:

- Search repositories
- List GitHub issues
- Retrieve pull request details
- Search Jira tickets
- Read Confluence pages
- Search Slack messages

Mutating actions such as creating issues, posting Slack messages, editing documentation, or modifying repositories will require explicit confirmation and stricter policy controls.

---

## Known Limitations

The following are deliberate limitations at the end of Phase 1:

- No authentication: every local caller can currently access all registry endpoints.
- No authorization: admin/developer/viewer roles do not yet exist.
- No pagination or filtering on `GET /api/v1/servers`.
- No health checks against registered MCP server URLs.
- No validation that a registered URL actually hosts an MCP server.
- No tool metadata exists yet.
- No credential management exists yet.
- No external APIs or MCP servers are connected.
- No audit log records are stored.
- No metrics endpoint exists.
- No React frontend exists.
- Duplicate-name detection currently relies on checking PostgreSQL error text.

The duplicate-name implementation is sufficient for the current local phase but will be improved later by checking PostgreSQL’s typed error code `23505` through `pgconn.PgError`.

---

## Planned Phases

### Phase 2: Tool Catalog

Phase 2 will create a persistent `mcp_tools` table and associate tools with registered MCP servers.

Planned fields:

```text
mcp_tools
- id
- server_id
- name
- description
- input_schema_json
- enabled
- risk_level
- created_at
- updated_at
```

Planned endpoints:

```text
GET  /api/v1/servers/{serverID}/tools
POST /api/v1/servers/{serverID}/tools
GET  /api/v1/servers/{serverID}/tools/{toolID}
PATCH /api/v1/servers/{serverID}/tools/{toolID}
DELETE /api/v1/servers/{serverID}/tools/{toolID}
```

Risk levels:

| Risk level | Examples | Future policy |
|---|---|---|
| `low` | Search repositories, list issues, read pages | Allowed for authorized developers |
| `medium` | Create Jira issue, post Slack message | Confirmation required |
| `high` | Delete repositories, bulk update tickets, production changes | Denied by default or requires elevated approval |

### Phase 3: Authentication and RBAC

Planned implementation:

- JWT bearer-token authentication
- Local development users
- `admin`, `developer`, and `viewer` roles
- Server-level permissions
- Tool-level invocation permissions
- Admin-only registry management

### Phase 4: Invocation Gateway

Planned request lifecycle:

```text
Incoming tool request
    ↓
Authenticate caller
    ↓
Authorize server and tool access
    ↓
Validate arguments against tool JSON Schema
    ↓
Evaluate tool risk and policy
    ↓
Create audit record
    ↓
Invoke MCP server or adapter
    ↓
Redact sensitive data
    ↓
Store outcome, latency, and errors
    ↓
Return structured result
```

### Phase 5: First live integration

Initial target: GitHub read-only tools.

Candidate tools:

- `search_repositories`
- `list_issues`
- `get_issue`
- `list_pull_requests`
- `get_pull_request`

A future option is registering and routing to the existing Jira MCP server through this gateway.

### Phase 6: Observability

Planned observability features:

- PostgreSQL audit records
- Request IDs
- Structured logs
- Prometheus metrics
- OpenTelemetry traces
- Server health status
- Tool success rates
- Error rates
- p50, p95, and p99 invocation latency
- Most-used tool dashboard

### Phase 7: React UI

Planned screens:

- MCP server catalog
- Server detail page
- Tool catalog
- Tool input-schema viewer
- JSON invocation sandbox
- Confirmation dialog for write-capable tools
- Invocation history
- Admin server registration form
- Lightweight observability dashboard

### Phase 8: Delivery and polish

Planned engineering improvements:

- GitHub Actions pipeline
- Dockerfile for backend
- Dockerfile for frontend
- Integration tests with Testcontainers
- OpenAPI documentation
- Deployment to cloud infrastructure
- Screenshots and demo video
- Architecture decision records
- Security and operations documentation

---

## Troubleshooting

### PostgreSQL container does not start

Check logs:

```bash
docker compose logs postgres
```

Check whether port `5432` is already in use:

```bash
lsof -i :5432
```

If another local PostgreSQL instance uses the port, either stop it or change the host mapping in `docker-compose.yml`.

---

### Application cannot connect to PostgreSQL

Verify the container is healthy:

```bash
docker compose ps
```

Verify the database connection string:

```bash
set -a
source .env
set +a

echo "$DATABASE_URL"
```

Verify the connection manually:

```bash
docker exec -it mcp-gateway-postgres \
  psql -U mcp_gateway -d mcp_gateway \
  -c "SELECT 1;"
```

---

### `migrate: command not found`

Install `golang-migrate` and update `PATH`:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

export PATH="$PATH:$(go env GOPATH)/bin"

migrate -version
```

---

### Migration reports `dirty database version`

A migration may have failed partway through. Inspect the migration state:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

For local-only development, you can reset the database:

```bash
docker compose down -v
docker compose up -d postgres

set -a
source .env
set +a

migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  up
```

> Do not use destructive reset commands against a shared or production database.

---

### `go run ./cmd/api` cannot load `.env`

Run the command from `backend/`:

```bash
cd backend
go run ./cmd/api
```

The application attempts to load both `../.env` and `.env`, allowing it to work when launched from the backend directory or repository root depending on your command.

---

### Port 8080 is already in use

Find the existing process:

```bash
lsof -i :8080
```

Stop that process or change the value in `.env`:

```dotenv
HTTP_PORT=8081
```

Then restart the application.

---

### Tests fail after changing imports

Verify all Go internal imports use the correct module path:

```text
github.com/iashu2k/mcp-gateway/backend
```

Then run:

```bash
cd backend
go mod tidy
gofmt -w .
go test ./...
```

---

## Contributing Workflow

Use small, phase-focused commits.

Suggested workflow:

```bash
git checkout -b feat/tool-catalog

# Implement one focused change.

cd backend
gofmt -w .
go mod tidy
go test ./...
go vet ./...

cd ..
git status
git add .
git commit -m "feat: add MCP tool catalog"
git push -u origin feat/tool-catalog
```

Suggested commit convention:

```text
feat: add MCP server registry CRUD API
feat: add MCP tool catalog
feat: add JWT authentication middleware
feat: add tool invocation audit log
fix: validate server URL before persistence
test: cover invalid tool risk levels
docs: document local development workflow
chore: configure Docker Compose PostgreSQL
```

---

## Project Status

```text
Phase 0: Complete
Phase 1: Complete
Phase 2: Ready to begin
```

The next milestone is to add the MCP Tool Catalog so each registered server can advertise its available tools, input schemas, enablement status, and risk level.