# MCP Gateway

A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations.

The project is inspired by the idea of an internal “USB-C for AI agents”: a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history.

> **Current status:** Phase 4 complete — Server Registry, Tool Catalog, JWT authentication, RBAC, and a policy-controlled, audited tool-invocation gateway are implemented and verified locally.

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
- Durable invocation audit records
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
├── Audit logging
└── Observability

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
│ Authentication API                   │
│ Server Registry API                  │
│ Tool Catalog API                     │
│ Invocation API                       │
│ JSON Schema argument validation      │
│ In-process mock tool executor        │
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

### Invocation request flow

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
Policy check — only low-risk tools in Phase 4
        │
        ▼
Validate arguments against stored JSON Schema
        │
        ▼
Create tool_invocations audit record (status: running)
        │
        ▼
Execute via registered mock executor
        │
        ▼
Update audit record: succeeded or failed
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
| Configuration | Environment variables + `godotenv` | Local configuration and secrets loading |
| Logging | Go `log/slog` + chi `WrapResponseWriter` | Structured logs with request ID, status, bytes, latency |
| API testing | `curl`, `jq`, Go `testing` | Manual API verification and automated tests |
| Containerization | Docker Compose | Local PostgreSQL environment |
| Frontend | React + TypeScript | Future discovery catalog and sandbox |
| Observability | Prometheus + OpenTelemetry | Future metrics, traces, and health visibility |
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
- Unit tests for server, tool, JWT, auth middleware, and invocation service behavior

### Not yet implemented

- OAuth/OIDC identity-provider integration
- Server-level and tool-level per-user permissions
- Medium/high-risk invocation confirmation flows
- Real MCP transport or external API invocation
- GitHub/Jira/Slack/Confluence live integrations
- Invocation history list endpoints
- Credential reference storage
- Prometheus metrics and OpenTelemetry traces
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
| Phase 5 | Next | First Live Integration | GitHub read-only invocation through the gateway |
| Phase 6 | Planned | Observability | Invocation history API, Prometheus metrics, OpenTelemetry traces |
| Phase 7 | Planned | React UI | Catalog, sandbox, history, and administration |
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
│   │   │   └── mock_executor.go
│   │   │
│   │   ├── httpapi/
│   │   │   ├── auth_handler.go
│   │   │   ├── auth_middleware.go
│   │   │   ├── auth_middleware_test.go
│   │   │   ├── errors.go
│   │   │   ├── handlers.go
│   │   │   ├── invocation_handler.go
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
│   │   │   ├── invocation_repository.go
│   │   │   ├── server_repository.go
│   │   │   ├── tool_repository.go
│   │   │   └── user_repository.go
│   │   │
│   │   └── service/
│   │       ├── auth_service.go
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
| `internal/executor` | Tool execution implementations (currently the deterministic mock executor) |
| `internal/httpapi` | Routes, handlers, middleware, JSON decoding, responses, HTTP error mapping |
| `internal/platform/database` | PostgreSQL connection-pool setup and health verification |
| `internal/repository` | Parameterized SQL and PostgreSQL persistence operations |
| `internal/service` | Business rules, validation, authentication, schema validation, invocation orchestration |
| `migrations` | Ordered and versioned database schema evolution |

Dependency direction:

