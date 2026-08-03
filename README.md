# MCP Gateway

A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations.

The project is inspired by the idea of an internal "USB-C for AI agents": a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history with full observability and a modern web interface.

> **Current status:** Phase 7 complete — the gateway now has a full-featured React UI for server discovery, tool invocation, and observability.

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
- [Phase 7: React UI](#phase-7-react-ui)
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
                    │ Catalog • Sandbox • History  │
                    │ Metrics • Admin (Vite+TS)    │
                    └──────────────┬──────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go MCP Gateway                            │
│                                                                 │
│ Server Registry • Tool Catalog • JWT Auth • RBAC               │
│ Invocation Gateway • JSON Schema Validation • Audit Records    │
│ Live GitHub Executor • Mock Executor • Upstream Error Capture  │
│ Prometheus Metrics • OpenTelemetry Tracing • History API       │
│ CORS Middleware • Structured Logging                           │
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
      │ Users • Servers • Tools • Invocation Audit Records       │
      └─────────────────────────────────────────────────────────┘
```

---

## Why This Project

This project demonstrates practical full-stack and AI-platform engineering skills rather than only building a standalone LLM application.

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
- Observability: Prometheus metrics, OpenTelemetry tracing, invocation history API
- **Modern React UI with Vite, TypeScript, and Tailwind CSS v4**
- **Component-based architecture with React Query for data fetching**
- **Protected routes and authentication context**
- Secure password handling
- Structured request logging with status codes
- CORS configuration for local development
- Containerized local development
- Production-style testing and deployment practices

The long-term architecture separates deterministic infrastructure from agentic reasoning.

```text
React UI (Vite + TypeScript)
├── Login page (JWT authentication)
├── Server catalog (list, search, filter)
├── Server detail page (tools list)
├── Tool detail page (schema viewer, invoke sandbox)
├── Invocation history (table, filters, export)
├── Metrics dashboard (Prometheus data visualization)
└── Admin panel (user management, server/tool CRUD)

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
├── CORS middleware
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
Developer / API Client / React UI
         │
         │ HTTP / JSON / CORS
         ▼
┌──────────────────────────────────────┐
│ Go MCP Gateway                       │
│                                      │
│ chi Router                           │
│ CORS middleware                      │
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

### React UI architecture

```text
React App (Vite + TypeScript)
        │
        ├── Authentication Context
        │   ├── JWT token management
        │   ├── User state
        │   └── Protected routes
        │
        ├── React Query
        │   ├── Server catalog queries
        │   ├── Tool detail queries
        │   ├── Invocation mutations
        │   └── History queries
        │
        ├── API Client (Axios)
        │   ├── Auth interceptor
        │   ├── Error handling
        │   └── 401 redirect
        │
        └── Pages
            ├── Login
            ├── Servers (list, detail)
            ├── Tools (detail, invoke sandbox)
            ├── Invocations (history, filters)
            └── Admin (future)
```

### Invocation request flow with observability

```text
POST /api/v1/servers/{serverID}/tools/{toolID}/invoke
        │
        ▼
CORS preflight (OPTIONS) → 200 OK
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
        │
        ▼
React UI displays result in sandbox
```

---

## Technology Stack

| Area | Technology | Purpose |
|---|---|---|
| **Frontend framework** | React 18 + TypeScript | Component-based UI |
| **Build tool** | Vite | Fast dev server, HMR, optimized builds |
| **Styling** | Tailwind CSS v4 | Utility-first CSS with Oxide engine |
| **Routing** | React Router v6 | Client-side routing |
| **Data fetching** | TanStack React Query | Caching, synchronization, mutations |
| **HTTP client** | Axios | Interceptors, error handling |
| **Icons** | Lucide React | Consistent icon library |
| **Date formatting** | date-fns | Lightweight date utilities |
| Backend language | Go | Concurrent, strongly typed gateway implementation |
| HTTP routing | `go-chi/chi` | Lightweight REST routing and middleware |
| CORS | `go-chi/cors` | Cross-origin resource sharing |
| Database | PostgreSQL 16 | Users, servers, tools, invocation audits |
| PostgreSQL driver | `pgx/v5` | Native PostgreSQL driver and connection pool |
| Migrations | `golang-migrate` | Version-controlled schema migrations |
| JSON storage | PostgreSQL `jsonb` | Tool input schemas and invocation payloads |
| Schema validation | `santhosh-tekuri/jsonschema/v6` | JSON Schema draft 2020-12 argument validation |
| Authentication | JWT with HS256 | Local development access tokens |
| JWT library | `github.com/golang-jwt/jwt/v5` | JWT signing, parsing, verification, claims validation |
| Password hashing | `golang.org/x/crypto/bcrypt` | Password hash generation and comparison |
| GitHub integration | `google/go-github/v62` | Typed Go client for GitHub REST API v3 |
| Metrics | `prometheus/client_golang` | Prometheus metrics collection and exposition |
| Tracing | `go.opentelemetry.io/otel` | OpenTelemetry distributed tracing |
| Configuration | Environment variables + `godotenv` | Local configuration and secrets loading |
| Logging | Go `log/slog` + chi `WrapResponseWriter` | Structured logs with request ID, status, bytes, latency |
| API testing | `curl`, `jq`, Go `testing` | Manual API verification and automated tests |
| Containerization | Docker Compose | Local PostgreSQL environment |
| Observability UI | Grafana | Future metrics and trace visualization |
| MCP integration | Official Go MCP SDK | Future real MCP discovery and invocation |

---

## Current Features

### Completed

#### Backend (Phases 0-6)

- Go module at `github.com/iashu2k/mcp-gateway/backend`
- Docker Compose-managed PostgreSQL 16 database
- Environment-based configuration
- PostgreSQL connection pool using `pgxpool`
- Health endpoint with database connectivity validation
- Structured JSON application logs
- Request logging with request ID, method, path, HTTP status, bytes written, and duration
- HTTP request ID and timeout middleware
- **CORS middleware for React UI integration**
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

#### Frontend (Phase 7)

- **Vite + React 18 + TypeScript project**
- **Tailwind CSS v4 with Oxide engine** (no config file needed)
- **JWT authentication with login page**
  - Email/password form with validation
  - Error handling for invalid credentials
  - Automatic redirect to servers page on success
  - Token stored in localStorage
- **Protected routes with authentication context**
  - `useAuth` hook for user state and token management
  - Automatic user loading on app start
  - 401 interceptor redirects to login
  - Role-based navigation (admin sees Admin link)
- **Server catalog with list and detail pages**
  - Grid layout with server cards
  - Status badges (active/inactive/unhealthy)
  - Owner team display
  - Empty state for no servers
  - Server detail page with metadata and tools list
- **Tool detail page with input schema viewer**
  - Tool metadata display (title, description, risk level, enabled status)
  - JSON schema viewer with syntax highlighting
  - Risk level color coding (green/yellow/red)
- **Invoke sandbox with JSON editor and live results**
  - JSON textarea for arguments input
  - Validation for invalid JSON
  - Invoke button with loading state
  - Result display with formatted JSON
  - Error handling for invocation failures
- **Invocation history page with filters**
  - Table view with status icons (checkmark/X/clock)
  - Filter by status (all, succeeded, failed, running)
  - Pagination control (25/50/100 per page)
  - Duration display in milliseconds
  - Formatted timestamps
  - Empty state for no invocations
- **React Query for data fetching and caching**
  - Automatic refetching on window focus
  - Optimistic updates for mutations
  - Loading and error states
  - Query invalidation after mutations
- **Responsive layout with Tailwind**
  - Navigation bar with logo and links
  - User info display with role badge
  - Logout button
  - Mobile-friendly design
- **API client with auth token interceptor**
  - Automatic Bearer token injection
  - 401 error handling with redirect to login
  - Base URL configuration via environment variable
- **Error handling and loading states**
  - Toast-style error messages
  - Loading spinners for async operations
  - Empty states for no data

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
- **Admin panel UI (server/tool CRUD forms, user management)**
- **Metrics dashboard UI with charts (Prometheus data visualization)**
- **Password change UI**
- **Server/tool search and advanced filtering in UI**
- **Invocation detail modal in history page**
- **Export invocation history to CSV**
- CI/CD pipeline, production deployment

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
| Phase 7 | Complete | React UI | Vite + React + TypeScript UI with auth, catalog, sandbox, history |
| Phase 8 | Next | Delivery and Polish | Admin UI, metrics dashboard, CI, containers, deployment |

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
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   └── services.ts
│   │   │
│   │   ├── components/
│   │   │   └── Layout.tsx
│   │   │
│   │   ├── hooks/
│   │   │   └── useAuth.tsx
│   │   │
│   │   ├── pages/
│   │   │   ├── login/
│   │   │   │   └── LoginPage.tsx
│   │   │   ├── servers/
│   │   │   │   ├── ServersPage.tsx
│   │   │   │   └── ServerDetailPage.tsx
│   │   │   ├── tools/
│   │   │   │   └── ToolDetailPage.tsx
│   │   │   ├── invocations/
│   │   │   │   └── InvocationsPage.tsx
│   │   │   └── admin/ (future)
│   │   │
│   │   ├── types/
│   │   │   └── index.ts
│   │   │
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   ├── index.css
│   │   └── vite-env.d.ts
│   │
│   ├── public/
│   ├── .env
│   ├── .env.example
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── README.md
│
├── docs/
├── infra/
├── scripts/
│   └── seed_users.sql
├── .env.example
├── .gitignore
├── docker-compose.yml
└── README.md
```

### Package responsibilities

#### Backend

| Package | Responsibility |
|---|---|
| `cmd/api` | Application startup, dependency wiring, configuration loading, graceful shutdown |
| `cmd/passwordhash` | Local utility to generate bcrypt hashes for development user seeding |
| `internal/auth` | JWT claims, signing, parsing, signature verification, expiration validation |
| `internal/config` | Environment-variable loading and validation |
| `internal/domain` | Core domain models, request models, constants, response structures |
| `internal/executor` | Tool execution implementations: mock (deterministic) and GitHub (live REST) |
| `internal/httpapi` | Routes, handlers, middleware, JSON decoding, responses, HTTP error mapping, CORS |
| `internal/observability` | Prometheus metrics and OpenTelemetry tracing |
| `internal/platform/database` | PostgreSQL connection-pool setup and health verification |
| `internal/repository` | Parameterized SQL and PostgreSQL persistence operations |
| `internal/service` | Business rules, validation, authentication, schema validation, invocation orchestration, history queries |
| `migrations` | Ordered and versioned database schema evolution |

#### Frontend

| Directory | Responsibility |
|---|---|
| `src/api` | Axios client configuration and API service functions |
| `src/components` | Reusable UI components (Layout, etc.) |
| `src/hooks` | Custom React hooks (useAuth, etc.) |
| `src/pages` | Page-level components organized by feature |
| `src/types` | TypeScript type definitions |
| `src/utils` | Utility functions (future) |

Dependency direction:

```text
React UI
   ↓
HTTP Handler / Middleware / CORS
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
- Node.js 18 or newer
- npm or yarn
- Docker Desktop or Docker Engine with Docker Compose
- Git
- `curl`
- `jq`
- PostgreSQL client tools, including `psql`
- `golang-migrate`

Verify:

```bash
go version
node --version
npm --version
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

### Backend setup

#### Create local configuration

```bash
cp .env.example .env
```

#### Configure JWT settings

```bash
openssl rand -base64 48
```

Set the generated value in `.env`:

```dotenv
JWT_SECRET=PASTE_A_LONG_RANDOM_VALUE_HERE
JWT_ISSUER=mcp-gateway
JWT_TTL_MINUTES=60
```

#### Configure GitHub integration (optional)

For public repositories, you can run unauthenticated (60 requests/hour). For higher rate limits or private repos, create a fine-grained personal access token with only "Issues: Read" permission:

```dotenv
GITHUB_TOKEN=github_pat_your_token_here
```

Leave empty for unauthenticated public access:

```dotenv
GITHUB_TOKEN=
```

#### Start PostgreSQL

```bash
docker compose up -d postgres
docker compose ps
```

#### Apply migrations

```bash
set -a
source .env
set +a

migrate -path backend/migrations -database "$DATABASE_URL" up
migrate -path backend/migrations -database "$DATABASE_URL" version
```

Expected after Phase 7:

```text
4
```

#### Seed local users

```bash
cd backend
go run ./cmd/passwordhash "AdminPass123"
go run ./cmd/passwordhash "DeveloperPass123"
go run ./cmd/passwordhash "ViewerPass123"
cd ..

# Create seed file with the generated hashes
cat > scripts/seed_users.sql << 'EOF'
INSERT INTO users (id, email, display_name, password_hash, role, active)
VALUES
  (
    'f8f3ace3-3c02-4ca3-86c0-11517ae1bee3',
    'admin@mcp-gateway.local',
    'Gateway Admin',
    '$2a$10$GS/jgCP2fk9wmiD6x47io.rl79UHog1/5fP1UHhfa5QiipaHFTo6a',
    'admin',
    true
  ),
  (
    '265d8f73-0221-45c5-9890-dcaa1dfc62ee',
    'developer@mcp-gateway.local',
    'Gateway Developer',
    '$2a$10$I0osJrIi/SYP2v16tcyC1etDQjFRUmcCSqNn5T0vTj7TDB97WVGGG',
    'developer',
    true
  ),
  (
    '28a6ffda-9a76-4a7d-90fd-8aa1670b1660',
    'viewer@mcp-gateway.local',
    'Gateway Viewer',
    '$2a$10$gzFskJR2pzBNWWyE5a4ghukZRc.M36I52Mmn4bpjrVGjr6Vf62vCC',
    'viewer',
    true
  )
ON CONFLICT (email) DO UPDATE
SET password_hash = EXCLUDED.password_hash;
EOF

psql "$DATABASE_URL" -f scripts/seed_users.sql
```

Local development users:

| Email | Password | Role |
|---|---|---|
| `admin@mcp-gateway.local` | `AdminPass123` | `admin` |
| `developer@mcp-gateway.local` | `DeveloperPass123` | `developer` |
| `viewer@mcp-gateway.local` | `ViewerPass123` | `viewer` |

> Local development credentials only. Never use them in any deployed environment.

### Frontend setup

#### Install dependencies

```bash
cd frontend
npm install
```

#### Create environment config

```bash
cp .env.example .env
```

Edit `frontend/.env`:

```
VITE_API_BASE_URL=http://localhost:8080
```

### Start the backend

From the repository root:

```bash
go run ./backend/cmd/api
```

Or from `backend/`:

```bash
cd backend
go run ./cmd/api
```

The API runs at `http://localhost:8080`.

### Start the frontend

In a separate terminal:

```bash
cd frontend
npm run dev
```

The UI runs at `http://localhost:5173`.

---

## Environment Variables

### Backend

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

### Frontend

```
VITE_API_BASE_URL=http://localhost:8080
```

| Variable | Required | Description |
|---|---:|---|
| `VITE_API_BASE_URL` | Yes | Base URL for the Go API |

`.env` files in both `backend/` and `frontend/` are local-only and must not be committed.

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

### Login (API)

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mcp-gateway.local",
    "password": "AdminPass123"
  }' | jq
