# MCP Gateway

A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations.

The project is inspired by the idea of an internal “USB-C for AI agents”: a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history.

> **Current status:** Phase 3 complete — the MCP Server Registry, MCP Tool Catalog, JWT authentication, and role-based access control are implemented and validated locally.

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
│ Policy Enforcement -  Invocation Proxy -  Audit Logs -  Metrics   │
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
      │ Users -  Servers -  Tools -  Permissions -  Audit Records   │
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
- JWT authentication
- Role-based access control
- Secure password handling
- Tool invocation policy enforcement
- Auditability and observability
- React-based developer tooling
- Containerized local development
- Production-style testing and deployment practices

The long-term architecture separates deterministic infrastructure from agentic reasoning.

```text
Go Gateway
├── Authentication
├── Authorization
├── Server registry
├── Tool catalog
├── Validation
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

---

## System Architecture

### Current architecture

```text
Developer / API Client
         │
         │ HTTP / JSON
         ▼
┌─────────────────────────────────────┐
│ Go MCP Gateway                      │
│                                     │
│ chi Router                          │
│ Request ID middleware               │
│ Timeout middleware                  │
│ JWT authentication middleware       │
│ RBAC middleware                     │
│ Structured logs                     │
│ Health endpoint                     │
│ Authentication API                  │
│ Server Registry API                 │
│ Tool Catalog API                    │
│ Request validation                  │
└────────────────┬────────────────────┘
                 │
                 │ pgx connection pool
                 ▼
┌─────────────────────────────────────┐
│ PostgreSQL 16                       │
│                                     │
│ users                               │
│ mcp_servers                         │
│ mcp_tools                           │
│ schema_migrations                   │
└─────────────────────────────────────┘
```

### Authentication request flow

```text
POST /api/v1/auth/login
        │
        ▼
Validate email and password
        │
        ▼
Fetch user by email from PostgreSQL
        │
        ▼
bcrypt password comparison
        │
        ▼
Issue HS256 signed JWT
        │
        ▼
Client sends Authorization: Bearer <token>
        │
        ▼
JWT validation middleware
        │
        ▼
Authenticated user stored in request context
        │
        ▼
Role middleware verifies allowed role
        │
        ▼
Protected API handler
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
| Database | PostgreSQL 16 | Users, server registry, tools, permissions, audit records |
| PostgreSQL driver | `pgx/v5` | Native PostgreSQL driver and connection pool |
| Migrations | `golang-migrate` | Version-controlled schema migrations |
| JSON storage | PostgreSQL `jsonb` | Tool input schemas and future structured metadata |
| Authentication | JWT with HS256 | Local development access tokens |
| JWT library | `github.com/golang-jwt/jwt/v5` | JWT signing, parsing, verification, and claims validation |
| Password hashing | `golang.org/x/crypto/bcrypt` | Secure password hash generation and comparison |
| Configuration | Environment variables + `godotenv` | Local configuration and secrets loading |
| Logging | Go `log/slog` | Structured JSON logs |
| API testing | `curl`, `jq`, Go `testing` | Manual API verification and automated tests |
| Containerization | Docker Compose | Local PostgreSQL environment |
| Frontend | React + TypeScript | Future discovery catalog and sandbox |
| Observability | Prometheus + OpenTelemetry | Future metrics, traces, and health visibility |
| MCP integration | Official Go MCP SDK | Future MCP discovery and tool invocation |

---

## Current Features

### Completed

