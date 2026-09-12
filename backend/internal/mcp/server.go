// Package mcp exposes the governed tool catalog as an MCP server over
// Streamable HTTP (Phase 9 inbound transport). It is a thin adapter: all
// policy, validation, execution, and auditing stay in the existing
// service/executor layers. See docs/phase-9-decisions.md.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// ServerLister enumerates registered MCP servers. Satisfied by
// repository.ServerRepository.
type ServerLister interface {
	List(ctx context.Context) ([]domain.MCPServer, error)
}

// ToolLister enumerates tools for one server. Satisfied by
// repository.ToolRepository.
type ToolLister interface {
	ListByServerID(ctx context.Context, serverID uuid.UUID) ([]domain.MCPTool, error)
}

// Invoker runs one governed invocation. Satisfied by
// service.InvocationService — the policy chain (role check, server active,
// tool enabled, low risk, JSON Schema validation), executor routing, and the
// audit record all live behind this single call.
type Invoker interface {
	Invoke(
		ctx context.Context,
		user domain.AuthenticatedUser,
		serverID string,
		toolID string,
		request domain.InvokeToolRequest,
	) (domain.ToolInvocationResponse, error)
}

// UserFromContext extracts the authenticated user placed on the request
// context by httpapi.RequireAuthentication. It is injected (rather than
// imported) because httpapi imports this package to mount the handler —
// importing back would create an import cycle.
type UserFromContext func(ctx context.Context) (domain.AuthenticatedUser, bool)

// Deps wires the gateway catalog into an MCP server handler.
type Deps struct {
	Servers         ServerLister
	Tools           ToolLister
	Invoker         Invoker
	UserFromContext UserFromContext
	Logger          *slog.Logger
}

// NewHandler returns an http.Handler serving MCP over Streamable HTTP.
//
// The handler is stateless (D8): every request builds a fresh mcp.Server and
// registers the catalog read from Postgres at that moment, so admin CRUD is
// reflected on the next tools/list with no extra wiring. Mount it behind
// RequireAuthentication, outside any chi Timeout middleware (D10).
func NewHandler(deps Deps) http.Handler {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	return sdkmcp.NewStreamableHTTPHandler(
		func(r *http.Request) *sdkmcp.Server {
			server := sdkmcp.NewServer(&sdkmcp.Implementation{
				Name:    "mcp-gateway",
				Version: "0.9.0",
			}, nil)
			deps.registerCatalogTools(r.Context(), server)
			return server
		},
		&sdkmcp.StreamableHTTPOptions{Stateless: true},
	)
}

// registerCatalogTools implements tools/list: advertise every exposable tool
// (D7) under its namespaced name (D3), with the stored JSON Schema passed
// through untouched.
func (d Deps) registerCatalogTools(ctx context.Context, server *sdkmcp.Server) {
	servers, err := d.Servers.List(ctx)
	if err != nil {
		d.Logger.Error("mcp: list servers for catalog", "error", err)
		return
	}

	for _, srv := range servers {
		if srv.Status != domain.ServerStatusActive {
			continue
		}

		tools, err := d.Tools.ListByServerID(ctx, srv.ID)
		if err != nil {
			d.Logger.Error("mcp: list tools for catalog",
				"server_id", srv.ID,
				"server_name", srv.Name,
				"error", err,
			)
			continue
		}

		for _, tool := range tools {
			if !exposable(srv, tool) {
				continue
			}
			server.AddTool(gatewayTool(srv, tool), d.callToolHandler())
		}
	}
}

// gatewayTool maps a catalog entry to an MCP tool definition. InputSchema is
// the stored draft 2020-12 schema; the SDK requires it to marshal to an
// object with "type": "object", which CRUD validation already enforces.
func gatewayTool(server domain.MCPServer, tool domain.MCPTool) *sdkmcp.Tool {
	return &sdkmcp.Tool{
		Name:        NamespacedName(server.Name, tool.Name),
		Title:       tool.Title,
		Description: fmt.Sprintf("[%s] %s", server.Name, tool.Description),
		InputSchema: json.RawMessage(tool.InputSchema),
	}
}

// callToolHandler implements tools/call. It resolves the namespaced name and
// delegates to the existing invocation service — no policy logic is
// duplicated here.
func (d Deps) callToolHandler() sdkmcp.ToolHandler {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
		user, ok := d.UserFromContext(ctx)
		if !ok {
			// The route sits behind RequireAuthentication, so this is a
			// wiring bug rather than a client error.
			return errorResult("authentication context is missing"), nil
		}

		serverName, toolName, ok := splitNamespacedName(req.Params.Name)
		if !ok {
			return errorResult(fmt.Sprintf(
				"tool name %q is not namespaced as serverName__toolName",
				req.Params.Name,
			)), nil
		}

		serverID, toolID, err := d.resolve(ctx, serverName, toolName)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		arguments := json.RawMessage(req.Params.Arguments)
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}

		resp, err := d.Invoker.Invoke(ctx, user, serverID, toolID, domain.InvokeToolRequest{
			Arguments: arguments,
		})
		if err != nil {
			// Policy denials, validation failures, and lookup errors all
			// surface as tool-level error results, per MCP convention.
			return errorResult(err.Error()), nil
		}

		// The invocation service records executor failures in the audit row
		// and returns them as a failed response with a nil error.
		if resp.Status == domain.InvocationStatusFailed {
			return errorResult(fmt.Sprintf(
				"tool execution failed (invocation %s is recorded in history)",
				resp.InvocationID,
			)), nil
		}

		result := &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(resp.Result)}},
		}
		if len(resp.Result) > 0 {
			result.StructuredContent = json.RawMessage(resp.Result)
		}
		return result, nil
	}
}

// resolve maps a namespaced call target to concrete server/tool IDs. It does
// not filter by status, enablement, or risk: InvocationService.Invoke
// re-applies the full policy chain, so resolution stays a pure lookup.
func (d Deps) resolve(ctx context.Context, serverName, toolName string) (string, string, error) {
	servers, err := d.Servers.List(ctx)
	if err != nil {
		return "", "", fmt.Errorf("resolve server %q: %w", serverName, err)
	}

	for _, srv := range servers {
		if srv.Name != serverName {
			continue
		}

		tools, err := d.Tools.ListByServerID(ctx, srv.ID)
		if err != nil {
			return "", "", fmt.Errorf("resolve tool %q: %w", toolName, err)
		}
		for _, tool := range tools {
			if tool.Name == toolName {
				return srv.ID.String(), tool.ID.String(), nil
			}
		}
		return "", "", fmt.Errorf("unknown tool %q on server %q", toolName, serverName)
	}

	return "", "", fmt.Errorf("unknown server %q", serverName)
}

// errorResult builds an isError tool result with a single text block.
func errorResult(message string) *sdkmcp.CallToolResult {
	var result sdkmcp.CallToolResult
	result.SetError(errors.New(message))
	return &result
}
