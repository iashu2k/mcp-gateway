package executor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// startFakeUpstream serves an in-memory MCP server over Streamable HTTP,
// exercising the real client transport path end to end.
func startFakeUpstream(t *testing.T) *httptest.Server {
	t.Helper()

	upstream := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "fake-upstream",
		Version: "0.0.1",
	}, nil)

	upstream.AddTool(&sdkmcp.Tool{
		Name:        "ping",
		Description: "replies with pong",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	}, func(
		ctx context.Context,
		req *sdkmcp.CallToolRequest,
	) (*sdkmcp.CallToolResult, error) {
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: "pong"}},
		}, nil
	})

	upstream.AddTool(&sdkmcp.Tool{
		Name:        "fail",
		Description: "always returns a tool-level error",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	}, func(
		ctx context.Context,
		req *sdkmcp.CallToolRequest,
	) (*sdkmcp.CallToolResult, error) {
		var result sdkmcp.CallToolResult
		result.SetError(context.DeadlineExceeded)
		return &result, nil
	})

	server := httptest.NewServer(sdkmcp.NewStreamableHTTPHandler(
		func(*http.Request) *sdkmcp.Server { return upstream },
		&sdkmcp.StreamableHTTPOptions{Stateless: true},
	))
	t.Cleanup(server.Close)

	return server
}

func upstreamServer(t *testing.T, url string) domain.MCPServer {
	t.Helper()

	return domain.MCPServer{
		ID:               uuid.New(),
		Name:             "fake-upstream",
		BaseURL:          url,
		TransportType:    domain.TransportStreamableHTTP,
		Status:           domain.ServerStatusActive,
		ConnectionConfig: json.RawMessage(`{}`),
	}
}

func TestMCPExecutorExecutesToolAgainstLiveUpstream(t *testing.T) {
	fake := startFakeUpstream(t)
	executor := NewMCPExecutor(5 * time.Second)
	t.Cleanup(executor.Close)

	result, err := executor.Execute(
		context.Background(),
		upstreamServer(t, fake.URL),
		domain.MCPTool{Name: "ping"},
		json.RawMessage(`{}`),
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(string(result), "pong") {
		t.Fatalf("expected pong in result, got %s", result)
	}
}

func TestMCPExecutorMapsToolErrorToUpstreamError(t *testing.T) {
	fake := startFakeUpstream(t)
	executor := NewMCPExecutor(5 * time.Second)
	t.Cleanup(executor.Close)

	_, err := executor.Execute(
		context.Background(),
		upstreamServer(t, fake.URL),
		domain.MCPTool{Name: "fail"},
		json.RawMessage(`{}`),
	)
	if err == nil {
		t.Fatal("expected error for isError tool result")
	}
	if !strings.Contains(err.Error(), ErrMCPUpstream.Error()) {
		t.Fatalf("expected ErrMCPUpstream, got %v", err)
	}
}

func TestMCPExecutorDiscoversTools(t *testing.T) {
	fake := startFakeUpstream(t)
	executor := NewMCPExecutor(5 * time.Second)
	t.Cleanup(executor.Close)

	discovered, err := executor.DiscoverTools(
		context.Background(),
		upstreamServer(t, fake.URL),
	)
	if err != nil {
		t.Fatalf("DiscoverTools: %v", err)
	}

	if len(discovered) != 2 {
		t.Fatalf("expected 2 discovered tools, got %d", len(discovered))
	}

	byName := map[string]domain.MCPTool{}
	for _, tool := range discovered {
		byName[tool.Name] = tool
	}

	ping, ok := byName["ping"]
	if !ok {
		t.Fatal("ping not discovered")
	}
	if ping.Source != domain.ToolSourceDiscovered {
		t.Fatalf("expected source=discovered, got %q", ping.Source)
	}
	if !ping.Enabled || ping.RiskLevel != domain.ToolRiskLow {
		t.Fatalf("unexpected governance defaults: %+v", ping)
	}
}

func TestMCPExecutorFailsWhenHeaderEnvVarMissing(t *testing.T) {
	executor := NewMCPExecutor(5 * time.Second)
	executor.lookupEnv = func(string) string { return "" }
	t.Cleanup(executor.Close)

	server := upstreamServer(t, "http://127.0.0.1:1/mcp")
	server.ConnectionConfig = json.RawMessage(
		`{"headers":{"Authorization":"MISSING_TOKEN_ENV"}}`,
	)

	_, err := executor.Execute(
		context.Background(),
		server,
		domain.MCPTool{Name: "ping"},
		json.RawMessage(`{}`),
	)
	if err == nil || !strings.Contains(err.Error(), "MISSING_TOKEN_ENV") {
		t.Fatalf("expected missing-env error, got %v", err)
	}
}

func TestMCPExecutorUpstreamDown(t *testing.T) {
	executor := NewMCPExecutor(500 * time.Millisecond)
	t.Cleanup(executor.Close)

	_, err := executor.Execute(
		context.Background(),
		upstreamServer(t, "http://127.0.0.1:1/mcp"),
		domain.MCPTool{Name: "ping"},
		json.RawMessage(`{}`),
	)
	if err == nil {
		t.Fatal("expected dial error for unreachable upstream")
	}
	if !strings.Contains(err.Error(), ErrMCPUpstream.Error()) {
		t.Fatalf("expected ErrMCPUpstream, got %v", err)
	}
}
