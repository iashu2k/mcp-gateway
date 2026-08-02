# MCP Gateway

A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations.

The project is inspired by the idea of an internal “USB-C for AI agents”: a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history.

> **Current status:** Phase 2 complete — MCP Server Registry and MCP Tool Catalog are implemented, tested, and validated locally.

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
- [Phase 2: MCP Tool Catalog](#phase-2-mcp-tool-catalog)
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

- Developers and AI agents cannot easily discover which tools exist.
- API credentials may be distributed across scripts, applications, and local environments.
- Access control is inconsistent across tools.
- Tool invocations are difficult to audit.
- Teams cannot easily understand tool reliability, latency, errors, or usage.
- Mutating operations can be invoked without sufficiently clear policy controls.

MCP Gateway addresses this by acting as a control plane for internal MCP servers.

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
│  Server Registry -  Tool Catalog -  Authentication -  RBAC         │
│  Policy Enforcement -  Invocation Proxy -  Audit Logs -  Metrics   │
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
      │ Servers -  Tools -  Permissions -  Audit Records -  Metadata │
      └─────────────────────────────────────────────────────────┘
```

---

## Why This Project

This project demonstrates practical backend and AI-platform engineering skills rather than only building a standalone LLM application.

It focuses on:

- Go backend development and service design
- PostgreSQL schema design and migrations
- API gateway patterns
- MCP server and tool registration
- JSON Schema-driven tool definitions
- Authentication and authorization
- Secure external-service credential handling
- Tool invocation policy enforcement
- Auditability and observability
- React-based developer tooling
- Containerized local development
- Production-style testing and deployment practices

The long-term architecture intentionally separates deterministic infrastructure from agentic reasoning.

```text
Go Gateway
├── Authentication
├── Authorization
├── Validation
├── Tool catalog
├── Policy checks
├── Tool routing
├── Audit logging
└── Observability

Optional Python Agent Service
├── LangGraph workflows
├── LLM orchestration
├── Retrieval-augmented generation
├── Evaluation pipelines
└── Agent planning and reasoning
```

The gateway remains reliable, predictable, and policy-driven even when future AI agents consume its tools.

---

## System Architecture

### Current architecture

```text
Developer / API Client
         │
         │ HTTP / JSON
         ▼
┌────────────────────────────────────┐
│ Go MCP Gateway                     │
│                                    │
│ chi Router                         │
│ Request ID middleware              │
│ Timeout middleware                 │
│ Structured logging                 │
│ Health endpoint                    │
│ Server Registry API                │
│ Tool Catalog API                   │
│ Request validation                 │
└────────────────┬───────────────────┘
                 │
                 │ pgx connection pool
                 ▼
┌────────────────────────────────────┐
│ PostgreSQL 16                      │
│                                    │
│ mcp_servers                        │
│ mcp_tools                          │
│ schema_migrations                  │
└────────────────────────────────────┘
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
| Database | PostgreSQL 16 | Persistent registry, tool catalog, permissions, audit records, and metadata |
| PostgreSQL driver | `pgx/v5` | Native PostgreSQL driver and connection pool for Go |
| Migrations | `golang-migrate` | Version-controlled database schema changes |
| JSON storage | PostgreSQL `jsonb` | Tool input schemas and future structured metadata |
| Configuration | Environment variables + `godotenv` | Local configuration and secret loading |
| Logging | Go `log/slog` | Structured JSON logs |
| API testing | `curl`, `jq`, Go `testing` | Endpoint validation and automated tests |
| Containerization | Docker Compose | Reproducible local PostgreSQL environment |
| Frontend | React + TypeScript | Future discovery catalog and testing sandbox |
| Observability | Prometheus + OpenTelemetry | Future metrics, traces, and service health visibility |
| MCP integration | Official Go MCP SDK | Future MCP discovery and tool invocation |

---

## Current Features

### Completed

- Go module initialized at `github.com/iashu2k/mcp-gateway/backend`
- Docker Compose-managed PostgreSQL 16 database
- Environment-based configuration
- PostgreSQL connection pool using `pgxpool`
- Health endpoint with database connectivity validation
- Structured JSON application logs
- HTTP request ID middleware
- HTTP timeout middleware
- Graceful HTTP shutdown
- Version-controlled PostgreSQL schema migrations
- MCP server registry with CRUD operations
- MCP tool catalog with CRUD operations
- Server-to-tool foreign key relationship
- Cascading tool deletion when a parent server is deleted
- JSON Schema storage and validation for tool inputs
- Tool risk classification: `low`, `medium`, and `high`
- Tool enable/disable state
- Field-level validation errors
- Duplicate server-name prevention
- Duplicate per-server tool-name prevention
- Typed PostgreSQL unique-constraint handling
- Unit tests for server and tool service validation
- Successful local `go test ./...` execution

### Not yet implemented

- MCP tool discovery from remote MCP servers
- Authentication
- JWT validation
- Role-based access control
- Server-level and tool-level permissions
- Tool invocation proxying
- GitHub/Jira/Slack/Confluence integrations
- Credential reference storage
- Audit records
- Metrics and distributed tracing
- React UI
- CI/CD pipeline
- Production deployment

---

## Development Roadmap

| Phase | Status | Focus | Main Outcome |
|---|---|---|---|
| Phase 0 | Complete | Foundation | Go API, PostgreSQL, Docker Compose, health endpoint |
| Phase 1 | Complete | MCP Server Registry | Persistent CRUD API for registered MCP servers |
| Phase 2 | Complete | MCP Tool Catalog | Per-server tool metadata, JSON schemas, risk levels, and enablement |
| Phase 3 | Next | Authentication and RBAC | JWT authentication and role-based authorization |
| Phase 4 | Planned | Invocation Gateway | Validate, authorize, proxy, and log tool calls |
| Phase 5 | Planned | First Live Integration | GitHub or Jira invocation through the gateway |
| Phase 6 | Planned | Observability | Audit logs, Prometheus metrics, OpenTelemetry traces |
| Phase 7 | Planned | React UI | Discovery catalog, JSON sandbox, invocation history |
| Phase 8 | Planned | Delivery and Polish | CI, tests, containers, deployment, documentation |

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
│   │   │   ├── server.go
│   │   │   └── tool.go
│   │   │
│   │   ├── httpapi/
│   │   │   ├── errors.go
│   │   │   ├── handlers.go
│   │   │   ├── router.go
│   │   │   ├── router_test.go
│   │   │   ├── server_handler.go
│   │   │   └── tool_handler.go
│   │   │
│   │   ├── platform/
│   │   │   └── database/
│   │   │       └── postgres.go
│   │   │
│   │   ├── repository/
│   │   │   ├── server_repository.go
│   │   │   └── tool_repository.go
│   │   │
│   │   └── service/
│   │       ├── server_service.go
│   │       ├── server_service_test.go
│   │       ├── tool_service.go
│   │       └── tool_service_test.go
│   │
│   ├── migrations/
│   │   ├── 000001_create_mcp_servers.down.sql
│   │   ├── 000001_create_mcp_servers.up.sql
│   │   ├── 000002_create_mcp_tools.down.sql
│   │   └── 000002_create_mcp_tools.up.sql
│   │
│   ├── go.mod
│   └── go.sum
│
├── docs/
├── frontend/
├── infra/
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
| `internal/domain` | Core domain structs, request models, constants, and shared concepts |
| `internal/httpapi` | Routes, handlers, JSON decoding, JSON responses, and HTTP error mapping |
| `internal/platform/database` | PostgreSQL connection-pool creation and verification |
| `internal/repository` | Parameterized SQL queries and PostgreSQL persistence |
| `internal/service` | Validation, defaults, business rules, and use cases |
| `migrations` | Ordered, versioned database schema evolution |

Dependency direction:

```text
HTTP Handler
    ↓
Service
    ↓
Repository
    ↓
PostgreSQL
```

HTTP handlers do not contain SQL. Services do not need to know HTTP response codes. Repositories do not need to know routing details.

---

## Prerequisites

Install:

- Go 1.23 or newer
- Docker Desktop or Docker Engine with Docker Compose
- Git
- `curl`
- `jq`
- `golang-migrate`

Verify:

```bash
go version
docker --version
docker compose version
git --version
curl --version
jq --version
```

Install `golang-migrate`:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Ensure installed Go binaries are on `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"

migrate -version
```

---

## Local Setup

### Clone the project

```bash
git clone https://github.com/iashu2k/mcp-gateway.git
cd mcp-gateway
```

### Create local configuration

```bash
cp .env.example .env
```

### Start PostgreSQL

```bash
docker compose up -d postgres
docker compose ps
```

Expected:

```text
mcp-gateway-postgres ... Up (healthy)
```

### Load environment variables

Run from the repository root:

```bash
set -a
source .env
set +a
```

### Apply migrations

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  up
```

Verify migration state:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

Expected after Phase 2:

```text
2
```

### Download backend dependencies

```bash
cd backend
go mod download
go mod tidy
cd ..
```

### Start the API

```bash
cd backend
go run ./cmd/api
```

The API runs at:

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
| `APP_ENV` | Yes | Runtime environment name, currently `development` |
| `HTTP_PORT` | Yes | Port used by the Go HTTP API |
| `POSTGRES_DB` | Yes | PostgreSQL database name |
| `POSTGRES_USER` | Yes | PostgreSQL user |
| `POSTGRES_PASSWORD` | Yes | Local PostgreSQL password |
| `DATABASE_URL` | Yes | Connection URL used by the Go service and migration CLI |

### Credential safety

`.env` is local-only and must never be committed.

Future credentials such as `GITHUB_TOKEN`, `JIRA_TOKEN`, and `SLACK_BOT_TOKEN` will be supplied through environment variables for local development and a secrets manager for deployed environments.

---

## Running the Project

### Start infrastructure

```bash
docker compose up -d postgres
```

### Run the backend

```bash
cd backend
go run ./cmd/api
```

### Verify API health

```bash
curl -s http://localhost:8080/health | jq
```

Expected shape:

```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2026-08-01T00:00:00Z"
}
```

### Verify API root

```bash
curl -s http://localhost:8080/api/v1/ | jq
```

Expected:

```json
{
  "message": "MCP Gateway API"
}
```

### Stop services

Stop the Go backend with:

```text
Ctrl+C
```

Stop PostgreSQL while keeping its data:

```bash
docker compose down
```

Stop PostgreSQL and permanently remove local data:

```bash
docker compose down -v
```

> Warning: `docker compose down -v` destroys the local PostgreSQL volume and all registered servers and tools.

---

## Database Migrations

Database migrations are located in:

```text
backend/migrations/
```

Current migrations:

```text
000001_create_mcp_servers.up.sql
000001_create_mcp_servers.down.sql

