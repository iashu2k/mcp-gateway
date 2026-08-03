package executor

import (
	"context"
	"encoding/json"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// RouterExecutor dispatches to the right executor based on the
// registered server name. "github" servers hit the live GitHub API;
// everything else falls back to the deterministic mock executor.
type RouterExecutor struct {
	mock   *MockExecutor
	github *GitHubExecutor
}

func NewRouterExecutor(
	mock *MockExecutor,
	github *GitHubExecutor,
) *RouterExecutor {
	return &RouterExecutor{mock: mock, github: github}
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

	return e.mock.Execute(ctx, server, tool, arguments)
}
