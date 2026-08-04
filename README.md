# MCP Gateway

> 🚀 A self-hosted MCP Gateway for centrally registering, discovering, governing, and observing internal AI-tool integrations — complete with a Go backend, React UI, and full Docker deployment.

The project is inspired by the idea of an internal "USB-C for AI agents": a unified platform where developers and AI agents can discover approved Model Context Protocol (MCP) servers, inspect their tools, invoke approved capabilities through centralized controls, and obtain audit-ready execution history with full observability.

> ✅ **Current status:** Phase 8 complete — the gateway is production-ready with an admin panel, metrics dashboard, password management, Docker deployment, and CI pipeline.

---

## 📑 Table of Contents

- [Project Vision](#-project-vision)
- [Why This Project](#-why-this-project)
- [System Architecture](#-system-architecture)
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
                    │       React Web UI           │
                    │ Catalog • Sandbox • History  │
                    │ Metrics • Admin Panel        │
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

## 💡 Why This Project

This project demonstrates practical full-stack and AI-platform engineering skills rather than only building a standalone LLM application.

It focuses on:

- Go backend development and service design
- PostgreSQL schema design and migrations
- API gateway patterns and JSON Schema-driven validation
- JWT authentication and role-based access control
- Policy-controlled tool invocation with durable audit records
- Live third-party API integration (GitHub REST)
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
Developer / API Client / React UI
         │
         │ HTTP / JSON / CORS
         ▼
┌──────────────────────────────────────┐
│ Go MCP Gateway                       │
│                                      │
│ chi Router • CORS • Request ID       │
│ Status-aware logging • Timeouts      │
│ JWT auth • RBAC middleware           │
│ Health + Metrics endpoints           │
│ Server Registry • Tool Catalog       │
│ Invocation API • History API         │
│ Password change endpoint             │
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
                 ▼
┌──────────────────────────────────────┐
│ PostgreSQL 16                        │
│ users • mcp_servers • mcp_tools      │
│ tool_invocations • schema_migrations │
└──────────────────────────────────────┘
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
    └── Admin (server/tool CRUD)
```

### Docker deployment architecture

```text
┌─────────────────────────────────────────────┐
│ docker compose                               │
│                                             │
│  ┌──────────┐   ┌──────────┐   ┌─────────┐  │
│  │ web      │──▶│ api      │──▶│postgres │  │
│  │ nginx:80 │   │ go:8080  │   │ pg:5432 │  │
│  └──────────┘   └──────────┘   └─────────┘  │
│       │                                      │
│  User → http://localhost:3000                │
│  (nginx proxies /api, /health, /metrics)     │
└─────────────────────────────────────────────┘
```

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

### Backend (Phases 0–6, 8)

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
- ✅ **Password change endpoint with current-password verification**
- ✅ Unit tests for services, JWT, and middleware

### Frontend (Phases 7–8)

- ✅ Vite + React 18 + TypeScript with Tailwind CSS v4
- ✅ JWT login with protected routes and auth context
- ✅ Server catalog with status badges and detail pages
- ✅ Tool detail with schema viewer and invoke sandbox
- ✅ Invocation history with status filters and pagination
- ✅ **Admin panel: server CRUD (list, create, edit, delete with confirmation)**
- ✅ **Admin panel: tool CRUD with JSON schema editor**
- ✅ **Metrics dashboard with live charts (Recharts + Prometheus parsing)**
- ✅ **Profile page with password change form**
- ✅ React Query caching, loading states, and error handling
- ✅ Responsive layout with role-based navigation

### Infrastructure (Phase 8)

- ✅ **Backend Dockerfile (multi-stage Go build → Alpine)**
- ✅ **Frontend Dockerfile (Node build → nginx with SPA fallback + API proxy)**
- ✅ **Full-stack Docker Compose (postgres + api + web)**
- ✅ **GitHub Actions CI: backend tests/vet/fmt, frontend lint/build, Docker image builds**

### Not yet implemented

- OAuth/OIDC identity-provider integration
- Per-server/per-tool user permissions
- Medium/high-risk invocation confirmation flows
- Additional executors (Jira, Slack, Confluence)
- MCP protocol transport (stdio/SSE/streamable HTTP)
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
│   │   ├── executor/                    # Mock + GitHub + router executors
│   │   ├── httpapi/                     # Handlers, middleware, router, CORS
│   │   ├── observability/               # Prometheus + OpenTelemetry
│   │   ├── platform/database/           # PostgreSQL connection pool
│   │   ├── repository/                  # SQL persistence layer
│   │   └── service/                     # Business logic & validation
│   ├── migrations/                      # 4 versioned SQL migrations
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
├── scripts/seed_users.sql               # Development user seed
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
| `api` | Multi-stage Go build | 8080 | Gateway API |
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
go run ./cmd/api              # http://localhost:8080
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

### Frontend

| Variable | Required | Description |
|---|---:|---|
| `VITE_API_BASE_URL` | Yes | Base URL for the Go API (dev mode only; nginx proxies in Docker) |

`.env` files are local-only and must never be committed.

---

## 🗄️ Database Migrations

```text
000001_create_mcp_servers.up/down.sql
000002_create_mcp_tools.up/down.sql
000003_create_users.up/down.sql
000004_create_tool_invocations.up/down.sql
```

```bash
migrate -path backend/migrations -database "$DATABASE_URL" up
migrate -path backend/migrations -database "$DATABASE_URL" version   # → 4
```

---

## 👥 Authentication and Roles

| Role | Catalog reads | Catalog mutations | Tool invocation | Invocation history |
|---|---:|---:|---:|---:|
| `admin` | ✅ | ✅ | ✅ (low-risk) | All users |
| `developer` | ✅ | ❌ | ✅ (low-risk) | Own only |
| `viewer` | ✅ | ❌ | ❌ | ❌ |

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
| `POST` / `PATCH` / `DELETE` | `/servers[...]` | Server mutations |
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

---

## ✅ Validation and Testing

### Backend

```bash
cd backend
gofmt -w . && go mod tidy
go test ./... && go vet ./...
```

### Frontend

```bash
cd frontend
npm run lint && npm run build
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

### Why a Prometheus text parser instead of a metrics backend

For a self-contained demo, parsing `/metrics` directly in the UI avoids running Prometheus + Grafana infrastructure. The parser normalizes the exposition format into chart-ready data. In production, swap this for Grafana dashboards backed by a real Prometheus server.

### Why delete confirmation inline instead of a modal

The admin server list uses an inline confirm/cancel pattern: one click arms the delete, a second confirms. This prevents accidental deletion without the complexity of modal state management.

### Why bcrypt hash updates don't revoke existing tokens

Password changes update the hash but existing JWTs remain valid until expiry (≤60 min TTL). Token revocation requires a denylist or token versioning — deliberately deferred as a documented limitation.

---

## ⚠️ Known Limitations

- Only two GitHub tools; no Jira/Slack/Confluence executors
- Only `low`-risk tools invocable; no confirmation flow for medium/high
- Roles are global; no per-server/per-tool permissions
- GitHub token is env-only; no credential reference store
- Metrics UI parses `/metrics` directly (no historical data, no Grafana)
- Traces go to stdout only (no Jaeger/Tempo backend)
- JWT in localStorage (XSS-vulnerable); no refresh tokens or revocation
- No MCP protocol transport; integrations are direct REST calls
- No rate limiting, CSRF protection, or security headers yet

---

## 🔧 Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `go.mod requires go >= 1.26.5` in Docker | Base image too old | Use `golang:1.26-alpine` in Dockerfile |
| CORS error in dev | Missing middleware | Ensure CORS middleware is first in `router.go` |
| 401 on login | Wrong password hash | Regenerate hash with `cmd/passwordhash`, update DB (escape `$` as `\$`) |
| Blank page in Docker | SPA routing | Ensure nginx `try_files $uri $uri/ /index.html;` |
| API calls 404 in Docker | Missing proxy | Verify nginx `location /api/` block proxies to `api:8080` |
| Tailwind not applying | Missing import | Ensure `@import "tailwindcss"` is in `index.css` |
| Stuck in login redirect | Stale token | `localStorage.clear()` in browser console |

---

## 🤝 Contributing Workflow

```bash
git checkout -b feat/my-feature

# Backend
cd backend && gofmt -w . && go test ./... && go vet ./...

# Frontend
cd ../frontend && npm run lint && npm run build

git add . && git commit -m "feat: my feature" && git push -u origin feat/my-feature
# CI runs automatically on PR
```

Commit convention:

```text
feat: add admin panel for server crud
fix: correct CORS middleware ordering
test: cover password change endpoint
docs: update phase 8 readme
chore: add docker multi-stage builds
ci: add github actions pipeline
```

---

## 🏁 Project Status

```text
Phase 0: ✅ Complete     Phase 4: ✅ Complete     Phase 8: ✅ Complete
Phase 1: ✅ Complete     Phase 5: ✅ Complete
Phase 2: ✅ Complete     Phase 6: ✅ Complete
Phase 3: ✅ Complete     Phase 7: ✅ Complete
```

**All 8 phases complete.** The gateway is a fully functional, containerized, CI-validated full-stack platform ready for demo and portfolio presentation.

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
```

The complete flow — UI → API → policy → execution → audit → metrics — works end-to-end in a single `docker compose up`.