000002_create_mcp_tools.up.sql
000002_create_mcp_tools.down.sql
```

### Apply migrations

```bash
set -a
source .env
set +a

migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  up
```

### View migration version

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

### Roll back the newest migration

Use only in local development:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  down 1
```

### Inspect database tables

```bash
docker exec -it mcp-gateway-postgres \
  psql -U mcp_gateway -d mcp_gateway
```

Then:

```sql
\dt
\d mcp_servers
\d mcp_tools

SELECT * FROM mcp_servers;
SELECT * FROM mcp_tools;
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

Example:

```bash
curl -s http://localhost:8080/health | jq
```

---

## MCP Server Registry API

### Create a server

```http
POST /api/v1/servers
Content-Type: application/json
```

Example:

```bash
curl -s -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "github-mcp",
    "description": "GitHub tools exposed through the MCP Gateway",
    "baseUrl": "http://localhost:3001",
    "transportType": "streamable_http",
    "ownerTeam": "developer-platform"
  }' | jq
```

Successful response: `201 Created`

```json
{
  "id": "d2b8ed6d-8a64-4bfc-9672-cb9f37139172",
  "name": "github-mcp",
  "description": "GitHub tools exposed through the MCP Gateway",
  "baseUrl": "http://localhost:3001",
  "transportType": "streamable_http",
  "status": "active",
  "ownerTeam": "developer-platform",
  "createdAt": "2026-08-01T20:00:00-04:00",
  "updatedAt": "2026-08-01T20:00:00-04:00"
}
```

### List servers

```http
GET /api/v1/servers
```

```bash
curl -s http://localhost:8080/api/v1/servers | jq
```

### Get one server

```http
GET /api/v1/servers/{serverID}
```

```bash
curl -s "http://localhost:8080/api/v1/servers/$SERVER_ID" | jq
```

### Update a server

```http
PATCH /api/v1/servers/{serverID}
Content-Type: application/json
```

```bash
curl -s -X PATCH "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "inactive"
  }' | jq