```

### Login (UI)

1. Navigate to `http://localhost:5173`
2. You'll be redirected to `/login`
3. Enter `admin@mcp-gateway.local` / `AdminPass123`
4. Click "Sign in"
5. You'll be redirected to the servers page

### Save tokens (API)

```bash
export ADMIN_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@mcp-gateway.local","password":"AdminPass123"}' \
    | jq -r '.accessToken'
)"

export DEVELOPER_TOKEN="$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"developer@mcp-gateway.local","password":"DeveloperPass123"}' \
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

Path normalization replaces UUIDs with `:id` to avoid high cardinality.

### Invocation history API

Query audit records via REST endpoints:

```http
GET /api/v1/invocations?serverId={uuid}&toolId={uuid}&status=succeeded&limit=50&offset=0
Authorization: Bearer <access-token>
```

**Role-based access:**

- **Admin**: Sees all invocations from all users
- **Developer**: Sees only their own invocations
- **Viewer**: No access to history endpoints

### OpenTelemetry tracing

The gateway initializes an OpenTelemetry tracer provider with a stdout exporter for local development. In production, this can be swapped for an OTLP exporter to send traces to Jaeger, Tempo, or other backends.

---

## Phase 7: React UI

**Status:** Complete

### Goal

Build a modern, responsive web interface for discovering servers, browsing tools, invoking them in a sandbox, and viewing invocation history. The UI uses Vite for fast development, React for component-based architecture, TypeScript for type safety, and Tailwind CSS v4 for styling.

### Technology choices

**Vite:**
- Lightning-fast HMR (Hot Module Replacement)
- Optimized production builds with Rollup
- Native ES modules support
- Built-in TypeScript support
- Plugin ecosystem

**React 18 + TypeScript:**
- Component-based architecture
- Type safety for props, state, and API responses
- Rich ecosystem (React Router, React Query, etc.)
- Excellent developer experience

**Tailwind CSS v4:**
- Utility-first CSS framework
- Oxide engine (Rust-based) for faster builds
- No config file needed (auto-detects content)
- Smaller bundle size with automatic tree-shaking
- New `@import "tailwindcss"` syntax

**React Query (TanStack Query):**
- Automatic caching and synchronization
- Background refetching
- Optimistic updates
- Loading and error states
- Query invalidation

**React Router v6:**
- Declarative routing
- Protected routes with authentication
- Nested routes with layouts
- URL parameter support

### Architecture

```text
React App
├── Authentication Context
│   ├── JWT token management (localStorage)
│   ├── User state (role, email, displayName)
│   ├── Protected routes wrapper
│   └── 401 interceptor → redirect to login
│
├── API Client (Axios)
│   ├── Base URL from environment variable
│   ├── Auth token interceptor (adds Bearer token)
│   ├── Error interceptor (handles 401)
│   └── Type-safe service functions
│
├── React Query
│   ├── QueryClient with default options
│   ├── Queries: servers, tools, invocations
│   ├── Mutations: login, invoke tool
│   └── Automatic refetching and caching
│
├── Layout Component
│   ├── Navigation bar with logo
│   ├── Links: Servers, History, Admin (role-based)
│   ├── User info display
│   └── Logout button
│
└── Pages
    ├── LoginPage
    │   ├── Email/password form
    │   ├── Error display
    │   └── Loading state
    │
    ├── ServersPage
    │   ├── Grid layout with server cards
    │   ├── Status badges
    │   ├── Empty state
    │   └── "Add Server" button (admin only)
    │
    ├── ServerDetailPage
    │   ├── Server metadata display
    │   ├── Tools list with risk level badges
    │   ├── Enabled/disabled status
    │   └── Back navigation
    │
    ├── ToolDetailPage
    │   ├── Tool metadata display
    │   ├── JSON schema viewer
    │   ├── Invoke sandbox
    │   │   ├── JSON textarea for arguments
    │   │   ├── Invoke button with loading state
    │   │   └── Result display
    │   └── Error handling
    │
    └── InvocationsPage
        ├── Table with status icons
        ├── Filter by status
        ├── Pagination control
        ├── Duration display
        └── Formatted timestamps
