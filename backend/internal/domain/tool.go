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

type MCPTool struct {
	ID          uuid.UUID       `json:"id"`
	ServerID    uuid.UUID       `json:"serverId"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	RiskLevel   string          `json:"riskLevel"`
	Enabled     bool            `json:"enabled"`
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