```

### Delete a server

```http
DELETE /api/v1/servers/{serverID}
```

```bash
curl -i -X DELETE "http://localhost:8080/api/v1/servers/$SERVER_ID"
```

Deleting a server also deletes its associated tools through the `mcp_tools.server_id` foreign-key cascade.

---

## MCP Tool Catalog API

MCP tools allow clients and language models to interact with external systems, including APIs, databases, and internal workflows. Each tool record stores its expected arguments in an `inputSchema` JSON object. 

### Tool endpoints

```text
POST   /api/v1/servers/{serverID}/tools
GET    /api/v1/servers/{serverID}/tools
GET    /api/v1/servers/{serverID}/tools/{toolID}
PATCH  /api/v1/servers/{serverID}/tools/{toolID}
DELETE /api/v1/servers/{serverID}/tools/{toolID}
```

### Create a tool

```http
POST /api/v1/servers/{serverID}/tools
Content-Type: application/json
```

Example:

```bash
curl -s -X POST "http://localhost:8080/api/v1/servers/$SERVER_ID/tools" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "list_issues",
    "title": "List GitHub Issues",
    "description": "Returns issues from a specified GitHub repository",
    "inputSchema": {
      "type": "object",
      "properties": {
        "owner": {
          "type": "string",
          "description": "GitHub organization or user name"
        },
        "repo": {
          "type": "string",
          "description": "GitHub repository name"
        },
        "state": {
          "type": "string",
          "enum": ["open", "closed", "all"],
          "default": "open"
        },
        "per_page": {
          "type": "integer",
          "minimum": 1,
          "maximum": 100,
          "default": 30
        }
      },
      "required": ["owner", "repo"],
      "additionalProperties": false
    },
    "riskLevel": "low",
    "enabled": true
  }' | jq