```

### Key features

**Authentication flow:**
1. User visits `http://localhost:5173`
2. Redirected to `/login` if not authenticated
3. Enters email/password
4. API returns JWT token
5. Token stored in localStorage
6. Redirected to `/servers`
7. All subsequent requests include Bearer token
8. On 401 error, redirected back to `/login`

**Server discovery:**
- Grid layout with responsive cards
- Server name, description, status, owner team
- Click to view server detail page
- Admin sees "Add Server" button

**Tool invocation:**
- Tool detail page shows metadata and input schema
- JSON textarea for entering arguments
- "Invoke Tool" button triggers mutation
- Result displayed in formatted JSON
- Loading state during invocation
- Error handling for failures

**Invocation history:**
- Table view with all invocations
- Status icons (checkmark for success, X for failure, clock for running)
- Filter by status dropdown
- Pagination control
- Duration in milliseconds
- Formatted timestamps

**CORS configuration:**
- Backend allows requests from `http://localhost:5173`
- Preflight OPTIONS requests return 200 OK
- Credentials included in requests

### Verified user flows

**1. Login:**
```
Navigate to http://localhost:5173
→ Redirected to /login
→ Enter admin@mcp-gateway.local / AdminPass123
→ Click "Sign in"
→ Redirected to /servers
```

