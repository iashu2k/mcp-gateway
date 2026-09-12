# MCP Gateway

> 🚀 A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations — complete with a Go backend, React UI, native MCP protocol support, and full Docker deployment.

The project is inspired by the idea of an internal "USB-C for AI agents": a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history with full observability.

> ✅ **Current status:** Phase 9 complete — the gateway speaks native MCP in both directions. AI agents connect over Streamable HTTP, discover governed tools, and invoke them through the full policy chain; outbound MCP upstreams are auto-discovered and executed live. Production-ready with admin panel, metrics dashboard, Docker deployment, and CI.

---

## 📑 Table of Contents

- [Project Vision](#-project-vision)
- [Why This Project](#-why-this-project)
- [System Architecture](#-system-architecture)
- [MCP Transport: How It Works](#-mcp-transport-how-it-works)
- [Technology Stack](#-technology-stack)
- [Current Features](#-current-features)
- [Development Roadmap](#-development-roadmap)
- [Repository Structure](#-repository-structure)
- [Prerequisites](#-prerequisites)
- [Quick Start with Docker](#-quick-start-with-docker)
- [Local Development Setup](#-local-development-setup)
- [Environment Variables](#-environment-variables)
- [Database Migrations](#-database-migrations)
- [Authentication and Roles](#-authentication-and-roles)
- [API Reference](#-api-reference)
- [Phase Summaries](#-phase-summaries)
- [Validation and Testing](#-validation-and-testing)
- [Design Decisions](#-design-decisions)
- [Known Limitations](#-known-limitations)
- [Future Enhancements](#-future-enhancements)
- [Troubleshooting](#-troubleshooting)
- [Contributing Workflow](#-contributing-workflow)

---

## 🎯 Project Vision

Internal organizations increasingly expose capabilities through APIs, automations, and MCP servers: GitHub repository operations, Jira issue management, Confluence search, Slack messaging, deployment workflows, analytics tools, and more.

Without a centralized gateway, teams often face several problems:

- 🔍 Developers and AI agents cannot easily discover which tools exist
- 🔑 API credentials may be distributed across scripts, applications, and local environments
- 🛡️ Access control is inconsistent across tools
- 📝 Tool invocations are difficult to audit
- 📊 Teams cannot easily understand tool reliability, latency, errors, or usage
- ⚠️ Mutating operations can be invoked without sufficiently clear policy controls

MCP Gateway addresses this by acting as a secure control plane for internal MCP servers.

```text
                    ┌─────────────────────────────┐
                    │       React Web UI          │
                    │ Catalog • Sandbox • History │
                    │ Metrics • Admin Panel       │
                    └──────────────┬──────────────┘
                                   │
        AI Agents / MCP Clients    │    (Claude Desktop, MCP Inspector,
                    │              │     agent frameworks)
                    │ MCP          │
                    ▼              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go MCP Gateway                           │
│                                                                 │
│ /mcp Streamable HTTP endpoint • Server Registry • Tool Catalog  │
│ JWT Auth • RBAC • Invocation Gateway • JSON Schema Validation   │
│ Audit Records • Prometheus Metrics • OTel Tracing • History API │
└───────┬──────────────┬──────────────────┬───────────────────────┘
        │              │                  │
        ▼              ▼                  ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────────┐
│ GitHub REST   │ │ Live MCP      │ │ Jira / Slack MCP  │
│ API (live)    │ │ upstreams     │ │ servers (mock)    │
└───────────────┘ │ (auto-tools)  │ └───────────────────┘
                  └───────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────┐
│ PostgreSQL                                              │
│ Users • Servers • Tools • Invocation Audit Records      │
└─────────────────────────────────────────────────────────┘
```

---

## 💡 Why This Project

This project demonstrates practical full-stack and AI-platform engineering skills rather than only building a standalone LLM application.

It focuses on:

- Go backend development and service design
- Native MCP protocol implementation, both server-side and client-side (official `modelcontextprotocol/go-sdk`)
- PostgreSQL schema design and migrations
- API gateway patterns and JSON Schema-driven validation
- JWT authentication and role-based access control
- Policy-controlled tool invocation with durable audit records
- Live third-party integrations (GitHub REST, live MCP upstreams)
- Observability: Prometheus metrics, OpenTelemetry tracing, history API
- Modern React UI with Vite, TypeScript, and Tailwind CSS v4
- Admin panel with full CRUD for servers and tools
- Metrics dashboard with live Prometheus data visualization
- Docker multi-stage builds for backend and frontend
- CI pipeline with automated testing, linting, and image builds
- Production-style engineering practices throughout

---

## 🏗️ System Architecture

### Current architecture

```text
Developer / API Client / React UI / AI Agents (MCP)
         │                              │
         │ HTTP / JSON / CORS           │ MCP over Streamable HTTP
         ▼                              ▼
┌────────────────────────────────────────────────────┐
│ Go MCP Gateway                                     │
│                                                    │
│ chi Router • CORS • Request ID                     │
│ Status-aware logging • Scoped timeouts             │
│ JWT auth • RBAC middleware • WWW-Authenticate      │
│ Health + Metrics endpoints                         │
│ Server Registry • Tool Catalog                     │
│ Invocation API • History API                       │
│ Password change endpoint                           │
│ JSON Schema argument validation                    │
│                                                    │
│ MCP transport (Phase 9)                            │
│ └── /mcp Streamable HTTP endpoint                  │
│     (stateless, catalog-driven, namespaced tools)  │
│                                                    │
│ Observability                                      │
│ ├── Prometheus metrics                             │
│ ├── OpenTelemetry tracing                          │
│ └── Structured logging                             │
│                                                    │
│ Executor Router                                    │
│ ├── MockExecutor (in-process)                      │
│ ├── GitHubExecutor (live REST API)                 │
│ └── MCPExecutor (live MCP upstreams,               │
│     session-cached, auto-discovered tools)         │
└──────────────────────┬─────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────┐
│ PostgreSQL 16                                      │
│ users • mcp_servers (+connection_config)           │
│ mcp_tools (+source) • tool_invocations             │
│ schema_migrations                                  │
└────────────────────────────────────────────────────┘
```

### React UI architecture

```text
React App (Vite + TypeScript)
├── Authentication Context
│   ├── JWT token management
│   ├── User state • Protected routes
│   └── 401 interceptor → redirect to login
├── React Query (caching, mutations)
├── API Client (Axios interceptors)
└── Pages
    ├── Login
    ├── Servers (catalog, detail)
    ├── Tools (schema viewer, invoke sandbox)
    ├── Invocations (history, filters)
    ├── Metrics (live dashboard with charts)
    ├── Profile (password change)
    └── Admin (server/tool CRUD, connection config editor)
```

### Docker deployment architecture

```text
┌─────────────────────────────────────────────┐
│ docker compose                              │
│                                             │
│  ┌──────────┐   ┌──────────┐   ┌─────────┐  │
│  │ web      │──▶│ api      │──▶│postgres │  │
│  │ nginx:80 │   │ go:8080  │   │ pg:5432 │  │
│  └──────────┘   └──────────┘   └─────────┘  │
│       │                                     │
│  User → http://localhost:3000               │
│  (nginx proxies /api, /health, /metrics)    │
└─────────────────────────────────────────────┘
```

---

## 🔌 MCP Transport: How It Works

Phase 9 made the gateway a first-class MCP citizen in **both** directions, using the official [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) over the Streamable HTTP transport. (SSE was deprecated by the MCP spec and is intentionally not implemented; stdio is deferred.)

### Inbound: the gateway as an MCP server

AI agents (Claude Desktop, MCP Inspector, agent frameworks) connect to a single endpoint, `http://localhost:8080/mcp`, and see the whole governed catalog as one logical MCP server.

```text
MCP Client                      Go MCP Gateway
   │                                  │
   │  POST /mcp  initialize           │
   │ ────────────────▶  RequireAuthentication (JWT Bearer, RFC 6750 challenge)
   │                   NewHandler builds a FRESH mcp.Server per request:
   │                     1. List active servers from Postgres
   │                     2. List their tools; keep enabled + low-risk only
   │                     3. Register each as  serverName__toolName
   │                        with the stored JSON Schema as inputSchema
   │ ◀────────────────  serverInfo: mcp-gateway, capabilities: tools
   │                                  │
   │  POST /mcp  tools/list           │
   │ ────────────────▶  (catalog read live from Postgres — admin CRUD
   │ ◀────────────────   is reflected on the very next call)
   │                                  │
   │  POST /mcp  tools/call           │
   │    {name: "github__list_issues", arguments: {...}}
   │ ────────────────▶  split namespace → resolve server/tool IDs
   │                    → InvocationService.Invoke (UNCHANGED path):
   │                      role check → server active → tool enabled
   │                      → low-risk only → JSON Schema validation
   │                      → executor router → durable audit row
   │ ◀────────────────  content blocks (+ structuredContent)
   │                    isError:true on policy denial or execution failure
```

Key properties:

- **Stateless per-request server construction** — no session state, no cache invalidation, works across replicas; the newest MCP protocol revision requires stateless mode.
- **Zero duplicated policy logic** — `tools/call` delegates to the same `InvocationService` as the REST invoke endpoint. A governance bypass is impossible by construction.
- **Namespace namespacing (`serverName__toolName`)** prevents collisions between servers and makes governance ownership visible to agents.
- **Exposure rule** — a tool is visible iff its server is `active`, the tool is `enabled`, and `risk_level = 'low'` — identical to the REST policy.
- **Unknown tools** are rejected by the SDK at the protocol layer (`-32602 invalid params`); execution failures surface as `isError` tool results per MCP convention.
- **Auth** is the existing JWT (`POST /api/v1/auth/login` → `Authorization: Bearer`). 401s carry a `WWW-Authenticate: Bearer` challenge. Full MCP OAuth 2.1 is deferred (see Future Enhancements).
- The chi 15-second request timeout is scoped to `/api/v1` only, so long-lived MCP streams are never cut.

### Outbound: the gateway as an MCP client

Registering a server with `transport_type = streamable_http` turns the gateway into an MCP client:

```text
Admin creates server (base_url, connection_config)
        │
        ▼
ServerService.syncDiscoveredTools
        │  1. MCPExecutor dials the upstream (StreamableClientTransport,
        │     lazy, session cached by server ID, dial/call timeout from
        │     MCP_UPSTREAM_TIMEOUT_MS)
        │  2. tools/list (paginated)
        │  3. Upsert into mcp_tools with source='discovered' in ONE
        │     transaction — on conflict, only name/description/schema
        │     refresh; admin-set risk_level/enabled always win
        │  4. Discovery failure keeps the server row and logs the error
        ▼
Invocation of a discovered tool:
  RouterExecutor precedence:
    1. name = "github"                        → GitHubExecutor (REST)
    2. streamable_http + source='discovered'  → MCPExecutor (live MCP)
    3. otherwise                              → MockExecutor
        │
        ▼
  MCPExecutor.Execute: forwarded arguments → upstream tools/call
  → content blocks normalized into the audit JSON
  → isError / transport failures recorded as failed audit rows
  → metrics: mcp_gateway_upstream_requests_total{service="mcp", ...}
```

Credential hygiene for outbound upstreams: `connection_config` stores **environment variable references, never secrets**:

```json
{"headers": {"Authorization": "UPSTREAM_TOKEN_ENV"}}
```

Service-layer validation enforces `^[A-Z_][A-Z0-9_]*$` on every header value, so a raw token cannot be persisted; the executor resolves values from the process environment at dial time.

### Try it

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"developer@mcp-gateway.local","password":"DeveloperPass123"}' \
  | jq -r .accessToken)

npx @modelcontextprotocol/inspector   # Transport: Streamable HTTP
                                      # URL: http://localhost:8080/mcp
                                      # Header: Authorization: Bearer $TOKEN
# → tools/list shows github__list_issues, mock-tools__echo, ...
# → tools/call github__list_issues {"owner":"golang","repo":"go","per_page":3}
# → History page shows the audited invocation; /metrics increments
```

The end-to-end gate script `scripts/c1-gate-check.sh` automates 14 checks across both transports (auth rejection, exposure filtering, policy denial, audit rows, protocol errors).

---

## 🛠️ Technology Stack

### Frontend

| Area | Technology | Purpose |
|---|---|---|
| Framework | React 18 + TypeScript | Component-based UI with type safety |
| Build tool | Vite | Fast dev server, HMR, optimized builds |
| Styling | Tailwind CSS v4 | Utility-first CSS with Oxide engine |
| Routing | React Router v6 | Client-side routing with protected routes |
| Data fetching | TanStack React Query | Caching, synchronization, mutations |
| Charts | Recharts | Metrics dashboard visualizations |
| HTTP client | Axios | Interceptors, error handling |
| Icons | Lucide React | Consistent icon library |
| Dates | date-fns | Lightweight date formatting |

### Backend

| Area | Technology | Purpose |
|---|---|---|
| Language | Go 1.26 | Concurrent, strongly typed gateway |
| HTTP routing | go-chi/chi | Lightweight REST routing and middleware |
| CORS | go-chi/cors | Cross-origin resource sharing |
| MCP protocol | modelcontextprotocol/go-sdk | Tier-1 MCP SDK: Streamable HTTP server + client |
| Database | PostgreSQL 16 + pgx/v5 | Persistence and connection pooling |
| Migrations | golang-migrate | Version-controlled schema evolution |
| Schema validation | santhosh-tekuri/jsonschema/v6 | JSON Schema draft 2020-12 validation |
| Authentication | golang-jwt/v5 + bcrypt | JWT tokens and password hashing |
| GitHub integration | google/go-github/v62 | Typed GitHub REST API client |
| Metrics | prometheus/client_golang | Prometheus metrics exposition |
| Tracing | go.opentelemetry.io/otel | OpenTelemetry distributed tracing |

### Infrastructure

| Area | Technology | Purpose |
|---|---|---|
| Containers | Docker multi-stage builds | Small production images |
| Orchestration | Docker Compose | Full-stack local deployment |
| Web server | nginx | Static file serving + API proxy |
| CI | GitHub Actions | Test, lint, build automation |

---

## ✨ Current Features

### Backend (Phases 0–6, 8–9)

- ✅ Go API with chi router, middleware, and graceful shutdown
- ✅ PostgreSQL persistence with version-controlled migrations
- ✅ MCP server registry with full CRUD
- ✅ MCP tool catalog with JSON Schema validation and risk levels
- ✅ JWT authentication with bcrypt password hashing
- ✅ Role-based access control (`admin`, `developer`, `viewer`)
- ✅ Policy-controlled invocation gateway (low-risk only)
- ✅ Live GitHub REST executor with executor routing
- ✅ Durable audit records with `running → succeeded/failed` lifecycle
- ✅ Prometheus metrics (`/metrics`) with path normalization
- ✅ Invocation history API with role-based filtering and pagination
- ✅ OpenTelemetry tracing (stdout exporter for local dev)
- ✅ CORS middleware for React UI integration
- ✅ Password change endpoint with current-password verification
- ✅ **Native MCP server: governed catalog over Streamable HTTP at `/mcp`**
- ✅ **Outbound MCP executor with registration-time tool auto-discovery**
- ✅ **Tool provenance tracking (`manual` vs `discovered`) with governance-preserving sync**
- ✅ **Env-var credential references for upstream auth (no secrets in the DB)**
- ✅ Unit + concurrency tests (services, JWT, middleware, MCP executor, MCP endpoint)

### Frontend (Phases 7–9)

- ✅ Vite + React 18 + TypeScript with Tailwind CSS v4
- ✅ JWT login with protected routes and auth context
- ✅ Server catalog with status badges and detail pages
- ✅ Tool detail with schema viewer and invoke sandbox
- ✅ Invocation history with status filters and pagination
- ✅ Admin panel: server CRUD (list, create, edit, delete with confirmation)
- ✅ Admin panel: tool CRUD with JSON schema editor
- ✅ **Admin panel: connection config editor for MCP upstreams**
- ✅ Metrics dashboard with live charts (Recharts + Prometheus parsing)
- ✅ Profile page with password change form
- ✅ React Query caching, loading states, and error handling
- ✅ Responsive layout with role-based navigation

### Infrastructure (Phase 8)

- ✅ Backend Dockerfile (multi-stage Go build → Alpine)
- ✅ Frontend Dockerfile (Node build → nginx with SPA fallback + API proxy)
- ✅ Full-stack Docker Compose (postgres + api + web)
- ✅ GitHub Actions CI: backend tests/vet/fmt, frontend lint/build, Docker image builds

### Not yet implemented

- OAuth/OIDC identity-provider integration
- Full MCP authorization spec (OAuth 2.1 resource server); `/mcp` bridges the existing JWT
- stdio transport for MCP upstreams (SSE deprecated by the MCP spec, never implemented)
- Per-server/per-tool user permissions
- Medium/high-risk invocation confirmation flows
- Additional REST executors (Jira, Slack, Confluence)
- Grafana/Jaeger backends for metrics and traces
- Invocation detail modal and CSV export in history page
- Refresh tokens and token revocation

---

## 🗺️ Development Roadmap

| Phase | Status | Focus | Main Outcome |
|---|---|---|---|
| Phase 0 | ✅ Complete | Foundation | Go API, PostgreSQL, Docker Compose, health endpoint |
| Phase 1 | ✅ Complete | Server Registry | Persistent CRUD API for MCP server metadata |
| Phase 2 | ✅ Complete | Tool Catalog | Per-server tools, input schemas, risk levels |
| Phase 3 | ✅ Complete | Auth & RBAC | JWT authentication, bcrypt, role protection |
| Phase 4 | ✅ Complete | Invocation Gateway | Policy-checked, schema-validated, audited invocations |
| Phase 5 | ✅ Complete | GitHub Integration | Real GitHub REST executor with audit trail |
| Phase 6 | ✅ Complete | Observability | Prometheus metrics, tracing, history API |
| Phase 7 | ✅ Complete | React UI | Vite + React UI with catalog, sandbox, history |
| Phase 8 | ✅ Complete | Delivery & Polish | Admin panel, metrics dashboard, Docker, CI |
| Phase 9 | ✅ Complete | Native MCP Transport | `/mcp` endpoint, outbound MCPExecutor, tool auto-discovery |

---

## 📁 Repository Structure

```text
mcp-gateway/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go                  # Application entrypoint
│   │   └── passwordhash/main.go         # bcrypt hash utility
│   ├── internal/
│   │   ├── auth/                        # JWT signing & validation
│   │   ├── config/                      # Environment configuration
│   │   ├── domain/                      # Core models & request types
│   │   ├── executor/                    # Mock + GitHub + MCP + router executors
│   │   ├── httpapi/                     # Handlers, middleware, router, CORS
│   │   ├── mcp/                         # MCP Streamable HTTP endpoint + catalog adapter
│   │   ├── observability/               # Prometheus + OpenTelemetry
│   │   ├── platform/database/           # PostgreSQL connection pool
│   │   ├── repository/                  # SQL persistence layer (+ discovery upsert)
│   │   └── service/                     # Business logic & validation (+ discovery sync)
│   ├── migrations/                      # 5 versioned SQL migrations
│   ├── Dockerfile                       # Multi-stage Go build
│   ├── .dockerignore
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── api/                         # Axios client + service functions
│   │   ├── components/Layout.tsx        # Navigation shell
│   │   ├── hooks/useAuth.tsx            # Auth context
│   │   ├── pages/
│   │   │   ├── login/                   # Login page
│   │   │   ├── servers/                 # Catalog + detail
│   │   │   ├── tools/                   # Detail + invoke sandbox
│   │   │   ├── invocations/             # History table
│   │   │   ├── metrics/                 # Live metrics dashboard
│   │   │   ├── profile/                 # Password change
│   │   │   └── admin/                   # Server/tool CRUD forms
│   │   ├── types/                       # TypeScript definitions
│   │   ├── utils/prometheus.ts          # Prometheus text parser
│   │   └── App.tsx                      # Router setup
│   ├── Dockerfile                       # Node build → nginx
│   ├── nginx.conf                       # SPA fallback + API proxy
│   ├── vite.config.ts
│   └── package.json
│
├── .github/workflows/ci.yml             # CI pipeline
├── docs/phase-9-decisions.md            # MCP transport architecture decision log (D1–D11)
├── scripts/seed_users.sql               # Development user seed
├── scripts/c1-gate-check.sh             # 14-check end-to-end MCP gate script
├── docker-compose.yml                   # Full-stack orchestration
├── .env.example
└── README.md
```

---

## 📋 Prerequisites

For **Docker deployment** (recommended):

- Docker Desktop or Docker Engine with Docker Compose

For **local development**:

- Go 1.26+, Node.js 18+, Docker Compose, Git, `curl`, `jq`, `psql`, `golang-migrate`

---

## 🐳 Quick Start with Docker

The fastest way to run the entire stack:

### 1. Clone and configure

```bash
git clone https://github.com/iashu2k/mcp-gateway.git
cd mcp-gateway
cp .env.example .env
```

Generate a JWT secret and set it in `.env`:

```bash
openssl rand -base64 48
```

### 2. Build and start all services

```bash
docker compose up --build
```

This starts three containers:

| Service | Image | Port | Purpose |
|---|---|---|---|
| `postgres` | postgres:16-alpine | 5432 | Database with health checks |
| `api` | Multi-stage Go build | 8080 | Gateway API + `/mcp` endpoint |
| `web` | Node build → nginx | 3000 | React UI with API proxy |

### 3. Run migrations and seed users

```bash
set -a; source .env; set +a

migrate -path backend/migrations -database "$DATABASE_URL" up
psql "$DATABASE_URL" -f scripts/seed_users.sql
```

### 4. Open the UI

Navigate to **http://localhost:3000** and sign in:

| Email | Password | Role |
|---|---|---|
| `admin@mcp-gateway.local` | `AdminPass123` | `admin` |
| `developer@mcp-gateway.local` | `DeveloperPass123` | `developer` |
| `viewer@mcp-gateway.local` | `ViewerPass123` | `viewer` |

> ⚠️ Local development credentials only. Never use them in any deployed environment.

The nginx container proxies `/api`, `/health`, and `/metrics` to the Go service — no CORS issues in Docker mode.

---

## 💻 Local Development Setup

### Backend

```bash
cp .env.example .env          # configure JWT_SECRET
docker compose up -d postgres # database only
set -a; source .env; set +a
migrate -path backend/migrations -database "$DATABASE_URL" up
psql "$DATABASE_URL" -f scripts/seed_users.sql

cd backend
go run ./cmd/api              # http://localhost:8080  (REST + /mcp)
```

### Frontend

```bash
cd frontend
npm install
npm run dev                   # http://localhost:5173
```

In dev mode, the UI at `:5173` talks to the API at `:8080` via the backend's CORS middleware.

---

## 🔐 Environment Variables

### Backend

| Variable | Required | Description |
|---|---:|---|
| `APP_ENV` | Yes | Runtime environment name |
| `HTTP_PORT` | Yes | Port for the Go HTTP API |
| `POSTGRES_DB` / `POSTGRES_USER` / `POSTGRES_PASSWORD` | Yes | PostgreSQL credentials |
| `DATABASE_URL` | Yes | Connection URL for app and migrations |
| `JWT_SECRET` | Yes | HS256 signing secret (≥32 chars) |
| `JWT_ISSUER` | Yes | Expected issuer claim |
| `JWT_TTL_MINUTES` | Yes | Access-token lifetime |
| `GITHUB_TOKEN` | No | GitHub PAT (empty = unauthenticated, 60 req/hr) |
| `MCP_UPSTREAM_TIMEOUT_MS` | No | Outbound MCP upstream dial/call timeout (default `10000`) |

### Frontend

| Variable | Required | Description |
|---|---:|---|
| `VITE_API_BASE_URL` | Yes | Base URL for the Go API (dev mode only; nginx proxies in Docker) |

`.env` files are local-only and must never be committed. The same rule applies to upstream MCP credentials: `connection_config` holds environment variable *names*, and the referenced variables are provided through the process environment.

---

## 🗄️ Database Migrations

```text
000001_create_mcp_servers.up/down.sql
000002_create_mcp_tools.up/down.sql
000003_create_users.up/down.sql
000004_create_tool_invocations.up/down.sql
000005_add_outbound_mcp_support.up/down.sql
```

```bash
migrate -path backend/migrations -database "$DATABASE_URL" up
migrate -path backend/migrations -database "$DATABASE_URL" version   # → 5
```

Migration 000005 adds `mcp_servers.connection_config` (JSONB, credential references) and `mcp_tools.source` (`manual` | `discovered`) for Phase 9 outbound support.

---

## 👥 Authentication and Roles

| Role | Catalog reads | Catalog mutations | Tool invocation (REST + MCP) | Invocation history |
|---|---:|---:|---:|---:|
| `admin` | ✅ | ✅ | ✅ (low-risk) | All users |
| `developer` | ✅ | ❌ | ✅ (low-risk) | Own only |
| `viewer` | ✅ | ❌ | ❌ | ❌ |

The same role rules apply on both transports: a viewer's MCP `tools/call` is denied by the shared policy chain with an `isError` result, and no audit row is created.

### Password change

Users can change their own password via the Profile page or API:

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"currentPassword":"AdminPass123","newPassword":"NewSecurePass456"}'
```

The endpoint verifies the current password against the stored bcrypt hash, enforces an 8-character minimum, and updates the hash atomically.

---

## 📡 API Reference

Base URL: `http://localhost:8080/api/v1`

### Public

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/health` | Gateway + database health |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/auth/login` | Authenticate, receive JWT |

### MCP endpoint

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` `GET` `DELETE` | `/mcp` | MCP over Streamable HTTP (JWT required). `initialize`, `tools/list`, `tools/call` for the governed catalog |

See [MCP Transport: How It Works](#-mcp-transport-how-it-works) for the full flow.

### Authenticated (all roles)

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/auth/me` | Current identity |
| `POST` | `/auth/change-password` | Change own password |
| `GET` | `/servers` • `/servers/{id}` | Server catalog reads |
| `GET` | `/servers/{id}/tools` • `/tools/{id}` | Tool catalog reads |

### Admin + developer

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/servers/{id}/tools/{id}/invoke` | Invoke low-risk tool |
| `GET` | `/invocations` • `/invocations/{id}` | History (role-filtered) |

### Admin only

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` / `PATCH` / `DELETE` | `/servers[...]` | Server mutations (creating/updating a `streamable_http` server triggers tool auto-discovery) |
| `POST` / `PATCH` / `DELETE` | `/servers/{id}/tools[...]` | Tool mutations |

---

## 📖 Phase Summaries

<details>
<summary><strong>Phases 0–3: Foundation → Registry → Catalog → Auth</strong></summary>

- **Phase 0:** Go module, Docker Compose PostgreSQL, health endpoint, structured logging, graceful shutdown
- **Phase 1:** Server registry CRUD with unique-name constraints and validation
- **Phase 2:** Tool catalog with JSON Schema storage, risk levels (`low`/`medium`/`high`), enablement
- **Phase 3:** Users table, bcrypt hashing, HS256 JWT issuance/verification, role middleware
</details>

<details>
<summary><strong>Phases 4–6: Invocation → GitHub → Observability</strong></summary>

- **Phase 4:** Policy chain (auth → role → server active → tool enabled → low-risk → schema validation), audit lifecycle with `ON DELETE RESTRICT` foreign keys, deterministic mock executor
- **Phase 5:** `GitHubExecutor` with `go-github/v62` for `list_issues` and `search_repositories`, executor routing by server name, upstream error capture in audit records
- **Phase 6:** Prometheus metrics (HTTP, invocations, upstream, DB pool), invocation history API with role-based filtering and pagination, OpenTelemetry initialization
</details>

<details>
<summary><strong>Phases 7–8: React UI → Delivery & Polish</strong></summary>

- **Phase 7:** Vite + React + TypeScript UI, Tailwind CSS v4, login with JWT, server catalog, tool invoke sandbox, invocation history, React Query, CORS middleware
- **Phase 8:**
  - 🛠️ **Admin panel** — server CRUD with delete confirmation, tool CRUD with JSON schema editor, typed forms
  - 📊 **Metrics dashboard** — Recharts visualizations fed by a Prometheus text parser, auto-refresh every 15s
  - 🔑 **Password change** — backend endpoint with current-password verification + profile UI
  - 🐳 **Docker** — multi-stage backend Dockerfile (Go → Alpine), frontend Dockerfile (Node → nginx with SPA fallback and API proxy), full-stack compose
  - ⚙️ **CI** — GitHub Actions with Go test/vet/fmt, frontend lint/build, Docker image builds
</details>

<details>
<summary><strong>Phase 9: Native MCP Transport</strong></summary>

- 🔌 **Inbound** — `/mcp` Streamable HTTP endpoint (official `modelcontextprotocol/go-sdk`, stateless per-request server) exposing the governed catalog with namespaced tools; `tools/call` reuses the existing invocation service — policy chain, schema validation, and audit trail unchanged
- 🔑 **Auth** — JWT bearer bridge with RFC 6750 `WWW-Authenticate` challenges; MCP OAuth 2.1 deferred
- 📡 **Outbound** — `MCPExecutor` with lazy, session-cached connections to live MCP upstreams, env-var credential references, and configurable timeouts
- 🔍 **Discovery** — registration-time `tools/list` sync writes `source='discovered'` tools in a single transaction, preserving admin-set risk/enabled governance on conflict
- 📊 **Observability** — upstream metrics extended with `service="mcp"`; MCP calls audited and metered identically to REST
- ✅ Race-tested concurrent MCP client sessions; 14-check end-to-end gate script (`scripts/c1-gate-check.sh`); full decision log in `docs/phase-9-decisions.md`
</details>

---

## ✅ Validation and Testing

### Backend

```bash
cd backend
gofmt -w . && go mod tidy
go test -race ./... && go vet ./...
```

### Frontend

```bash
cd frontend
npm run lint && npm run build
```

### MCP end-to-end gate

```bash
bash scripts/c1-gate-check.sh
# 14 checks: auth rejection + Bearer challenge, initialize, exposure
# filtering (disabled / medium-risk / inactive-server tools hidden),
# tools/call with audit row, viewer policy denial, protocol-level
# unknown-tool rejection
```

### Full-stack integration

```bash
docker compose up --build
# → http://localhost:3000 → login → invoke tool → check history → view metrics
```

### CI pipeline

Every push and PR runs:

- 🔍 Go: `gofmt` check, `go vet`, `go test -race` (with PostgreSQL service), `go build`
- 🎨 Frontend: `npm ci`, `npm run lint`, `npm run build`
- 🐳 Docker: backend and frontend image builds

---

## 🧭 Design Decisions

### Why nginx proxy in Docker, CORS in dev

Two modes, each idiomatic:

- **Dev:** Vite dev server (`:5173`) → Go API (`:8080`) cross-origin, handled by chi CORS middleware
- **Docker:** nginx serves static files and proxies `/api` same-origin — no CORS, no exposed API port needed

### Why a stateless MCP server per request

The `/mcp` endpoint builds a fresh MCP server per request from the live Postgres catalog (SDK stateless mode). Admin CRUD is reflected on the next `tools/list` with zero cache invalidation, there is no session state to leak across replicas, and the newest MCP protocol revision requires it.

### Why namespaced tool names

Tools are exposed as `serverName__toolName` so auto-discovered tools from multiple upstreams can never collide, and agents can see governance ownership at a glance.

### Why credential references, not secrets, in connection_config

Outbound upstream credentials are stored as environment variable names (`{"headers": {"Authorization": "UPSTREAM_TOKEN_ENV"}}`), enforced by service validation (`^[A-Z_][A-Z0-9_]*$`). The database never holds a raw upstream secret; values are resolved from the process environment at dial time.

### Why tool provenance (`source`) drives executor routing

Every pre-Phase-9 row defaulted to `transport_type='streamable_http'`, so transport alone cannot distinguish a live MCP upstream from a mock-executed demo server. The `discovered` marker — written only by a successful upstream `tools/list` — is the explicit signal that routes a server to the `MCPExecutor`. Existing behavior is preserved by construction.

### Why a Prometheus text parser instead of a metrics backend

For a self-contained demo, parsing `/metrics` directly in the UI avoids running Prometheus + Grafana infrastructure. The parser normalizes the exposition format into chart-ready data. In production, swap this for Grafana dashboards backed by a real Prometheus server.

### Why delete confirmation inline instead of a modal

The admin server list uses an inline confirm/cancel pattern: one click arms the delete, a second confirms. This prevents accidental deletion without the complexity of modal state management.

### Why bcrypt hash updates don't revoke existing tokens

Password changes update the hash but existing JWTs remain valid until expiry (≤60 min TTL). Token revocation requires a denylist or token versioning — deliberately deferred as a documented limitation.

---

## ⚠️ Known Limitations

- Only two GitHub tools; no Jira/Slack/Confluence REST executors
- Only `low`-risk tools invocable; no confirmation flow for medium/high
- Roles are global; no per-server/per-tool permissions
- `/mcp` auth is a JWT bridge, not full MCP OAuth 2.1 (no protected-resource metadata or dynamic client registration)
- No stdio upstreams; no SSE transport (deprecated by the MCP spec)
- Discovery sync runs only on server create/update (no periodic re-sync or upstream health checks)
- GitHub token is env-only; no credential reference store
- Metrics UI parses `/metrics` directly (no historical data, no Grafana)
- Traces go to stdout only (no Jaeger/Tempo backend)
- JWT in localStorage (XSS-vulnerable); no refresh tokens or revocation
- No rate limiting, CSRF protection, or security headers yet

---

## 🌱 Future Enhancements

Roughly in priority order, each already scoped with known tradeoffs (see `docs/phase-9-decisions.md`):

- **MCP OAuth 2.1 authorization** — replace the JWT bridge with a real resource server: protected-resource metadata, `WWW-Authenticate` challenges with `resource_metadata`, dynamic client registration for agent clients
- **stdio upstream support** — subprocess lifecycle management (needs a non-Alpine base image or sidecar model; deferred from Phase 9 deliberately)
- **Periodic re-discovery + upstream health checks** — background reconciler that re-syncs `tools/list`, marks stale tools, and flips server status to `unhealthy` on repeated dial failures
- **Per-server/per-tool permissions** — replace global roles with grants; the audit schema already carries user IDs
- **Medium/high-risk confirmation flows** — pending-approval invocation state with admin review in the UI
- **Refresh tokens + revocation** — token versioning or denylist; enables longer-lived agent credentials for `/mcp`
- **Per-server MCP endpoints** — `/mcp/{serverName}` views for scoped agent access
- **Rich content passthrough** — forward non-text MCP content blocks (images, resources) instead of the current text-first normalization
- **Rate limiting, security headers, CSRF** — baseline hardening for internet-facing deployment
- **Grafana/Jaeger backends** — replace the in-UI Prometheus parser and stdout traces with real backends

---

## 🔧 Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `go.mod requires go >= 1.26.5` in Docker | Base image too old | Use `golang:1.26-alpine` in Dockerfile |
| CORS error in dev | Missing middleware | Ensure CORS middleware is first in `router.go` |
| 401 on login | Wrong password hash | Regenerate hash with `cmd/passwordhash`, update DB (escape `$` as `\$`) |
| 401 from MCP Inspector | Missing/expired JWT | Re-login via `/api/v1/auth/login`; check `WWW-Authenticate` header |
| MCP client disconnects after 15s | Timeout middleware on `/mcp` | Timeout must be scoped to `/api/v1` only (see router.go) |
| Registered upstream has no tools | Discovery failed (auth/network) | Check API logs for `tool discovery failed`; verify `connectionConfig` env vars are set on the API process |
| Blank page in Docker | SPA routing | Ensure nginx `try_files $uri $uri/ /index.html;` |
| API calls 404 in Docker | Missing proxy | Verify nginx `location /api/` block proxies to `api:8080` |
| Tailwind not applying | Missing import | Ensure `@import "tailwindcss"` is in `index.css` |
| Stuck in login redirect | Stale token | `localStorage.clear()` in browser console |

---

## 🤝 Contributing Workflow

```bash
git checkout -b feat/my-feature

# Backend
cd backend && gofmt -w . && go test -race ./... && go vet ./...

# Frontend
cd ../frontend && npm run lint && npm run build

git add . && git commit -m "feat: my feature" && git push -u origin feat/my-feature
# CI runs automatically on PR
```

Commit convention:

```text
feat: add admin panel for server crud
fix: correct CORS middleware ordering
test: cover mcp transport inbound and outbound paths
docs: update phase 9 readme
chore: add docker multi-stage builds
ci: add github actions pipeline
```

---

## 🏁 Project Status

```text
Phase 0: ✅ Complete     Phase 4: ✅ Complete     Phase 8: ✅ Complete
Phase 1: ✅ Complete     Phase 5: ✅ Complete     Phase 9: ✅ Complete
Phase 2: ✅ Complete     Phase 6: ✅ Complete
Phase 3: ✅ Complete     Phase 7: ✅ Complete
```

**All 10 phases (0–9) complete.** The gateway is a fully functional, containerized, CI-validated full-stack platform that speaks native MCP in both directions — ready for demo and portfolio presentation.

---

## 🎬 Demo Walkthrough

```bash
# 1. Start the stack
docker compose up --build

# 2. Open http://localhost:3000
# 3. Login: admin@mcp-gateway.local / AdminPass123
# 4. Servers → github → "List GitHub Issues"
# 5. Invoke with: {"owner":"golang","repo":"go","per_page":3}
# 6. View real GitHub issues in the result panel
# 7. History → see the audited invocation with duration
# 8. Metrics → watch charts update with the new invocation
# 9. Admin → add/edit servers and tools via the UI
# 10. Profile → change your password

# 11. MCP: get a token and connect MCP Inspector to http://localhost:8080/mcp
#     (Streamable HTTP, header: Authorization: Bearer $TOKEN)
# 12. tools/list → only enabled low-risk tools, namespaced (github__list_issues)
# 13. Call github__list_issues with {"owner":"golang","repo":"go","per_page":3}
# 14. History → the MCP-invoked call appears, audited identically to REST

# 15. Outbound: Admin → register a server with transport streamable_http
#     pointing at a live MCP server → its tools auto-appear (source=discovered)
# 16. Invoke a discovered tool → routed live through the MCPExecutor
```

The complete flow — UI and MCP clients → API → policy → execution → audit → metrics — works end-to-end in a single `docker compose up`.