- Go module at `github.com/iashu2k/mcp-gateway/backend`
- Docker Compose-managed PostgreSQL 16 database
- Environment-based configuration
- PostgreSQL connection pool using `pgxpool`
- Health endpoint with database connectivity validation
- Structured JSON application logs
- HTTP request ID and timeout middleware
- Graceful HTTP shutdown
- Version-controlled PostgreSQL schema migrations
- MCP server registry with CRUD operations
- MCP tool catalog with CRUD operations
- Server-to-tool foreign key relationship
- Cascading tool deletion when a server is deleted
- JSON Schema storage and validation for tool inputs
- Tool risk classifications: `low`, `medium`, and `high`
- Tool enabled/disabled state
- Field-level validation responses
- Duplicate server-name prevention
- Duplicate tool-name prevention within a server
- PostgreSQL typed unique-constraint handling
- Local user table with active/inactive status
- Bcrypt password-hash verification
- Signed JWT access-token issuance
- JWT issuer and expiration validation
- Bearer-token authentication middleware
- `admin`, `developer`, and `viewer` roles
- Admin-only server/tool mutations
- Authenticated catalog reads
- Current-user endpoint
- Unit tests for server, tool, JWT, and authentication middleware behavior

### Not yet implemented

- OAuth/OIDC identity-provider integration
- MCP Protected Resource Metadata endpoint
- Server-level user permissions
- Tool-level user permissions
- MCP tool discovery from remote MCP servers
- Remote server health checks
- Tool invocation proxying
- GitHub/Jira/Slack/Confluence integrations
- External credential-reference storage
- Audit records
- Prometheus metrics
- OpenTelemetry traces
- React UI
- CI/CD pipeline
- Production deployment

---

## Development Roadmap

| Phase | Status | Focus | Main Outcome |
|---|---|---|---|
| Phase 0 | Complete | Foundation | Go API, PostgreSQL, Docker Compose, health endpoint |
| Phase 1 | Complete | MCP Server Registry | Persistent CRUD API for MCP server metadata |
| Phase 2 | Complete | MCP Tool Catalog | Per-server tools, input schemas, risk levels, enablement |
| Phase 3 | Complete | Authentication and RBAC | JWT authentication, local users, bcrypt, role protection |
| Phase 4 | Next | Invocation Gateway | Validate, authorize, proxy, and audit tool calls |
| Phase 5 | Planned | First Live Integration | GitHub or Jira invocation through the gateway |
| Phase 6 | Planned | Observability | Audit logs, Prometheus metrics, OpenTelemetry traces |
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
│   │   │   ├── server.go
│   │   │   ├── tool.go
│   │   │   └── user.go
│   │   │
│   │   ├── httpapi/
│   │   │   ├── auth_handler.go
│   │   │   ├── auth_middleware.go
│   │   │   ├── auth_middleware_test.go
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
│   │   │   ├── tool_repository.go
│   │   │   └── user_repository.go
│   │   │
│   │   └── service/
│   │       ├── auth_service.go
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
│   │   └── 000003_create_users.up.sql
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
| `internal/domain` | Core domain models, request models, constants, and response structures |
| `internal/httpapi` | Routes, handlers, middleware, JSON decoding, responses, and HTTP error mapping |
| `internal/platform/database` | PostgreSQL connection-pool setup and health verification |
| `internal/repository` | Parameterized SQL and PostgreSQL persistence operations |
| `internal/service` | Business rules, validation, authentication, and application use cases |
| `migrations` | Ordered and versioned database schema evolution |

Dependency direction:

```text
HTTP Handler / Middleware
           ↓
        Service
           ↓
       Repository
           ↓
       PostgreSQL
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
```

Ensure Go-installed binaries are accessible:

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

### Configure JWT settings

Generate a long local development secret:

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

Expected:

```text
mcp-gateway-postgres ... Up (healthy)
```

### Load environment variables

Run this from the repository root:

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

Verify the migration state:

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

Expected after Phase 3:

```text
3
```

### Download backend dependencies

```bash
cd backend
go mod download
go mod tidy
cd ..
```

### Seed local users

Generate bcrypt password hashes:

```bash
cd backend

go run ./cmd/passwordhash "AdminPass123!"
go run ./cmd/passwordhash "DeveloperPass123!"
go run ./cmd/passwordhash "ViewerPass123!"
```