**2. Browse servers:**
```
View server cards in grid layout
→ See GitHub and mock-tools servers
→ Click on GitHub server
→ View server detail page with tools list
```

**3. Invoke tool:**
```
Click on "List GitHub Issues" tool
→ View tool detail page
→ Enter JSON arguments: {"owner": "golang", "repo": "go", "per_page": 3}
→ Click "Invoke Tool"
→ See loading state
→ View result with real GitHub issues
```

**4. View history:**
```
Click "History" in navigation
→ View invocation history table
→ Filter by "succeeded" status
→ See duration and timestamps
→ Verify role-based access (developer sees only own invocations)
```

### Phase 7 acceptance criteria

- [x] Vite + React + TypeScript project initialized
- [x] Tailwind CSS v4 configured (no config file needed)
- [x] JWT authentication with login page
- [x] Protected routes with auth context
- [x] Server catalog with list and detail pages
- [x] Tool detail page with input schema viewer
- [x] Invoke sandbox with JSON editor and live results
- [x] Invocation history page with filters
- [x] React Query for data fetching and caching
- [x] Responsive layout with Tailwind
- [x] API client with auth token interceptor
- [x] Error handling and loading states
- [x] CORS middleware on backend
- [x] Role-based navigation (admin sees Admin link)
- [x] Empty states for no data
- [x] Formatted timestamps with date-fns
- [x] Icons with Lucide React

