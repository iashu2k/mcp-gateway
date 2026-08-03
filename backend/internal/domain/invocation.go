package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	InvocationStatusRunning   = "running"
	InvocationStatusSucceeded = "succeeded"
	InvocationStatusFailed    = "failed"
	InvocationStatusDenied    = "denied"
)

type ToolInvocation struct {
	ID               uuid.UUID       `json:"id"`
	ServerID         uuid.UUID       `json:"serverId"`
	ToolID           uuid.UUID       `json:"toolId"`
	UserID           uuid.UUID       `json:"userId"`
	Status           string          `json:"status"`
	RequestArguments json.RawMessage `json:"requestArguments"`
	ResponsePayload  json.RawMessage `json:"responsePayload,omitempty"`
	ErrorCode        *string         `json:"errorCode,omitempty"`
	ErrorMessage     *string         `json:"errorMessage,omitempty"`
	DurationMS       *int64          `json:"durationMs,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	CompletedAt      *time.Time      `json:"completedAt,omitempty"`
}

type InvokeToolRequest struct {
	Arguments json.RawMessage `json:"arguments"`
}

type ToolInvocationResponse struct {
	InvocationID uuid.UUID       `json:"invocationId"`
	ServerID     uuid.UUID       `json:"serverId"`
	ToolID       uuid.UUID       `json:"toolId"`
	ToolName     string          `json:"toolName"`
	Status       string          `json:"status"`
	Result       json.RawMessage `json:"result,omitempty"`
	DurationMS   *int64          `json:"durationMs,omitempty"`
	CompletedAt  *time.Time      `json:"completedAt,omitempty"`
}