Create `scripts/seed_users.sql` locally from the example file:

```bash
cp scripts/seed_users.sql.example scripts/seed_users.sql
```

Replace each placeholder bcrypt hash in `scripts/seed_users.sql` with the corresponding generated hash.

Run the seed script:

```bash
set -a
source .env
set +a

psql "$DATABASE_URL" -f scripts/seed_users.sql
```

Verify seeded users:

```bash
psql "$DATABASE_URL" \
  -c "SELECT id, email, display_name, role, active FROM users ORDER BY role;"
```

Local development users:

| Email | Password | Role |
|---|---|---|
| `admin@mcp-gateway.local` | `AdminPass123!` | `admin` |
| `developer@mcp-gateway.local` | `DeveloperPass123!` | `developer` |
| `viewer@mcp-gateway.local` | `ViewerPass123!` | `viewer` |

> These credentials are for local development only. Do not use them in any deployed environment.

### Start the API

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

### Credential safety

The `.env` file and `scripts/seed_users.sql` are local-only files and must not be committed.

Future service credentials such as `GITHUB_TOKEN`, `JIRA_TOKEN`, and `SLACK_BOT_TOKEN` should be supplied through environment variables locally and a secrets manager in deployed environments.

---

## Running the Project

### Start infrastructure

```bash
docker compose up -d postgres
```

### Start backend

```bash
cd backend
go run ./cmd/api
```

### Health check

The health endpoint is intentionally public:

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

### Unauthenticated catalog access

Catalog access now requires a bearer token:

```bash
curl -i http://localhost:8080/api/v1/servers
```

Expected:

```text
HTTP/1.1 401 Unauthorized
```

### Stop local services

Stop the backend:

```text
Ctrl+C
```

Stop PostgreSQL while preserving data:

```bash
docker compose down
```

Remove PostgreSQL and all local data:

```bash
docker compose down -v
```

> Warning: `docker compose down -v` destroys all locally stored users, servers, and tools.

---

## Database Migrations

Migration files are stored in:

```text
backend/migrations/
```

Current migrations:

```text
000001_create_mcp_servers.up.sql
000001_create_mcp_servers.down.sql

000002_create_mcp_tools.up.sql
000002_create_mcp_tools.down.sql

000003_create_users.up.sql
000003_create_users.down.sql
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

### Check migration version

```bash
migrate \
  -path backend/migrations \
  -database "$DATABASE_URL" \
  version
```

### Roll back the newest migration

Only use this in local development:

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

Then run:

```sql
\dt
\d users
\d mcp_servers
\d mcp_tools

SELECT id, email, display_name, role, active FROM users;
SELECT * FROM mcp_servers;
SELECT * FROM mcp_tools;
```

---

## Authentication and Roles

### Local authentication design

Phase 3 implements local development authentication using email/password login and signed JWT access tokens.

```text
Password submitted by client
        ↓
bcrypt compares it to stored password hash
        ↓
Gateway verifies active user status
        ↓
Gateway generates an HS256-signed JWT
        ↓
Client sends token as Bearer authentication
        ↓
Gateway validates signature, issuer, expiration, subject, and role
```

Passwords are never stored in plaintext. The API does not serialize the `password_hash` field.

### JWT claims

The gateway includes these claims in issued access tokens:

| Claim | Meaning |
|---|---|
| `sub` | Authenticated user UUID |
| `iss` | JWT issuer, configured by `JWT_ISSUER` |
| `iat` | Token issuance time |
| `exp` | Token expiration time |
| `email` | User email |
| `displayName` | User display name |
| `role` | `admin`, `developer`, or `viewer` |

The application explicitly requires HS256 when parsing a token, validates the configured issuer, requires an expiration claim, checks that the subject is a valid UUID, and rejects unknown roles.

### Role model

| Role | Catalog reads | Server/tool creation | Server/tool update | Server/tool deletion | Future tool invocation |
|---|---:|---:|---:|---:|---:|
| `admin` | Yes | Yes | Yes | Yes | Yes, subject to policy |
| `developer` | Yes | No | No | No | Low-risk approved tools |
| `viewer` | Yes | No | No | No | No |
| Unauthenticated | No | No | No | No | No |

### Bearer token format

Send the token in the HTTP `Authorization` header:

```text
Authorization: Bearer <access-token>
```

Example:

```bash
curl -s http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json
```

Request:

```json
{
  "email": "admin@mcp-gateway.local",
  "password": "AdminPass123!"
}
```

Example:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mcp-gateway.local",
    "password": "AdminPass123!"
  }' | jq
```