---

## Validation and Testing

### Backend

```bash
cd backend

gofmt -w .
go mod tidy
go test ./...
go vet ./...
```

### Frontend

```bash
cd frontend

npm run build
npm run lint
```

### Integration test

1. Start backend: `go run ./backend/cmd/api`
2. Start frontend: `cd frontend && npm run dev`
3. Navigate to `http://localhost:5173`
4. Login with `admin@mcp-gateway.local` / `AdminPass123`
5. Browse servers
6. Invoke a tool
7. View invocation history

---

## Design Decisions

### Why Vite over Create React App

Vite provides:

- **Faster dev server startup** — Native ES modules, no bundling in dev
- **Instant HMR** — Changes reflect immediately without full reload
- **Optimized builds** — Rollup-based production builds with code splitting
- **Modern tooling** — Built-in TypeScript, JSX, CSS support
- **Smaller bundle size** — Tree-shaking and minification out of the box

Create React App is deprecated and slower for modern development.

### Why Tailwind CSS v4

Tailwind v4 offers:

- **No config file** — Auto-detects content files
- **Faster builds** — Oxide engine (Rust-based)
- **Smaller bundle** — Automatic tree-shaking
- **Simpler setup** — Just `@import "tailwindcss"` in CSS
- **Better DX** — Fewer configuration files to manage

