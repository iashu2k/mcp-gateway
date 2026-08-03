package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

type fakeInvocationStore struct {
	created domain.ToolInvocation
}

func (f *fakeInvocationStore) CreateRunning(
	_ context.Context,
	invocation domain.ToolInvocation,
) (domain.ToolInvocation, error) {
	invocation.ID = uuid.New()
	invocation.Status = domain.InvocationStatusRunning
	f.created = invocation

	return invocation, nil
}

func (f *fakeInvocationStore) MarkSucceeded(
	_ context.Context,
	invocationID uuid.UUID,
	result json.RawMessage,
	durationMS int64,
	completedAt time.Time,
) (domain.ToolInvocation, error) {
	invocation := f.created
	invocation.ID = invocationID
	invocation.Status = domain.InvocationStatusSucceeded
	invocation.ResponsePayload = result
	invocation.DurationMS = &durationMS
	invocation.CompletedAt = &completedAt

	return invocation, nil
}

func (f *fakeInvocationStore) MarkFailed(
	_ context.Context,
	invocationID uuid.UUID,
	errorCode string,
	errorMessage string,
	durationMS int64,
	completedAt time.Time,
) (domain.ToolInvocation, error) {
	invocation := f.created
	invocation.ID = invocationID
	invocation.Status = domain.InvocationStatusFailed
	invocation.ErrorCode = &errorCode
	invocation.ErrorMessage = &errorMessage
	invocation.DurationMS = &durationMS
	invocation.CompletedAt = &completedAt

	return invocation, nil
}

type fakeInvocationServerReader struct {
	server domain.MCPServer
	err    error
}

func (f fakeInvocationServerReader) GetByID(
	_ context.Context,
	_ uuid.UUID,
) (domain.MCPServer, error) {
	return f.server, f.err
}

type fakeInvocationToolReader struct {
	tool domain.MCPTool
	err  error
}

func (f fakeInvocationToolReader) GetByID(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) (domain.MCPTool, error) {
	return f.tool, f.err
}

type fakeSchemaValidator struct {
	err error
}

func (f fakeSchemaValidator) Validate(
	_ json.RawMessage,
	_ json.RawMessage,
) error {
	return f.err
}

type fakeExecutor struct {
	result json.RawMessage
	err    error
}

func (f fakeExecutor) Execute(
	_ context.Context,
	_ domain.MCPServer,
	_ domain.MCPTool,
	_ json.RawMessage,
) (json.RawMessage, error) {
	return f.result, f.err
}

func TestInvocationServiceInvokesLowRiskTool(t *testing.T) {
	serverID := uuid.New()
	toolID := uuid.New()
	userID := uuid.New()

	store := &fakeInvocationStore{}

	service := NewInvocationService(
		store,
		fakeInvocationServerReader{
			server: domain.MCPServer{
				ID:     serverID,
				Status: "active",
			},
		},
		fakeInvocationToolReader{
			tool: domain.MCPTool{
				ID:          toolID,
				ServerID:    serverID,
				Name:        "echo",
				InputSchema: json.RawMessage(`{"type":"object"}`),
				RiskLevel:   domain.ToolRiskLow,
				Enabled:     true,
			},
		},
		fakeSchemaValidator{},
		fakeExecutor{
			result: json.RawMessage(`{"message":"hello"}`),
		},
	)

	response, err := service.Invoke(
		context.Background(),
		domain.AuthenticatedUser{
			ID:   userID,
			Role: domain.RoleDeveloper,
		},
		serverID.String(),
		toolID.String(),
		domain.InvokeToolRequest{
			Arguments: json.RawMessage(`{"message":"hello"}`),
		},
	)
	if err != nil {
		t.Fatalf("invoke tool: %v", err)
	}

	if response.Status != domain.InvocationStatusSucceeded {
		t.Fatalf(
			"expected status %q, got %q",
			domain.InvocationStatusSucceeded,
			response.Status,
		)
	}

	if store.created.UserID != userID {
		t.Fatalf(
			"expected user ID %q, got %q",
			userID,
			store.created.UserID,
		)
	}
}

func TestInvocationServiceRejectsViewer(t *testing.T) {
	service := NewInvocationService(
		&fakeInvocationStore{},
		fakeInvocationServerReader{},
		fakeInvocationToolReader{},
		fakeSchemaValidator{},
		fakeExecutor{},
	)

	_, err := service.Invoke(
		context.Background(),
		domain.AuthenticatedUser{
			ID:   uuid.New(),
			Role: domain.RoleViewer,
		},
		uuid.NewString(),
		uuid.NewString(),
		domain.InvokeToolRequest{
			Arguments: json.RawMessage(`{}`),
		},
	)

	if !errors.Is(err, ErrInvocationForbidden) {
		t.Fatalf("expected ErrInvocationForbidden, got %v", err)
	}
}

func TestInvocationServiceRejectsNonLowRiskTool(t *testing.T) {
	serverID := uuid.New()
	toolID := uuid.New()

	service := NewInvocationService(
		&fakeInvocationStore{},
		fakeInvocationServerReader{
			server: domain.MCPServer{
				ID:     serverID,
				Status: "active",
			},
		},
		fakeInvocationToolReader{
			tool: domain.MCPTool{
				ID:        toolID,
				ServerID:  serverID,
				RiskLevel: domain.ToolRiskMedium,
				Enabled:   true,
			},
		},
		fakeSchemaValidator{},
		fakeExecutor{},
	)

	_, err := service.Invoke(
		context.Background(),
		domain.AuthenticatedUser{
			ID:   uuid.New(),
			Role: domain.RoleAdmin,
		},
		serverID.String(),
		toolID.String(),
		domain.InvokeToolRequest{
			Arguments: json.RawMessage(`{}`),
		},
	)

	if !errors.Is(err, ErrToolRiskNotAllowed) {
		t.Fatalf("expected ErrToolRiskNotAllowed, got %v", err)
	}
}
