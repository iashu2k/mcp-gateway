package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	TransportStreamableHTTP = "streamable_http"
	TransportSSE            = "sse"
	TransportStdio          = "stdio"

	ServerStatusActive    = "active"
	ServerStatusInactive  = "inactive"
	ServerStatusUnhealthy = "unhealthy"
)

type MCPServer struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	BaseURL       string    `json:"baseUrl"`
	TransportType string    `json:"transportType"`
	Status        string    `json:"status"`
	OwnerTeam     string    `json:"ownerTeam"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateMCPServerRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	BaseURL       string `json:"baseUrl"`
	TransportType string `json:"transportType"`
	OwnerTeam     string `json:"ownerTeam"`
}

type UpdateMCPServerRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	BaseURL       *string `json:"baseUrl"`
	TransportType *string `json:"transportType"`
	Status        *string `json:"status"`
	OwnerTeam     *string `json:"ownerTeam"`
}
