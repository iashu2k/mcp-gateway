package mcp

import (
	"strings"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// namespaceSeparator joins server and tool names for MCP exposure (D3).
// Parsing cuts at the first occurrence (D9): tool names may contain the
// separator, server names may not.
const namespaceSeparator = "__"

// NamespacedName returns the tool name advertised over MCP, e.g.
// "github__list_issues".
func NamespacedName(serverName, toolName string) string {
	return serverName + namespaceSeparator + toolName
}

// splitNamespacedName parses an MCP tool name back into its server and tool
// components. ok is false when the name is not namespaced or either side is
// empty.
func splitNamespacedName(name string) (serverName, toolName string, ok bool) {
	serverName, toolName, found := strings.Cut(name, namespaceSeparator)
	if !found || serverName == "" || toolName == "" {
		return "", "", false
	}
	return serverName, toolName, true
}

// exposable reports whether a tool may be advertised via MCP (D7). It mirrors
// the REST invocation policy exactly: active server, enabled tool, low risk.
func exposable(server domain.MCPServer, tool domain.MCPTool) bool {
	return server.Status == domain.ServerStatusActive &&
		tool.Enabled &&
		tool.RiskLevel == domain.ToolRiskLow
}
