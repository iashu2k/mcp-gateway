CREATE TABLE tool_invocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE RESTRICT,
    tool_id UUID NOT NULL REFERENCES mcp_tools(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    status TEXT NOT NULL,
    request_arguments JSONB NOT NULL,
    response_payload JSONB,
    error_code TEXT,
    error_message TEXT,
    duration_ms BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    CONSTRAINT tool_invocations_status_check CHECK (
        status IN ('running', 'succeeded', 'failed', 'denied')
    ),
    CONSTRAINT tool_invocations_duration_non_negative CHECK (
        duration_ms IS NULL OR duration_ms >= 0
    )
);

CREATE INDEX idx_tool_invocations_server_id
    ON tool_invocations(server_id);

CREATE INDEX idx_tool_invocations_tool_id
    ON tool_invocations(tool_id);

CREATE INDEX idx_tool_invocations_user_id
    ON tool_invocations(user_id);

CREATE INDEX idx_tool_invocations_status
    ON tool_invocations(status);

CREATE INDEX idx_tool_invocations_created_at
    ON tool_invocations(created_at DESC);