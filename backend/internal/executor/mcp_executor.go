package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// ErrMCPUpstream marks failures talking to a live MCP upstream. The
// invocation service records it in the audit row, same as ErrGitHubUpstream.
var ErrMCPUpstream = errors.New("mcp upstream request failed")

// MCPExecutor executes tools on live MCP servers over Streamable HTTP
// (Phase 9.4). It also implements service.ToolDiscoverer for
// registration-time tools/list sync (D4).
//
// Client sessions are cached by server ID and created lazily on first use.
// A session whose call fails is evicted so the next call dials fresh.
type MCPExecutor struct {
	timeout    time.Duration
	lookupEnv  func(string) string
	mu         sync.Mutex
	sessions   map[uuid.UUID]*cachedSession
	httpClient *http.Client
}

type cachedSession struct {
	session *sdkmcp.ClientSession
	cancel  context.CancelFunc
}

func NewMCPExecutor(timeout time.Duration) *MCPExecutor {
	return &MCPExecutor{
		timeout:    timeout,
		lookupEnv:  os.Getenv,
		sessions:   make(map[uuid.UUID]*cachedSession),
		httpClient: &http.Client{},
	}
}

// Execute forwards already-validated arguments to the upstream tool and
// normalizes the result into the gateway's audit JSON shape.
func (e *MCPExecutor) Execute(
	ctx context.Context,
	server domain.MCPServer,
	tool domain.MCPTool,
	arguments json.RawMessage,
) (json.RawMessage, error) {
	session, err := e.sessionFor(server)
	if err != nil {
		return nil, err
	}

	// Strict upstreams reject "arguments": null; always send an object.
	if len(arguments) == 0 {
		arguments = json.RawMessage(`{}`)
	}

	callCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	result, err := session.CallTool(callCtx, &sdkmcp.CallToolParams{
		Name:      tool.Name,
		Arguments: arguments,
	})
	if err != nil {
		// Transport/protocol failure: the session may be broken, so drop it.
		e.evict(server.ID)
		return nil, fmt.Errorf("%w: %v", ErrMCPUpstream, err)
	}

	return normalizeCallResult(result)
}

// DiscoverTools runs tools/list against a live upstream and maps the result
// to catalog entries marked source='discovered' (D4). Governance fields get
// the same defaults as manual creation (enabled, low risk); on conflict the
// upsert preserves admin-set risk_level/enabled.
func (e *MCPExecutor) DiscoverTools(
	ctx context.Context,
	server domain.MCPServer,
) ([]domain.MCPTool, error) {
	session, err := e.sessionFor(server)
	if err != nil {
		return nil, err
	}

	listCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	discovered := make([]domain.MCPTool, 0)
	var params *sdkmcp.ListToolsParams

	// Follow pagination until the upstream stops returning a cursor.
	for {
		result, err := session.ListTools(listCtx, params)
		if err != nil {
			e.evict(server.ID)
			return nil, fmt.Errorf("%w: %v", ErrMCPUpstream, err)
		}

		for _, tool := range result.Tools {
			schema, err := json.Marshal(tool.InputSchema)
			if err != nil || string(schema) == "null" {
				schema = json.RawMessage(`{"type":"object","properties":{}}`)
			}

			discovered = append(discovered, domain.MCPTool{
				ServerID:    server.ID,
				Name:        tool.Name,
				Title:       tool.Title,
				Description: tool.Description,
				InputSchema: schema,
				RiskLevel:   domain.ToolRiskLow,
				Enabled:     true,
				Source:      domain.ToolSourceDiscovered,
			})
		}

		if result.NextCursor == "" {
			break
		}
		params = &sdkmcp.ListToolsParams{Cursor: result.NextCursor}
	}

	return discovered, nil
}

// Close shuts down every cached session. Wire to graceful shutdown.
func (e *MCPExecutor) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for id, entry := range e.sessions {
		entry.cancel()
		_ = entry.session.Close()
		delete(e.sessions, id)
	}
}

// sessionFor returns the cached session for a server, dialing lazily on
// first use. Sessions are created on a background-derived context with the
// dial timeout; the cancel func is retained so eviction/shutdown can tear
// the session down. Request-scoped cancellation therefore never kills a
// shared session, and per-call timeouts are applied at CallTool/ListTools.
func (e *MCPExecutor) sessionFor(
	server domain.MCPServer,
) (*sdkmcp.ClientSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if entry, ok := e.sessions[server.ID]; ok {
		return entry.session, nil
	}

	roundTripper, err := e.roundTripperFor(server)
	if err != nil {
		return nil, err
	}

	transport := &sdkmcp.StreamableClientTransport{
		Endpoint:   server.BaseURL,
		HTTPClient: &http.Client{Transport: roundTripper},
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    "mcp-gateway",
		Version: "0.9.0",
	}, nil)

	dialCtx, cancel := context.WithTimeout(context.Background(), e.timeout)
	session, err := client.Connect(dialCtx, transport, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("%w: %v", ErrMCPUpstream, err)
	}

	e.sessions[server.ID] = &cachedSession{session: session, cancel: cancel}
	return session, nil
}

func (e *MCPExecutor) evict(serverID uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if entry, ok := e.sessions[serverID]; ok {
		entry.cancel()
		_ = entry.session.Close()
		delete(e.sessions, serverID)
	}
}

// roundTripperFor resolves connection_config headers. Values are environment
// variable NAMES (enforced at CRUD time, D5 security note); the secret value
// is read from the process environment here, never stored in the database.
func (e *MCPExecutor) roundTripperFor(
	server domain.MCPServer,
) (http.RoundTripper, error) {
	base := e.httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	var config struct {
		Headers map[string]string `json:"headers"`
	}

	if len(server.ConnectionConfig) > 0 {
		if err := json.Unmarshal(server.ConnectionConfig, &config); err != nil {
			return nil, fmt.Errorf(
				"%w: decode connection_config: %v",
				ErrMCPUpstream,
				err,
			)
		}
	}

	resolved := make(map[string]string, len(config.Headers))
	for header, envName := range config.Headers {
		value := e.lookupEnv(envName)
		if value == "" {
			return nil, fmt.Errorf(
				"%w: environment variable %s (header %q) is not set",
				ErrMCPUpstream,
				envName,
				header,
			)
		}
		resolved[header] = value
	}

	return headerRoundTripper{base: base, headers: resolved}, nil
}

type headerRoundTripper struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t headerRoundTripper) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	clone := req.Clone(req.Context())
	for name, value := range t.headers {
		clone.Header.Set(name, value)
	}
	return t.base.RoundTrip(clone)
}

// normalizeCallResult flattens MCP content blocks into the gateway's audit
// JSON. A tool-level isError result maps to an error so the invocation
// service records a failed audit row (D6).
func normalizeCallResult(
	result *sdkmcp.CallToolResult,
) (json.RawMessage, error) {
	texts := make([]string, 0, len(result.Content))
	for _, content := range result.Content {
		if text, ok := content.(*sdkmcp.TextContent); ok {
			texts = append(texts, text.Text)
		} else {
			texts = append(texts, "[non-text content block omitted]")
		}
	}

	if result.IsError {
		return nil, fmt.Errorf("%w: %s", ErrMCPUpstream, strings.Join(texts, "\n"))
	}

	payload := map[string]any{"content": texts}
	if result.StructuredContent != nil {
		payload["structuredContent"] = result.StructuredContent
	}

	return json.Marshal(payload)
}
