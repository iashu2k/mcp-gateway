package executor

import (
	"context"
	"encoding/json"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// RouterExecutor dispatches to the right executor (D11 precedence):
//
//  1. server name "github"              -> GitHubExecutor (live REST)
//  2. streamable_http server whose tools were auto-discovered from a live
//     upstream (source='discovered')    -> MCPExecutor (live MCP call)
//  3. everything else                   -> MockExecutor (deterministic)
//
// Rule 2 requires the discovered marker because every pre-Phase-9 row
// defaults to transport_type='streamable_http'; transport alone would
// hijack existing mock/demo servers.
type RouterExecutor struct {
	mock   *MockExecutor
	github *GitHubExecutor
	mcp    *MCPExecutor
}

func NewRouterExecutor(
	mock *MockExecutor,
	github *GitHubExecutor,
	mcp *MCPExecutor,
) *RouterExecutor {
	return &RouterExecutor{mock: mock, github: github, mcp: mcp}
}

func (e *RouterExecutor) Execute(
	ctx context.Context,
	server domain.MCPServer,
	tool domain.MCPTool,
	arguments json.RawMessage,
) (json.RawMessage, error) {
	if server.Name == "github" {
		return e.github.Execute(ctx, server, tool, arguments)
	}

	if e.mcp != nil &&
		server.TransportType == domain.TransportStreamableHTTP &&
		tool.Source == domain.ToolSourceDiscovered {
		return e.mcp.Execute(ctx, server, tool, arguments)
	}

	return e.mock.Execute(ctx, server, tool, arguments)
}
