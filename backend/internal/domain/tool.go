package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ToolRiskLow    = "low"
	ToolRiskMedium = "medium"
	ToolRiskHigh   = "high"
)

// Tool provenance (Phase 9, D11). Source is system-managed: tools created via
// admin CRUD are 'manual'; tools synced from a live upstream via tools/list
// discovery are 'discovered'. It is intentionally absent from the Create and
// Update request types so admins cannot forge it.
const (
	ToolSourceManual     = "manual"
	ToolSourceDiscovered = "discovered"
)

type MCPTool struct {
	ID          uuid.UUID       `json:"id"`
	ServerID    uuid.UUID       `json:"serverId"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	RiskLevel   string          `json:"riskLevel"`
	Enabled     bool            `json:"enabled"`
	Source      string          `json:"source"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type CreateMCPToolRequest struct {
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	RiskLevel   string          `json:"riskLevel"`
	Enabled     *bool           `json:"enabled"`
}

type UpdateMCPToolRequest struct {
	Name        *string          `json:"name"`
	Title       *string          `json:"title"`
	Description *string          `json:"description"`
	InputSchema *json.RawMessage `json:"inputSchema"`
	RiskLevel   *string          `json:"riskLevel"`
	Enabled     *bool            `json:"enabled"`
}
