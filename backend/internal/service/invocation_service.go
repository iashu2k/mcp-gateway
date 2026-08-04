package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/executor"
	"github.com/iashu2k/mcp-gateway/backend/internal/observability"
)

var (
	ErrInvocationForbidden = errors.New("user is not allowed to invoke tools")
	ErrServerInactive      = errors.New("MCP server is not active")
	ErrToolDisabled        = errors.New("MCP tool is disabled")
	ErrToolRiskNotAllowed  = errors.New("MCP tool risk level is not allowed")
)

type InvocationStore interface {
	CreateRunning(
		ctx context.Context,
		invocation domain.ToolInvocation,
	) (domain.ToolInvocation, error)

	MarkSucceeded(
		ctx context.Context,
		invocationID uuid.UUID,
		responsePayload json.RawMessage,
		durationMS int64,
		completedAt time.Time,
	) (domain.ToolInvocation, error)

	MarkFailed(
		ctx context.Context,
		invocationID uuid.UUID,
		errorCode string,
		errorMessage string,
		durationMS int64,
		completedAt time.Time,
	) (domain.ToolInvocation, error)
}

type InvocationServerReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.MCPServer, error)
}

type InvocationToolReader interface {
	GetByID(
		ctx context.Context,
		serverID uuid.UUID,
		toolID uuid.UUID,
	) (domain.MCPTool, error)
}

type InvocationService struct {
	invocations InvocationStore
	servers     InvocationServerReader
	tools       InvocationToolReader
	validator   SchemaValidator
	executor    executor.ToolExecutor
	now         func() time.Time
}

func NewInvocationService(
	invocations InvocationStore,
	servers InvocationServerReader,
	tools InvocationToolReader,
	validator SchemaValidator,
	executor executor.ToolExecutor,
) *InvocationService {
	return &InvocationService{
		invocations: invocations,
		servers:     servers,
		tools:       tools,
		validator:   validator,
		executor:    executor,
		now:         time.Now,
	}
}