The utility-first approach keeps styling co-located with components.

### Why React Query

React Query solves common data fetching problems:

- **Caching** — Automatic caching with configurable stale time
- **Background refetching** — Keeps data fresh automatically
- **Optimistic updates** — UI updates immediately, rolls back on error
- **Loading/error states** — Built-in state management
- **Query invalidation** — Refetch related queries after mutations

Without React Query, you'd need to manage caching, loading states, and synchronization manually.

### Why CORS over Vite proxy

CORS middleware is more explicit and works in production:

- **Production-ready** — Works when frontend and backend are deployed separately
- **Explicit configuration** — Clear which origins are allowed
- **Standard approach** — Industry-standard for cross-origin requests

Vite proxy is convenient for local development but doesn't solve production CORS.

### Why localStorage for JWT token

localStorage is simple and works for this use case:

- **Simple API** — `localStorage.getItem/setItem`
- **Persistent** — Survives page refreshes
- **Synchronous** — No async complexity

For production with higher security requirements, consider httpOnly cookies or session storage.

---

## Known Limitations

Intentional at the end of Phase 7:

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
- **Admin panel UI not implemented (server/tool CRUD forms)**
- **Metrics dashboard UI not implemented (Prometheus data visualization)**
- **No password change UI**
- **No server/tool search in UI**
- **No invocation detail modal in history page**
- **No export to CSV in history page**
- **JWT token stored in localStorage (vulnerable to XSS)**
- No CI/CD pipeline, production deployment
- No MCP protocol transport (stdio/SSE/streamable HTTP); all integrations are direct REST API calls

