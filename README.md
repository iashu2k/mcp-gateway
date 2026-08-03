# MCP Gateway

A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations.

The project is inspired by the idea of an internal “USB-C for AI agents”: a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history with full observability.

> **Current status:** Phase 6 complete — the gateway now has full observability with Prometheus metrics, invocation history API, and OpenTelemetry tracing.

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
- [Authentication and Roles](#authentication-and-roles)
- [API Reference](#api-reference)
- [Phase 0: Foundation](#phase-0-foundation)
- [Phase 1: MCP Server Registry](#phase-1-mcp-server-registry)
- [Phase 2: MCP Tool Catalog](#phase-2-mcp-tool-catalog)
- [Phase 3: Authentication and RBAC](#phase-3-authentication-and-rbac)
- [Phase 4: Invocation Gateway](#phase-4-invocation-gateway)
- [Phase 5: Live GitHub Integration](#phase-5-live-github-integration)
- [Phase 6: Observability](#phase-6-observability)
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

MCP Gateway addresses this by acting as a secure control plane for internal MCP servers.

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
│ Server Registry -  Tool Catalog -  JWT Auth -  RBAC               │
│ Invocation Gateway -  JSON Schema Validation -  Audit Records    │
│ Live GitHub Executor -  Mock Executor -  Upstream Error Capture  │
│ Prometheus Metrics -  OpenTelemetry Tracing -  History API       │
└─────────────┬──────────────────┬──────────────────┬─────────────┘
              │                  │                  │
              ▼                  ▼                  ▼
      ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
      │ GitHub REST   │  │ Jira MCP      │  │ Slack MCP     │
      │ API (live)    │  │ Server/API    │  │ Server/API    │
      └───────────────┘  └───────────────┘  └───────────────┘
              │
              ▼
      ┌─────────────────────────────────────────────────────────┐
      │ PostgreSQL                                               │
      │ Users -  Servers -  Tools -  Invocation Audit Records       │
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
- JSON Schema-driven tool definitions and argument validation
- JWT authentication
- Role-based access control
- Policy-controlled tool invocation
- Live third-party API integration (GitHub REST)
- Executor routing and abstraction
- Durable invocation audit records with upstream error capture
- **Observability: Prometheus metrics, OpenTelemetry tracing, invocation history API**
- Secure password handling
- Structured request logging with status codes
- Containerized local development
- Production-style testing and deployment practices

The long-term architecture separates deterministic infrastructure from agentic reasoning.

```text
Go Gateway
├── Authentication
├── Authorization
├── Server registry
├── Tool catalog
├── Argument validation
├── Policy checks
├── Tool routing and execution
│   ├── Mock executor (deterministic testing)
│   └── Live API executors (GitHub, future Jira/Slack)
├── Audit logging
├── Observability
│   ├── Prometheus metrics
│   ├── OpenTelemetry traces
│   └── Invocation history API
└── Health checks

Optional Python Agent Service
├── LangGraph workflows
├── LLM orchestration
├── Retrieval-augmented generation
├── Evaluation pipelines
└── Agent planning and reasoning
```

---

## System Architecture

### Current architecture

```text
Developer / API Client
         │
         │ HTTP / JSON
         ▼
┌──────────────────────────────────────┐
│ Go MCP Gateway                       │
│                                      │
│ chi Router                           │
│ Request ID middleware                │
│ Status-aware request logging         │
│ Timeout middleware                   │
│ JWT authentication middleware        │
│ RBAC middleware                      │
│ Health endpoint                      │
│ Metrics endpoint (/metrics)          │
│ Authentication API                   │
│ Server Registry API                  │
│ Tool Catalog API                     │
│ Invocation API                       │
│ Invocation History API               │
│ JSON Schema argument validation      │
│                                      │
│ Observability                        │
│ ├── Prometheus metrics               │
│ ├── OpenTelemetry tracing            │
│ └── Structured logging               │
│                                      │
│ Executor Router                      │
│ ├── MockExecutor (in-process)        │
│ └── GitHubExecutor (live REST API)   │
└────────────────┬─────────────────────┘
                 │
                 │ pgx connection pool
                 ▼
┌──────────────────────────────────────┐
│ PostgreSQL 16                        │
│                                      │
│ users                                │
│ mcp_servers                          │
│ mcp_tools                            │
│ tool_invocations                     │
│ schema_migrations                    │
└──────────────────────────────────────┘
```

### Invocation request flow with observability

```text
POST /api/v1/servers/{serverID}/tools/{toolID}/invoke
        │
        ▼
JWT authentication middleware
        │
        ▼
Role check: admin or developer
        │
        ▼
Load server — must exist and be active
        │
        ▼
Load tool — must exist and be enabled
        │
        ▼
Policy check — only low-risk tools
        │
        ▼
Validate arguments against stored JSON Schema
        │
        ▼
Create tool_invocations audit record (status: running)
        │
        ▼
RouterExecutor dispatches by server name:
    ├── "github"  → GitHubExecutor → live GitHub REST API
    └── others    → MockExecutor   → deterministic mock
        │
        ▼
Update audit record: succeeded or failed
    ├── success → response_payload, duration_ms, completed_at
    └── failure → error_code, error_message (upstream errors captured)
        │
        ▼
Record Prometheus metrics:
    ├── invocations_total (server, tool, status)
    ├── invocation_duration_seconds (server, tool)
    └── upstream_requests_total (service, status)
        │
        ▼
Return structured invocation response
```

### Target architecture

```text
React Discovery UI
        │
        ▼
Go MCP Gateway API
├── Server Registry
├── Tool Catalog
├── JWT / OIDC Authentication
├── Role-Based Access Control
├── Server and Tool Permissions
├── Policy Engine
├── Tool Invocation Proxy
│   ├── GitHub REST Executor (live)
│   ├── Jira Executor (planned)
│   ├── Slack Executor (planned)
│   └── MCP Client Executor (planned)
├── Audit Log Service
├── Observability
│   ├── Prometheus metrics (implemented)
│   ├── OpenTelemetry traces (initialized)
│   └── Grafana dashboards (planned)
└── MCP Client / Adapter Layer
        │
        ├── GitHub MCP Server or REST API
        ├── Jira MCP Server
        ├── Slack MCP Server
        ├── Confluence MCP Server
        └── Internal engineering tools
```

---

## Technology Stack

| Area | Technology | Purpose |
|---|---|---|
| Backend language | Go | Concurrent, strongly typed gateway implementation |
| HTTP routing | `go-chi/chi` | Lightweight REST routing and middleware |
| Database | PostgreSQL 16 | Users, servers, tools, invocation audits |
| PostgreSQL driver | `pgx/v5` | Native PostgreSQL driver and connection pool |
| Migrations | `golang-migrate` | Version-controlled schema migrations |
| JSON storage | PostgreSQL `jsonb` | Tool input schemas and invocation payloads |
| Schema validation | `santhosh-tekuri/jsonschema/v6` | JSON Schema draft 2020-12 argument validation |
| Authentication | JWT with HS256 | Local development access tokens |
| JWT library | `github.com/golang-jwt/jwt/v5` | JWT signing, parsing, verification, claims validation |
| Password hashing | `golang.org/x/crypto/bcrypt` | Password hash generation and comparison |
| GitHub integration | `google/go-github/v62` | Typed Go client for GitHub REST API v3 |
| **Metrics** | `prometheus/client_golang` | Prometheus metrics collection and exposition |
| **Tracing** | `go.opentelemetry.io/otel` | OpenTelemetry distributed tracing |
| Configuration | Environment variables + `godotenv` | Local configuration and secrets loading |
| Logging | Go `log/slog` + chi `WrapResponseWriter` | Structured logs with request ID, status, bytes, latency |
| API testing | `curl`, `jq`, Go `testing` | Manual API verification and automated tests |
| Containerization | Docker Compose | Local PostgreSQL environment |
| Frontend | React + TypeScript | Future discovery catalog and sandbox |
| Observability UI | Grafana | Future metrics and trace visualization |
| MCP integration | Official Go MCP SDK | Future real MCP discovery and invocation |

---

## Current Features

### Completed

- Go module at `github.com/iashu2k/mcp-gateway/backend`
- Docker Compose-managed PostgreSQL 16 database
- Environment-based configuration
- PostgreSQL connection pool using `pgxpool`
- Health endpoint with database connectivity validation
- Structured JSON application logs
- Request logging with request ID, method, path, HTTP status, bytes written, and duration
- HTTP request ID and timeout middleware
- Graceful HTTP shutdown
- Version-controlled PostgreSQL schema migrations
- MCP server registry with CRUD operations
- MCP tool catalog with CRUD operations
- Server-to-tool foreign key with cascading deletion
- JSON Schema storage (`jsonb`) and structural validation for tool inputs
- Tool risk classifications: `low`, `medium`, and `high`
- Tool enabled/disabled state
- Field-level validation responses
- Duplicate server-name and per-server tool-name prevention
- Typed PostgreSQL unique-constraint handling
- Local user table with bcrypt password hashes
- Signed JWT access-token issuance with issuer/expiration validation
- Bearer-token authentication middleware
- `admin`, `developer`, and `viewer` roles
- Admin-only catalog mutations
- **Phase 4: policy-controlled tool invocation**
  - Authenticated invocation endpoint for `admin` and `developer`
  - Server-active and tool-enabled policy checks
  - Low-risk-only invocation policy
  - Full JSON Schema validation of invocation arguments
  - Durable `tool_invocations` audit records with `running → succeeded/failed` lifecycle
  - Deterministic in-process mock executor (`echo`, `list_issues`)
  - Audit foreign keys using `ON DELETE RESTRICT` to preserve accountability history
- **Phase 5: live GitHub REST integration**
  - Real GitHub REST API executor for public repositories
  - Executor routing: `github` servers hit live API, others use mock executor
  - `list_issues` and `search_repositories` tools with live GitHub data
  - Slimmed response payloads (only essential fields stored in audit)
  - Upstream error capture in audit records (`error_code: execution_failed`)
  - Unauthenticated access (60 req/hr) or token-based (5000 req/hr)
- **Phase 6: observability**
  - Prometheus metrics endpoint at `/metrics`
  - HTTP request metrics (count, duration) by method/path/status
  - Invocation metrics (count, duration) by server/tool/status
  - Upstream API metrics (GitHub requests by success/error)
  - Database connection pool metrics
  - Invocation history list endpoint with filtering (server, tool, user, status, date)
  - Invocation history detail endpoint
  - Pagination support (limit/offset)
  - Role-based history access (non-admin users see only their own invocations)
  - OpenTelemetry tracing initialized (stdout exporter for local dev)
- Unit tests for server, tool, JWT, auth middleware, and invocation service behavior

### Not yet implemented

- OAuth/OIDC identity-provider integration
- Server-level and tool-level per-user permissions
- Medium/high-risk invocation confirmation flows
- Additional GitHub tools (`get_issue`, `list_pull_requests`, etc.)
- Jira/Slack/Confluence live integrations
- MCP protocol transport (stdio/SSE/streamable HTTP)
- Grafana dashboards for metrics visualization
- Jaeger/Tempo for trace visualization
- Credential reference storage (GitHub token currently env-only)
- React UI, CI/CD pipeline, production deployment

---

## Development Roadmap

| Phase | Status | Focus | Main Outcome |
|---|---|---|---|
| Phase 0 | Complete | Foundation | Go API, PostgreSQL, Docker Compose, health endpoint |
| Phase 1 | Complete | MCP Server Registry | Persistent CRUD API for MCP server metadata |
| Phase 2 | Complete | MCP Tool Catalog | Per-server tools, input schemas, risk levels, enablement |
| Phase 3 | Complete | Authentication and RBAC | JWT authentication, local users, bcrypt, role protection |
| Phase 4 | Complete | Invocation Gateway | Policy-checked, schema-validated, audited mock invocations |
| Phase 5 | Complete | Live GitHub Integration | Real GitHub REST executor with audit trail |
| Phase 6 | Complete | Observability | Prometheus metrics, OpenTelemetry tracing, invocation history API |
| Phase 7 | Next | React UI | Catalog, sandbox, history, and administration interface |
| Phase 8 | Planned | Delivery and Polish | CI, containers, deployment, demo, documentation |

---

## Repository Structure

```text
mcp-gateway/
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go
│   │   └── passwordhash/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── jwt_test.go
│   │   │
│   │   ├── config/
│   │   │   └── config.go
│   │   │
│   │   ├── domain/
│   │   │   ├── invocation.go
│   │   │   ├── server.go
│   │   │   ├── tool.go
│   │   │   └── user.go
│   │   │
│   │   ├── executor/
│   │   │   ├── github_executor.go
│   │   │   ├── mock_executor.go
│   │   │   └── router_executor.go
│   │   │
│   │   ├── httpapi/
│   │   │   ├── auth_handler.go
│   │   │   ├── auth_middleware.go
│   │   │   ├── auth_middleware_test.go
│   │   │   ├── errors.go
│   │   │   ├── handlers.go
│   │   │   ├── invocation_handler.go
│   │   │   ├── invocation_history_handler.go
│   │   │   ├── router.go
│   │   │   ├── router_test.go
│   │   │   ├── server_handler.go
│   │   │   └── tool_handler.go
│   │   │
│   │   ├── observability/
│   │   │   ├── metrics.go
│   │   │   └── tracing.go
│   │   │
│   │   ├── platform/
│   │   │   └── database/
│   │   │       └── postgres.go
│   │   │
│   │   ├── repository/
│   │   │   ├── invocation_repository.go
│   │   │   ├── server_repository.go
│   │   │   ├── tool_repository.go
│   │   │   └── user_repository.go
│   │   │
│   │   └── service/
│   │       ├── auth_service.go
│   │       ├── invocation_history_service.go
│   │       ├── invocation_service.go
│   │       ├── invocation_service_test.go
│   │       ├── schema_validator.go
│   │       ├── server_service.go
│   │       ├── server_service_test.go
│   │       ├── tool_service.go
│   │       └── tool_service_test.go
│   │
│   ├── migrations/
│   │   ├── 000001_create_mcp_servers.down.sql
│   │   ├── 000001_create_mcp_servers.up.sql
│   │   ├── 000002_create_mcp_tools.down.sql
│   │   ├── 000002_create_mcp_tools.up.sql
│   │   ├── 000003_create_users.down.sql
│   │   ├── 000003_create_users.up.sql
│   │   ├── 000004_create_tool_invocations.down.sql
│   │   └── 000004_create_tool_invocations.up.sql
│   │
│   ├── go.mod
│   └── go.sum
│
├── docs/
├── frontend/
├── infra/
├── scripts/
│   └── seed_users.sql.example
├── .env.example
├── .gitignore
├── docker-compose.yml
└── README.md
```

### Package responsibilities

| Package | Responsibility |
|---|---|
| `cmd/api` | Application startup, dependency wiring, configuration loading, graceful shutdown |
| `cmd/passwordhash` | Local utility to generate bcrypt hashes for development user seeding |
| `internal/auth` | JWT claims, signing, parsing, signature verification, expiration validation |
| `internal/config` | Environment-variable loading and validation |
| `internal/domain` | Core domain models, request models, constants, response structures |
| `internal/executor` | Tool execution implementations: mock (deterministic) and GitHub (live REST) |
| `internal/httpapi` | Routes, handlers, middleware, JSON decoding, responses, HTTP error mapping |
| `internal/observability` | Prometheus metrics and OpenTelemetry tracing |
| `internal/platform/database` | PostgreSQL connection-pool setup and health verification |
| `internal/repository` | Parameterized SQL and PostgreSQL persistence operations |
| `internal/service` | Business rules, validation, authentication, schema validation, invocation orchestration, history queries |
| `migrations` | Ordered and versioned database schema evolution |

Dependency direction:

```text
HTTP Handler / Middleware
           ↓
        Service
           ↓
  Repository / Executor Router
           ↓
  PostgreSQL / Mock / Live APIs
           ↓
  Observability (metrics, traces, logs)
```

---

## Prerequisites

Install:

- Go 1.23 or newer
- Docker Desktop or Docker Engine with Docker Compose
- Git
- `curl`
- `jq`
- PostgreSQL client tools, including `psql`
- `golang-migrate`

Verify:

```bash
go version
docker --version
docker compose version
git --version
curl --version
jq --version
psql --version
```

Install `golang-migrate`:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
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

### Configure JWT settings

```bash
openssl rand -base64 48
```

Set the generated value in `.env`:

```dotenv
JWT_SECRET=PASTE_A_LONG_RANDOM_VALUE_HERE
JWT_ISSUER=mcp-gateway
JWT_TTL_MINUTES=60
```

### Configure GitHub integration (optional)

For public repositories, you can run unauthenticated (60 requests/hour). For higher rate limits or private repos, create a fine-grained personal access token with only "Issues: Read" permission:

```dotenv
GITHUB_TOKEN=github_pat_your_token_here
```

Leave empty for unauthenticated public access:

```dotenv
GITHUB_TOKEN=
```

### Start PostgreSQL

```bash
docker compose up -d postgres
docker compose ps
```

### Apply migrations

```bash
set -a
source .env
set +a

migrate -path backend/migrations -database "$DATABASE_URL" up
migrate -path backend/migrations -database "$DATABASE_URL" version
```

Expected after Phase 6:

```text
4
```

### Seed local users

```bash
cd backend
go run ./cmd/passwordhash "AdminPass123!"
go run ./cmd/passwordhash "DeveloperPass123!"
go run ./cmd/passwordhash "ViewerPass123!"
cd ..

cp scripts/seed_users.sql.example scripts/seed_users.sql
# Replace the placeholder hashes in scripts/seed_users.sql

psql "$DATABASE_URL" -f scripts/seed_users.sql
```

Local development users:

| Email | Password | Role |
|---|---|---|
| `admin@mcp-gateway.local` | `AdminPass123!` | `admin` |
| `developer@mcp-gateway.local` | `DeveloperPass123!` | `developer` |
| `viewer@mcp-gateway.local` | `ViewerPass123!` | `viewer` |

> Local development credentials only. Never use them in any deployed environment.

### Start the API

From the repository root:

```bash
go run ./backend/cmd/api
```

Or from `backend/`:

```bash
cd backend
go run ./cmd/api
```

The service runs at `http://localhost:8080`.

---

## Environment Variables

```dotenv
APP_ENV=development
HTTP_PORT=8080

POSTGRES_DB=mcp_gateway
POSTGRES_USER=mcp_gateway
POSTGRES_PASSWORD=mcp_gateway_dev_password
DATABASE_URL=postgres://mcp_gateway:mcp_gateway_dev_password@localhost:5432/mcp_gateway?sslmode=disable

JWT_SECRET=replace-with-a-long-random-secret
JWT_ISSUER=mcp-gateway
JWT_TTL_MINUTES=60

GITHUB_TOKEN=
```

| Variable | Required | Description |
|---|---:|---|
| `APP_ENV` | Yes | Runtime environment name |
| `HTTP_PORT` | Yes | Port used by the Go HTTP API |
| `POSTGRES_DB` | Yes | PostgreSQL database name |
| `POSTGRES_USER` | Yes | PostgreSQL user |
| `POSTGRES_PASSWORD` | Yes | Local PostgreSQL password |
| `DATABASE_URL` | Yes | Connection URL for the Go app and migration CLI |
| `JWT_SECRET` | Yes | At least 32-character HS256 signing secret |
| `JWT_ISSUER` | Yes | Expected issuer claim for generated and accepted tokens |
| `JWT_TTL_MINUTES` | Yes | Access-token lifetime in minutes |
| `GITHUB_TOKEN` | No | GitHub personal access token (empty = unauthenticated public access) |

`.env` and `scripts/seed_users.sql` are local-only files and must not be committed.

---

## Database Migrations

Current migrations:

```text
000001_create_mcp_servers.up/down.sql
000002_create_mcp_tools.up/down.sql
000003_create_users.up/down.sql
000004_create_tool_invocations.up/down.sql
```

### Apply and check

```bash
set -a
source .env
set +a

migrate -path backend/migrations -database "$DATABASE_URL" up
migrate -path backend/migrations -database "$DATABASE_URL" version
```

### Inspect tables

```bash
psql "$DATABASE_URL" -c "\dt"
psql "$DATABASE_URL" -c "SELECT id, email, role, active FROM users;"
psql "$DATABASE_URL" -c "SELECT id, name, status FROM mcp_servers;"
psql "$DATABASE_URL" -c "SELECT id, name, risk_level, enabled FROM mcp_tools;"
psql "$DATABASE_URL" -c "SELECT id, status, duration_ms FROM tool_invocations ORDER BY created_at DESC LIMIT 5;"
```

---

## Authentication and Roles

### Role model

| Role | Catalog reads | Catalog mutations | Tool invocation | Invocation history |
|---|---:|---:|---:|---:|
| `admin` | Yes | Yes | Yes (low-risk) | All users |
| `developer` | Yes | No | Yes (low-risk) | Own invocations only |
| `viewer` | Yes | No | No | No |
| Unauthenticated | No | No | No | No |

### Login

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mcp-gateway.local",
    "password": "AdminPass123!"
  }' | jq
```

### Save tokens

```bash
export ADMIN_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@mcp-gateway.local","password":"AdminPass123!"}' \
    | jq -r '.accessToken'
)"

export DEVELOPER_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"developer@mcp-gateway.local","password":"DeveloperPass123!"}' \
    | jq -r '.accessToken'
)"
```

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Public endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/health` | Gateway and PostgreSQL health check |
| `GET` | `/metrics` | Prometheus metrics exposition |
| `GET` | `/api/v1/` | API root message |
| `POST` | `/api/v1/auth/login` | Authenticate and receive JWT |

### Authenticated endpoints (all roles)

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/api/v1/auth/me` | Current authenticated identity |
| `GET` | `/api/v1/servers` | List MCP servers |
| `GET` | `/api/v1/servers/{serverID}` | Get MCP server |
| `GET` | `/api/v1/servers/{serverID}/tools` | List tools for server |
| `GET` | `/api/v1/servers/{serverID}/tools/{toolID}` | Get tool |

### Admin + developer endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/api/v1/servers/{serverID}/tools/{toolID}/invoke` | Invoke an enabled low-risk tool |
| `GET` | `/api/v1/invocations` | List invocation history (filtered by role) |
| `GET` | `/api/v1/invocations/{invocationID}` | Get invocation details (filtered by role) |

### Admin-only endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/api/v1/servers` | Register MCP server |
| `PATCH` | `/api/v1/servers/{serverID}` | Update MCP server |
| `DELETE` | `/api/v1/servers/{serverID}` | Delete MCP server |
| `POST` | `/api/v1/servers/{serverID}/tools` | Register tool |
| `PATCH` | `/api/v1/servers/{serverID}/tools/{toolID}` | Update tool |
| `DELETE` | `/api/v1/servers/{serverID}/tools/{toolID}` | Delete tool |

---

## Phase 0: Foundation

**Status:** Complete

- Git repository and Go module initialization
- Environment configuration
- Docker Compose PostgreSQL service
- PostgreSQL connection pool
- Health endpoint
- Request-ID and timeout middleware
- Structured JSON logging
- Graceful shutdown
- Baseline tests

---

## Phase 1: MCP Server Registry

**Status:** Complete

### Endpoints

```text
POST   /api/v1/servers
GET    /api/v1/servers
GET    /api/v1/servers/{serverID}
PATCH  /api/v1/servers/{serverID}
DELETE /api/v1/servers/{serverID}
```

### Server model

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `name` | Unique human-readable server identifier |
| `description` | Server purpose and capabilities |
| `baseUrl` | Future MCP server connection URL |
| `transportType` | `streamable_http`, `sse`, or `stdio` |
| `status` | `active`, `inactive`, or `unhealthy`; defaults to `active` on create |
| `ownerTeam` | Team responsible for server operations |
| `createdAt` / `updatedAt` | UTC timestamps |

---

## Phase 2: MCP Tool Catalog

**Status:** Complete

### Endpoints

```text
POST   /api/v1/servers/{serverID}/tools
GET    /api/v1/servers/{serverID}/tools
GET    /api/v1/servers/{serverID}/tools/{toolID}
PATCH  /api/v1/servers/{serverID}/tools/{toolID}
DELETE /api/v1/servers/{serverID}/tools/{toolID}
```

### Tool model

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `serverId` | Parent server UUID (`ON DELETE CASCADE`) |
| `name` | Tool identifier, unique within parent server |
| `title` | Optional display name |
| `description` | Human-readable tool purpose |
| `inputSchema` | JSON Schema object (`jsonb`) for accepted arguments |
| `riskLevel` | `low`, `medium`, or `high`; defaults to `low` |
| `enabled` | Whether the tool can be invoked; defaults to `true` |

### Tool risk levels

| Risk level | Examples | Current invocation policy |
|---|---|---|
| `low` | Search repos, list issues, read docs | Invocable by admin/developer |
| `medium` | Create issue, send message, open PR | Rejected (`403`) pending confirmation flow |
| `high` | Delete repo, bulk ticket changes, production settings | Rejected (`403`) pending elevated approval |

---

## Phase 3: Authentication and RBAC

**Status:** Complete

- `users` table with bcrypt password hashes
- Local seed users with `admin`, `developer`, and `viewer` roles
- Login endpoint issuing HS256 JWTs
- Explicit signing-method, issuer, expiration, subject, and role validation
- Bearer-token middleware with request-context identity
- Role middleware for admin-only mutations
- `/auth/me` current-identity endpoint
- Auth and middleware tests

### Route protection summary

```text
Public:               /health, /metrics, /api/v1/, POST /auth/login
Authenticated:        /auth/me, server/tool catalog reads
Admin + developer:    POST .../tools/{toolID}/invoke, GET /invocations
Admin only:           server/tool catalog mutations
```

---

## Phase 4: Invocation Gateway

**Status:** Complete

### Goal

Provide a policy-controlled, audited path for invoking a registered tool — without yet connecting to any external system. A deterministic in-process mock executor proves the entire lifecycle end to end before Phase 5 adds live integrations.

### Endpoint

```http
POST /api/v1/servers/{serverID}/tools/{toolID}/invoke
Authorization: Bearer <access-token>
Content-Type: application/json
```

### Invocation policy chain

1. **Authentication** — valid, unexpired JWT.
2. **Role** — `admin` or `developer`; `viewer` receives `403`.
3. **Server status** — server must exist and be `active`, otherwise `409 server_inactive`.
4. **Tool state** — tool must exist under the server and be `enabled`, otherwise `409 tool_disabled`.
5. **Risk level** — only `low` in Phase 4, otherwise `403 tool_risk_not_allowed`.
6. **Argument validation** — arguments must be a JSON object satisfying the tool's stored JSON Schema, otherwise `400 validation_error`.

### Audit lifecycle

```text
Validate request
    ↓
INSERT tool_invocations (status = 'running')
    ↓
Execute via ToolExecutor
    ↓
UPDATE tool_invocations
    ├── success → status 'succeeded', response_payload, duration_ms, completed_at
    └── failure → status 'failed', error_code, error_message, duration_ms, completed_at
```

Audit foreign keys use `ON DELETE RESTRICT`: deleting a server, tool, or user is blocked while audit records reference it.

---

## Phase 5: Live GitHub Integration

**Status:** Complete

### Goal

Replace the mock path for one registered server with a real, read-only GitHub REST integration. The lifecycle proven in Phase 4 stays identical; only the executor gains a second implementation that makes live API calls.

### Executor routing

The `RouterExecutor` dispatches to the appropriate executor based on the registered server's `name` field:

| Server name | Executor | Behavior |
|---|---|---|
| `github` | `GitHubExecutor` | Makes live GitHub REST API calls using `go-github` |
| Any other | `MockExecutor` | Returns deterministic mock responses |

### GitHub executor

Uses `google/go-github/v62` for typed access to the GitHub REST API v3. Authentication is optional:

- **Unauthenticated** (`GITHUB_TOKEN=` empty): 60 requests/hour per IP, public repos only
- **Authenticated** (`GITHUB_TOKEN=github_pat_...`): 5000 requests/hour, can access private repos

Supported tools:

| Tool name | Arguments | Live behavior |
|---|---|---|
| `list_issues` | `{ "owner", "repo", "state?", "per_page?" }` | Lists issues from a GitHub repository |
| `search_repositories` | `{ "query", "per_page?" }` | Searches repositories by name/description |

---

## Phase 6: Observability

**Status:** Complete

### Goal

Add full observability to the gateway: Prometheus metrics for monitoring, OpenTelemetry tracing for debugging, and an invocation history API for querying audit records.

### Prometheus metrics

The gateway exposes metrics at `/metrics` in Prometheus exposition format. Metrics are collected for:

**HTTP requests:**
- `mcp_gateway_http_requests_total` — Total HTTP requests by method, path, status
- `mcp_gateway_http_request_duration_seconds` — HTTP request duration histogram by method, path

**Tool invocations:**
- `mcp_gateway_invocations_total` — Total invocations by server, tool, status
- `mcp_gateway_invocation_duration_seconds` — Invocation duration histogram by server, tool

**Upstream API calls:**
- `mcp_gateway_upstream_requests_total` — Total upstream API requests by service (github, etc.), status

**Database:**
- `mcp_gateway_database_connections_open` — Number of open database connections (gauge, updated every 15s)

Path normalization replaces UUIDs with `:id` to avoid high cardinality:

```text
/api/v1/servers/c3b00a6e-78b3-4e8d-b40f-dc7e6149625d/tools/2d81df9b-b69a-41a1-9fe9-3c1535ce8b62/invoke
→ /api/v1/servers/:id/tools/:id/invoke
```

### Invocation history API

Query audit records via REST endpoints:

```http
GET /api/v1/invocations?serverId={uuid}&toolId={uuid}&status=succeeded&limit=50&offset=0
Authorization: Bearer <access-token>
```

**Query parameters:**

| Parameter | Type | Description |
|---|---|---|
| `serverId` | UUID | Filter by server ID (optional) |
| `toolId` | UUID | Filter by tool ID (optional) |
| `status` | string | Filter by status: `running`, `succeeded`, `failed`, `denied` (optional) |
| `limit` | integer | Max results (default 50, max 100) |
| `offset` | integer | Pagination offset (default 0) |

**Role-based access:**

- **Admin**: Sees all invocations from all users
- **Developer**: Sees only their own invocations
- **Viewer**: No access to history endpoints

**Get specific invocation:**

```http
GET /api/v1/invocations/{invocationID}
Authorization: Bearer <access-token>
```

Non-admin users can only view their own invocations.

### OpenTelemetry tracing

The gateway initializes an OpenTelemetry tracer provider with a stdout exporter for local development. In production, this can be swapped for an OTLP exporter to send traces to Jaeger, Tempo, or other backends.

Traces are created for:
- HTTP requests (via middleware)
- Tool invocations (span from policy check to execution)
- Upstream API calls (GitHub requests)

### Verified metrics output

After making several invocations:

```bash
curl -s http://localhost:8080/metrics | grep mcp_gateway
```

Expected output:

```text
mcp_gateway_http_requests_total{method="POST",path="/api/v1/servers/:id/tools/:id/invoke",status="200"} 5
mcp_gateway_http_request_duration_seconds_bucket{method="POST",path="/api/v1/servers/:id/tools/:id/invoke",le="0.5"} 5
mcp_gateway_invocations_total{server="github",tool="list_issues",status="succeeded"} 3
mcp_gateway_invocations_total{server="mock-tools",tool="echo",status="succeeded"} 2
mcp_gateway_invocation_duration_seconds_bucket{server="github",tool="list_issues",le="0.5"} 3
mcp_gateway_upstream_requests_total{service="github",status="success"} 3
```

### Verified invocation history

```bash
# Admin sees all invocations
curl -s "http://localhost:8080/api/v1/invocations?limit=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Filter by server
curl -s "http://localhost:8080/api/v1/invocations?serverId=$SERVER_ID&status=succeeded" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Developer sees only their own
curl -s "http://localhost:8080/api/v1/invocations?limit=10" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" | jq
```

Response:

```json
{
  "count": 5,
  "data": [
    {
      "id": "cb112ed5-92fc-488a-80d6-e60be4af1e7f",
      "serverId": "c3b00a6e-78b3-4e8d-b40f-dc7e6149625d",
      "toolId": "2d81df9b-b69a-41a1-9fe9-3c1535ce8b62",
      "userId": "265d8f73-0221-45c5-9890-dcaa1dfc62ee",
      "status": "succeeded",
      "requestArguments": {
        "owner": "golang",
        "repo": "go",
        "state": "open",
        "per_page": 5
      },
      "responsePayload": {
        "count": 4,
        "issues": [...]
      },
      "durationMs": 404,
      "createdAt": "2026-08-03T17:46:38.835182-04:00",
      "completedAt": "2026-08-03T17:46:38.835182-04:00"
    }
  ]
}
```

### Phase 6 acceptance criteria

- [x] Prometheus metrics endpoint at `/metrics`
- [x] HTTP request metrics (count, duration) by method/path/status
- [x] Invocation metrics (count, duration) by server/tool/status
- [x] Upstream API metrics (GitHub requests by success/error)
- [x] Database connection pool metrics
- [x] Path normalization to avoid high cardinality
- [x] Invocation history list endpoint with filtering (server, tool, status)
- [x] Invocation history detail endpoint
- [x] Pagination support (limit/offset, max 100)
- [x] Role-based history access (admin sees all, developer sees own)
- [x] OpenTelemetry tracing initialized (stdout exporter)
- [x] `go test ./...` and `go vet ./...` pass

---

## Validation and Testing

```bash
cd backend

gofmt -w .
go mod tidy
go test ./...
go vet ./...
```

### Metrics smoke test

```bash
# Make a few invocations
curl -s -X POST \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID/invoke" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"arguments": {"owner": "golang", "repo": "go", "per_page": 3}}' | jq

# Check metrics
curl -s http://localhost:8080/metrics | grep mcp_gateway
```

### Invocation history smoke test

```bash
# List recent invocations
curl -s "http://localhost:8080/api/v1/invocations?limit=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Filter by status
curl -s "http://localhost:8080/api/v1/invocations?status=failed" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

---

## Design Decisions

### Why Prometheus for metrics

Prometheus is the de facto standard for metrics in cloud-native systems:

- Pull-based model (gateway exposes `/metrics`, Prometheus scrapes)
- Rich query language (PromQL) for aggregations
- Wide ecosystem (Grafana, Alertmanager, etc.)
- Go client library is mature and well-maintained

Alternative approaches considered:

- **StatsD/DataDog:** Requires external agent, adds operational complexity
- **Custom metrics endpoint:** Reinvents the wheel, lacks ecosystem

### Why path normalization for metrics

High-cardinality metrics (unique label combinations) can cause performance issues in Prometheus. By replacing UUIDs with `:id`, we reduce the number of unique time series:

```text
Without normalization:
mcp_gateway_http_requests_total{path="/api/v1/servers/c3b00a6e.../invoke",status="200"} 1
mcp_gateway_http_requests_total{path="/api/v1/servers/a1b2c3d4.../invoke",status="200"} 1
# ... thousands of unique paths

With normalization:
mcp_gateway_http_requests_total{path="/api/v1/servers/:id/tools/:id/invoke",status="200"} 5000
```

This keeps Prometheus memory usage reasonable and queries fast.

### Why role-based history access

Invocation history contains sensitive information:

- What tools were invoked
- What arguments were passed
- What results were returned
- When invocations occurred

Restricting non-admin users to their own invocations prevents information leakage:

- Developers can't see what other developers are doing
- Viewers can't access any history (they can't invoke tools anyway)
- Admins have full visibility for debugging and auditing

### Why OpenTelemetry for tracing

OpenTelemetry is the CNCF standard for observability:

- Vendor-neutral (works with Jaeger, Tempo, Zipkin, etc.)
- Automatic instrumentation via middleware
- Context propagation across service boundaries
- Future-proof (can add spans to any operation)

The stdout exporter is used for local development because it requires no external services. In production, swap for OTLP exporter to send traces to a backend.

---

## Known Limitations

Intentional at the end of Phase 6:

- Only two GitHub tools (`list_issues`, `search_repositories`); no `get_issue`, `list_pull_requests`, etc.
- GitHub token stored in environment variable, not a credential reference store
- No rate limit tracking or backoff logic (relies on GitHub's 403 response)
- Only `low` risk tools are invocable; no confirmation flow exists for medium/high
- Roles are global; no per-server or per-tool user permissions
- No Grafana dashboards (metrics are exposed but not visualized)
- No Jaeger/Tempo backend (traces go to stdout only)
- No alerting (Prometheus metrics are collected but no alert rules defined)
- Schema compilation happens per request (no cache)
- No OAuth/OIDC, refresh tokens, or token revocation
- No React UI, CI/CD, or deployment
- No MCP protocol transport (stdio/SSE/streamable HTTP); all integrations are direct REST API calls

---

## Planned Phases

### Phase 7: React UI (Next)

Build the discovery and sandbox interface:

- MCP server catalog with search and filtering
- Tool detail pages with input schema viewer
- JSON invocation sandbox with live results
- **Invocation history page with filtering and export**
- **Metrics dashboard using Prometheus data**
- Admin management (users, servers, tools)
- Observability dashboard (metrics, traces, health)

### Phase 8: Delivery and Polish

Prepare for production deployment:

- GitHub Actions CI (test, lint, build)
- Backend and frontend Dockerfiles
- Integration tests using Testcontainers
- OpenAPI documentation
- **Grafana dashboards for metrics visualization**
- **Jaeger/Tempo for trace visualization**
- **Alertmanager for alerting rules**
- Cloud deployment (AWS/GCP/Azure)
- Architecture decision records
- Demo video and screenshots
- Security and operations documentation

---

## Troubleshooting

### Metrics endpoint returns empty

Make sure you've made at least one request after starting the server. Metrics are initialized with zero values and only appear after the first observation.

### Invocation history returns empty for developer

Developers can only see their own invocations. If a developer hasn't invoked any tools, the response will be `{"count":0,"data":[]}`.

### Traces not appearing in stdout

The stdout exporter batches traces. You may need to wait a few seconds or make multiple requests before traces appear. For immediate output, use `sdktrace.WithSyncer(exporter)` instead of `WithBatcher`.

### GitHub API returns 403 rate limit exceeded

You've exceeded the unauthenticated limit (60 requests/hour). Either:

1. Wait for the rate limit to reset (check `X-RateLimit-Reset` header in the error)
2. Create a fine-grained personal access token and set `GITHUB_TOKEN` in `.env`

### Route returns 404 for tool endpoints

Trailing-slash mismatch. If routes are registered as `/{serverID}/tools/`, you must call `.../tools/` (with trailing slash). Alternatively, add `middleware.StripSlashes` to the router and register routes without trailing slashes.

### `DATABASE_URL is required` when running from `backend/`

The `.env` file is at the repository root. Either:

1. Run from the root: `go run ./backend/cmd/api`
2. Update `main.go` to load `../.env`:
   ```go
   _ = godotenv.Load()
   _ = godotenv.Load("../.env")
   ```

### Reset all server and tool records (keeps users)

```bash
psql "$DATABASE_URL" -c "TRUNCATE TABLE mcp_tools, mcp_servers CASCADE;"
```

Note: existing `tool_invocations` rows referencing those rows will block the truncate because of `RESTRICT`. Remove them first if needed:

```bash
psql "$DATABASE_URL" -c "TRUNCATE TABLE tool_invocations, mcp_tools, mcp_servers CASCADE;"
```

### Full local reset

```bash
docker compose down -v
docker compose up -d postgres
set -a; source .env; set +a
migrate -path backend/migrations -database "$DATABASE_URL" up
# Re-seed users
```

---

## Contributing Workflow

```bash
git checkout -b feat/react-ui

cd backend
gofmt -w .
go mod tidy
go test ./...
go vet ./...

cd ..
git add .
git commit -m "feat: add react ui for server catalog and invocation sandbox"
git push -u origin feat/react-ui
```

Suggested commit convention:

```text
feat: add observability with prometheus metrics
feat: add react ui for server catalog
fix: validate tool input schema
test: cover invalid JWT and role checks
docs: document phase 6 observability
chore: configure Docker Compose PostgreSQL
```

---

## Project Status

```text
Phase 0: Complete
Phase 1: Complete
Phase 2: Complete
Phase 3: Complete
Phase 4: Complete
Phase 5: Complete
Phase 6: Complete
Phase 7: Ready to begin — React UI
```

The next milestone is the **React UI**: a web interface for discovering servers, browsing tools, invoking them in a sandbox, viewing invocation history, and visualizing metrics.

---

## Demo: Observability in Action

Here's a complete walkthrough of Phase 6 observability features:

### 1. Make several invocations

```bash
# GitHub invocation
curl -s -X POST \
  "http://localhost:8080/api/v1/servers/$GITHUB_SERVER_ID/tools/$LIST_ISSUES_TOOL_ID/invoke" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"arguments": {"owner": "golang", "repo": "go", "per_page": 3}}' | jq

# Mock invocation
curl -s -X POST \
  "http://localhost:8080/api/v1/servers/$MOCK_SERVER_ID/tools/$ECHO_TOOL_ID/invoke" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"arguments": {"message": "test"}}' | jq
```

### 2. Check Prometheus metrics

```bash
curl -s http://localhost:8080/metrics | grep mcp_gateway_invocations_total
```

Output:

```text
mcp_gateway_invocations_total{server="github",tool="list_issues",status="succeeded"} 1
mcp_gateway_invocations_total{server="mock-tools",tool="echo",status="succeeded"} 1
```

### 3. Query invocation history

```bash
# All invocations (admin)
curl -s "http://localhost:8080/api/v1/invocations?limit=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.data[] | {id, status, tool: .toolId, duration: .durationMs}'

# Only GitHub invocations
curl -s "http://localhost:8080/api/v1/invocations?serverId=$GITHUB_SERVER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.data[] | {id, status, duration: .durationMs}'

# Only failed invocations
curl -s "http://localhost:8080/api/v1/invocations?status=failed" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.data[] | {id, error: .errorMessage}'
```

### 4. Verify role-based access

```bash
# Developer sees only their own invocations
curl -s "http://localhost:8080/api/v1/invocations?limit=10" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" | jq '.count'

# Compare with admin count (should be higher if other users invoked tools)
curl -s "http://localhost:8080/api/v1/invocations?limit=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.count'
```

This demonstrates the complete observability stack: metrics for monitoring, history API for querying, and role-based access control for security.