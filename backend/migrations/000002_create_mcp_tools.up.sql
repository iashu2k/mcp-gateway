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

    CONSTRAINT mcp_tools_name_not_blank CHECK (length(trim(name)) > 0),
    CONSTRAINT mcp_tools_description_not_blank CHECK (
        length(trim(description)) > 0
    ),
    CONSTRAINT mcp_tools_risk_level_check CHECK (
        risk_level IN ('low', 'medium', 'high')
    ),
    CONSTRAINT mcp_tools_server_name_unique UNIQUE (server_id, name)
);

CREATE INDEX idx_mcp_tools_server_id ON mcp_tools(server_id);
CREATE INDEX idx_mcp_tools_server_enabled ON mcp_tools(server_id, enabled);
CREATE INDEX idx_mcp_tools_risk_level ON mcp_tools(risk_level);
CREATE INDEX idx_mcp_tools_input_schema_gin
    ON mcp_tools
    USING GIN (input_schema);