---

## Planned Phases

### Phase 8: Delivery and Polish (Next)

Complete the production-ready features:

- **Admin panel UI**
  - Server CRUD forms (create, edit, delete)
  - Tool CRUD forms with JSON schema editor
  - User management (list, create, edit roles)
  - Password change UI

- **Metrics dashboard UI**
  - Charts for invocation counts by server/tool
  - Latency histograms
  - Error rate trends
  - Database connection pool metrics
  - Integration with Prometheus data

- **Enhanced history page**
  - Invocation detail modal with full request/response
  - Export to CSV
  - Advanced filtering (date range, user, server, tool)
  - Pagination with page numbers

- **Search and filtering**
  - Server catalog search by name/description
  - Tool search within server detail page
  - Filter by risk level, enabled status

- **CI/CD pipeline**
  - GitHub Actions workflow
  - Automated tests (backend + frontend)
  - Linting and formatting checks
  - Docker image builds
  - Deployment to staging/production

- **Docker images**
  - Backend Dockerfile (multi-stage build)
  - Frontend Dockerfile (nginx for static files)
  - Docker Compose for full stack
  - Production-ready configuration

- **Documentation**
  - OpenAPI/Swagger spec for API
  - Architecture decision records
  - Deployment guide
  - Security best practices
  - Demo video and screenshots

- **Security hardening**
  - Move JWT token to httpOnly cookie
  - Add CSRF protection
  - Rate limiting on API endpoints
  - Input sanitization
  - Security headers (CSP, X-Frame-Options, etc.)

- **Production deployment**
  - Cloud platform setup (AWS/GCP/Azure)
  - Environment-specific configuration
  - Secrets management
  - Monitoring and alerting
  - Backup and disaster recovery

---

## Troubleshooting

### CORS errors in browser console

**Symptom:** `Access to fetch at 'http://localhost:8080/api/v1/...' from origin 'http://localhost:5173' has been blocked by CORS policy`

**Fix:** Ensure CORS middleware is added to `router.go` **before** other middleware:

```go
router.Use(cors.Handler(cors.Options{
	AllowedOrigins:   []string{"http://localhost:5173"},
	AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
	AllowCredentials: true,
}))
```

### Login returns 401 invalid_credentials

**Symptom:** `{"error":"invalid_credentials","message":"email or password is incorrect"}`

**Fix:** Verify the password hash in the database matches the password:

```bash
# Generate new hash
cd backend
go run ./cmd/passwordhash "AdminPass123"

# Update database (escape $ with \$)
psql "$DATABASE_URL" -c "
  UPDATE users
  SET password_hash = '\$2a\$10\$GS/jgCP2fk9wmiD6x47io.rl79UHog1/5fP1UHhfa5QiipaHFTo6a'
  WHERE email = 'admin@mcp-gateway.local';
"
```