```

Successful response: `201 Created`

```json
{
  "id": "2514e1c1-d6ad-4192-87c7-d79d6b30410c",
  "serverId": "d2b8ed6d-8a64-4bfc-9672-cb9f37139172",
  "name": "list_issues",
  "title": "List GitHub Issues",
  "description": "Returns issues from a specified GitHub repository",
  "inputSchema": {
    "type": "object",
    "required": ["owner", "repo"],
    "properties": {
      "owner": {
        "type": "string",
        "description": "GitHub organization or user name"
      },
      "repo": {
        "type": "string",
        "description": "GitHub repository name"
      },
      "state": {
        "type": "string",
        "enum": ["open", "closed", "all"],
        "default": "open"
      },
      "per_page": {
        "type": "integer",
        "minimum": 1,
        "maximum": 100,
        "default": 30
      }
    },
    "additionalProperties": false
  },
  "riskLevel": "low",
  "enabled": true,
  "createdAt": "2026-08-01T20:02:20.562005-04:00",
  "updatedAt": "2026-08-01T20:02:20.562005-04:00"
}
```

### List tools for a server

```http
GET /api/v1/servers/{serverID}/tools
```

```bash
curl -s "http://localhost:8080/api/v1/servers/$SERVER_ID/tools" | jq
```

Verified Phase 2 response:

```json
{
  "count": 1,
  "data": [
    {
      "id": "2514e1c1-d6ad-4192-87c7-d79d6b30410c",
      "serverId": "d2b8ed6d-8a64-4bfc-9672-cb9f37139172",
      "name": "list_issues",
      "title": "List GitHub Issues",
      "description": "Returns issues from a specified GitHub repository",
      "inputSchema": {
        "type": "object",
        "required": ["owner", "repo"],
        "properties": {
          "owner": {
            "type": "string",
            "description": "GitHub organization or user name"
          },
          "repo": {
            "type": "string",
            "description": "GitHub repository name"
          }
        },
        "additionalProperties": false
      },
      "riskLevel": "low",
      "enabled": true
    }
  ]
}
```

### Get one tool

```http
GET /api/v1/servers/{serverID}/tools/{toolID}
```

```bash
curl -s \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID" | jq
```

### Partially update a tool

```http
PATCH /api/v1/servers/{serverID}/tools/{toolID}
Content-Type: application/json
```

```bash
curl -s -X PATCH \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "riskLevel": "medium",
    "enabled": false
  }' | jq