func (s *InvocationService) Invoke(
	ctx context.Context,
	user domain.AuthenticatedUser,
	serverID string,
	toolID string,
	request domain.InvokeToolRequest,
) (domain.ToolInvocationResponse, error) {
	// Role check
	if user.Role != domain.RoleAdmin && user.Role != domain.RoleDeveloper {
		return domain.ToolInvocationResponse{}, ErrInvocationForbidden
	}

	serverUUID, toolUUID, err := parseInvocationIDs(serverID, toolID)
	if err != nil {
		return domain.ToolInvocationResponse{}, err
	}

	server, err := s.servers.GetByID(ctx, serverUUID)
	if err != nil {
		return domain.ToolInvocationResponse{}, err
	}

	if server.Status != "active" {
		return domain.ToolInvocationResponse{}, ErrServerInactive
	}

	tool, err := s.tools.GetByID(ctx, serverUUID, toolUUID)
	if err != nil {
		return domain.ToolInvocationResponse{}, err
	}

	if !tool.Enabled {
		return domain.ToolInvocationResponse{}, ErrToolDisabled
	}

	if tool.RiskLevel != domain.ToolRiskLow {
		return domain.ToolInvocationResponse{}, ErrToolRiskNotAllowed
	}

	request.Arguments = json.RawMessage(
		strings.TrimSpace(string(request.Arguments)),
	)

	if len(request.Arguments) == 0 {
		return domain.ToolInvocationResponse{}, ValidationError{
			Fields: []FieldError{
				{
					Field:   "arguments",
					Message: "arguments is required",
				},
			},
		}
	}

	if err := s.validator.Validate(tool.InputSchema, request.Arguments); err != nil {
		if errors.Is(err, ErrInvalidToolArguments) {
			return domain.ToolInvocationResponse{}, ValidationError{
				Fields: []FieldError{
					{
						Field:   "arguments",
						Message: err.Error(),
					},
				},
			}
		}

		return domain.ToolInvocationResponse{}, err
	}

	// Create audit record
	invocation, err := s.invocations.CreateRunning(ctx, domain.ToolInvocation{
		ServerID:         server.ID,
		ToolID:           tool.ID,
		UserID:           user.ID,
		RequestArguments: request.Arguments,
	})
	if err != nil {
		slog.Error("failed to create invocation audit record",
			"server_id", serverUUID,
			"tool_id", toolUUID,
			"user_id", user.ID,
			"error", err,
		)
		return domain.ToolInvocationResponse{}, err
	}

	slog.Info("tool invocation started",
		"invocation_id", invocation.ID,
		"server_id", server.ID,
		"server_name", server.Name,
		"tool_id", tool.ID,
		"tool_name", tool.Name,
		"user_id", user.ID,
	)

	startedAt := s.now()

	// Execute tool
	result, err := s.executor.Execute(ctx, server, tool, request.Arguments)

	completedAt := s.now().UTC()
	duration := completedAt.Sub(startedAt)
	durationMS := duration.Milliseconds()

	if err != nil {
		failed, markErr := s.invocations.MarkFailed(
			ctx,
			invocation.ID,
			"execution_failed",
			"tool execution failed",
			durationMS,
			completedAt,
		)
		if markErr != nil {
			slog.Error("failed to record invocation failure",
				"invocation_id", invocation.ID,
				"error", markErr,
			)
			return domain.ToolInvocationResponse{}, fmt.Errorf(
				"record failed invocation: %w",
				markErr,
			)
		}

		slog.Error("tool execution failed",
			"invocation_id", invocation.ID,
			"server_name", server.Name,
			"tool_name", tool.Name,
			"duration_ms", durationMS,
			"error", err,
		)

		// Metrics
		observability.InvocationsTotal.WithLabelValues(
			server.Name,
			tool.Name,
			"failed",
		).Inc()

		observability.InvocationDuration.WithLabelValues(
			server.Name,
			tool.Name,
		).Observe(duration.Seconds())

		// Track upstream errors separately
		if server.Name == "github" {
			observability.UpstreamRequestsTotal.WithLabelValues(
				"github",
				"error",
			).Inc()
		}

		return domain.ToolInvocationResponse{
			InvocationID: failed.ID,
			ServerID:     failed.ServerID,
			ToolID:       failed.ToolID,
			ToolName:     tool.Name,
			Status:       failed.Status,
			DurationMS:   failed.DurationMS,
			CompletedAt:  failed.CompletedAt,
		}, nil
	}

	succeeded, err := s.invocations.MarkSucceeded(
		ctx,
		invocation.ID,
		result,
		durationMS,
		completedAt,
	)
	if err != nil {
		slog.Error("failed to record invocation success",
			"invocation_id", invocation.ID,
			"error", err,
		)
		return domain.ToolInvocationResponse{}, fmt.Errorf(
			"record successful invocation: %w",
			err,
		)
	}

	slog.Info("tool invocation completed",
		"invocation_id", invocation.ID,
		"server_name", server.Name,
		"tool_name", tool.Name,
		"status", "succeeded",
		"duration_ms", durationMS,
	)

	// Metrics
	observability.InvocationsTotal.WithLabelValues(
		server.Name,
		tool.Name,
		"succeeded",
	).Inc()

	observability.InvocationDuration.WithLabelValues(
		server.Name,
		tool.Name,
	).Observe(duration.Seconds())

	// Track upstream success
	if server.Name == "github" {
		observability.UpstreamRequestsTotal.WithLabelValues(
			"github",
			"success",
		).Inc()
	}

	return invocationResponse(succeeded, tool.Name), nil
}

func parseInvocationIDs(
	serverID string,
	toolID string,
) (uuid.UUID, uuid.UUID, error) {
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidServerID
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToolID
	}

	return serverUUID, toolUUID, nil
}

func invocationResponse(
	invocation domain.ToolInvocation,
	toolName string,
) domain.ToolInvocationResponse {
	return domain.ToolInvocationResponse{
		InvocationID: invocation.ID,
		ServerID:     invocation.ServerID,
		ToolID:       invocation.ToolID,
		ToolName:     toolName,
		Status:       invocation.Status,
		Result:       invocation.ResponsePayload,
		DurationMS:   invocation.DurationMS,
		CompletedAt:  invocation.CompletedAt,
	}
}
