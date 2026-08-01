CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    base_url TEXT NOT NULL,
    transport_type TEXT NOT NULL DEFAULT 'streamable_http',
    status TEXT NOT NULL DEFAULT 'active',
    owner_team TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_transport_type CHECK (
        transport_type IN ('streamable_http', 'sse', 'stdio')
    ),
    CONSTRAINT valid_server_status CHECK (
        status IN ('active', 'inactive', 'unhealthy')
    )
);

CREATE INDEX idx_mcp_servers_status ON mcp_servers(status);
CREATE INDEX idx_mcp_servers_owner_team ON mcp_servers(owner_team);