Successful response:

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "tokenType": "Bearer",
  "expiresAt": "2026-08-01T22:00:00Z",
  "user": {
    "id": "USER_UUID",
    "email": "admin@mcp-gateway.local",
    "displayName": "Gateway Admin",
    "role": "admin"
  }
}
```

### Save an admin token

```bash
export ADMIN_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "admin@mcp-gateway.local",
      "password": "AdminPass123!"
    }' | jq -r '.accessToken'
)"
```

Confirm a token was returned without printing it in full:

```bash
echo "$ADMIN_TOKEN" | cut -c1-30
```

### Current authenticated user

```http
GET /api/v1/auth/me
Authorization: Bearer <access-token>
```

Example:

```bash
curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Developer token

```bash
export DEVELOPER_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "developer@mcp-gateway.local",
      "password": "DeveloperPass123!"
    }' | jq -r '.accessToken'
)"
```

### Viewer token

```bash
export VIEWER_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "viewer@mcp-gateway.local",
      "password": "ViewerPass123!"
    }' | jq -r '.accessToken'
)"
```

### Expected authorization behavior

Developer can list servers:

```bash
curl -i http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $DEVELOPER_TOKEN"
```

Expected: `200 OK`.

Developer cannot create a server:

```bash
curl -i -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $DEVELOPER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "unauthorized-server",
    "description": "This request must be denied",
    "baseUrl": "http://localhost:3003",
    "transportType": "streamable_http",
    "ownerTeam": "developer-platform"
  }'
```

Expected:

```text
HTTP/1.1 403 Forbidden
```

Invalid credentials:

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mcp-gateway.local",
    "password": "wrong-password"
  }'
```

Expected:

```text
HTTP/1.1 401 Unauthorized
```

---

## API Reference

Base URL:

```text
http://localhost:8080/api/v1
```

### Public endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/health` | Gateway and PostgreSQL health check |
| `GET` | `/api/v1/` | API root message |
| `POST` | `/api/v1/auth/login` | Authenticate local user and receive JWT |

### Authenticated endpoints

| Method | Endpoint | Allowed roles | Purpose |
|---|---|---|---|
| `GET` | `/api/v1/auth/me` | Admin, developer, viewer | Current authenticated identity |
| `GET` | `/api/v1/servers` | Admin, developer, viewer | List MCP servers |
| `GET` | `/api/v1/servers/{serverID}` | Admin, developer, viewer | Get MCP server |
| `GET` | `/api/v1/servers/{serverID}/tools` | Admin, developer, viewer | List tools for server |
| `GET` | `/api/v1/servers/{serverID}/tools/{toolID}` | Admin, developer, viewer | Get tool |

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

## MCP Server Registry API

All endpoints in this section require:

```text
Authorization: Bearer <access-token>
```

### Create server

```bash
curl -s -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "github-mcp",
    "description": "GitHub tools exposed through the MCP Gateway",
    "baseUrl": "http://localhost:3001",
    "transportType": "streamable_http",
    "ownerTeam": "developer-platform"
  }' | jq
```

### List servers

```bash
curl -s http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Get server

```bash
curl -s "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Update server