```text
HTTP Handler / Middleware
           ↓
        Service
           ↓
  Repository / Executor
           ↓
  PostgreSQL / Mock Tools
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

Expected after Phase 4:

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

| Role | Catalog reads | Catalog mutations | Tool invocation |
|---|---:|---:|---:|
| `admin` | Yes | Yes | Yes (low-risk in Phase 4) |
| `developer` | Yes | No | Yes (low-risk in Phase 4) |
| `viewer` | Yes | No | No |
| Unauthenticated | No | No | No |

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

export VIEWER_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"viewer@mcp-gateway.local","password":"ViewerPass123!"}' \
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
Public:               /health, /api/v1/, POST /auth/login
Authenticated:        /auth/me, server/tool catalog reads
Admin + developer:    POST .../tools/{toolID}/invoke
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

Request body:

```json
{
  "arguments": {
    "message": "Hello from Phase 4"
  }
}
```

The request mirrors MCP's `tools/call` shape, where callers supply a tool name/identifier and an arguments object.

### Invocation policy chain

Every invocation must pass all of these checks before executing:

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

Audit foreign keys use `ON DELETE RESTRICT`: deleting a server, tool, or user is blocked while audit records reference it. This preserves the accountability trail until an explicit retention policy is introduced.

### `tool_invocations` table

| Field | Description |
|---|---|
| `id` | UUID primary key |
| `server_id` / `tool_id` / `user_id` | Referenced resources (`RESTRICT`) |
| `status` | `running`, `succeeded`, `failed`, or `denied` |
| `request_arguments` | Validated invocation arguments (`jsonb`) |
| `response_payload` | Tool result (`jsonb`, nullable) |
| `error_code` / `error_message` | Failure classification (nullable) |
| `duration_ms` | Execution duration |
| `created_at` / `completed_at` | Lifecycle timestamps |

### Mock executor

The Phase 4 executor is in-process and deterministic. It makes **no network calls** and reads **no credentials**.

Supported tools:

| Tool name | Arguments | Mock behavior |
|---|---|---|
| `echo` | `{ "message": string }` | Returns the message verbatim |
| `list_issues` | `{ "owner", "repo", "state?", "per_page?" }` | Returns two stable mock issues with `"mock": true` |

Any other tool name returns an unsupported-tool error, which is recorded as a failed invocation.

### JSON Schema validation details

Arguments are validated with `santhosh-tekuri/jsonschema/v6` (draft 2020-12 default). Both the stored schema and the incoming arguments are parsed with `jsonschema.UnmarshalJSON` and passed to `compiler.AddResource` / `schema.Validate` as decoded JSON values.

> Implementation note: passing an `io.Reader` (e.g., `bytes.Reader`) to `AddResource` is a v5 API pattern and fails at runtime in v6 with `invalid jsonType *bytes.Reader`. Always feed v6 decoded documents via `jsonschema.UnmarshalJSON`.

### Verified manual walkthrough

Register a mock server (admin), then a low-risk echo tool:

```bash
curl -s -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mock-tools",
    "description": "Local deterministic mock tools for invocation testing",
    "baseUrl": "http://localhost:9999",
    "transportType": "streamable_http",
    "ownerTeam": "developer-platform"
  }' | jq
```

```bash
curl -s -X POST "http://localhost:8080/api/v1/servers/$SERVER_ID/tools" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "echo",
    "title": "Echo",
    "description": "Returns the supplied message for deterministic gateway testing",
    "inputSchema": {
      "type": "object",
      "properties": {
        "message": { "type": "string", "minLength": 1 }
      },
      "required": ["message"],
      "additionalProperties": false
    },
    "riskLevel": "low",
    "enabled": true
  }' | jq
```

Invoke as a developer:

```bash
curl -s -X POST \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID/invoke" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"arguments": {"message": "Hello from Phase 4"}}' | jq
```

Verified response:

```json
{
  "invocationId": "INVOCATION_UUID",
  "serverId": "SERVER_UUID",
  "toolId": "TOOL_UUID",
  "toolName": "echo",
  "status": "succeeded",
  "result": { "message": "Hello from Phase 4" },
  "durationMs": 9,
  "completedAt": "2026-08-03T15:18:00Z"
}
```

Verify the audit record:

```bash
psql "$DATABASE_URL" -c "
  SELECT status, request_arguments, response_payload, duration_ms
  FROM tool_invocations
  ORDER BY created_at DESC
  LIMIT 1;
