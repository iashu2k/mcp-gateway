package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
)

type fakeToolStore struct {
	createFn func(
		ctx context.Context,
		tool domain.MCPTool,
	) (domain.MCPTool, error)
}

func (f fakeToolStore) Create(
	ctx context.Context,
	tool domain.MCPTool,
) (domain.MCPTool, error) {
	return f.createFn(ctx, tool)
}

func (f fakeToolStore) ListByServerID(
	context.Context,
	uuid.UUID,
) ([]domain.MCPTool, error) {
	return nil, nil
}

func (f fakeToolStore) GetByID(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (domain.MCPTool, error) {
	return domain.MCPTool{}, repository.ErrToolNotFound
}

func (f fakeToolStore) Update(
	context.Context,
	domain.MCPTool,
) (domain.MCPTool, error) {
	return domain.MCPTool{}, nil
}

func (f fakeToolStore) Delete(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) error {
	return nil
}

type fakeServerReader struct {
	server domain.MCPServer
	err    error
}

func (f fakeServerReader) GetByID(
	context.Context,
	uuid.UUID,
) (domain.MCPServer, error) {
	return f.server, f.err
}

func TestCreateToolUsesDefaultRiskAndEnabledState(t *testing.T) {
	serverID := uuid.New()

	toolStore := fakeToolStore{
		createFn: func(
			_ context.Context,
			tool domain.MCPTool,
		) (domain.MCPTool, error) {
			if tool.RiskLevel != domain.ToolRiskLow {
				t.Fatalf(
					"expected risk level %q, got %q",
					domain.ToolRiskLow,
					tool.RiskLevel,
				)
			}

			if !tool.Enabled {
				t.Fatal("expected tool to default to enabled")
			}

			if tool.ServerID != serverID {
				t.Fatalf(
					"expected server ID %q, got %q",
					serverID,
					tool.ServerID,
				)
			}

			tool.ID = uuid.New()
			return tool, nil
		},
	}

	serverReader := fakeServerReader{
		server: domain.MCPServer{ID: serverID},
	}

	service := NewToolService(toolStore, serverReader)

	tool, err := service.Create(context.Background(), serverID.String(), domain.CreateMCPToolRequest{
		Name:        "list_issues",
		Description: "List GitHub issues in a repository",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string"},
				"repo": {"type": "string"}
			},
			"required": ["owner", "repo"]
		}`),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tool.ID == uuid.Nil {
		t.Fatal("expected created tool ID")
	}
}

func TestCreateToolRejectsSchemaWithoutObjectRoot(t *testing.T) {
	serverID := uuid.New()

	toolStore := fakeToolStore{
		createFn: func(
			_ context.Context,
			tool domain.MCPTool,
		) (domain.MCPTool, error) {
			return tool, nil
		},
	}

	serverReader := fakeServerReader{
		server: domain.MCPServer{ID: serverID},
	}

	service := NewToolService(toolStore, serverReader)

	_, err := service.Create(context.Background(), serverID.String(), domain.CreateMCPToolRequest{
		Name:        "invalid_tool",
		Description: "A tool with an invalid input schema",
		InputSchema: json.RawMessage(`{
			"type": "array"
		}`),
	})

	var validationError ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestCreateToolRejectsInvalidRiskLevel(t *testing.T) {
	serverID := uuid.New()

	toolStore := fakeToolStore{
		createFn: func(
			_ context.Context,
			tool domain.MCPTool,
		) (domain.MCPTool, error) {
			return tool, nil
		},
	}

	serverReader := fakeServerReader{
		server: domain.MCPServer{ID: serverID},
	}

	service := NewToolService(toolStore, serverReader)

	_, err := service.Create(context.Background(), serverID.String(), domain.CreateMCPToolRequest{
		Name:        "list_issues",
		Description: "List GitHub issues in a repository",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		RiskLevel: "critical",
	})

	var validationError ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}