```bash
curl -s -X PATCH "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "inactive"
  }' | jq
```

### Delete server

```bash
curl -i -X DELETE "http://localhost:8080/api/v1/servers/$SERVER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Deleting a server also removes its associated tools through the `mcp_tools.server_id` foreign-key cascade.

---

## MCP Tool Catalog API

MCP tools allow clients and language models to interact with external systems, including APIs, databases, and internal workflows. Each record stores the expected tool arguments as an `inputSchema` JSON object.

### Create a tool

```bash
curl -s -X POST "http://localhost:8080/api/v1/servers/$SERVER_ID/tools" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
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
        }
      },
      "required": ["owner", "repo"],
      "additionalProperties": false
    },
    "riskLevel": "low",
    "enabled": true
  }' | jq
```

### List tools for a server

```bash
curl -s "http://localhost:8080/api/v1/servers/$SERVER_ID/tools" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Get one tool

```bash
curl -s \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Update tool

```bash
curl -s -X PATCH \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "riskLevel": "medium",
    "enabled": false
  }' | jq
```

### Delete tool

```bash
curl -i -X DELETE \
  "http://localhost:8080/api/v1/servers/$SERVER_ID/tools/$TOOL_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

## API Error Format

Errors use this structured JSON shape:

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
| `400` | `invalid_json` | Malformed JSON or unknown request field |
| `400` | `validation_error` | Submitted request fields are invalid |
| `400` | `invalid_server_id` | `serverID` is not a UUID |
| `400` | `invalid_tool_id` | `toolID` is not a UUID |
| `401` | `missing_or_invalid_authorization` | Missing or malformed Bearer header |
| `401` | `invalid_token` | JWT signature, claims, or format is invalid |
| `401` | `expired_token` | JWT is expired |
| `401` | `invalid_credentials` | Email/password combination is invalid |
| `403` | `forbidden` | Authenticated user lacks required role |
| `404` | `server_not_found` | Requested MCP server does not exist |
| `404` | `tool_not_found` | Tool does not exist under the requested server |
| `409` | `duplicate_server_name` | Another server already uses that name |
| `409` | `duplicate_tool_name` | Tool name already exists for the parent server |
| `500` | `internal_error` | Unexpected application or database failure |

---

## Phase 0: Foundation

**Status:** Complete

### Goal

Create a reliable local development baseline before adding product functionality.

### Completed work

- Git repository and Go module initialization
- Go API entry point
- Environment configuration
- Docker Compose PostgreSQL service
- PostgreSQL connection pool
- Health endpoint
- Request-ID middleware
- Timeout middleware
- Structured JSON logging
- Graceful shutdown
- Baseline tests

---

## Phase 1: MCP Server Registry

**Status:** Complete

### Goal

Create a persistent source of truth for MCP servers known to the gateway.

### Implemented endpoints

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
| `description` | Server purpose and supported capabilities |
| `baseUrl` | Future MCP server connection URL |
| `transportType` | `streamable_http`, `sse`, or `stdio` |
| `status` | `active`, `inactive`, or `unhealthy` |
| `ownerTeam` | Team responsible for server operations |
| `createdAt` | UTC record creation timestamp |
| `updatedAt` | UTC last-update timestamp |

### Phase 1 acceptance criteria

- [x] Migration creates `mcp_servers`
- [x] Server CRUD endpoints work
- [x] URL, transport, status, and required fields validate
- [x] Duplicate server names return `409`
- [x] Unknown server IDs return `404`
- [x] Unit tests pass

---

## Phase 2: MCP Tool Catalog

**Status:** Complete

### Goal

Associate discoverable tool metadata with each registered MCP server.

### Implemented endpoints

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
| `serverId` | Parent server UUID |
| `name` | Tool identifier, unique within parent server |
| `title` | Optional display name |
| `description` | Human-readable tool purpose |
| `inputSchema` | JSON Schema object for accepted arguments |
| `riskLevel` | `low`, `medium`, or `high` |
| `enabled` | Whether the tool can be exposed later |
| `createdAt` | UTC creation timestamp |
| `updatedAt` | UTC last-update timestamp |