```

Only fields supplied in the request are updated.

### Delete a tool

```http
DELETE /api/v1/servers/{serverID}/tools/{toolID}
```

```bash
curl -i -X DELETE \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID"
```

Successful response:

```text
HTTP/1.1 204 No Content
```

---

## Tool Data Model

### `mcp_tools` table

```sql
CREATE TABLE mcp_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL,
    input_schema JSONB NOT NULL,
    risk_level TEXT NOT NULL DEFAULT 'low',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT mcp_tools_server_name_unique UNIQUE (server_id, name)
);
```

| Field | Description |
|---|---|
| `id` | UUID primary key for the tool |
| `server_id` | Parent MCP server UUID |
| `name` | Machine-readable tool name, unique within its parent server |
| `title` | Optional display label |
| `description` | Human-readable tool explanation |
| `input_schema` | JSON Schema for tool arguments, stored as PostgreSQL `jsonb` |
| `risk_level` | Policy classification: `low`, `medium`, or `high` |
| `enabled` | Whether the tool is available for future discovery and invocation |
| `created_at` | UTC creation timestamp |
| `updated_at` | UTC timestamp of last modification |

### Tool risk levels

| Risk level | Examples | Future policy |
|---|---|---|
| `low` | Search repositories, list issues, read documentation | Allow for authorized developers |
| `medium` | Create Jira issue, send Slack message, open pull request | Require explicit confirmation |
| `high` | Delete repository, alter production configuration, bulk-change tickets | Deny by default or require elevated approval |

### Tool input-schema requirements

The current gateway validates that `inputSchema`:

- Is included in the request.
- Contains syntactically valid JSON.
- Has a JSON object as its root.
- Declares `"type": "object"`.
- Uses an object for `properties`, when present.
- Uses an array of strings for `required`, when present.

The gateway stores schema data as `jsonb`. PostgreSQL stores `jsonb` in a decomposed binary representation and supports indexing, making it appropriate for future schema search and JSON-based queries. 

---

## API Error Format

All known API failures return structured JSON:

```json
{
  "error": "validation_error",
  "message": "request validation failed",
  "details": [
    {
      "field": "inputSchema",
      "message": "must declare \"type\": \"object\""
    }
  ]
}
```

| HTTP status | Error code | Meaning |
|---:|---|---|
| `400` | `invalid_json` | Malformed JSON or unrecognized request field |
| `400` | `validation_error` | One or more submitted fields are invalid |
| `400` | `invalid_server_id` | `serverID` is not a valid UUID |
| `400` | `invalid_tool_id` | `toolID` is not a valid UUID |
| `404` | `server_not_found` | The requested MCP server does not exist |
| `404` | `tool_not_found` | The requested tool does not exist under the specified server |
| `409` | `duplicate_server_name` | Another registered server already uses the name |
| `409` | `duplicate_tool_name` | A tool with the same name already exists under this server |
| `500` | `internal_error` | Unexpected application or persistence failure |

---

## Phase 0: Foundation

**Status:** Complete

### Goal

Create a reliable local development baseline before adding product functionality.

### Implemented components

- Git repository initialization
- Go module initialization
- Go API entry point
- Environment configuration
- Docker Compose PostgreSQL service
- PostgreSQL connection pool
- Health endpoint
- Request-ID middleware
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

All internal Go imports use this module path.

#### PostgreSQL with Docker Compose

PostgreSQL runs independently from the Go service:

```text
Go API:      localhost:8080
PostgreSQL:  localhost:5432
```

This allows the Go service to run directly with `go run` while PostgreSQL remains reproducible through Docker Compose.

#### Connection pooling

The application creates a single `pgxpool.Pool` at startup and shares it across repositories.

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

#### Graceful shutdown

The API listens for `SIGINT` and `SIGTERM`.

During shutdown:

1. The server stops accepting new requests.
2. Existing requests have up to 10 seconds to finish.
3. The PostgreSQL pool closes.
4. The process exits cleanly.

### Phase 0 validation

```bash
docker compose up -d postgres