"
```

### Phase 4 acceptance criteria

- [x] Migration creates `tool_invocations` with `RESTRICT` foreign keys
- [x] Developer can invoke an enabled, active, low-risk tool
- [x] Arguments are validated against the tool's stored JSON Schema
- [x] Missing/invalid/extra arguments return `400 validation_error`
- [x] Viewer invocation returns `403`
- [x] Medium/high-risk tools return `403 tool_risk_not_allowed`
- [x] Disabled tools return `409 tool_disabled`
- [x] Inactive servers return `409 server_inactive`
- [x] Successful invocations persist `succeeded` audit rows with result and latency
- [x] Failed executions persist `failed` audit rows with error details
- [x] Mock executor returns deterministic output with no network access
- [x] Request logs include HTTP status codes and response sizes
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

### Invocation smoke test

```bash
curl -s -X POST \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID/invoke" \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"arguments": {"message": "smoke test"}}' | jq
```

### Structured request logging

Every request logs method, path, request ID, status, bytes, and duration:

```json
{"time":"...","level":"INFO","msg":"http request completed","request_id":"...","method":"POST","path":"/api/v1/servers/.../invoke","status":200,"bytes":312,"duration_ms":9}
```

Status capture uses chi's `middleware.NewWrapResponseWriter`; a status of `0` (handler never wrote one explicitly) is logged as `200`.

---

## Design Decisions

### Why a mock executor before live integrations

The invocation path has many failure modes — authentication, authorization, policy, schema validation, audit persistence — that are independent of any third-party API. Proving them against a deterministic executor means Phase 5's GitHub integration only adds one new variable: the external call itself.

### Why `ON DELETE RESTRICT` for audit records

Server/tool deletion cascades made sense for the catalog (`mcp_tools` → `mcp_servers`), because orphaned catalog rows are useless. Audit rows are the opposite: they are the record of what happened, so deleting referenced resources is blocked rather than silently erasing history.

### Why validate arguments against stored schemas

The gateway is the policy boundary. Validating arguments against each tool's stored `inputSchema` guarantees that every invocation — and its audit record — contains exactly the shape the tool declared, before any executor or external system sees it.

### Why `jsonschema/v6`

Draft 2020-12 support, active maintenance, and a compiled-schema API that maps cleanly onto stored `jsonb` documents. Schemas are parsed once per invocation via `jsonschema.UnmarshalJSON`; a compiled-schema cache is a Phase 6 optimization.

---

## Known Limitations

Intentional at the end of Phase 4:

- Invocation executes against an in-process mock, not real MCP servers or APIs.
- Only `low` risk tools are invocable; no confirmation flow exists for medium/high.
- Roles are global; no per-server or per-tool user permissions.
- No invocation history list endpoints yet (query PostgreSQL directly).
- Failed invocations return a generic API error; details live in the audit row and logs.
- Schema compilation happens per request (no cache).
- No metrics, tracing, or rate limiting.
- No OAuth/OIDC, refresh tokens, or token revocation.
- No React UI, CI/CD, or deployment.

---

## Planned Phases

### Phase 5: First Live Integration (Next)

Replace the mock path for one server with a real GitHub read-only adapter:

- `search_repositories`, `list_issues`, `get_issue`
- GitHub token via environment variable, never stored in the database
- Executor selection by server type (`mock` vs `github`)
- External call latency and errors recorded in the same audit table

### Phase 6: Observability

- Invocation history list/detail endpoints
- Prometheus metrics (`/metrics`)
- OpenTelemetry traces across the invocation chain
- Tool success/error rates and p95 latency

### Phase 7: React UI

Catalog, tool detail with schema viewer, JSON invocation sandbox, invocation history, admin management.

### Phase 8: Delivery and Polish

GitHub Actions CI, Dockerfiles, Testcontainers integration tests, OpenAPI docs, deployment, demo video.

---

## Troubleshooting

### Invocation returns `invocation_failed` with no audit row

The failure occurred before the audit insert: check the backend terminal — the handler logs the root cause via `slog.Error("tool invocation failed", ...)`. A known cause was passing an `io.Reader` to `jsonschema/v6` `AddResource`; use `jsonschema.UnmarshalJSON` to decode documents first.

### `compile tool input schema ... invalid jsonType *bytes.Reader`

You are on `jsonschema/v6` but using the v5 `AddResource(url, io.Reader)` pattern. See the v6-compatible `schema_validator.go` described in Phase 4.

### Route returns 404 / non-JSON response on tool creation

Trailing-slash mismatch. `r.Post("/{serverID}/tools/", ...)` matches `.../tools/`, not `.../tools`. Either call the trailing-slash URL or add `middleware.StripSlashes` to the router.

### Request logs lack status codes

Wrap the writer: `middleware.NewWrapResponseWriter(w, r.ProtoMajor)` and log `ww.Status()` / `ww.BytesWritten()` after `next.ServeHTTP`.

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
git checkout -b feat/github-integration

cd backend
gofmt -w .
go mod tidy
go test ./...
go vet ./...

cd ..
git add .
git commit -m "feat: add github read-only tool executor"
git push -u origin feat/github-integration
```

---

## Project Status

```text
Phase 0: Complete
Phase 1: Complete
Phase 2: Complete
Phase 3: Complete
Phase 4: Complete
Phase 5: Ready to begin — GitHub live integration
```

The next milestone is the **first live integration**: a GitHub read-only executor behind the existing invocation gateway, with real tool results flowing through the same policy, schema-validation, and audit pipeline proven in Phase 4.