### Tool risk levels

| Risk level | Examples | Future policy |
|---|---|---|
| `low` | Search repositories, list issues, read documentation | Allow for authorized developers |
| `medium` | Create Jira issue, send Slack message, open pull request | Require explicit confirmation |
| `high` | Delete repository, bulk-change tickets, alter production settings | Deny by default or require elevated approval |

### Phase 2 acceptance criteria

- [x] Migration creates `mcp_tools`
- [x] Tools belong to valid parent servers
- [x] Tool input schemas require a JSON object with `"type": "object"`
- [x] `input_schema` persists as PostgreSQL `jsonb`
- [x] Per-server tool-name uniqueness is enforced
- [x] Tool CRUD endpoints work
- [x] Tool risk and enabled state validate
- [x] Server deletion cascades to tool deletion
- [x] Unit tests pass

---

## Phase 3: Authentication and RBAC

**Status:** Complete

### Goal

Protect the gateway’s registry and catalog using local development authentication and role-based authorization.

### Implemented components

- `users` table and migration
- Local development users with email, display name, role, active state, and bcrypt password hash
- Password-hash generation utility
- Local user seed workflow
- Login endpoint
- Signed HS256 JWT access tokens
- Required JWT issuer validation
- Required JWT expiration validation
- Explicit HS256 signing-method validation
- Bearer-token parsing middleware
- Request-context storage of authenticated identity
- Role authorization middleware
- `admin`, `developer`, and `viewer` roles
- Authenticated `/auth/me` endpoint
- Catalog read protection
- Admin-only server and tool mutations
- Authentication and middleware tests

### User data model

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

| Field | Description |
|---|---|
| `id` | UUID user identifier |
| `email` | Unique login identifier |
| `password_hash` | Bcrypt hash; never serialized in API responses |
| `display_name` | User-facing display name |
| `role` | `admin`, `developer`, or `viewer` |
| `active` | Whether the account can log in |
| `created_at` | UTC creation timestamp |
| `updated_at` | UTC last-update timestamp |

### Route protection model

```text
Public:
GET  /health
GET  /api/v1/
POST /api/v1/auth/login

Authenticated:
GET  /api/v1/auth/me
GET  /api/v1/servers
GET  /api/v1/servers/{serverID}
GET  /api/v1/servers/{serverID}/tools
GET  /api/v1/servers/{serverID}/tools/{toolID}

Admin only:
POST   /api/v1/servers
PATCH  /api/v1/servers/{serverID}
DELETE /api/v1/servers/{serverID}
POST   /api/v1/servers/{serverID}/tools
PATCH  /api/v1/servers/{serverID}/tools/{toolID}
DELETE /api/v1/servers/{serverID}/tools/{toolID}
```

### Phase 3 acceptance criteria

- [x] Migration creates `users`
- [x] Passwords are stored as bcrypt hashes, not plaintext
- [x] Local seed users can log in
- [x] Login returns signed JWT access token and user metadata
- [x] Invalid credentials return `401`
- [x] Expired/invalid/malformed tokens return `401`
- [x] `/auth/me` returns authenticated identity
- [x] Catalog reads require valid authentication
- [x] Admins can mutate servers and tools
- [x] Developers and viewers can read catalogs
- [x] Developers and viewers receive `403` for catalog mutation attempts
- [x] Health endpoint remains public
- [x] Unit tests, formatting, and static checks pass

### MCP authorization direction

Phase 3 uses a local JWT issuer to establish authentication and role boundaries during development. The eventual MCP-facing authorization model should move to OAuth/OIDC: MCP protected servers are expected to expose OAuth Protected Resource Metadata so clients can discover authorization servers and obtain appropriate access tokens. 

