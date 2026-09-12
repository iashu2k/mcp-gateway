# Phase 9 Decision Log — Native MCP Transport

> Status: **accepted** (C0). Date: 2026-09-12. Scope: Phase 9 inbound + outbound MCP transport per `phase-9-handover-native-mcp-transport.md`.
> SDK verified against `github.com/modelcontextprotocol/go-sdk` v1.7.x source on 2026-09-12 (supports spec versions through 2026-07-28).
> D5 amended and D11 added on 2026-09-12 after reviewing `000001_create_mcp_servers.up.sql` and `server_service.go`.

| # | Decision | Outcome |
|---|----------|---------|
| D1 | MCP endpoint shape | Single `/mcp` endpoint exposing the whole governed catalog as one MCP server |
| D2 | Auth on `/mcp` | JWT bearer bridge via existing `RequireAuthentication` middleware; MCP OAuth 2.1 deferred to Phase 10+ |
| D3 | Tool naming | Namespaced `serverName__toolName` |
| D4 | Outbound discovery | Auto `tools/list` sync on registration when `transport_type=streamable_http`; manual CRUD stays for other servers; discovered tools get `source='discovered'` |
| D5 | DB schema change | **Amended:** reuse existing `transport_type`; migration `000005` adds `connection_config jsonb` to `mcp_servers` and `source text` to `mcp_tools` only |
| D6 | Streaming audit semantics | Finalize audit row when the upstream call completes (after stream close); capture upstream error text on failure, same pattern as `GitHubExecutor` |
| D7 | MCP-exposed tools | All enabled, low-risk tools on active servers; no new flags in Phase 9 |
| D8 | Inbound server topology | Stateless per-request server construction (see below) |
| D9 | Namespace parsing constraint | Split on the first `__`; server names containing `__` are unsupported |
| D10 | HTTP timeout re-scoping | Global 15s `middleware.Timeout` moves into the `/api/v1` group; `/mcp` gets no chi timeout |
| D11 | Outbound executor routing | `MCPExecutor` handles only `streamable_http` servers with discovered tools; name-based routing stays first; mock stays the fallback |

---

## D1 — Single `/mcp` endpoint

**Decision:** One endpoint, `POST/GET/DELETE /mcp`, serves the entire governed catalog as a single logical MCP server.

**Rationale:** The gateway's value proposition is "one governed front door." Per-server endpoints (`/mcp/{serverName}`) multiply auth/policy surface and are a later enhancement if needed.

**Consequences:** Tool name collisions across servers must be handled by namespacing (D3).

## D2 — JWT bridge for MVP

**Decision:** `/mcp` is mounted behind the existing `RequireAuthentication(tokenService)` middleware. Clients obtain a token via `POST /api/v1/auth/login` and send `Authorization: Bearer <token>`.

**Rationale:** Reuses the tested auth path; the authenticated user flows through request context into tool handlers, so invocations are attributed to real users in the audit trail.

**Consequences:** Full MCP auth spec (OAuth 2.1 resource server, protected-resource metadata, `WWW-Authenticate` challenges) is deferred. Phase 9.2 adds `WWW-Authenticate` on 401s and docs for non-browser clients only.

## D3 — Namespaced tool names

**Decision:** Tools are advertised as `serverName__toolName` (e.g. `github__list_issues`).

**Rationale:** Prevents collisions once outbound-discovered tools from multiple servers share the catalog (D4); self-documenting for agents reading `tools/list`.

**Consequences:** See D9 for the parsing constraint.

## D4 — Auto-discovery on outbound registration

**Decision:** When a server is registered/updated with `transport_type=streamable_http`, the gateway connects, runs `tools/list`, and upserts discovered tools (name, description, input schema) into the catalog, marked with `source='discovered'`.

**Consequences:** Manual edits to discovered tools risk being overwritten by the next sync; registration-time discovery failure keeps the server row and records the error (per-server discovery state lives in Phase 9.4 implementation notes).

## D5 — Migration 000005 (amended 2026-09-12)

**Original plan (superseded):** add `transport text not null default 'rest'` (check: `rest`, `streamable-http`) + `connection_config jsonb`.

**Why amended:** the schema already models transport — `mcp_servers.transport_type TEXT NOT NULL DEFAULT 'streamable_http'` with `CHECK (transport_type IN ('streamable_http', 'sse', 'stdio'))`, and `domain.MCPServer.TransportType` plus `server_service.go` validation already use it. A new hyphenated `transport` column would duplicate it and fight the existing CHECK constraint. There is also no `rest` value: every existing row (github, mock demo servers) defaults to `streamable_http`, which forces the routing rule in D11.

**Amended migration `000005_add_outbound_mcp_support.up/down.sql`:**