cd backend
go test ./...
go vet ./...
go run ./cmd/api
```

In a second terminal:

```bash
curl -s http://localhost:8080/health | jq
curl -s http://localhost:8080/api/v1/ | jq
```

---

## Phase 1: MCP Server Registry

**Status:** Complete

### Goal

Create the persistent source of truth for MCP servers known to the gateway.

An MCP server record describes an external or internal system that can expose tools through a supported transport.

Examples:

- GitHub tool server
- Jira tool server
- Slack tool server
- Confluence tool server
- Internal deployment automation server
- Internal documentation search server
- Security remediation tool server

### Server endpoints

```text
POST   /api/v1/servers
GET    /api/v1/servers
GET    /api/v1/servers/{serverID}
PATCH  /api/v1/servers/{serverID}
DELETE /api/v1/servers/{serverID}
```

### `mcp_servers` data model

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `name` | Unique human-readable server identifier |
| `description` | Explanation of the server’s purpose |
| `baseUrl` | Future MCP server connection location |
| `transportType` | `streamable_http`, `sse`, or `stdio` |
| `status` | `active`, `inactive`, or `unhealthy` |
| `ownerTeam` | Team responsible for server operations |
| `createdAt` | UTC creation timestamp |
| `updatedAt` | UTC last-update timestamp |

### Validation rules

| Field | Rules |
|---|---|
| `name` | Required, trimmed, at most 100 characters, globally unique |
| `description` | Required, trimmed, at most 1000 characters |
| `baseUrl` | Required for HTTP/SSE transports; must be an absolute HTTP/HTTPS URL |
| `transportType` | `streamable_http`, `sse`, or `stdio` |
| `status` | `active`, `inactive`, or `unhealthy` |
| `ownerTeam` | Required, trimmed, at most 100 characters |

### Phase 1 acceptance criteria

- [x] PostgreSQL migration creates `mcp_servers`
- [x] Create endpoint persists a server and returns `201 Created`
- [x] List endpoint returns records and a count
- [x] Get-by-ID endpoint returns one server
- [x] Patch endpoint updates submitted fields only
- [x] Delete endpoint returns `204 No Content`
- [x] Duplicate names return `409 Conflict`
- [x] Invalid payloads return field-level `400 Bad Request` errors
- [x] Invalid UUIDs return `400 Bad Request`
- [x] Unknown servers return `404 Not Found`
- [x] Tests pass with `go test ./...`

---

## Phase 2: MCP Tool Catalog

**Status:** Complete

### Goal

Add discoverable tool metadata for each registered MCP server.

A tool is a capability exposed by an MCP server. It has a unique name within that server, a description, a JSON Schema describing accepted arguments, a risk classification, and an enablement state.

The initial verified tool is:

```text
Server: github-mcp
Tool: list_issues
Risk: low
Enabled: true
```

### Implemented components

- `mcp_tools` PostgreSQL table
- Foreign key from `mcp_tools.server_id` to `mcp_servers.id`
- Cascading delete of tools when a server is deleted
- JSONB storage for `input_schema`
- GIN index on input schemas
- Unique `(server_id, name)` constraint
- Tool CRUD endpoints
- Create and partial-update request models
- Input-schema validation
- Risk-level validation
- Enable/disable support
- Parent server existence checks
- Duplicate tool-name protection per server
- Typed PostgreSQL unique-constraint detection
- Unit tests for tool defaults and validation

### Tool endpoints

```text
POST   /api/v1/servers/{serverID}/tools
GET    /api/v1/servers/{serverID}/tools
GET    /api/v1/servers/{serverID}/tools/{toolID}
PATCH  /api/v1/servers/{serverID}/tools/{toolID}
DELETE /api/v1/servers/{serverID}/tools/{toolID}
```

### Tool validation rules

| Field | Rules |
|---|---|
| `name` | Required, trimmed, maximum 100 characters, unique per server |
| `title` | Optional, trimmed, maximum 150 characters |
| `description` | Required, trimmed, maximum 2000 characters |
| `inputSchema` | Required valid JSON object with `"type": "object"` |
| `riskLevel` | `low`, `medium`, or `high`; defaults to `low` |
| `enabled` | Boolean; defaults to `true` |

### Phase 2 verified output

The tool catalog was successfully queried after adding the GitHub issue-listing tool:

```json
{
  "count": 1,
  "data": [
    {
      "id": "2514e1c1-d6ad-4192-87c7-d79d6b30410c",
      "serverId": "d2b8ed6d-8a64-4bfc-9672-cb9f37139172",
      "name": "list_issues",
      "title": "List GitHub Issues",
      "description": "Returns issues from a specified GitHub repository",
      "inputSchema": {
        "type": "object",
        "required": ["owner", "repo"],
        "properties": {
          "owner": {
            "type": "string",
            "description": "GitHub organization or user name"
          },
          "repo": {
            "type": "string",
            "description": "GitHub repository name"
          },
          "state": {
            "type": "string",
            "enum": ["open", "closed", "all"],
            "default": "open"
          },
          "per_page": {
            "type": "integer",
            "minimum": 1,
            "maximum": 100,
            "default": 30
          }
        },
        "additionalProperties": false
      },
      "riskLevel": "low",
      "enabled": true,
      "createdAt": "2026-08-01T20:02:20.562005-04:00",
      "updatedAt": "2026-08-01T20:02:20.562005-04:00"
    }
  ]
}
```

### Phase 2 acceptance criteria

- [x] PostgreSQL migration creates `mcp_tools`
- [x] A tool can be created for an existing server
- [x] Tool input schemas require a JSON object root
- [x] Tools list correctly under their parent server
- [x] A specific tool can be fetched by ID
- [x] Partial updates support risk and enabled-state changes
- [x] Duplicate names under one server return `409 Conflict`
- [x] The same tool name can exist under different servers
- [x] Invalid schema root types return `400 Bad Request`
- [x] Invalid risk levels return `400 Bad Request`
- [x] Unknown parent servers return `404 Not Found`
- [x] Unknown tools return `404 Not Found`
- [x] Tool deletion returns `204 No Content`
- [x] Parent server deletion cascades to delete associated tools
- [x] `go test ./...` passes

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

### Format Go code

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

curl -s http://localhost:8080/api/v1/servers | jq

curl -s \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools" | jq
```

