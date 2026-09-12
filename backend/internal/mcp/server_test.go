package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// --- fakes -----------------------------------------------------------------

type fakeServerLister struct {
	servers []domain.MCPServer
}

func (f fakeServerLister) List(
	context.Context,
) ([]domain.MCPServer, error) {
	return f.servers, nil
}

type fakeToolLister struct {
	toolsByServer map[uuid.UUID][]domain.MCPTool
}

func (f fakeToolLister) ListByServerID(
	_ context.Context,
	serverID uuid.UUID,
) ([]domain.MCPTool, error) {
	return f.toolsByServer[serverID], nil
}

type fakeInvoker struct {
	mu    sync.Mutex
	calls int
}

func (f *fakeInvoker) Invoke(
	_ context.Context,
	_ domain.AuthenticatedUser,
	_ string,
	_ string,
	_ domain.InvokeToolRequest,
) (domain.ToolInvocationResponse, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()

	return domain.ToolInvocationResponse{
		InvocationID: uuid.New(),
		Status:       domain.InvocationStatusSucceeded,
		Result:       json.RawMessage(`{"ok":true}`),
	}, nil
}

// --- fixture ---------------------------------------------------------------

func newTestHandler() (*httptest.Server, *fakeInvoker) {
	serverID := uuid.New()

	deps := Deps{
		Servers: fakeServerLister{servers: []domain.MCPServer{
			{
				ID:           serverID,
				Name:         "demo",
				Status:       domain.ServerStatusActive,
				TransportType: domain.TransportStreamableHTTP,
			},
			{
				ID:     uuid.New(),
				Name:   "inactive",
				Status: domain.ServerStatusInactive,
			},
		}},
		Tools: fakeToolLister{toolsByServer: map[uuid.UUID][]domain.MCPTool{
			serverID: {
				{
					ID:          uuid.New(),
					ServerID:    serverID,
					Name:        "echo",
					Description: "echoes input",
					InputSchema: json.RawMessage(
						`{"type":"object","properties":{"message":{"type":"string"}}}`,
					),
					RiskLevel: domain.ToolRiskLow,
					Enabled:   true,
					Source:    domain.ToolSourceManual,
				},
				{
					ID:          uuid.New(),
					ServerID:    serverID,
					Name:        "hidden",
					Description: "disabled tool",
					InputSchema: json.RawMessage(`{"type":"object"}`),
					RiskLevel:   domain.ToolRiskLow,
					Enabled:     false,
					Source:      domain.ToolSourceManual,
				},
			},
		}},
		Invoker: &fakeInvoker{},
		UserFromContext: func(context.Context) (domain.AuthenticatedUser, bool) {
			return domain.AuthenticatedUser{
				ID:   uuid.New(),
				Role: domain.RoleDeveloper,
			}, true
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	server := httptest.NewServer(NewHandler(deps))
	return server, deps.Invoker.(*fakeInvoker)
}

func connectClient(
	t *testing.T,
	url string,
) *sdkmcp.ClientSession {
	t.Helper()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(
		context.Background(),
		&sdkmcp.StreamableClientTransport{Endpoint: url},
		nil,
	)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	return session
}

// --- tests -----------------------------------------------------------------

func TestToolsListFiltersViaSDK(t *testing.T) {
	t.Parallel()

	server, _ := newTestHandler()
	defer server.Close()

	session := connectClient(t, server.URL)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	if len(result.Tools) != 1 {
		names := make([]string, 0, len(result.Tools))
		for _, tool := range result.Tools {
			names = append(names, tool.Name)
		}
		t.Fatalf("expected 1 tool, got %v", names)
	}

	if result.Tools[0].Name != "demo__echo" {
		t.Fatalf("expected namespaced demo__echo, got %q", result.Tools[0].Name)
	}
}

func TestCallToolFlowsThroughInvoker(t *testing.T) {
	t.Parallel()

	server, invoker := newTestHandler()
	defer server.Close()

	session := connectClient(t, server.URL)

	result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "demo__echo",
		Arguments: json.RawMessage(`{"message":"hi"}`),
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected isError result: %+v", result.Content)
	}
	if invoker.calls != 1 {
		t.Fatalf("expected invoker to be called once, got %d", invoker.calls)
	}
}

// TestConcurrentClientSessions exercises simultaneous MCP sessions against
// the stateless per-request server construction. Run with -race: any shared
// state mistake in registerCatalogTools or the handler wiring fails here.
func TestConcurrentClientSessions(t *testing.T) {
	t.Parallel()

	server, invoker := newTestHandler()
	defer server.Close()

	const clients = 8

	var wg sync.WaitGroup
	errs := make(chan error, clients)

	for i := range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine gets its own session, like distinct agents.
			client := sdkmcp.NewClient(&sdkmcp.Implementation{
				Name:    fmt.Sprintf("client-%d", i),
				Version: "0.0.1",
			}, nil)

			session, err := client.Connect(
				context.Background(),
				&sdkmcp.StreamableClientTransport{Endpoint: server.URL},
				nil,
			)
			if err != nil {
				errs <- fmt.Errorf("client %d connect: %w", i, err)
				return
			}
			defer func() { _ = session.Close() }()

			if _, err := session.ListTools(context.Background(), nil); err != nil {
				errs <- fmt.Errorf("client %d list: %w", i, err)
				return
			}

			if _, err := session.CallTool(
				context.Background(),
				&sdkmcp.CallToolParams{
					Name:      "demo__echo",
					Arguments: json.RawMessage(`{"message":"concurrent"}`),
				},
			); err != nil {
				errs <- fmt.Errorf("client %d call: %w", i, err)
				return
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	if invoker.calls != clients {
		t.Fatalf("expected %d invocations, got %d", clients, invoker.calls)
	}
}