```sql
-- up
ALTER TABLE mcp_servers
  ADD COLUMN connection_config jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE mcp_tools
  ADD COLUMN source text NOT NULL DEFAULT 'manual'
    CONSTRAINT valid_tool_source CHECK (source IN ('manual', 'discovered'));

-- down
ALTER TABLE mcp_tools DROP COLUMN source;
ALTER TABLE mcp_servers DROP COLUMN connection_config;
```

**Security note:** `connection_config` stores credential *references* (env var names), never raw tokens. Goes in the README security section in Phase 9.7.

## D6 — Audit finalization on stream close

**Decision:** `succeeded` only when the upstream stream closes cleanly; partial results with a clean close still count as succeeded and the partial content is captured. Failures capture upstream error text, same as the existing GitHub executor pattern.

## D7 — Exposure rule

**Decision:** A tool is MCP-visible iff `server.status = 'active' AND tool.enabled AND tool.risk_level = 'low'` — identical to the existing REST invocation policy, so no new flags and no behavior drift between transports.

## D8 — Stateless per-request MCP server

**Decision:** `NewStreamableHTTPHandler` with `StreamableHTTPOptions{Stateless: true}`; the `getServer` callback builds a fresh `mcp.Server` per request and registers the catalog read from Postgres at that moment.

**Rationale:**
1. The SDK only accepts the newest protocol revision (2026-07-28) in stateless mode; stateful mode targets older revisions (verified in SDK docs 2026-09-12).
2. Catalog freshness for free: admin CRUD is reflected on the next `tools/list` without wiring `AddTool`/`RemoveTools` calls into the tool service.
3. No session state to leak across a future multi-replica deployment.

**Consequences:** No server-initiated requests to clients (sampling/elicitation) — irrelevant for this gateway. No `Mcp-Session-Id` semantics to manage.

## D9 — Namespace parsing constraint

**Decision:** `splitNamespacedName` cuts at the **first** `__`. Tool names may contain `__` (suffix side is safe); server names containing `__` would be unresolvable and are therefore unsupported.

**Consequences:** Low risk today (seeded names are simple). A follow-up should add a server-name validation rule rejecting `__` in `validateServer` (do it with the Phase 9.3 server-model changes).

## D10 — Timeout middleware re-scoping

**Decision:** `middleware.Timeout(15 * time.Second)` moves off the global middleware stack into the `/api/v1` route group only.

**Rationale:** chi's Timeout cancels the request context after 15s. Applied globally it would kill streamable GET connections and long-running `tools/call` executions on `/mcp`. REST behavior is unchanged because `/api/v1` keeps the identical 15s budget.

**Consequences:** `/mcp` has no chi-level timeout; upstream-bound calls are bounded by executor/transport timeouts instead (Phase 9.4 adds `MCP_UPSTREAM_TIMEOUT_MS`).

## D11 — Outbound executor routing precedence (added 2026-09-12)

**Decision:** `RouterExecutor` precedence in Phase 9.4:

1. Server name `github` → `GitHubExecutor` (unchanged)
2. `transport_type = 'streamable_http'` **and** the server's tools carry `source = 'discovered'` → `MCPExecutor`
3. Everything else → `MockExecutor` (unchanged fallback)

**Rationale:** all pre-Phase-9 rows default to `transport_type='streamable_http'`, so routing on transport alone would silently redirect existing mock/demo servers to live upstream dials and break Phase 5 behavior and the README demo. Tool provenance (`source`) is the explicit opt-in signal that a server is a real MCP upstream.

**Consequences:** registering a streamable-http server without successful discovery keeps it on the mock path — acceptable and self-healing (next successful discovery flips it).

---

## SDK facts verified 2026-09-12 (pin: v1.7.x)

- `mcp.NewStreamableHTTPHandler(getServer func(*http.Request) *mcp.Server, opts *mcp.StreamableHTTPOptions)` returns an `http.Handler` — mounts directly into chi.
- `Server.AddTool(t *mcp.Tool, h mcp.ToolHandler)` is the low-level registration path: no schema inference, no argument validation — caller's responsibility. `Tool.InputSchema` is `any` and accepts `json.RawMessage`; it must JSON-marshal to an object with `"type": "object"` or `AddTool` panics (our CRUD validation already enforces this).
- `ToolHandler` = `func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error)`; `req.Params.Name` is the tool name and `req.Params.Arguments` is `json.RawMessage`.
- `CallToolResult.SetError(err)` produces an `isError: true` result with text content; use it for policy denials and execution failures (protocol-level JSON-RPC errors are reserved for malformed requests).
- Client side (Phase 9.4): `mcp.StreamableClientTransport`; `mcp.NewInMemoryTransports()` gives in-process client/server pairs for unit tests.