---

## Validation and Testing

### Format code

```bash
cd backend
gofmt -w .
```

### Resolve dependency metadata

```bash
cd backend
go mod tidy
```

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

### Complete local quality check

```bash
cd backend

gofmt -w .
go mod tidy
go test ./...
go vet ./...
```

### Authentication smoke test

```bash
export ADMIN_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "admin@mcp-gateway.local",
      "password": "AdminPass123!"
    }' | jq -r '.accessToken'
)"

curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

curl -s http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

---

## Design Decisions

### Why Go

The gateway is an I/O-heavy control-plane service. It authenticates callers, queries PostgreSQL, invokes external MCP services or APIs, enforces policy, stores audits, and emits operational signals.

Go is a strong fit because it provides:

- Static typing
- Fast startup
- Straightforward container deployment
- Lightweight concurrency through goroutines
- Strong HTTP standard library
- Explicit error handling
- Good suitability for API gateways and platform services

Python remains appropriate for future LangGraph workflows, RAG pipelines, agent evaluation, and LLM reasoning.

### Why JWT for this phase

JWT provides a self-contained signed token format appropriate for a local development authentication flow. The gateway verifies the token signature and required claims without querying the user table on every authenticated request.

This phase validates:

- HS256 algorithm is used explicitly
- Issuer matches configured `JWT_ISSUER`
- Expiration is present and valid
- Subject is a valid UUID
- Role is known and supported

Future production work should use OIDC/OAuth and asymmetric signing with a trusted identity provider rather than a shared HMAC secret.

### Why bcrypt

Passwords must not be stored as plaintext. Bcrypt uses an adaptive password hashing algorithm, and its comparison function checks a submitted plaintext candidate against the stored hash. 

### Why local JWT before OIDC

A local development login flow makes the authorization model immediately testable without requiring external identity-provider configuration.

The transition path is:

```text
Phase 3:
Local PostgreSQL users + bcrypt + HS256 JWT

Future:
Keycloak/Auth0/Okta + OAuth/OIDC + JWKS + RS256/ES256 tokens
```

### Why layered backend design

```text
Handlers: HTTP concerns
Middleware: authentication and authorization enforcement
Services: business rules and use cases
Repositories: persistence
Domain: shared application structures
```

This structure makes later policy, audit, and invocation work easier to test and evolve.

---

## Known Limitations

The following are intentional at the end of Phase 3:

- Authentication is local-development only.
- JWT signing uses a shared HS256 secret rather than asymmetric keys.
- No token refresh endpoint exists.
- No logout, token revocation list, or session management exists.
- User registration and user-management endpoints do not exist.
- No password reset or account-recovery workflow exists.
- No OAuth/OIDC provider integration exists.
- No MCP OAuth Protected Resource Metadata endpoint exists.
- No server-level user permissions exist; roles are global.
- No tool-level user permissions exist; roles are global.
- No automatic MCP tool discovery exists.
- No server health checks exist.
- No external-tool invocation path exists.
- No external API credential-reference store exists.
- No audit log records or metrics exist.
- No React UI, CI/CD workflow, or deployed environment exists.

---

## Planned Phases

### Phase 4: Invocation Gateway

Phase 4 will introduce a controlled path for invoking a registered tool.

Planned lifecycle:

```text
Tool invocation request
        ↓
Authenticate caller
        ↓
Authorize caller role and future server/tool permission
        ↓
Confirm server is active
        ↓
Confirm tool is enabled
        ↓
Validate arguments against tool input schema
        ↓
Evaluate risk level and write-operation policy
        ↓
Create audit record
        ↓
Invoke a mock MCP server or adapter
        ↓
Redact sensitive values
        ↓
Store response, latency, and failure details
        ↓
