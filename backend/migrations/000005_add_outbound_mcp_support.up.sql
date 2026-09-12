-- Phase 9.3 (amended D5): outbound MCP support.
-- mcp_servers.connection_config holds endpoint extras (header credential
-- REFERENCES, e.g. env var names) — never raw secrets. base_url already
-- stores the upstream endpoint.
-- mcp_tools.source marks tool provenance: 'manual' (admin CRUD, default) or
-- 'discovered' (synced from a live upstream via tools/list in Phase 9.4).
-- D11: discovery is the explicit signal that routes a streamable_http server
-- to the MCPExecutor instead of the mock fallback.

ALTER TABLE mcp_servers
    ADD COLUMN connection_config jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE mcp_tools
    ADD COLUMN source text NOT NULL DEFAULT 'manual'
    CONSTRAINT valid_tool_source CHECK (source IN ('manual', 'discovered'));