---

## Design Decisions

### Why Go

The gateway is an I/O-heavy control-plane service. It will authenticate callers, query PostgreSQL, call external MCP servers and APIs, enforce policy, persist audits, and expose metrics.

Go is a strong fit because it provides:

- Static typing
- Fast startup
- Straightforward container deployment
- Lightweight concurrency through goroutines
- Strong standard-library HTTP support
- Explicit error handling
- Good fit for API gateways, platform services, and developer tooling

Python remains a strong future choice for LangGraph workflows, RAG, evaluation pipelines, and agent reasoning.

### Why PostgreSQL

PostgreSQL supports the project’s future relational and structured-data needs:

- Server, tool, user, permission, and audit relationships
- JSON/JSONB tool argument schemas
- Indexes for tool discovery and usage history
- Transactions for policy-controlled invocations
- Strong data constraints
- Mature Docker and cloud support

### Why `jsonb` for tool schemas

MCP tool definitions include input schemas that describe the arguments clients may send. PostgreSQL `jsonb` preserves this structured metadata in a queryable, indexable form.

The current migration adds a GIN index:

```sql
CREATE INDEX idx_mcp_tools_input_schema_gin
    ON mcp_tools
    USING GIN (input_schema);
```

This supports future capabilities such as:

- Filtering tools based on schema properties
- Finding tools that require particular inputs
- Validating tool-catalog conventions
- Building schema-aware search in the React UI

### Why layered backend design

The backend separates concerns:

```text
Handlers: HTTP request/response concerns
Services: validation and business rules
Repositories: database persistence
Domain: shared application structures
```

This makes later additions such as authentication, policy checks, audit events, and MCP clients easier to test and evolve.

### Why read-only tools first

The first live integration will prioritize low-risk actions:

- Search repositories
- List GitHub issues
- Get pull request details
- Search Jira tickets
- Read Confluence pages
- Search Slack messages

Mutating operations such as creating issues, posting Slack messages, editing documentation, or modifying repositories will require explicit confirmation and stricter policies.

---

## Known Limitations

The following are intentional at the end of Phase 2:

- No authentication: every local caller can access all APIs.
- No authorization: no admin, developer, or viewer role boundaries exist yet.
- No pagination, filtering, or search on server/tool list endpoints.
- No automatic discovery from remote MCP servers yet.
- No server health checks against registered URLs.
- No validation that a registered URL hosts a live MCP server.
- No real tool invocation path yet.
- No external API credentials or credential-reference table.
- No audit log records.
- No metrics endpoint.
- No OpenTelemetry traces.
- No React user interface.
- No CI/CD pipeline.
- No deployed environment.

---

## Planned Phases

### Phase 3: Authentication and RBAC

Phase 3 will protect the gateway with JWT-based authentication and role-based authorization.

Planned roles:

| Role | Permissions |
|---|---|
| `admin` | Manage servers, tools, future credentials, and access policies |
| `developer` | Browse approved servers/tools and invoke permitted tools in later phases |
| `viewer` | Browse server and tool catalog only |