Return structured invocation result
```

Start with a mock low-risk `echo` or `list_issues` tool before connecting to a live third-party API.

### Phase 5: First Live Integration

Initial integration target: GitHub read-only tools.

Candidate tools:

- `search_repositories`
- `list_issues`
- `get_issue`
- `list_pull_requests`
- `get_pull_request`

A later integration can route through an existing Jira MCP server.

### Phase 6: Observability

Planned features:

- Persistent invocation audit records
- Request IDs
- Structured invocation logs
- Prometheus metrics
- OpenTelemetry traces
- Server health status
- Tool success and error rates
- p50, p95, and p99 latency
- Tool usage dashboard

### Phase 7: React UI

Planned screens:

- MCP server discovery catalog
- Server details
- Tool catalog
- Tool input-schema viewer
- JSON invocation sandbox
- Confirmation dialog for medium/high-risk tools
- Invocation history
- Admin management
- Observability dashboard

### Phase 8: Delivery and Polish

Planned improvements:

- GitHub Actions CI
- Backend Dockerfile
- Frontend Dockerfile
- Integration tests using Testcontainers
- OpenAPI documentation
- Cloud deployment
- Architecture decision records
- Screenshots and demo video
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

Check the connection string:

```bash
set -a
source .env
set +a

echo "$DATABASE_URL"
```

Test directly:

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

### Login returns `401 invalid_credentials`

Confirm users exist:

```bash
set -a
source .env
set +a

psql "$DATABASE_URL" \
  -c "SELECT email, role, active FROM users ORDER BY email;"
```

If no users exist, regenerate bcrypt hashes and run the local seed file again.

### Authenticated request returns `401 invalid_token`

Possible causes:

- Token has expired.
- `JWT_SECRET` changed after the token was issued.
- `JWT_ISSUER` changed after the token was issued.
- The header does not use `Authorization: Bearer <token>` format.

Login again to receive a fresh token:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mcp-gateway.local",
    "password": "AdminPass123!"
  }' | jq
```

### Request returns `403 forbidden`

Your token is valid, but the assigned role cannot perform the operation.

Use an admin token for catalog mutations:

```bash
curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Reset local server and tool records

To remove all tool and server entries while keeping tables, migrations, and users:

```bash
set -a
source .env
set +a

psql "$DATABASE_URL" \
  -c "TRUNCATE TABLE mcp_tools, mcp_servers CASCADE;"
```

Verify:

```bash
psql "$DATABASE_URL" \
  -c "SELECT COUNT(*) AS server_count FROM mcp_servers;"

psql "$DATABASE_URL" \
  -c "SELECT COUNT(*) AS tool_count FROM mcp_tools;"
```

### Reset the entire local database

This removes all local data, including users, servers, tools, and migration state:

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

Then seed development users again.

### `.env` does not load

Run the backend from `backend/`:

```bash
cd backend
go run ./cmd/api
```

### Port 8080 is already in use

```bash
lsof -i :8080
```

Stop the process or update:

```dotenv
HTTP_PORT=8081
```

Then restart the backend.

### Tests fail after import changes

All internal imports must use:

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
git checkout -b feat/invocation-gateway

cd backend
gofmt -w .
go mod tidy
go test ./...
go vet ./...

cd ..
git add .
git commit -m "feat: add policy-controlled tool invocation"
git push -u origin feat/invocation-gateway
```

Suggested commit convention:

```text
feat: add MCP server registry CRUD API
feat: add MCP tool catalog
feat: add JWT authentication and RBAC
feat: add tool invocation audit records
fix: validate tool input schema
test: cover invalid JWT and role checks
docs: document phase 3 authentication
chore: configure Docker Compose PostgreSQL
```

---

## Project Status

```text
Phase 0: Complete
Phase 1: Complete
Phase 2: Complete
Phase 3: Complete
Phase 4: Ready to begin
```

The next milestone is the **Invocation Gateway**: a policy-controlled, audited path that allows an authenticated caller to invoke a registered low-risk tool through the MCP Gateway.