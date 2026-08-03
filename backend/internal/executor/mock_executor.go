package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

var ErrUnsupportedTool = errors.New("tool is not supported by the mock executor")

type ToolExecutor interface {
	Execute(
		ctx context.Context,
		server domain.MCPServer,
		tool domain.MCPTool,
		arguments json.RawMessage,
	) (json.RawMessage, error)
}

type MockExecutor struct{}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

func (e *MockExecutor) Execute(
	_ context.Context,
	server domain.MCPServer,
	tool domain.MCPTool,
	arguments json.RawMessage,
) (json.RawMessage, error) {
	switch tool.Name {
	case "echo":
		return executeEcho(arguments)

	case "list_issues":
		return executeListIssues(server, arguments)

	default:
		return nil, fmt.Errorf(
			"%w: %s",
			ErrUnsupportedTool,
			tool.Name,
		)
	}
}

func executeEcho(arguments json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, fmt.Errorf("decode echo arguments: %w", err)
	}

	return json.Marshal(map[string]any{
		"message": input.Message,
	})
}

func executeListIssues(
	server domain.MCPServer,
	arguments json.RawMessage,
) (json.RawMessage, error) {
	var input struct {
		Owner   string `json:"owner"`
		Repo    string `json:"repo"`
		State   string `json:"state"`
		PerPage int    `json:"per_page"`
	}

	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, fmt.Errorf("decode list_issues arguments: %w", err)
	}

	if input.State == "" {
		input.State = "open"
	}

	if input.PerPage == 0 {
		input.PerPage = 30
	}

	return json.Marshal(map[string]any{
		"server": server.Name,
		"repository": map[string]string{
			"owner": input.Owner,
			"repo":  input.Repo,
		},
		"state":   input.State,
		"perPage": input.PerPage,
		"mock":    true,
		"issues": []map[string]any{
			{
				"number": 101,
				"title":  "Mock issue: improve gateway observability",
				"state":  input.State,
			},
			{
				"number": 102,
				"title":  "Mock issue: add MCP transport adapter",
				"state":  input.State,
			},
		},
	})
}