### Frontend can't connect to backend

**Symptom:** Network error or `ERR_CONNECTION_REFUSED`

**Fix:** Ensure backend is running and `VITE_API_BASE_URL` is correct:

```bash
# Check backend is running
curl http://localhost:8080/health

# Check frontend .env
cat frontend/.env
# Should contain: VITE_API_BASE_URL=http://localhost:8080
```

### Tailwind classes not working

**Symptom:** No styling applied, components look unstyled

**Fix:** Ensure `@import "tailwindcss"` is in `src/index.css` and the file is imported in `main.tsx`:

```typescript
// src/main.tsx
import './index.css'
```

### 401 redirect loop

**Symptom:** Constantly redirected to login page

**Fix:** Clear localStorage and login again:

```javascript
// In browser console
localStorage.clear()
```

Then login again with correct credentials.

---

## Contributing Workflow

```bash
git checkout -b feat/admin-panel

# Backend changes
cd backend
gofmt -w .
go mod tidy
go test ./...
go vet ./...

# Frontend changes
cd ../frontend
npm run lint
npm run build

cd ..
git add .
git commit -m "feat: add admin panel for server and tool management"
git push -u origin feat/admin-panel
```

Suggested commit convention:

```text
feat: add react ui for server catalog
feat: add admin panel for server crud
feat: add metrics dashboard with charts
fix: validate tool input schema
test: cover invalid JWT and role checks
docs: document phase 7 react ui
chore: configure vite and tailwind
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
Phase 7: Complete
Phase 8: Ready to begin — Delivery and Polish
```

The next milestone is **Delivery and Polish**: admin panel UI, metrics dashboard, CI/CD pipeline, Docker images, and production deployment.

---

## Demo: Full Stack Walkthrough

Here's a complete end-to-end walkthrough of the system:

### 1. Start the stack

```bash
# Terminal 1: PostgreSQL
docker compose up -d postgres

# Terminal 2: Backend
cd backend
go run ./cmd/api

# Terminal 3: Frontend
cd frontend
npm run dev
```

### 2. Login via UI

1. Navigate to `http://localhost:5173`
2. Redirected to `/login`
3. Enter `admin@mcp-gateway.local` / `AdminPass123`
4. Click "Sign in"
5. Redirected to `/servers`

### 3. Browse servers

1. See grid of server cards (GitHub, mock-tools)
2. Click on "github" server
3. View server detail page with tools list
4. See "List GitHub Issues" and "Search Repositories" tools

### 4. Invoke a tool

1. Click on "List GitHub Issues"
2. View tool detail page with input schema
3. Enter arguments in JSON editor:
   ```json
   {
     "owner": "golang",
     "repo": "go",
     "state": "open",
     "per_page": 3
   }
   ```
4. Click "Invoke Tool"
5. See loading state
6. View result with real GitHub issues

### 5. View invocation history

1. Click "History" in navigation
2. See table with all invocations
3. Filter by "succeeded" status
4. See duration (e.g., 404ms)
5. See formatted timestamp

### 6. Check metrics

```bash
curl -s http://localhost:8080/metrics | grep mcp_gateway_invocations_total
```

Output:

```text
mcp_gateway_invocations_total{server="github",tool="list_issues",status="succeeded"} 1
```

### 7. Verify audit trail

```bash
psql "$DATABASE_URL" -c "
  SELECT 
    ti.status,
    s.name AS server,
    t.name AS tool,
    ti.duration_ms,
    ti.created_at
  FROM tool_invocations ti
  JOIN mcp_servers s ON s.id = ti.server_id
  JOIN mcp_tools t ON t.id = ti.tool_id
  ORDER BY ti.created_at DESC
  LIMIT 5;
"
```

This demonstrates the complete flow: UI → API → policy → execution → audit → metrics — all working end-to-end with a modern React interface.
