package service

import (
	"testing"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

func TestUpstreamServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server domain.MCPServer
		tool   domain.MCPTool
		want   string
	}{
		{
			name:   "github server reports github",
			server: domain.MCPServer{Name: "github"},
			tool:   domain.MCPTool{Source: domain.ToolSourceManual},
			want:   "github",
		},
		{
			name: "discovered streamable_http tool reports mcp",
			server: domain.MCPServer{
				Name:          "self-upstream",
				TransportType: domain.TransportStreamableHTTP,
			},
			tool: domain.MCPTool{Source: domain.ToolSourceDiscovered},
			want: "mcp",
		},
		{
			name: "manual streamable_http tool has no upstream (mock path)",
			server: domain.MCPServer{
				Name:          "demo-tools",
				TransportType: domain.TransportStreamableHTTP,
			},
			tool: domain.MCPTool{Source: domain.ToolSourceManual},
			want: "",
		},
		{
			name: "stdio server has no upstream",
			server: domain.MCPServer{
				Name:          "local-tools",
				TransportType: domain.TransportStdio,
			},
			tool: domain.MCPTool{Source: domain.ToolSourceDiscovered},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := upstreamServiceName(tt.server, tt.tool); got != tt.want {
				t.Fatalf("upstreamServiceName() = %q, want %q", got, tt.want)
			}
		})
	}
}