Planned implementation:

- `users` table
- Role assignments
- Local development login endpoint
- Password hashing
- JWT issuance
- JWT authentication middleware
- Protected route groups
- Admin-only server and tool mutation endpoints
- User identity exposed in request context
- Secure JWT secret configuration

The long-term MCP authorization model will evolve toward OAuth-compatible resource-server patterns. MCP’s authorization specification treats protected MCP servers as OAuth resources that accept access tokens; this phase begins with local JWTs to establish identity and RBAC before external identity-provider integration. 

### Phase 4: Invocation Gateway

Planned request lifecycle:

```text
Incoming tool invocation
    ↓
Authenticate caller
    ↓
Authorize server and tool access
    ↓
Verify tool is enabled
    ↓
Validate arguments against JSON Schema
    ↓
Evaluate risk and policy
    ↓
Create audit record
    ↓
Invoke MCP server or external adapter
    ↓
Redact sensitive data
    ↓
Store result, latency, and errors
    ↓
Return structured response
```

### Phase 5: First Live Integration

Initial integration target: GitHub read-only tools.

Candidate tools:

- `search_repositories`
- `list_issues`
- `get_issue`
- `list_pull_requests`
- `get_pull_request`

A later option is registering and routing to an existing Jira MCP server through this gateway.

### Phase 6: Observability

Planned features:

- Persistent audit records
- Request IDs
- Structured logs
- Prometheus metrics
- OpenTelemetry traces
- Server health status
- Tool success/error rates
- p50, p95, and p99 invocation latency
- Most-used tool dashboard

### Phase 7: React UI

Planned screens:

- Server discovery catalog
- Server detail page
- Tool catalog
- Tool input-schema viewer
- JSON invocation sandbox
- Confirmation dialog for medium/high-risk tools
- Invocation history
- Admin server/tool management
- Lightweight observability dashboard

### Phase 8: Delivery and Polish

Planned improvements:

- GitHub Actions CI pipeline
- Backend Dockerfile
- Frontend Dockerfile
- Integration tests using Testcontainers
- OpenAPI documentation
- Deployment to cloud infrastructure
- Screenshots and demo video
- Architecture decision records
- Security and operations documentation

---

## Troubleshooting

### PostgreSQL container does not start

```bash
docker compose logs postgres
```

Check whether port `5432` is occupied:

```bash
lsof -i :5432
```

### Application cannot connect to PostgreSQL

```bash
docker compose ps
```

Load and inspect the local connection string:

```bash
set -a
source .env
set +a

echo "$DATABASE_URL"
```

Check the database directly:

```bash
docker exec -it mcp-gateway-postgres \
  psql -U mcp_gateway -d mcp_gateway \
  -c "SELECT 1;"
```

### `migrate: command not found`

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

export PATH="$PATH:$(go env GOPATH)/bin"

migrate -version
```

### Dirty migration state

Inspect migration state:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

For local development only, reset everything:

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

> Do not perform destructive resets against shared or production databases.

### `.env` does not load

Run the backend from `backend/`:

```bash
cd backend
go run ./cmd/api
```

The application attempts to load both `../.env` and `.env`.

### Port 8080 is already in use

```bash
lsof -i :8080
```

Stop the process or change:

```dotenv
HTTP_PORT=8081
```

Then restart the backend.

### Tests fail after import changes

Ensure imports use:

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

```bash
git checkout -b feat/authentication-rbac

cd backend
gofmt -w .
go mod tidy
go test ./...
go vet ./...

cd ..
git add .
git commit -m "feat: add JWT authentication and RBAC"
git push -u origin feat/authentication-rbac
```

Suggested commit convention:

```text
feat: add MCP server registry CRUD API
feat: add MCP tool catalog
feat: add JWT authentication middleware
feat: add role-based access control
feat: add tool invocation audit records
fix: validate tool input schema
test: cover invalid tool risk levels
docs: document phase 2 tool catalog
chore: configure Docker Compose PostgreSQL
```

---

## Project Status

```text
Phase 0: Complete
Phase 1: Complete
Phase 2: Complete
Phase 3: Ready to begin
```

The next milestone is securing the gateway with **JWT authentication and role-based access control**, so that only administrators can mutate the server/tool catalog and future tool invocations can be tied to an authenticated identity.